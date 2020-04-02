package discovery

import (
	"context"
	"fmt"
	"github.com/Gufran/flightpath/catalog"
	"github.com/Gufran/flightpath/metrics"
	envoyapiv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	accesslogconfig "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslogfilter "github.com/envoyproxy/go-control-plane/envoy/config/filter/accesslog/v2"
	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
	duration "github.com/golang/protobuf/ptypes/duration"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/golang/protobuf/ptypes/wrappers"
	"strings"
	"time"
)

const XdsClusterName = "xds_cluster"

type SyncChans struct {
	cluster chan catalog.ClusterInfo
	tls     chan catalog.TLSInfo
	cleanup chan string
}

func NewSyncChans() *SyncChans {
	return &SyncChans{
		cluster: make(chan catalog.ClusterInfo),
		tls:     make(chan catalog.TLSInfo),
		cleanup: make(chan string),
	}
}

func (x *XDS) Start(ctx context.Context) {
	source := catalog.NewCatalog(ctx, x.Consul)

	ch := NewSyncChans()

	go source.DiscoverClusters(ch.cluster, ch.cleanup)
	go source.WatchTLS(x.ServiceName, ch.tls)

	if x.Debug.Enable {
		StartDebugServer(x.Debug.Port, x.Envoy.NodeName, x.Cache)
	}

	go synchronize(ctx, x.Cache, ch, x.Envoy)
}

func synchronize(ctx context.Context, snc cache.SnapshotCache, ch *SyncChans, envoyConfig *EnvoyConfig) {
	// TLS info is absolutely necessary and since we know that we
	// are registered as a connect enabled service and guaranteed
	// to receive a certificate pair, we'll just wait for it to
	// arrive before setting up the shop.
	certs := <-ch.tls

	tick := 1 * time.Second
	timer := time.NewTimer(tick)
	resetTimer := func() {
		timer.Reset(tick)
	}

	knownClusters := map[string]catalog.ClusterInfo{}

	for {
		metrics.Incr("discovery.sync.loop", nil)
		select {
		case <-ctx.Done():
			return

		case cluster := <-ch.cluster:
			resetTimer()
			logger.WithField("cluster", cluster.Name()).Info("updating cluster entry")
			metrics.Incr("discovery.cluster.update", []string{"cluster:" + cluster.Name()})
			metrics.GaugeI("discovery.cluster.endpoints.count", len(cluster.Endpoints()), []string{"cluster:" + cluster.Name()})
			knownClusters[cluster.Name()] = cluster

		case certs = <-ch.tls:
			resetTimer()
			metrics.Incr("discovery.tls.update", nil)

		case name := <-ch.cleanup:
			resetTimer()
			metrics.Incr("discovery.cluster.cleanup", []string{"cluster:" + name})
			logger.WithField("cluster", name).Info("removing cluster from tracked list")
			if _, ok := knownClusters[name]; ok {
				delete(knownClusters, name)
			}

		case <-timer.C:
			metrics.Incr("discovery.cluster.flush", nil)
			metrics.GaugeI("discovery.cluster.batch_size", len(knownClusters), nil)

			logger.Info("flushing cluster configuration to xDS server")
			err := putCache(snc, envoyConfig, clustersList(knownClusters), certs)
			if err != nil {
				metrics.Incr("discovery.cluster.error.flush", nil)
				logger.WithError(err).Error("failed to update cluster information")
			}
		}
	}
}

func clustersList(cl map[string]catalog.ClusterInfo) []catalog.ClusterInfo {
	var result []catalog.ClusterInfo
	for _, c := range cl {
		result = append(result, c)
	}
	return result
}

func putCache(snc cache.SnapshotCache, envoyConfig *EnvoyConfig, clusters []catalog.ClusterInfo, tls catalog.TLSInfo) error {
	var (
		// NOTE: actual type is []envoyapiv2.Cluster
		clusterResource []cache.Resource
		// NOTE: actual type is []envoyapiv2.RouteConfiguration
		routeResource []cache.Resource
		// NOTE: actual type is []envoyapiv2.Listener
		listenerResource []cache.Resource
		// NOTE actual type is []envoyapiv2.ClusterLoadAssignment
		endpointResource []cache.Resource
	)

	defer metrics.Timed("discovery.cache.put_ns", time.Now(), nil)

	vhosts := vhostPool{
		vhs: []vhost{},
	}

	envoyListener, err := buildListener("flightpath", envoyConfig)
	if err != nil {
		return fmt.Errorf("failed to build cluster definition. %s", err)
	}

	listenerResource = []cache.Resource{
		envoyListener,
	}

	for _, service := range clusters {
		clusterConfig := buildCluster(service)

		if service.IsConnectEnabled() {
			clusterConfig.TransportSocket, err = buildTransportSocket(tls)
			if err != nil {
				return err
			}
		}

		vhosts.add(service)

		clusterResource = append(clusterResource, clusterConfig)
		endpointResource = append(endpointResource, &envoyapiv2.ClusterLoadAssignment{
			ClusterName: service.Name(),
			Endpoints:   buildEndpoints(service.Endpoints()),
		})
	}

	routeResource = []cache.Resource{
		&envoyapiv2.RouteConfiguration{
			Name:         "upstream",
			VirtualHosts: vhosts.collect(envoyConfig.ListenerPort),
		},
	}

	metrics.GaugeI("discovery.cache.put.clusters", len(clusterResource), nil)
	metrics.GaugeI("discovery.cache.put.endpoints", len(endpointResource), nil)
	metrics.GaugeI("discovery.cache.put.routes", len(routeResource), nil)
	metrics.GaugeI("discovery.cache.put.listener", len(listenerResource), nil)

	snap := cache.NewSnapshot(catalog.Hash(clusters), endpointResource, clusterResource, routeResource, listenerResource)
	return snc.SetSnapshot(envoyConfig.NodeName, snap)
}

func buildCluster(service catalog.ClusterInfo) *envoyapiv2.Cluster {
	cluster := &envoyapiv2.Cluster{
		Name:                          service.Name(),
		LbPolicy:                      envoyapiv2.Cluster_ROUND_ROBIN,
		RespectDnsTtl:                 true,
		DrainConnectionsOnHostRemoval: true,
		ClusterDiscoveryType: &envoyapiv2.Cluster_Type{
			Type: envoyapiv2.Cluster_EDS,
		},
		EdsClusterConfig: &envoyapiv2.Cluster_EdsClusterConfig{
			ServiceName: service.Name(),
			EdsConfig: &core.ConfigSource{
				ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
					ApiConfigSource: &core.ApiConfigSource{
						ApiType: core.ApiConfigSource_GRPC,
						GrpcServices: []*core.GrpcService{
							{
								TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
									EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
										ClusterName: XdsClusterName,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	settings, err := service.Settings()
	if err != nil {
		logger.WithError(err).
			WithField("service", service.Name()).
			Errorf("failed to load cluster settings")
		return cluster
	}

	cluster.ConnectTimeout = &duration.Duration{
		Seconds: settings.ConnTimeout,
	}

	cluster.PerConnectionBufferLimitBytes = &wrappers.UInt32Value{
		Value: settings.PerConnBufLimitBytes,
	}

	cluster.MaxRequestsPerConnection = &wrappers.UInt32Value{
		Value: settings.MaxReqPerConn,
	}

	cluster.UpstreamConnectionOptions = &envoyapiv2.UpstreamConnectionOptions{
		TcpKeepalive: &core.TcpKeepalive{
			KeepaliveProbes: &wrappers.UInt32Value{
				Value: settings.TcpKeepaliveProbes,
			},
			KeepaliveTime: &wrappers.UInt32Value{
				Value: settings.TcpKeepaliveTime,
			},
			KeepaliveInterval: &wrappers.UInt32Value{
				Value: settings.TcpKeepaliveInterval,
			},
		},
	}

	return cluster
}

func buildListener(name string, envoyConfig *EnvoyConfig) (*envoyapiv2.Listener, error) {
	filterChain, err := buildFilterChains(envoyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build ListenerFilterChain. %s", err)
	}

	l := &envoyapiv2.Listener{
		Name: name,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: uint32(envoyConfig.ListenerPort),
					},
				},
			},
		},
		FilterChains: filterChain,
	}

	l.DrainType = envoyapiv2.Listener_DEFAULT
	if envoyConfig.ListenerDrainType == "modified" {
		l.DrainType = envoyapiv2.Listener_MODIFY_ONLY
	}

	l.Transparent = &wrappers.BoolValue{
		Value: envoyConfig.ListenTransparent,
	}

	if envoyConfig.ListenTcpFastOpenQueueLength >= 0 {
		l.TcpFastOpenQueueLength = &wrappers.UInt32Value{
			Value: uint32(envoyConfig.ListenTcpFastOpenQueueLength),
		}
	}

	l.PerConnectionBufferLimitBytes = &wrappers.UInt32Value{
		Value: uint32(envoyConfig.ListenerPerConnBufLimitBytes),
	}

	return l, nil
}

func buildTransportSocket(tls catalog.TLSInfo) (*core.TransportSocket, error) {
	upstreamTls := &auth.UpstreamTlsContext{
		AllowRenegotiation: true,
		CommonTlsContext: &auth.CommonTlsContext{
			TlsCertificates: []*auth.TlsCertificate{
				buildTlsCertChain(tls),
			},
		},
	}

	tlsAny, err := ptypes.MarshalAny(upstreamTls)
	if err != nil {
		return nil, err
	}

	return &core.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &core.TransportSocket_TypedConfig{
			TypedConfig: tlsAny,
		},
	}, nil
}

func buildFilterChains(envoyConfig *EnvoyConfig) ([]*listener.FilterChain, error) {
	serviceTarget := &core.GrpcService{
		TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
			EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
				ClusterName: XdsClusterName,
			},
		},
	}

	rdsSource := &core.ConfigSource{
		ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
			ApiConfigSource: &core.ApiConfigSource{
				ApiType:                   core.ApiConfigSource_GRPC,
				GrpcServices:              []*core.GrpcService{serviceTarget},
				SetNodeOnFirstMessageOnly: true,
			},
		},
	}

	// See: https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log#config-access-log-default-format
	accessLogger := &accesslogconfig.FileAccessLog{
		Path: envoyConfig.HttpAccessLogPath,
		AccessLogFormat: &accesslogconfig.FileAccessLog_JsonFormat{
			JsonFormat: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"start_time":            pbStringValue(`%START_TIME%`),
					"method":                pbStringValue(`%REQ(:METHOD)%`),
					"path":                  pbStringValue(`%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%`),
					"protocol":              pbStringValue(`%PROTOCOL%`),
					"response_code":         pbStringValue(`%RESPONSE_CODE%`),
					"response_code_details": pbStringValue(`%RESPONSE_CODE_DETAILS%`),
					"time_to_first_byte":    pbStringValue(`%RESPONSE_DURATION%`),
					"upstream_cluster":      pbStringValue(`%UPSTREAM_CLUSTER%`),
					"response_flags":        pbStringValue(`%RESPONSE_FLAGS%`),
					"bytes_received":        pbStringValue(`%BYTES_RECEIVED%`),
					"bytes_sent":            pbStringValue(`%BYTES_SENT%`),
					"duration":              pbStringValue(`%DURATION%`),
					"upstream_service_time": pbStringValue(`%RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)%`),
					"x_forwarded_for":       pbStringValue(`%REQ(X-FORWARDED-FOR)%`),
					"user_agent":            pbStringValue(`%REQ(USER-AGENT)%`),
					"request_id":            pbStringValue(`%REQ(X-REQUEST-ID)%`),
					"authority":             pbStringValue(`%REQ(:AUTHORITY)%`),
					"upstream_host":         pbStringValue(`%UPSTREAM_HOST%`),
				},
			},
		},
	}

	alPbStr, err := ptypes.MarshalAny(accessLogger)
	if err != nil {
		return nil, err
	}

	manager := &hcm.HttpConnectionManager{
		ServerName:                "LadyLuck",
		CodecType:                 hcm.HttpConnectionManager_AUTO,
		StatPrefix:                "http",
		IdleTimeout:               &duration.Duration{Seconds: envoyConfig.HttpIdleTimeout},
		StreamIdleTimeout:         &duration.Duration{Seconds: envoyConfig.HttpStreamIdleTimeout},
		RequestTimeout:            &duration.Duration{Seconds: envoyConfig.HttpRequestTimeout},
		DrainTimeout:              &duration.Duration{Seconds: envoyConfig.HttpDrainTimeout},
		DelayedCloseTimeout:       &duration.Duration{Seconds: envoyConfig.HttpDelayedCloseTimeout},
		PreserveExternalRequestId: envoyConfig.HttpPreserveExtReqId,
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				RouteConfigName: "upstream",
				ConfigSource:    rdsSource,
			},
		},
		HttpFilters: []*hcm.HttpFilter{
			{
				Name: wellknown.Router,
			},
		},
		AccessLog: []*accesslogfilter.AccessLog{
			{
				Name: wellknown.FileAccessLog,
				ConfigType: &accesslogfilter.AccessLog_TypedConfig{
					TypedConfig: alPbStr,
				},
			},
		},
	}

	if envoyConfig.EnableTracing {
		var opName hcm.HttpConnectionManager_Tracing_OperationName
		if envoyConfig.TracingOpName == "ingress" {
			opName = hcm.HttpConnectionManager_Tracing_INGRESS
		} else {
			opName = hcm.HttpConnectionManager_Tracing_EGRESS
		}

		manager.Tracing = &hcm.HttpConnectionManager_Tracing{
			OperationName: opName,
			Verbose:       envoyConfig.TracingVerbose,
		}
	}

	mgrPbStr, err := ptypes.MarshalAny(manager)
	if err != nil {
		return nil, err
	}

	// TODO: probably a better idea to serve each cluster
	//   as a separate filterchain?
	return []*listener.FilterChain{
		{
			Filters: []*listener.Filter{
				{
					Name: wellknown.HTTPConnectionManager,
					ConfigType: &listener.Filter_TypedConfig{
						TypedConfig: mgrPbStr,
					},
				},
			},
		},
	}, nil
}

func buildEndpoints(endpoints []catalog.Endpoint) []*endpoint.LocalityLbEndpoints {
	var results []*endpoint.LocalityLbEndpoints
	for _, e := range endpoints {
		epdef := &endpoint.LocalityLbEndpoints{
			// TODO: configure LoadBalancingWeight from consul catalog
			//    service weight.
			LbEndpoints: []*endpoint.LbEndpoint{
				{
					HostIdentifier: &endpoint.LbEndpoint_Endpoint{
						Endpoint: &endpoint.Endpoint{
							Address: &core.Address{
								Address: &core.Address_SocketAddress{
									SocketAddress: &core.SocketAddress{
										Protocol: core.SocketAddress_TCP,
										Address:  e.Addr(),
										PortSpecifier: &core.SocketAddress_PortValue{
											PortValue: uint32(e.Port()),
										},
									},
								},
							},
						},
					},
				},
			},
		}

		results = append(results, epdef)
	}
	return results
}

func buildTlsCertChain(tls catalog.TLSInfo) *auth.TlsCertificate {
	return &auth.TlsCertificate{
		CertificateChain: &core.DataSource{
			Specifier: &core.DataSource_InlineString{
				InlineString: tls.Cert(),
			},
		},
		PrivateKey: &core.DataSource{
			Specifier: &core.DataSource_InlineString{
				InlineString: tls.PKey(),
			},
		},
	}
}

func buildRouteMatchSpec(r string) *route.RouteMatch {
	matcher := &route.RouteMatch{
		CaseSensitive: &wrappers.BoolValue{
			Value: false,
		},
	}

	if strings.HasSuffix(r, "/") || strings.HasSuffix(r, "*") {
		matcher.PathSpecifier = &route.RouteMatch_Prefix{
			Prefix: r,
		}
	} else {
		matcher.PathSpecifier = &route.RouteMatch_Path{
			Path: r,
		}
	}

	return matcher
}

func buildClusterRoutingAction(config *vhostInfo) *route.Route_Route {
	return &route.Route_Route{
		Route: &route.RouteAction{
			ClusterNotFoundResponseCode: route.RouteAction_SERVICE_UNAVAILABLE,
			ClusterSpecifier: &route.RouteAction_Cluster{
				Cluster: config.clusterName,
			},
		},
	}
}

type vhostInfo struct {
	clusterName  string
	endpointName string
	paths        []string
}

type vhost struct {
	routes   map[string][]vhostInfo
	settings *catalog.ClusterSettings
}

type vhostPool struct {
	vhs []vhost
}

func (v *vhostPool) add(c catalog.ClusterInfo) {
	settings, err := c.Settings()
	if err != nil {
		logger.WithError(err).WithField("cluster", c.Name()).
			Error("failed to load cluster settings. using default values for virtualhost")
		settings = &catalog.ClusterSettings{}
		settings.Canonicalize()
	}

	vhs := vhost{
		routes:   map[string][]vhostInfo{},
		settings: settings,
	}

	for _, e := range c.Endpoints() {
		for domain, paths := range e.RoutingInfo() {
			vhs.routes[domain] = append(vhs.routes[domain], vhostInfo{
				clusterName:  c.Name(),
				endpointName: e.Name(),
				paths:        paths,
			})
		}
	}

	v.vhs = append(v.vhs, vhs)
}

func (v *vhostPool) collect(proxyPort int) []*route.VirtualHost {
	var virtualHosts []*route.VirtualHost
	for _, vh := range v.vhs {
		for domain, routes := range vh.routes {
			target := &route.VirtualHost{
				Name:                       "vh-" + domain,
				// TODO: Envoy fails to match the domain if the client
				//   sends the Host header with port in it. for the time
				//   we match on the bare domain as well as on the domain
				//   with the port number on it, but this should be removed
				//   when the behaviour is improved in Envoy.
				// See: https://github.com/envoyproxy/envoy/issues/886
				Domains:                    []string{domain, fmt.Sprintf("%s:%d", domain, proxyPort)},
				IncludeRequestAttemptCount: true,
				Routes:                     []*route.Route{},
			}

			if vh.settings.RetryOn != "" {
				target.RetryPolicy = &route.RetryPolicy{
					RetryOn:                       vh.settings.RetryOn,
					HostSelectionRetryMaxAttempts: 3,
					NumRetries:                    &wrappers.UInt32Value{Value: vh.settings.RetryAttempts},
					PerTryTimeout:                 &duration.Duration{Seconds: vh.settings.RetryAttemptTimeout},
					RetryHostPredicate: []*route.RetryPolicy_RetryHostPredicate{
						{Name: "envoy.retry_host_predicates.previous_hosts"},
					},
					RetryBackOff: &route.RetryPolicy_RetryBackOff{
						BaseInterval: &duration.Duration{Seconds: vh.settings.RetryBackoffBase},
						MaxInterval:  &duration.Duration{Seconds: vh.settings.RetryBackoffMax},
					},
				}
			}

			for _, vhinfo := range routes {
				target.Routes = append(target.Routes, buildVirtualHostRoutes(&vhinfo)...)
			}

			virtualHosts = append(virtualHosts, target)
		}
	}

	return virtualHosts
}

func buildVirtualHostRoutes(config *vhostInfo) []*route.Route {
	var routes []*route.Route
	for idx, pat := range config.paths {
		routes = append(routes, &route.Route{
			Name:   fmt.Sprintf("%s.%s-%d", config.clusterName, config.endpointName, idx),
			Match:  buildRouteMatchSpec(pat),
			Action: buildClusterRoutingAction(config),
		})
	}
	return routes
}
