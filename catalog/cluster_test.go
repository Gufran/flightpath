package catalog

import (
	"github.com/hashicorp/consul/api"
	"reflect"
	"testing"
)

func TestGetRouteInfo(t *testing.T) {
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
