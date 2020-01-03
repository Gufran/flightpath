package catalog

import "fmt"

// Endpoint is the network address and port
// of the service instance.
type Endpoint struct {
	name        string
	serviceName string
	isConnect   bool
	addr        string
	port        int
	routing     map[string][]string
}

func (e *Endpoint) Name() string {
	return e.name
}

func (e *Endpoint) Addr() string {
	return e.addr
}

func (e *Endpoint) DNSName() string {
	if e.isConnect {
		return fmt.Sprintf("%s.%s", e.serviceName, ConnectZoneName)
	}
	return fmt.Sprintf("%s.%s", e.serviceName, ServiceZoneName)
}

func (e *Endpoint) Port() int {
	return e.port
}

func (e *Endpoint) RoutingInfo() map[string][]string {
	return e.routing
}
