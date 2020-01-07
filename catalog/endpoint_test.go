package catalog

import (
	"testing"
)

func TestEndpoint_DNSName(t *testing.T) {
	tests := []struct {
		endpoint Endpoint
		expect   string
	}{
		{
			endpoint: Endpoint{isConnect: false, serviceName: "one"},
			expect:   "one." + ServiceZoneName,
		},
		{
			endpoint: Endpoint{isConnect: true, serviceName: "two"},
			expect:   "two." + ConnectZoneName,
		},
	}

	for idx, test := range tests {
		result := test.endpoint.DNSName()
		if result != test.expect {
			t.Errorf("case %d: expected %s, got %s", idx, test.expect, result)
		}
	}
}
