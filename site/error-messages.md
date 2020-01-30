# Error Messages

Following is a list of all errors that can be raised by Flightpath.

!!! warning
    This list should only be used as a reference document. The goal of this list is not
    to provide accurate debugging information but merely to point you towards a potential
    cause of the error.
    
    You should always check logs for `error` attribute which contains the original error
    raised by the underlying component.

### Global Scope

==failed to initialize metrics subsystem==

:    Flightpath failed to initialize the metrics backend. It is possible that the metrics
     sink is not listening on specified port or Flightpath does not have sufficient permissions
     on socket file.

### Catalog Subsystem

==failed to fetch list of services from consul catalog==

:    Flightpath failed to list all services currently presently registered in consul
     catalog. The most likely reason is a network problem or a disbanded consul quorum.  
     If consul ACLs are in effect then the permissions on Flightpath token should also be reviewed.
     
==failed to filter connect target services==

:    Flightpath failed to tell sidecar proxies from main services.
     The most likely reason is a network problem or a disbanded consul quorum.
     If consul ACLs are in effect then the permissions on Flightpath token should also be reviewed.

==failed to fetch service definition==

:    Flightpath failed to read the service definition from consul catalog.
     The most likely reason is a network problem or a disbanded consul quorum.
     If consul ACLs are in effect then the permissions on Flightpath token should also be reviewed.
     
==failed to fetch TLS leaf certificates==

:    Flightpath failed to read the service definition from consul catalog.
     The most likely reason is a network problem or a disbanded consul quorum.
     If consul ACLs are in effect then the permissions on Flightpath token should also be reviewed.

### Discovery Subsystem

==GRPC server failed==

!!! bug
    The xDS gRPC server has shut down with an error.
    Flightpath may not have sufficient permissions to bind to the port configured for xDS.  
     
    [Click here to report the bug](https://github.com/Gufran/flightpath/issues/new?title=GRPC+server+failed)
     
==failed to deregister the service from consul catalog==

:    Flightpath failed to deregister itself from consul catalog.
     The most likely reason is a network problem or a disbanded consul quorum.
     If consul ACLs are in effect then the permissions on Flightpath token should also be reviewed.
     
==failed to stop the socket listener==

!!! bug
    The network listener for xDS has failed to stop.
    The attempt to close the listener may have been made before the listener could start
    listening on the network port or it may have already been closed.  
    
    [Click here to report the bug](https://github.com/Gufran/flightpath/issues/new?title=failed+to+stop+the+socket+listener)
     
==failed to update cluster information==

!!! bug
    Flightpath failed to build Envoy configuration or the configuration was built incorrectly.
    
    [Click here to report the bug](https://github.com/Gufran/flightpath/issues/new?title=failed+to+update+cluster+information)
    
==debug http server crashed==

:    Flightpath may not have sufficient permissions to bing to the port configured for the debug server

==failed to retrieve snapshot from XDS cache==

!!! bug
    Debug server failed to retrieve the active configuration from XDS server
    
    [Click here to report the bug](https://github.com/Gufran/flightpath/issues/new?title=failed+to+retrieve+snapshot+from+XDS+cache)
    
### Metrics Subsystem

==failed to report Gauge metrics==

:    The metrics subsystem failed to push metrics to the backend. Most likely reason is an unresponsive
     agent or network problem.

==failed to report Increment metrics==

:    The metrics subsystem failed to push metrics to the backend. Most likely reason is an unresponsive
     agent or network problem.

