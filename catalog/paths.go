package catalog

import (
	"context"
	"github.com/hashicorp/consul/api"
)

type RouteFinder interface {
	List(string, *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error)
}

type Route struct {
	cluster    string
	domains    []string
	pathPrefix []string
}

type RouteStorage struct {
	ctx    context.Context
	prefix string
	finder RouteFinder
}

func NewRouteStorage(ctx context.Context, prefix string, client *api.Client) *RouteStorage {
	return &RouteStorage{
		ctx:    ctx,
		prefix: prefix,
		finder: client.KV(),
	}
}

func (i *RouteStorage) WatchRoutes(routes chan <- []Route, errs chan <-error) {
	// TODO: work in progress
	//   evaluating the idea of managing the routes externally from consul kv
	//   rather than relying on the service metadata
}
