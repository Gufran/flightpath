package catalog

import (
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHash(t *testing.T) {
	tests := []struct {
		clusters []ClusterInfo
		hash     string
	}{
		{
			hash: "c08f6090b8",
			clusters: []ClusterInfo{
				&Cluster{
					services: []*api.CatalogService{
						{ID: "one"},
						{ID: "two"},
						{ID: "three"},
						{ID: "four"},
					},
				},
				&Cluster{
					services: []*api.CatalogService{
						{ID: "one"},
						{ID: "two"},
					},
				},
				&Cluster{
					services: []*api.CatalogService{
						{ID: "three"},
						{ID: "four"},
					},
				},
			},
		},
	}

	for idx, test := range tests {
		result := Hash(test.clusters)
		if result != test.hash {
			t.Errorf("case %d: unexpected hash %q, expected %q", idx, result, test.hash)
		}
	}
}

func TestCluster_Hash(t *testing.T) {
	tests := []struct {
		cluster *Cluster
		hash    string
	}{
		{
			cluster: &Cluster{
				services: []*api.CatalogService{
					{ID: "one"},
					{ID: "two"},
					{ID: "three"},
					{ID: "four"},
				},
			},
			hash: "fcb13b7761",
		},
		{
			cluster: &Cluster{
				services: []*api.CatalogService{
					{ID: "one"},
					{ID: "two"},
				},
			},
			hash: "30ae97492c",
		},
		{
			cluster: &Cluster{
				services: []*api.CatalogService{
					{ID: "one"},
					{ID: "four"},
				},
			},
			hash: "257c7fd00c",
		},
	}

	for idx, test := range tests {
		result := test.cluster.Hash()
		if result != test.hash {
			t.Errorf("case %d: unexpected cluster hash %q, expected %q", idx, result, test.hash)
		}
	}
}

func TestGetRoutingInfo(t *testing.T) {
	tests := []struct {
		service api.CatalogService
		expect  map[string][]string
	}{
		{
			service: api.CatalogService{
				ServiceMeta: map[string]string{
					"flightpath-route-1": "/only-uri",
				},
			},
			expect: map[string][]string{
				"*": {
					"/only-uri",
				},
			},
		},
		{
			service: api.CatalogService{
				ServiceMeta: map[string]string{
					"flightpath-route-1": "some-domain",
				},
			},
			expect: map[string][]string{
				"some-domain": {
					"/",
				},
			},
		},
		{
			service: api.CatalogService{
				ServiceMeta: map[string]string{
					"flightpath-route-1": "/uri-one",
					"flightpath-route-2": "/uri-prefix-one/*",
					"flightpath-route-3": "/uri-two",
					"flightpath-route-4": "/uri-two/subresource",
				},
			},
			expect: map[string][]string{
				"*": {
					"/uri-one",
					"/uri-prefix-one/*",
					"/uri-two",
					"/uri-two/subresource",
				},
			},
		},
		{
			service: api.CatalogService{
				ServiceMeta: map[string]string{
					"flightpath-route-1": "some-domain/uri-one",
					"flightpath-route-2": "some-domain/uri-prefix-one/*",
					"flightpath-route-3": "other-domain/uri-two",
					"flightpath-route-4": "other-domain/uri-two/subresource",
					"flightpath-route-5": "another-domain",
					"flightpath-route-6": "another-domain/*",
					"flightpath-route-7": "/*",
					"flightpath-route-8": "/prefix/*",
				},
			},
			expect: map[string][]string{
				"*": {
					"/*",
					"/prefix/*",
				},
				"some-domain": {
					"/uri-one",
					"/uri-prefix-one/*",
				},
				"other-domain": {
					"/uri-two",
					"/uri-two/subresource",
				},
				"another-domain": {
					"/",
					"/*",
				},
			},
		},
	}

	for idx, test := range tests {
		result := getRoutingInfo(&test.service)
		for k, v := range test.expect {
			if slice, ok := result[k]; !ok {
				t.Errorf("Case %d: Key %s is not present in map", idx, k)
			} else {
				assert.ElementsMatch(t, v, slice, "Case %d", idx)
			}
		}
	}
}

func TestCluster_Endpoints(t *testing.T) {
	tests := []struct {
		cluster Cluster
		expect  []Endpoint
	}{
		{
			cluster: Cluster{
				isConnect: false,
				services: []*api.CatalogService{
					{
						ID:          "case-1-id-1",
						ServiceName: "case-1-id-1",
						Address:     "1.1",
						ServicePort: 11,
						ServiceMeta: map[string]string{
							"flightpath-route-one": "just-domain",
							"flightpath-route-two": "domain/fixed-path",
						},
					},
					{
						ID:          "case-1-id-2",
						ServiceName: "case-1-id-2",
						Address:     "1.2",
						ServicePort: 12,
						ServiceMeta: map[string]string{
							"flightpath-route-two": "/path-prefix/",
						},
					},
				},
			},
			expect: []Endpoint{
				{
					name:        "case-1-id-1",
					serviceName: "case-1-id-1",
					isConnect:   false,
					addr:        "1.1",
					port:        11,
					routing: map[string][]string{
						"just-domain": {
							"/",
						},
						"domain": {
							"/fixed-path",
						},
					},
				},
				{
					name:        "case-1-id-2",
					serviceName: "case-1-id-2",
					isConnect:   false,
					addr:        "1.2",
					port:        12,
					routing: map[string][]string{
						"*": {
							"/path-prefix/",
						},
					},
				},
			},
		},
		{
			cluster: Cluster{
				isConnect: true,
				services: []*api.CatalogService{
					{
						ID:          "case-2-id-1",
						ServiceName: "case-2-id-1",
						Address:     "2.1",
						ServicePort: 21,
						ServiceMeta: map[string]string{},
					},
					{
						ID:          "case-2-id-2",
						ServiceName: "case-2-id-2",
						Address:     "2.2",
						ServicePort: 22,
						ServiceMeta: map[string]string{},
					},
				},
			},
			expect: []Endpoint{
				{
					name:        "case-2-id-1",
					serviceName: "case-2-id-1",
					isConnect:   true,
					addr:        "2.1",
					port:        21,
					routing:     map[string][]string{},
				},
				{
					name:        "case-2-id-2",
					serviceName: "case-2-id-2",
					isConnect:   true,
					addr:        "2.2",
					port:        22,
					routing:     map[string][]string{},
				},
			},
		},
	}

	for idx, test := range tests {
		result := test.cluster.Endpoints()
		assert.ElementsMatch(t, result, test.expect, "Case %d", idx)
	}
}

func TestDecodeClusterSettings(t *testing.T) {
	tests := []struct {
		cluster *Cluster
		expect  *ClusterSettings
	}{
		{
			cluster: &Cluster{
				services: []*api.CatalogService{
					{
						CreateIndex: 3,
						ServiceMeta: map[string]string{},
					},
					{
						CreateIndex: 1,
						ServiceMeta: map[string]string{},
					},
					{
						CreateIndex: 5,
						ServiceMeta: map[string]string{
							"flightpath-cluster-conn_timeout":           "16",
							"flightpath-cluster-tcp_keepalive_time":     "77",
							"flightpath-cluster-tcp_keepalive_interval": "56",
							"flightpath-retry-on":                       "5xx",
						},
					},
					{
						CreateIndex: 2,
						ServiceMeta: map[string]string{},
					},
				},
			},
			expect: &ClusterSettings{
				ConnTimeout:          16,
				PerConnBufLimitBytes: 32768,
				MaxReqPerConn:        10000,
				TcpKeepaliveProbes:   9,
				TcpKeepaliveTime:     77,
				TcpKeepaliveInterval: 56,
				RetryOn:              "5xx",
				RetryAttempts:        3,
				RetryAttemptTimeout:  5,
				RetryBackoffBase:     1,
				RetryBackoffMax:      6,
			},
		},
		{
			cluster: &Cluster{
				services: []*api.CatalogService{
					{
						CreateIndex: 3,
						ServiceMeta: map[string]string{},
					},
					{
						CreateIndex: 1,
						ServiceMeta: map[string]string{},
					},
					{
						CreateIndex: 5,
						ServiceMeta: map[string]string{
							"flightpath-cluster-conn_timeout":             "16",
							"flightpath-cluster-per_conn_buf_limit_bytes": "1212",
							"flightpath-cluster-max_req_per_conn":         "33",
							"flightpath-cluster-tcp_keepalive_probes":     "7",
							"flightpath-cluster-tcp_keepalive_time":       "77",
							"flightpath-cluster-tcp_keepalive_interval":   "56",
							"flightpath-retry-on":                         "5xx,gateway-error",
							"flightpath-retry-attempts":                   "4",
							"flightpath-retry-per_try_timeout":            "8",
							"flightpath-retry-backoff_base_interval":      "2",
							"flightpath-retry-backoff_max_interval":       "11",
						},
					},
					{
						CreateIndex: 2,
						ServiceMeta: map[string]string{},
					},
				},
			},
			expect: &ClusterSettings{
				ConnTimeout:          16,
				PerConnBufLimitBytes: 1212,
				MaxReqPerConn:        33,
				TcpKeepaliveProbes:   7,
				TcpKeepaliveTime:     77,
				TcpKeepaliveInterval: 56,
				RetryOn:              "5xx,gateway-error",
				RetryAttempts:        4,
				RetryAttemptTimeout:  8,
				RetryBackoffBase:     2,
				RetryBackoffMax:      11,
			},
		},
	}

	for idx, test := range tests {
		settings, err := test.cluster.Settings()
		if err != nil {
			t.Errorf("case %d: failed to decode cluster settings. %s", idx, err)
		}

		if !cmp.Equal(settings, test.expect) {
			t.Errorf("case %d: unexpected result. %s", idx, cmp.Diff(settings, test.expect))
		}
	}
}
