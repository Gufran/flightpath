package catalog

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
	"log"
	"time"
)

const FlightPathTag = "in-flightpath"

type ServiceFinder interface {
	Services(*api.QueryOptions) (map[string][]string, *api.QueryMeta, error)
	Service(string, string, *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error)
}

type CertFinder interface {
	ConnectCALeaf(string, *api.QueryOptions) (*api.LeafCert, *api.QueryMeta, error)
}

// Catalog interacts with consul and keeps a list
// of currently valid entities
type Catalog struct {
	ctx     context.Context
	catalog ServiceFinder
	connect CertFinder
}

// New creates a new instance of Catalog
func New(ctx context.Context, client *api.Client) *Catalog {
	return &Catalog{
		ctx:     ctx,
		catalog: client.Catalog(),
		connect: client.Agent(),
	}
}

// DiscoverClusters delivers the state-of-the-world list
// of clusters as available in the consul catalog.
// Errors encountered during the blocking loop are
// sent over the error channel. It is up to the consumer
// to decide what to do with those errors.
func (c *Catalog) DiscoverClusters(clusters chan<- ClusterInfo, cleanup chan<- string, errs chan<- error) {
	qopts := &api.QueryOptions{
		AllowStale:        false,
		RequireConsistent: true,
		WaitIndex:         0,
		WaitTime:          30 * time.Second,
	}

	activeWatchers := map[string]func(){}

	for {
		select {
		case <-c.ctx.Done():
			log.Println("Shutting down the target service watcher because the context is done")
			return

		default:
			services, meta, err := c.catalog.Services(qopts.WithContext(c.ctx))
			if err != nil {
				errs <- fmt.Errorf("failed to fetch list of services from consul catalog. %s", err)
				time.Sleep(3 * time.Second)
				break
			}

			if meta.LastIndex <= qopts.WaitIndex {
				break
			}

			qopts.WaitIndex = meta.LastIndex

			candidates := map[string]bool{}
			for name, tags := range services {
				for _, tag := range tags {
					if tag == FlightPathTag {
						candidates[name] = true
						log.Printf("service %s is marked as a discovery candidate", name)
						break
					}
				}
			}

			// Remove all services where there is a corresponding
			// connect proxy registered in catalog
			candidates, err = c.filterConnectTargets(candidates)
			if err != nil {
				errs <- fmt.Errorf("failed to filter connect target services. %s", err)
				time.Sleep(3 * time.Second)
				break
			}

			// Find services that dont' have an active watcher
			// and start a watcher for them.
			for name, isSidecar := range candidates {
				if _, ok := activeWatchers[name]; !ok {
					log.Printf("starting new watcher for service %s", name)
					ctx, cancel := context.WithCancel(c.ctx)
					activeWatchers[name] = cancel
					go c.watchService(ctx, name, isSidecar, clusters, errs)
				} else {
					log.Printf("watcher is already active for %s", name)
				}
			}

			// Find services that have an active watcher but
			// are no longer available in catalog and stop
			// their watcher
			for name, stop := range activeWatchers {
				if _, ok := candidates[name]; !ok {
					log.Printf("stopping watcher for service %s", name)

					stop()
					delete(activeWatchers, name)

					// Notify the cleanup channel asynchronously so that we
					// don't end up blocking if the other side is not
					// able to consume the channel fast enough
					go func(n string) { cleanup <- n }(name)
				} else {
					log.Printf("no active watcher for %#v", name)
				}
			}
		}
	}
}

func (c *Catalog) watchService(ctx context.Context, name string, isSidecar bool, clusters chan<- ClusterInfo, errs chan<- error) {
	qopts := &api.QueryOptions{
		AllowStale:        false,
		RequireConsistent: true,
		WaitIndex:         0,
		WaitTime:          30 * time.Second,
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("Shutting down the service watcher for %s because the context is done", name)
			return

		default:
			nodes, meta, err := c.catalog.Service(name, "", qopts.WithContext(c.ctx))
			if err != nil {
				errs <- err
				time.Sleep(3 * time.Second)
				break
			}

			if meta.LastIndex <= qopts.WaitIndex {
				break
			}

			qopts.WaitIndex = meta.LastIndex

			clusters <- &Cluster{
				name:      name,
				isConnect: isSidecar,
				services:  nodes,
			}
		}
	}
}

func isSidecarProxy(srvc *api.CatalogService) bool {
	if srvc.ServiceProxy == nil {
		return false
	}

	if srvc.ServiceProxy.DestinationServiceName == "" {
		return false
	}

	return true
}

func (c *Catalog) filterConnectTargets(candidates map[string]bool) (map[string]bool, error) {
	qopts := &api.QueryOptions{
		AllowStale:        false,
		RequireConsistent: true,
	}

	results := make(map[string]bool, len(candidates))
	for service := range candidates {
		results[service] = false
	}

	for name := range candidates {
		services, _, err := c.catalog.Service(name, "", qopts.WithContext(c.ctx))
		if err != nil {
			return nil, err
		}

		// interrogating just one service instance is sufficient
		// to make out whether or not there is a sidecar proxy
		// running for the service.
		svc := services[0]
		if isSidecarProxy(svc) {
			// If the service is a sidecar proxy then we
			// dont' want to watch the target service since
			// the target may not be able to receive traffic.
			// In this case we simply remove the target name
			// from results list.
			log.Printf("Service %s is a sidecar proxy for %s, removing %s from candidate list", svc.ServiceName, svc.ServiceProxy.DestinationServiceName, svc.ServiceProxy.DestinationServiceName)
			log.Printf("Service %s selected as a valid candidate for connect cluster discovery", svc.ServiceName)

			results[svc.ServiceName] = true
			delete(results, svc.ServiceProxy.DestinationServiceName)
		}
	}

	log.Printf("services selected for cluster discovery: %#v", results)
	return results, nil
}
