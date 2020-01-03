package catalog

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"log"
	"time"
)

type TLSInfo interface {
	Serial() string
	Cert() string
	PKey() string
	ServiceName() string
}

var _ TLSInfo = &tls{}

type tls struct {
	serial  string
	certPem string
	pkeyPem string
	service string
}

func (t *tls) Serial() string {
	return t.serial
}

func (t *tls) Cert() string {
	return t.certPem
}

func (t *tls) PKey() string {
	return t.pkeyPem
}

func (t *tls) ServiceName() string {
	return t.service
}

func (c *Catalog) WatchTLS(service string, cert chan<- TLSInfo, errs chan<- error) {
	qopts := &api.QueryOptions{
		AllowStale:        false,
		RequireConsistent: true,
		WaitIndex:         0,
		WaitTime:          30 * time.Second,
	}

	for {
		select {
		case <-c.ctx.Done():
			log.Println("Shutting down Leaf Certificate watcher because the context is done")
			return
		default:
			resp, meta, err := c.client.Agent().ConnectCALeaf(service, qopts.WithContext(c.ctx))
			if err != nil {
				errs <- fmt.Errorf("failed to query consul catalog for leaf certificates. %s", err)
				break
			}

			if meta.LastIndex <= qopts.WaitIndex {
				break
			}

			qopts.WaitIndex = meta.LastIndex
			cert <- &tls{
				serial:  resp.SerialNumber,
				certPem: resp.CertPEM,
				pkeyPem: resp.PrivateKeyPEM,
				service: service,
			}
		}
	}
}
