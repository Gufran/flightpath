package catalog

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
	"testing"
)

func TestCatalog_WatchTLS(t *testing.T) {
	serviceName := "test-service"
	wke := fmt.Errorf("error")

	cases := map[string][]ConnectCALeafResult{
		serviceName: {
			{
				cert: nil,
				meta: nil,
				err:  wke,
			},
			{
				cert: nil,
				meta: &api.QueryMeta{
					LastIndex: 0,
				},
				err: nil,
			},
			{
				cert: &api.LeafCert{
					CertPEM:       "cert material",
					PrivateKeyPEM: "pkey material",
				},
				meta: &api.QueryMeta{
					LastIndex: 1,
				},
				err: nil,
			},
		},
	}

	ctx, cancel := context.WithCancel(context.TODO())
	catalog := &Catalog{
		ctx:     ctx,
		connect: NewCertFinderMock(ctx, t, cases, true),
	}

	certChan := make(chan TLSInfo)
	errChan := make(chan error)

	doneChan := make(chan struct{})

	go func() {
		catalog.WatchTLS(serviceName, certChan, errChan)
		doneChan <- struct{}{}
	}()

	err := <-errChan
	if err.Error() != "failed to query consul catalog for leaf certificates. error" {
		t.Errorf("unexpected error message %q", err)
	}

	cert := <-certChan
	if cert.Cert() != "cert material" {
		t.Errorf("unexpected certificated material %q", cert.Cert())
	}

	if cert.PKey() != "pkey material" {
		t.Errorf("unexpected pkey material %q", cert.PKey())
	}

	close(certChan)
	close(errChan)

	cancel()
	<-doneChan
}
