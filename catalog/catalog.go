package catalog

import (
	"context"
	"fmt"
	"github.com/Gufran/flightpath/log"
	"github.com/hashicorp/consul/api"
	"time"
)

var logger = log.New("catalog")

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

// NewCatalog creates a new instance of Catalog
func NewCatalog(ctx context.Context, client *api.Client) *Catalog {
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
			logger.Info("cluster discovery loop has shut down")
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
						logger.WithField("service", name).Info("found discovery candidate service")
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

			logger.WithField("total_candidates", len(candidates)).
				WithField("candidates", candidates).
				Debug("candidate list is ready")

			logger.WithField("watchers", mapToSlice(activeWatchers)).
				Debug("watcher state")

			// Find services that dont' have an active watcher
			// and start a watcher for them.
			for name, isSidecar := range candidates {
				if _, ok := activeWatchers[name]; !ok {
					logger.WithField("service", name).Info("starting watcher for service")

					ctx, cancel := context.WithCancel(c.ctx)
					activeWatchers[name] = cancel
					go c.watchService(ctx, name, isSidecar, clusters, errs)
				}
			}

			// Find services that have an active watcher but
			// are no longer available in catalog and stop
			// their watcher
			for name, stop := range activeWatchers {
				if _, ok := candidates[name]; !ok {
					logger.WithField("service", name).Info("stopping watcher for service")

					stop()
					delete(activeWatchers, name)

					// Notify the cleanup channel asynchronously so that we
					// don't end up blocking if the other side is not
					// able to consume the channel fast enough
					go func(n string) { cleanup <- n }(name)
				}
			}
		}
	}
}

func mapToSlice(m map[string]func()) []string {
	var r []string
	for n := range m {
		r = append(r, n)
	}
	return r
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
			logger.WithField("service", name).Infof("service watcher loop has shut down")
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
			logger.WithField("service", svc.ServiceName).
				WithField("sidecar_destination", svc.ServiceProxy.DestinationServiceName).
				Info("sidecar selected for cluster discovery")

			results[svc.ServiceName] = true
			delete(results, svc.ServiceProxy.DestinationServiceName)
		}
	}

	return results, nil
}
