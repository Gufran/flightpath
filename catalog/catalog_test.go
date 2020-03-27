package catalog

import (
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/consul/api"
	"testing"
)

func TestMapToSlice(t *testing.T) {
	tests := []struct {
		in  map[string]func()
		out []string
	}{
		{
			in:  map[string]func(){},
			out: nil,
		},
		{
			in: map[string]func(){
				"one": func() {},
				"two": func() {},
			},
			out: []string{"one", "two"},
		},
	}

	for idx, test := range tests {
		result := mapToSlice(test.in)
		if !cmp.Equal(result, test.out) {
			t.Errorf("case %d: %s", idx, cmp.Diff(result, test.out))
		}
	}
}

func TestAllChecksPassing(t *testing.T) {
	tests := []struct {
		checks api.HealthChecks
		result bool
	}{
		{
			checks: api.HealthChecks{
				{Status: api.HealthCritical},
			},
			result: false,
		},
		{
			checks: api.HealthChecks{
				{Status: api.HealthMaint},
			},
			result: false,
		},
		{
			checks: api.HealthChecks{
				{Status: api.HealthWarning},
			},
			result: false,
		},
		{
			checks: api.HealthChecks{
				{Status: api.HealthPassing},
			},
			result: true,
		},
		{
			checks: api.HealthChecks{
				{Status: api.HealthPassing},
				{Status: api.HealthWarning},
				{Status: api.HealthCritical},
			},
			result: false,
		},
		{
			checks: api.HealthChecks{
				{Status: api.HealthPassing},
				{Status: api.HealthPassing},
				{Status: api.HealthCritical},
			},
			result: false,
		},
		{
			checks: api.HealthChecks{
				{Status: api.HealthPassing},
				{Status: api.HealthPassing},
				{Status: api.HealthPassing},
			},
			result: true,
		},
	}

	for idx, test := range tests {
		result := allChecksPassing(test.checks)
		if result != test.result {
			t.Errorf("case %d: result(%v) != expected(%v)", idx, result, test.result)
		}
	}
}

func TestFilterUnhealthyNodes(t *testing.T) {
	tests := []struct {
		services []*api.CatalogService
		result   []*api.CatalogService
	}{
		{
			services: []*api.CatalogService{
				{
					Checks: api.HealthChecks{
						{Status: api.HealthWarning},
					},
				},
			},
			result: nil,
		},
		{
			services: []*api.CatalogService{
				{
					ID: "healthy-service",
					Checks: api.HealthChecks{
						{Status: api.HealthPassing},
					},
				},
				{
					ID: "warning-service",
					Checks: api.HealthChecks{
						{Status: api.HealthWarning},
					},
				},
				{
					ID: "critical-service",
					Checks: api.HealthChecks{
						{Status: api.HealthCritical},
					},
				},
			},
			result: []*api.CatalogService{
				{
					ID: "healthy-service",
					Checks: api.HealthChecks{
						{Status: api.HealthPassing},
					},
				},
			},
		},
	}

	for idx, test := range tests {
		result := filterUnhealthyNodes(test.services)
		if !cmp.Equal(result, test.result) {
			t.Errorf("case %d: %s", idx, cmp.Diff(result, test.result))
		}
	}
}

func TestIsSidecarProxy(t *testing.T) {
	tests := []struct {
		service *api.CatalogService
		result  bool
	}{
		{
			service: &api.CatalogService{},
			result: false,
		},
		{
			service: &api.CatalogService{
				ServiceProxy: &api.AgentServiceConnectProxyConfig{
					DestinationServiceName: "",
				},
			},
			result: false,
		},
		{
			service: &api.CatalogService{
				ServiceProxy: &api.AgentServiceConnectProxyConfig{
					DestinationServiceName: "destination-service-name",
				},
			},
			result: true,
		},
	}

	for idx, test := range tests {
		result := isSidecarProxy(test.service)
		if result != test.result {
			t.Errorf("case %d: result(%v) != expected(%v)", idx, result, test.result)
		}
	}
}
