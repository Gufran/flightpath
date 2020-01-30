# Route Discovery

In order to route the traffic to internal services you need to add a tag and some metadata to the service definition before
Flightpath can identify it as a valid target.

 1. Tag the service with `in-flightpath`
 1. Add metadata `flightpath-route-main = the-domain.tld/path`
 
and you should be able to access the service on `http://thedomain.tld/path`
 
!!! info
    Flightpath can route traffic to any service. It does not matter if the service has connect sidecar or not, if a service
    is registered in consul catalog with proper tag and metadata Flightpath can route traffic to it.

Flightpath watches consul catalog for services tagged with `in-flightpath` tag. A service with this tag means it is expected
to receive traffic from edge.  

After discovering the services Flightpath looks for service metadata. Any metadata attribute that starts with `flightpath-route-`
is used for routing configuration.

!!! tip
    Flightpath only cares about the `flightpath-route` prefix on metadata key. You can configure as many routes as you want
    simply by adding more meta attributes with this prefix.
    
    For example one service can have `flightpath-route-main`, `flightpath-route-path-prefix`, `flightpath-route-only-auth`
    to configure three different routes with different domains and paths.

The value of a matching attribute must adhere to one of the following forms:

`domain.tld`

:   If the value is a fully qualified domain name then all the requests for this domain are routed to the service

`/path`

:   If the value starts with a forward slash `/` then it is assumed to be a path based match. In this case the
    domain is set to wildcard, e.g. every request to `/path` will be routed to the service regardless of the domain.

`domain.tld/path`

:   Only the requests to specified path on specified domain are routed to the service.

`domain.tld/path-prefix/`, `/path-prefix/` or `/path-prefix/*`

:   If the value has `/` or `*` as suffix it is assumed to be a prefix based match. In this case every request that
    matches the path prefix will be routed to the service, e.g. `domain.tld/path-prefix/one/two` or `*/path-prefix/one/two/three`
    if the domain is omitted.


