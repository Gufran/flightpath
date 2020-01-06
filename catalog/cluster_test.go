package catalog

import (
	"github.com/hashicorp/consul/api"
	"reflect"
	"testing"
)

func TestHash(t *testing.T) {
	tests := []struct {
		clusters []ClusterInfo
		hash     string
	}{
		{
			hash: "6d184a4d36feb2f642c738f1d212a2552bbc13b3",
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
			hash: "fcb13b7761df6dd7a574c02208f40a7e933106ce",
		},
		{
			cluster: &Cluster{
				services: []*api.CatalogService{
					{ID: "one"},
					{ID: "two"},
				},
			},
			hash: "30ae97492ce1da88d0e7117ace0a60a6f9e1e0bc",
		},
		{
			cluster: &Cluster{
				services: []*api.CatalogService{
					{ID: "one"},
					{ID: "four"},
				},
			},
			hash: "257c7fd00c8e62af9a17fcc5d52f566c7d5c5b75",
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
		if !reflect.DeepEqual(result, test.expect) {
			t.Errorf("case %d: unexpected result. Expected %#v != Result %#v", idx, test.expect, result)
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
					routing: map[string][]string{},
				},
				{
					name:        "case-2-id-2",
					serviceName: "case-2-id-2",
					isConnect:   true,
					addr:        "2.2",
					port:        22,
					routing: map[string][]string{},
				},
			},
		},
	}

	for idx, test := range tests {
		result := test.cluster.Endpoints()
		if !reflect.DeepEqual(result, test.expect) {
			t.Errorf("case %d: unexpected %#v, expected %#v", idx, result, test.expect)
		}
	}
}
