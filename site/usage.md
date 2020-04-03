# Usage

Following command line flags can be used to configure flightpath

==`-consul.host`==

:    Default `"127.0.0.1"`

     Network address to a consul agent

==`-consul.port`==

:    Default `"8500"`

     Port on which the consul agent is listening

==`-consul.proto`==

:    Default `"http"`

     Protocol used to connect with consul agent

==`-consul.token`==

:    Default `""`

     Consul token to use

==`-debug`==

:    Default `"false"`

     Start debug HTTP server on loopback interface

==`-debug.port`==

:    Default `"7180"`

     Network port to use for debug HTTP server

==`-dogstatsd.addr`==

:    Default `"127.0.0.1"`

     Address of the dogstatsd agent

==`-dogstatsd.namespace`==

:    Default `"flightpath"`

     Metrics namespace for dogstatsd

==`-dogstatsd.port`==

:    Default `"8125"`

     Port of the dogstatsd agent

==`-envoy.http.access-logs`==

:    Default `"/var/log/envoy/access.log"`

     Path to the file where envoy will write listener access logs

==`-envoy.http.delayed-close-timeout`==

:    Default `"1"`

     Number of seconds to wait for closing the connection after peer closes from their side

==`-envoy.http.drain-timeout`==

:    Default `"30"`

     Number of seconds to wait for HTTP/2 to shut down after sending GOAWAY frame

==`-envoy.http.idle-timeout`==

:    Default `"15"`

     Number of seconds after which an idle connection is cleaned up

==`-envoy.http.preserve-req-id`==

:    Default `"true"`

     Preserve external request ID if set in headers

==`-envoy.http.req-timeout`==

:    Default `"30"`

     Number of seconds to wait for the entire request to be received

==`-envoy.http.stream-idle-timeout`==

:    Default `"300"`

     Number of seconds after which an idle TCP connection is cleaned up

==`-envoy.listen.drain-type`==

:    Default `"default"`

     Method used to drain upstream connections. Valid options are 'default' and 'modified'

==`-envoy.listen.per-conn-buf-limit`==

:    Default `"1049000"`

     Soft limit in bytes on size of the listener’s new connection read and write buffers

==`-envoy.listen.port`==

:    Default `"9292"`

     Port used by Envoy Listener

==`-envoy.listen.tcp-fast-open-q-length`==

:    Default `"-1"`

     TFO queue length. -1 means the setting is not modified, 0 means TFO is disabled and 1 and higher value means TFO is enabled with queue size set to this value

==`-envoy.listen.transparent`==

:    Default `"true"`

     Set the listener as transparent socket

==`-envoy.tracing.enabled`==

:    Default `"false"`

     Enable request tracing on envoy

==`-envoy.tracing.op-name`==

:    Default `"egress"`

     Tracing operation name, valid values are 'ingress' or 'egress'

==`-envoy.tracing.verbose`==

:    Default `"false"`

     Add verbose information to traces

==`-log.format`==

:    Default `"json"`

     Format of the log message. Valid options are json and plain

==`-log.level`==

:    Default `"INFO"`

     Set log verbosity. Valid options are trace, debug, error, warn, info, fatal and panic

==`-metrics.runtime`==

:    Default `"true"`

     Expose runtime stats on memory and CPU

==`-metrics.sink`==

:    Default `""`

     Set the metrics sink. Valid options are 'dogstatsd' and 'stderr'

==`-name`==

:    Default `"flightpath"`

     Name used to register the flightpath service in Consul Catalog

==`-node-name`==

:    Default `"flightpath-edge"`

     Named of the Envoy node

==`-port`==

:    Default `"7171"`

     Port for XDS listener

==`-version`==

:    Default `"false"`

     Show version information

