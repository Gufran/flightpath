package discovery

import (
	"context"
	"fmt"
	"log"
	"net"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	dss "github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
)

type Config struct {
	ListenPort      int
	ConsulProto     string
	ConsulHost      string
	ConsulPort      int
	ConsulToken     string
	SelfName        string
	NodeName        string
	EnvoyListenPort int
}

func (c *Config) getConsulClient() (*consul.Client, error) {
	cfg := consul.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s://%s:%d", c.ConsulProto, c.ConsulHost, c.ConsulPort)
	cfg.Token = c.ConsulToken

	return consul.NewClient(cfg)
}

func Start(ctx context.Context, config *Config) (func(), error) {
	cc, err := config.getConsulClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client. %s", err)
	}

	// TODO: inject a logger that honors the global
	//   log level configuration
	apicache := cache.NewSnapshotCache(false, cache.IDHash{}, &ServerLog{})
	xds := dss.NewServer(apicache, nil)
	server := grpc.NewServer()

	sid, err := registerSelf(cc, config.SelfName, config.ListenPort)
	if err != nil {
		return nil, fmt.Errorf("failed to register the service in consul catalog. %s", err)
	}

	x := &XDS{
		Consul:            cc,
		Cache:             apicache,
		ServiceName:       config.SelfName,
		ProxyNodeName:     config.NodeName,
		ProxyListenerPort: config.EnvoyListenPort,
	}

	x.Start(ctx)

	sd.RegisterAggregatedDiscoveryServiceServer(server, xds)
	api.RegisterEndpointDiscoveryServiceServer(server, xds)
	api.RegisterClusterDiscoveryServiceServer(server, xds)
	api.RegisterRouteDiscoveryServiceServer(server, xds)
	api.RegisterListenerDiscoveryServiceServer(server, xds)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.ListenPort))
	if err != nil {
		return nil, fmt.Errorf("failed to start network listener on port %d. %s", config.ListenPort, err)
	}

	go func() {
		err := server.Serve(listener)
		if err != nil {
			log.Printf("GRPC server failed with error: %s", err)
		}
	}()

	return func() {
		err := deregisterSelf(cc, sid)
		if err != nil {
			log.Printf("failed to deregister the service from consul catalog. %s", err)
		}

		server.Stop()
		err = listener.Close()
		if err != nil {
			log.Printf("failed to stop the socket listener. %s", err)
		}
	}, nil
}

func registerSelf(cc *consul.Client, name string, port int) (string, error) {
	reg := &consul.AgentServiceRegistration{
		ID:   uuid.New().String(),
		Name: name,
		Kind: consul.ServiceKindTypical,
		Port: port,
		Connect: &consul.AgentServiceConnect{
			Native: true,
		},

		// TODO: register health checks at least for the TCP
		//  socket connection. Expand the health checks to
		//  cover the GRPC interface and some basic endpoints
		//  so that failures can be detected early on
	}

	err := cc.Agent().ServiceRegister(reg)
	if err != nil {
		return "", err
	}

	return reg.ID, nil
}

func deregisterSelf(cc *consul.Client, sid string) error {
	return cc.Agent().ServiceDeregister(sid)
}
