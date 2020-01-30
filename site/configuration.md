# Configuration

Flightpath has several configuration options you might want to tweak to get it working in your specific environment.

| Option |  Description |
|:--------|:------------|
| `-consul.host` |  Network address to a consul agent |
| `-consul.port` |  Port on which the consul agent is listening |
| `-consul.proto` |  Protocol used to connect with consul agent |
| `-consul.token` |  Consul token to use |
| `-debug` |  Start debug HTTP server on loopback interface |
| `-debug.port` |  Network port to use for debug HTTP server |
| `-dogstatsd` |  Enable publishing metrics to dogstatsd agent |
| `-dogstatsd.addr` |  Address of the dogstatsd agent |
| `-dogstatsd.namespace` |  Metrics namespace for dogstatsd |
| `-dogstatsd.port` |  Port of the dogstatsd agent |
| `-node-name` |  Named of the Envoy node |
| `-envoy.access-logs` |  Path to the file where envoy will write listener access logs |
| `-envoy.listen.port` |  Port used by Envoy Listener |
| `-log.format` |  Format of the log message. Valid options are json and plain |
| `-log.level` |  Set log verbosity. Valid options are trace, debug, error, warn, info, fatal and panic |
| `-name` |  Name used to register the flightpath service in Consul Catalog |
| `-port` |  Port for XDS listener |
| `-version` |  Show version information |


Flightpath primarily provides configuration to Envoy and the Envoy XDS protocol requires some information to be shared
between Envoy and the XDS server. 

We will use the following envoy configuration for reference:

```yaml
node:
  id: flightpath-edge
  cluster: flightpath

static_resources:
  clusters:
    - name: xds_cluster
      connect_timeout: 0.25s
      type: STATIC
      lb_policy: ROUND_ROBIN
      http2_protocol_options: {}
      upstream_connection_options:
        tcp_keepalive: {}
      load_assignment:
        cluster_name: xds_cluster
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: 127.0.0.1
                      port_value: 7171

dynamic_resources:
  lds_config:
    api_config_source:
      api_type: GRPC
      grpc_services:
        envoy_grpc:
          cluster_name: xds_cluster

  cds_config:
    api_config_source:
      api_type: GRPC
      grpc_services:
        envoy_grpc:
          cluster_name: xds_cluster

```

You can copy this configuration, or if you want you can also download the [internal/envoy-config.yaml][] file from github.

## Understanding Envoy Config

If you are not familiar with Envoy configuration then you should head over to the [official envoy documentation][] and at least get
an understanding of the basic building blocks.

This configuration declares that the envoy process is running as node `flightpath-edge` in cluster `flightpath`.
This information is presented to Flightpath when envoy tries to gather the configuration [^1].

There is also a static cluster registered with name `xds_cluster` and address `127.0.0.1:7171`, this is the network address
where envoy will try to reach the XDS server.

Rest of the configuration declares that the network listener and the upstream clusters can be discovered from `xds_cluster`.

This is just enough configuration for Envoy to be able to make connection with Flightpath and ask for more configuration

If you start Envoy with this configuration it will attempt to connect with an XDS server on `127.0.0.1:7171`. At this point
Envoy has not started a network listener so it is unable to accept any request. 

Now start Flightpath and you should see in Envoy logs that it has connected with the XDS server and loaded listener configuration.

## Understanding Flightpath Config

At this point if you have both the Flightpath and Envoy running you have a reasonably functional setup that can discover services
from consul catalog and route network traffic to them.  
This is because Envoy configuration we used in previous section uses the default values chosen by Flightpath.  

Default configuration is helpful to get up and running quickly but in a production like environment you might want to
change things a little and that is where it becomes necessary to understand how a configuration option in Flightpath maps
to that of Envoy.

The most critical configuration options available in Flightpath are


| Option |  Description |
|:--------|:------------|
| `-node-name` |  Named of the Envoy node |
| `-envoy.listen.port` |  Port used by Envoy Listener |
| `-envoy.access-logs` |  Path to the file where envoy will write listener access logs |


`-node-name`

:    This option directly maps to the node name set in Envoy configuration. Envoy presents this value to the XDS server
     and the server returns the configuration specifically tailored for the node with this name. How you choose this name
     depends very much on your Flightpath deployment.
     
     If you are running Flightpath on the same machine as Envoy then you can use the hostname of the machine to represent
     the node name.
     
     If you have Flightpath deployed as a seperate cluster from Envoy then you might want to choose a name that best
     describes the role of Envoy cluster, e.g. public-us-east-1, public-eu-central-1 etc.
     In this case you want to make sure that all Envoy processes that connect to the same Flightpath cluster are
     configured to use the same node name or they won't receive any configuration and therefore won't be able to serve
     traffic.

`-envoy.listen.port`

:    This is the network port Envoy is expected to use for its listener. This is configured in Flightpath because Envoy
     relies on Flightpath for listener discovery as well.
     
     You can add other static listeners as part of the Envoy configuration but consul based service discovery and routing
     will only work on this listener.


`-envoy.access-logs`

:    This is the absolute path to the location where you want Envoy to write access logs. This is set in Flightpath
     because this is configured on the listener and Flightpath configures the listener on Envoy.

     You can set this to `/dev/stdout` or `/dev/stderr` if you want the logs to go to standard devices but keep in mind
     that if you run Envoy as a systemd service you won't be able to stream to stdout or stderr.  


Apart from these configuration options there are things that Flightpath chooses to set on Envoy with no way to override them.
This is only a problem in short term while things are being changed and shuffled around. A later version of Flightpath
will provide methods to configure every aspect of Envoy.



[^1]: See [Envoy Node Configuration](https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/core/base.proto#envoy-api-msg-core-node) for details

[internal/envoy-config.yaml]: https://github.com/Gufran/flightpath/blob/master/internal/envoy-config.yaml
[official envoy documentation]: https://www.envoyproxy.io/docs/envoy/latest/configuration/configuration
