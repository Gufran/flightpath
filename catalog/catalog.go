package catalog

import (
	"context"
	"fmt"
	"github.com/Gufran/flightpath/log"
	"github.com/Gufran/flightpath/metrics"
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
func (c *Catalog) DiscoverClusters(clusters chan<- ClusterInfo, cleanup chan<- string) {
	qopts := &api.QueryOptions{
		AllowStale:        false,
		RequireConsistent: true,
		WaitIndex:         0,
		WaitTime:          30 * time.Second,
	}

	activeWatchers := map[string]func(){}

	for {
		metrics.Incr("catalog.discovery.clusters.loop", nil)
		select {
		case <-c.ctx.Done():
			logger.Info("cluster discovery loop has shut down")
			return

		default:
			services, meta, err := c.catalog.Services(qopts.WithContext(c.ctx))
			if err != nil {
				metrics.Incr("catalog.discovery.clusters.error.fetch", nil)
				logger.WithError(err).Error("failed to fetch list of services from consul catalog")
				time.Sleep(3 * time.Second)
				break
			}

			if meta.LastIndex <= qopts.WaitIndex {
				metrics.Incr("catalog.discovery.clusters.noop", nil)
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

			metrics.GaugeI("catalog.discovery.clusters.candidate", len(candidates), nil)

			// Remove all services where there is a corresponding
			// connect proxy registered in catalog
			candidates, err = c.filterConnectTargets(candidates)
			if err != nil {
				metrics.Incr("catalog.discovery.clusters.error.filter_connect", nil)
				logger.WithError(err).Error("failed to filter connect target services")
				time.Sleep(3 * time.Second)
				break
			}

			metrics.GaugeI("catalog.discovery.clusters.targets", len(candidates), nil)
			logger.WithField("total_candidates", len(candidates)).
				WithField("candidates", candidates).
				Debug("candidate list is ready")

			metrics.GaugeI("catalog.discovery.clusters.watchers", len(activeWatchers), nil)
			logger.WithField("watchers", mapToSlice(activeWatchers)).
				Debug("watcher state")

			// Find services that dont' have an active watcher
			// and start a watcher for them.
			for name, isSidecar := range candidates {
				if _, ok := activeWatchers[name]; !ok {
					metrics.Incr("catalog.discovery.clusters.watcher.new", []string{"service:"+name})
					logger.WithField("service", name).Info("starting watcher for service")

					ctx, cancel := context.WithCancel(c.ctx)
					activeWatchers[name] = cancel
					go c.watchService(ctx, name, isSidecar, clusters)
				}
			}

			// Find services that have an active watcher but
			// are no longer available in catalog and stop
			// their watcher
			for name, stop := range activeWatchers {
				if _, ok := candidates[name]; !ok {
					metrics.Incr("catalog.discovery.clusters.watcher.closing", []string{"service:"+name})
					logger.WithField("service", name).Info("stopping watcher for service")

					// Notify the cleanup channel asynchronously so that we
					// don't end up blocking if the other side is not
					// able to consume the channel fast enough
					go func(n string) { cleanup <- n }(name)

					stop()
					delete(activeWatchers, name)
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

func (c *Catalog) watchService(ctx context.Context, name string, isSidecar bool, clusters chan<- ClusterInfo) {
	qopts := &api.QueryOptions{
		AllowStale:        false,
		RequireConsistent: true,
		WaitIndex:         0,
		WaitTime:          30 * time.Second,
	}

	tags := []string{
		"service:"+name,
		fmt.Sprintf("is_sidecar:%s", isSidecar),
	}

	for {
		metrics.Incr("catalog.discovery.service.loop", tags)
		select {
		case <-ctx.Done():
			logger.WithField("service", name).Infof("service watcher loop has shut down")
			return

		default:
			nodes, meta, err := c.catalog.Service(name, "", qopts.WithContext(c.ctx))
			if err != nil {
				metrics.Incr("catalog.discovery.service.error.fetch", tags)
				logger.WithError(err).WithField("service", name).Error("failed to fetch service definition")
				time.Sleep(3 * time.Second)
				break
			}

			if meta.LastIndex <= qopts.WaitIndex {
				metrics.Incr("catalog.discovery.service.noop", tags)
				break
			}

			qopts.WaitIndex = meta.LastIndex

			metrics.Incr("catalog.discovery.service.updated", tags)
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
