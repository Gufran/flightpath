package catalog

import (
	"crypto/sha1"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/mitchellh/mapstructure"
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
	Settings() (*ClusterSettings, error)
}

var _ ClusterInfo = &Cluster{}

type Cluster struct {
	name      string
	isConnect bool
	services  []*api.CatalogService
}

type ClusterSettings struct {
	ConnTimeout          int64  `mapstructure:"flightpath-cluster-conn_timeout"`
	PerConnBufLimitBytes uint32 `mapstructure:"flightpath-cluster-per_conn_buf_limit_bytes"`
	MaxReqPerConn        uint32 `mapstructure:"flightpath-cluster-max_req_per_conn"`

	TcpKeepaliveProbes   uint32 `mapstructure:"flightpath-cluster-tcp_keepalive_probes"`
	TcpKeepaliveTime     uint32 `mapstructure:"flightpath-cluster-tcp_keepalive_time"`
	TcpKeepaliveInterval uint32 `mapstructure:"flightpath-cluster-tcp_keepalive_interval"`

	RetryOn             string `mapstructure:"flightpath-retry-on"`
	RetryAttempts       uint32 `mapstructure:"flightpath-retry-attempts"`
	RetryAttemptTimeout int64  `mapstructure:"flightpath-retry-per_try_timeout"`
	RetryBackoffBase    int64  `mapstructure:"flightpath-retry-backoff_base_interval"`
	RetryBackoffMax     int64  `mapstructure:"flightpath-retry-backoff_max_interval"`
}

// TODO: right now we don't care about new or old cluster
//   settings. If these settings change across deployments
//   we'll just pick the latest one and work with that.
//   An upcoming version will support creating multiple
//   clusters per service e.g. canary, strand, etc
//   which will stick with their own cluster settings.
func (c *Cluster) Settings() (*ClusterSettings, error) {
	var (
		settings        = map[string]string{}
		result          = new(ClusterSettings)
		latest   uint64 = 0
	)

	for _, s := range c.services {
		if latest < s.CreateIndex {
			latest = s.CreateIndex
			settings = s.ServiceMeta
		}
	}

	err := mapstructure.WeakDecode(settings, result)
	if err != nil {
		return nil, err
	}

	result.Canonicalize()
	return result, nil
}

func (cs *ClusterSettings) Canonicalize() {
	if cs.ConnTimeout == 0 {
		cs.ConnTimeout = 10
	}

	if cs.PerConnBufLimitBytes == 0 {
		cs.PerConnBufLimitBytes = 32768
	}

	if cs.MaxReqPerConn == 0 {
		cs.MaxReqPerConn = 10000
	}

	if cs.TcpKeepaliveProbes == 0 {
		cs.TcpKeepaliveProbes = 9
	}

	if cs.TcpKeepaliveTime == 0 {
		cs.TcpKeepaliveTime = 5 * 60
	}

	if cs.TcpKeepaliveInterval == 0 {
		cs.TcpKeepaliveInterval = 90
	}

	if cs.RetryOn != "" {
		if cs.RetryAttempts == 0 {
			cs.RetryAttempts = 3
		}

		if cs.RetryAttemptTimeout == 0 {
			cs.RetryAttemptTimeout = 5
		}

		if cs.RetryBackoffBase == 0 {
			cs.RetryBackoffBase = 1
		}

		if cs.RetryBackoffMax == 0 {
			cs.RetryBackoffMax = 6
		}
	}
}

func Hash(l []ClusterInfo) string {
	var ids []string
	for _, s := range l {
		ids = append(ids, s.Hash())
	}
	sort.Strings(ids)
	cid := strings.Join(ids, "")

	return fmt.Sprintf("%x", sha1.Sum([]byte(cid)))[:10]
}

func (c *Cluster) Hash() string {
	var ids []string
	for _, s := range c.services {
		ids = append(ids, s.ID)
	}
	sort.Strings(ids)
	cid := strings.Join(ids, "")

	return fmt.Sprintf("%x", sha1.Sum([]byte(cid)))[:10]
}

func (c *Cluster) Name() string {
	return c.name
}

func (c *Cluster) Endpoints() []Endpoint {
	var results []Endpoint
	for _, service := range c.services {
		routing := getRoutingInfo(service)
		results = append(results, Endpoint{
			name:        service.ID,
			serviceName: service.ServiceName,
			isConnect:   c.IsConnectEnabled(),
			addr:        service.Address,
			port:        service.ServicePort,
			routing:     routing,
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
