package discovery

import (
	"context"
	"fmt"
	"net"

	"github.com/Gufran/flightpath/log"
	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	dss "github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
)

var logger = log.New("discovery")

func getConsulClient(c *ConsulConfig) (*consul.Client, error) {
	cfg := consul.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s://%s:%d", c.Proto, c.Host, c.Port)
	cfg.Token = c.Token

	return consul.NewClient(cfg)
}

func Start(ctx context.Context, config *Config) (func(), error) {
	cc, err := getConsulClient(config.Consul)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client. %s", err)
	}

	apicache := cache.NewSnapshotCache(false, cache.IDHash{}, log.NewSrvLogger())
	xds := dss.NewServer(apicache, nil)
	server := grpc.NewServer()

	sid, err := registerSelf(cc, config.XDS.ServiceName, config.XDS.ListenPort)
	if err != nil {
		return nil, fmt.Errorf("failed to register the service in consul catalog. %s", err)
	}

	config.XDS.Init(cc, apicache)
	config.XDS.Start(ctx)

	sd.RegisterAggregatedDiscoveryServiceServer(server, xds)
	api.RegisterEndpointDiscoveryServiceServer(server, xds)
	api.RegisterClusterDiscoveryServiceServer(server, xds)
	api.RegisterRouteDiscoveryServiceServer(server, xds)
	api.RegisterListenerDiscoveryServiceServer(server, xds)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.XDS.ListenPort))
	if err != nil {
		return nil, fmt.Errorf("failed to start network listener on port %d. %s", config.XDS.ListenPort, err)
	}

	go func() {
		err := server.Serve(listener)
		if err != nil {
			logger.WithError(err).Error("GRPC server failed")
		}
	}()

	return func() {
		err := deregisterSelf(cc, sid)
		if err != nil {
			logger.WithError(err).Error("failed to deregister the service from consul catalog")
		}

		server.Stop()
		err = listener.Close()
		if err != nil {
			logger.WithError(err).Error("failed to stop the socket listener")
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
