package catalog

import (
	"github.com/Gufran/flightpath/metrics"
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

func (c *Catalog) WatchTLS(service string, cert chan<- TLSInfo) {
	qopts := &api.QueryOptions{
		AllowStale:        false,
		RequireConsistent: true,
		WaitIndex:         0,
		WaitTime:          30 * time.Second,
	}

	for {
		metrics.Incr("catalog.tls.loop", nil)
		select {
		case <-c.ctx.Done():
			log.Println("Shutting down Leaf Certificate watcher because the context is done")
			return
		default:
			resp, meta, err := c.connect.ConnectCALeaf(service, qopts.WithContext(c.ctx))
			if err != nil {
				metrics.Incr("catalog.tls.error.fetch", nil)
				logger.WithError(err).Error("failed to fetch TLS leaf certificates")
				break
			}

			if meta.LastIndex <= qopts.WaitIndex {
				metrics.Incr("catalog.tls.noop", nil)
				break
			}

			qopts.WaitIndex = meta.LastIndex

			metrics.Incr("catalog.tls.updated", nil)

			cert <- &tls{
				serial:  resp.SerialNumber,
				certPem: resp.CertPEM,
				pkeyPem: resp.PrivateKeyPEM,
				service: service,
			}
		}
	}
}
