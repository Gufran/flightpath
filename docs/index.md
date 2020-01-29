# Introduction

Flightpath is an [xDS server][] that can configure Envoy to act as an Edge proxy for [Consul Connect][] enabled services.

Flightpath registers itself in Consul Catalog as a Connect Native service and uses its own TLS certificate to configure
routing in Envoy. Using these certificates Envoy can communicate with Connect Sidecars as if the connection was
initiated by Flightpath.

Flightpath can be compared with [Fabio][] or [Traefik][] but it is not a proxy in itself. While both Fabio and Traefik manage
cluster discovery and traffic routing themselves, Flightpath is only responsible for discovering the routing information
from consul catalog and configuring Envoy to route the traffic.

  Configuration
  Observability
    Logs
    Metrics
      Datadog
      Statsd


[xDS server]: https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol 
[Consul Connect]: https://www.consul.io/docs/connect/index.html
[Traefik]: https://docs.traefik.io/
[Fabio]: https://fabiolb.net/