# Flightpath

![Test](https://github.com/Gufran/flightpath/workflows/Test/badge.svg?branch=master)

Flightpath is an Envoy Control Plane that integrates with Consul Connect and provides L7 routing
at edge.

At its core Flightpath is a gRPC server that implements [Envoy XDS v2 API interface][].
It registers itself in Consul Catalog as a [Connect Native service][] and shares its certificate with Envoy. Envoy uses
the certificates to connect with other services as if Envoy itself was registered as a Connect enabled service.

Flightpath watches Consul Catalog for services with tag `in-flightpath`. Services with matching tag are used to populate
the CDS and EDS gRPC streams.  
For routing information Flightpath relies on `route` meta attributes on the service. `route` attribute is used to specify
the protocol, domain, and the URL to match before the traffic can be routed to the service instance.



[Envoy XDS v2 API interface]: https://www.envoyproxy.io/docs/envoy/latest/api-v2/api
[Connect Native service]: https://www.consul.io/docs/connect/native.html
