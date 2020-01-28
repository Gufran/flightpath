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
	envoytype "github.com/envoyproxy/go-control-plane/envoy/type"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
	duration "github.com/golang/protobuf/ptypes/duration"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/golang/protobuf/ptypes/wrappers"
	consul "github.com/hashicorp/consul/api"
	"strings"
	"time"
)

const XdsClusterName = "xds_cluster"

type XDS struct {
	ServiceName       string
	ProxyNodeName     string
	Consul            *consul.Client
	Cache             cache.SnapshotCache
	ProxyListenerPort int
	WithDebugServer   bool
	DebugServerPort   int
}

func (x *XDS) Start(ctx context.Context) {
	source := catalog.NewCatalog(ctx, x.Consul)

	clusterChan := make(chan catalog.ClusterInfo)
	tlsChan := make(chan catalog.TLSInfo)
	cleanupChan := make(chan string)

	go source.DiscoverClusters(clusterChan, cleanupChan)
	go source.WatchTLS(x.ServiceName, tlsChan)

	if x.WithDebugServer {
		StartDebugServer(x.DebugServerPort, x.ProxyNodeName, x.Cache)
	}

	go synchronize(ctx, x.Cache, x.ProxyNodeName, x.ProxyListenerPort, clusterChan, tlsChan, cleanupChan)
}

func synchronize(ctx context.Context,
	snc cache.SnapshotCache,
	node string,
	proxyPort int,
	clusters <-chan catalog.ClusterInfo,
	tls <-chan catalog.TLSInfo,
	cleanup <-chan string) {

	// TLS info is absolutely necessary and since we know that we
	// are registered as a connect enabled service and guaranteed
	// to receive a certificate pair, we'll just wait for it to
	// arrive before setting up the shop.
	certs := <-tls

	tick := time.Tick(5 * time.Second)
	knownClusters := map[string]catalog.ClusterInfo{}

	for {
		metrics.Incr("discovery.sync.loop", nil)
		select {
		case <-ctx.Done():
			return

		case cluster := <-clusters:
			logger.WithField("cluster", cluster.Name()).Info("updating cluster entry")
			metrics.Incr("discovery.cluster.update", []string{"cluster:" + cluster.Name()})
			knownClusters[cluster.Name()] = cluster

		case certs = <-tls:
			metrics.Incr("discovery.tls.update", nil)
		// keep'em up to date

		case name := <-cleanup:
			metrics.Incr("discovery.cluster.cleanup", []string{"cluster:" + name})
			logger.WithField("cluster", name).Info("removing cluster from tracked list")
			if _, ok := knownClusters[name]; ok {
				delete(knownClusters, name)
			}

		case <-tick:
			metrics.Incr("discovery.cluster.flush", nil)
			metrics.GaugeI("discovery.cluster.batch_size", len(knownClusters), nil)

			err := putCache(snc, node, proxyPort, clustersList(knownClusters), certs)
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

func putCache(snc cache.SnapshotCache, node string, proxyPort int, clusters []catalog.ClusterInfo, tls catalog.TLSInfo) error {
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
		mappings: map[string][]vhostInfo{},
	}

	envoyListener, err := buildListener("flightpath", proxyPort)
	if err != nil {
		return fmt.Errorf("failed to build cluster definition. %s", err)
	}

	listenerResource = []cache.Resource{
		envoyListener,
	}

	for _, service := range clusters {
		clusterConfig := buildCluster(service.Name())

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
			VirtualHosts: vhosts.collect(proxyPort),
		},
	}

	metrics.GaugeI("discovery.cache.put.clusters", len(clusterResource), nil)
	metrics.GaugeI("discovery.cache.put.endpoints", len(endpointResource), nil)
	metrics.GaugeI("discovery.cache.put.routes", len(routeResource), nil)
	metrics.GaugeI("discovery.cache.put.listener", len(listenerResource), nil)

	snap := cache.NewSnapshot(catalog.Hash(clusters), endpointResource, clusterResource, routeResource, listenerResource)
	return snc.SetSnapshot(node, snap)
}

func buildCluster(serviceName string) *envoyapiv2.Cluster {
	return &envoyapiv2.Cluster{
		Name:                          serviceName,
		LbPolicy:                      envoyapiv2.Cluster_ROUND_ROBIN,
		RespectDnsTtl:                 true,
		DrainConnectionsOnHostRemoval: true,
		ConnectTimeout: &duration.Duration{
			Seconds: 10,
		},
		ClusterDiscoveryType: &envoyapiv2.Cluster_Type{
			Type: envoyapiv2.Cluster_EDS,
		},
		EdsClusterConfig: &envoyapiv2.Cluster_EdsClusterConfig{
			ServiceName: serviceName,
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
}

func buildListener(name string, port int) (*envoyapiv2.Listener, error) {
	filterChain, err := buildFilterChains()
	if err != nil {
		return nil, fmt.Errorf("failed to build ListenerFilterChain. %s", err)
	}

	return &envoyapiv2.Listener{
		Name: name,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: uint32(port),
					},
				},
			},
		},
		FilterChains: filterChain,
	}, nil
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

func buildFilterChains() ([]*listener.FilterChain, error) {
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
		// TODO: make the path configurable.
		//    systemd messes up with /dev/stdout and /dev/stderr of
		//    a process and converts them into a socket that streams
		//    logs to journald. It is not possible to execute `open()`
		//    on a socket file so the access logger crashes and Envoy
		//    can't start the listener.
		//    On a machine without systemd it would be possible to use
		//    /dev/stdout or /dev/stderr for Access Logging but with
		//    systemd we have to use a regular file.
		Path: "/var/log/envoy/access.log",
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
		ServerName: "LadyLuck",
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "http",
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

func buildClusterRoutingAction(cluster string) *route.Route_Route {
	return &route.Route_Route{
		Route: &route.RouteAction{
			ClusterNotFoundResponseCode: route.RouteAction_SERVICE_UNAVAILABLE,
			ClusterSpecifier: &route.RouteAction_Cluster{
				Cluster: cluster,
			},
		},
	}
}

type vhostInfo struct {
	clusterName  string
	endpointName string
	paths        []string
}
type vhostPool struct {
	mappings map[string][]vhostInfo
}

func (v *vhostPool) add(c catalog.ClusterInfo) {
	for _, e := range c.Endpoints() {
		for domain, paths := range e.RoutingInfo() {
			v.mappings[domain] = append(v.mappings[domain], vhostInfo{
				clusterName:  c.Name(),
				endpointName: e.Name(),
				paths:        paths,
			})
		}
	}
}

func (v *vhostPool) collect(proxyPort int) []*route.VirtualHost {
	var vhs []*route.VirtualHost
	for domain, routes := range v.mappings {
		target := &route.VirtualHost{
			Name:                       "vh-" + domain,
			Domains:                    []string{domain, fmt.Sprintf("%s:%d", domain, proxyPort)},
			IncludeRequestAttemptCount: true,
			Routes:                     []*route.Route{},
			RetryPolicy: &route.RetryPolicy{
				RetryOn:                       "gateway-error",
				HostSelectionRetryMaxAttempts: 3,
				NumRetries: &wrappers.UInt32Value{
					Value: 3,
				},
				PerTryTimeout: &duration.Duration{
					Seconds: 1,
				},
				RetryHostPredicate: []*route.RetryPolicy_RetryHostPredicate{
					{
						Name: "envoy.retry_host_predicates.previous_hosts",
					},
				},
			},
		}

		for _, vhinfo := range routes {
			// TODO: Envoy fails to match the domain if the client
			//   sends the Host header with port in it. for the time
			//   we match on the bare domain as well as on the domain
			//   with the port number on it, but this should be removed
			//   when the behaviour is improved in Envoy.
			// See: https://github.com/envoyproxy/envoy/issues/886
			target.Routes = append(target.Routes, buildVirtualHostRoutes(vhinfo.clusterName, vhinfo.endpointName, vhinfo.paths)...)
		}

		vhs = append(vhs, target)
	}

	return vhs
}

func buildVirtualHostRoutes(clusterName string, endpointName string, paths []string) []*route.Route {
	var routes []*route.Route
	for idx, pat := range paths {
		routes = append(routes, &route.Route{
			Name:   fmt.Sprintf("%s.%s-%d", clusterName, endpointName, idx),
			Match:  buildRouteMatchSpec(pat),
			Action: buildClusterRoutingAction(clusterName),
			Tracing: &route.Tracing{
				OverallSampling: &envoytype.FractionalPercent{
					Numerator:   1000,
					Denominator: envoytype.FractionalPercent_TEN_THOUSAND,
				},
			},
		})
	}
	return routes
}
