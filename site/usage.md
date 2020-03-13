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

:    Default `"true"`

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

==`-envoy.access-logs`==

:    Default `"/var/log/envoy/access.log"`

     Path to the file where envoy will write listener access logs

==`-envoy.listen.port`==

:    Default `"9292"`

     Port used by Envoy Listener

==`-log.format`==

:    Default `"json"`

     Format of the log message. Valid options are json and plain

==`-log.level`==

:    Default `"INFO"`

     Set log verbosity. Valid options are trace, debug, error, warn, info, fatal and panic

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

