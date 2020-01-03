package catalog

import (
	"crypto/sha1"
	"fmt"
	"github.com/hashicorp/consul/api"
	"sort"
	"strings"
)

const (
	ConnectZoneName = "connect.consul"
	ServiceZoneName = "service.consul"
)

// ClusterInfo represents a collection of service instances
// registered in consul catalog. A cluster represents
// only one type of service.
type ClusterInfo interface {
	Name() string
	Endpoints() []Endpoint
	IsConnectEnabled() bool
	Hash() string
}

var _ ClusterInfo = &Cluster{}

type Cluster struct {
	name      string
	isConnect bool
	services  []*api.CatalogService
}

func Hash(l []ClusterInfo) string {
	var ids []string
	for _, s := range l {
		ids = append(ids, s.Hash())
	}
	sort.Strings(ids)
	cid := strings.Join(ids, "")

	return fmt.Sprintf("%x", sha1.Sum([]byte(cid)))
}

func (c *Cluster) Hash() string {
	var ids []string
	for _, s := range c.services {
		ids = append(ids, s.ID)
	}
	sort.Strings(ids)
	cid := strings.Join(ids, "")

	return fmt.Sprintf("%x", sha1.Sum([]byte(cid)))
}

func (c *Cluster) Name() string {
	return c.name
}

func (c *Cluster) Endpoints() []Endpoint {
	var results []Endpoint
	for _, service := range c.services {
		routing := getRoutingInfo(service)
		results = append(results, Endpoint{
			name:    service.ID,
			serviceName: service.ServiceName,
			isConnect: c.IsConnectEnabled(),
			addr:    service.Address,
			port:    service.ServicePort,
			routing: routing,
		})
	}

	return results
}

func getRoutingInfo(service *api.CatalogService) map[string][]string {
	results := map[string][]string{}
	for k, v := range service.ServiceMeta {
		if !strings.HasPrefix(k, "flightpath-route") {
			continue
		}

		domain := "*"
		uriMatch := "/"
		if strings.HasPrefix(v, "/") {
			// value is a path match
			uriMatch = v
		} else {
			// value has a domain and potentially a path
			idx := strings.Index(v, "/")
			if idx == -1 {
				domain = v
			} else {
				domain = v[:idx]
				uriMatch = v[idx:]
			}
		}

		if _, ok := results[domain]; !ok {
			results[domain] = []string{}
		}

		results[domain] = append(results[domain], uriMatch)
	}

	return results
}

func (c *Cluster) IsConnectEnabled() bool {
	return c.isConnect
}
