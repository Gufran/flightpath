# Envoy Configuration

Flightpath can configure Envoy Upstream Cluster and Route definition using the service metadata attribute.
Following metadata attributes can be set on consul service:

## Cluster Configuration

==`flightpath-cluster-conn_timeout`==

:    Integer  
     Default: `10 seconds`  
     Used to set the `connect_timeout` attribute on [Cluster](https://www.envoyproxy.io/docs/envoy/v1.13.1/api-v2/api/v2/cluster.proto) configuration.
     

==`flightpath-cluster-per_conn_buf_limit_bytes`==

:    Integer  
     Default: `32768`  
     Used to set the `per_connection_buffer_limit_bytes` attribute on [Cluster](https://www.envoyproxy.io/docs/envoy/v1.13.1/api-v2/api/v2/cluster.proto) configuration.

==`flightpath-cluster-max_req_per_conn`==

:    Integer  
     Default: `10,000`  
     Used to set the `max_requests_per_connection` attribute on [Cluster](https://www.envoyproxy.io/docs/envoy/v1.13.1/api-v2/api/v2/cluster.proto) configuration.

==`flightpath-cluster-tcp_keepalive_probes`==

:    Integer  
     Default: `9`
     Used to configure `upstream_connection_options.tcp_keepalive.keepalive_probes` attribute on [Cluster](https://www.envoyproxy.io/docs/envoy/v1.13.1/api-v2/api/v2/cluster.proto)

==`flightpath-cluster-tcp_keepalive_time`==

:    Integer  
     Default: `300 Seconds`
     Used to configure `upstream_connection_options.tcp_keepalive.keepalive_time` attribute on [Cluster](https://www.envoyproxy.io/docs/envoy/v1.13.1/api-v2/api/v2/cluster.proto)

==`flightpath-cluster-tcp_keepalive_interval`==

:    Integer  
     Default: `90 Seconds`
     Used to configure `upstream_connection_options.tcp_keepalive.keepalive_interval` attribute on [Cluster](https://www.envoyproxy.io/docs/envoy/v1.13.1/api-v2/api/v2/cluster.proto)

==`flightpath-retry-on`==

:    String  
     Default: `""`
     Used to configure `retry_on` attribute on route's [RetryPolicy](https://www.envoyproxy.io/docs/envoy/v1.13.1/api-v3/config/route/v3/route_components.proto.html?highlight=retry_policy#envoy-v3-api-msg-config-route-v3-retrypolicy)  
     
     Valid policies are `5xx`, `gateway-error`, `reset`, `connect-failure`, `retriable-4xx`, `refused-stream`, `retriable-status-codes`, and `retriable-headers`.  
     Multiple policies can be specified by using a `,` (comma) delimited list.

==`flightpath-retry-attempts`==

:    Integer  
     Default: `3`  
     Used to configure `num_retries` attribute on route's [RetryPolicy](https://www.envoyproxy.io/docs/envoy/v1.13.1/api-v3/config/route/v3/route_components.proto.html?highlight=retry_policy#envoy-v3-api-msg-config-route-v3-retrypolicy)  
     
     This setting is ignored if `flightpath-retry-on` is not set.

==`flightpath-retry-per_try_timeout`==

:    Integer  
     Default: `5`  
     Used to configure `per_try_timeout` attribute on route's [RetryPolicy](https://www.envoyproxy.io/docs/envoy/v1.13.1/api-v3/config/route/v3/route_components.proto.html?highlight=retry_policy#envoy-v3-api-msg-config-route-v3-retrypolicy)  
     
     This setting is ignored if `flightpath-retry-on` is not set.

==`flightpath-retry-backoff_base_interval`==

:    Integer  
     Default: `1`  
     Used to configure `retry_back_off.base_interval` attribute on route's [RetryPolicy](https://www.envoyproxy.io/docs/envoy/v1.13.1/api-v3/config/route/v3/route_components.proto.html?highlight=retry_policy#envoy-v3-api-msg-config-route-v3-retrypolicy)  
     
     This setting is ignored if `flightpath-retry-on` is not set.

==`flightpath-retry-backoff_max_interval`==

:    Integer  
     Default: `6`  
     Used to configure `retry_back_off.max_interval` attribute on route's [RetryPolicy](https://www.envoyproxy.io/docs/envoy/v1.13.1/api-v3/config/route/v3/route_components.proto.html?highlight=retry_policy#envoy-v3-api-msg-config-route-v3-retrypolicy)  
     
     This setting is ignored if `flightpath-retry-on` is not set.

