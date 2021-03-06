# Observability

!!! caution
    Observability is an evolving component of Flightpath. Right now Flightpath is capable of writing logs to stdout in KV and
    JSON format and metrics are exposed only for Dogstatsd backend.
    
### Configuring Metrics

Metrics backend can be configured by starting flightpath with `-dogstatsd` command line parameter.

The backend can be configured using `-dogstatsd.*` parameters. See [Configuration][] for more details.

!!! note
    A future version of Flightpath will provide more backends for metrics as well as the ability to expose traces.

## Exposed Metrics

### Cluster Discovery Metrics

==`catalog.discovery.clusters.loop`==

:    Counter type  
     No Tags  
     
     Incremented on every iteration of cluster watcher loop.  
     A rapidly increasing value might indicate high activity in consul catalog or flapping consul quorum.


==`catalog.discovery.clusters.error.fetch`==

:    Counter type  
     No Tags  
     
     Incremented every time there is an error while attempting to fetch service definitions from consul catalog.
     
     It is recommended to raise alerts if this metric has a non-zero value.  
     Check logs from **catalog** subsystem for details on the error.
     

==`catalog.discovery.clusters.noop`==

:    Counter type  
     No Tags  
     
     Incremented every time the catalog watcher returns without updates.  
     
     
==`catalog.discovery.clusters.candidate`==

:    Gauge type  
     No Tags  
     
     Represents the number of services potentially eligible for Flightpath based routing.
     
==`catalog.discovery.clusters.error.filter_connect`==

:    Counter type  
     No Tags  
     
     Incremented if an error is encountered trying to filter out the connect target
     services that can only be reached via their sidecar.
     
     It is recommended to raise alert if this metric has a non zero value.  
     Check logs from **catalog** subsystem for more details on the error.
     
==`catalog.discovery.clusters.targets`==

:    Gauge type  
     No Tags  
     
     Represents the number of services that have been selected for Flightpath based routing
     
     
==`catalog.discovery.clusters.watchers`==

:    Gauge type
     No tags
     
     Represents the number of active consul watcher. Each target service has one corresponding watcher 
     

==`catalog.discovery.clusters.watcher.new`==

:    Counter type  
     **service:** Name of the service being watched
     
     Incremented every time a new consul watcher is started
     
==`catalog.discovery.clusters.watcher.closing`==

:    Counter type  
     **service:** Name of the service that was being watched
     
     Incremented every time a service watcher is shut down

### Service Discovery Metrics

==`catalog.discovery.service.loop`==

:    Counter type  
     **service:** Name of the service being watched  
     **is_sidecar:** Whether or not the service is a sidecar proxy  
     
     Incremented on every interation of service watcher loop.  
     A rapidly increasing value might indicate high activity in consul catalog or a flapping consul quorum.
     

==`catalog.discovery.service.error.fetch`==

:    Counter type  
     **service:** Name of the service being watched  
     **is_sidecar:** Whether or not the service is a sidecar proxy  
     
     Incremented every time there is an error while attempting to fetch service definition from consul quorum
     
     It is recommended to raise alert if this metric has non-zero value.  
     Check logs from **catalog** subsystem for details on the error.
     
     
==`catalog.discovery.service.noop`==

:    Counter type  
     **service:** Name of the service being watched  
     **is_sidecar:** Whether or not the service is a sidecar proxy  
     
     Incremented every time the service watcher returns without updates.

==`catalog.discovery.service.updated`==

:    Counter type  
     **service:** Name of the service being watched  
     **is_sidecar:** Whether or not the service is a sidecar proxy  
     
     Incremented every time the watched service is updated in catalog.

### TLS Discovery Metrics

==`catalog.tls.loop`==

:    Counter type  
     No tags
     
     Incremented on every iteration of the TLS certificate watcher loop.  
     A rapidly increasing value might indicate a flapping consul quorum.
     
==`catalog.tls.error.fetch`==

:    Counter type  
     No tags
     
     Incremented every time there is an error while attempting to fetch TLS certificates
     
     It is recommended to raise alert if this metric has a non-zero value.  
     Check logs from **catalog** subsystem for details on error.
     
==`catalog.tls.noop`==

:    Counter type  
     No tags
     
     Incremented every time the TLS watcher returns without updates.
     
     
==`catalog.tls.updated`==

:    Counter type  
     No tags
     
     Incremented every time the TLS certificate is updated.

### XDS Server Metrics

==`discovery.sync.loop`==

:    Counter type  
     No tags  
     
     Incremented on every iteration of configuration synchronization loop
     
     
==`discovery.cluster.update`==

:    Counter type  
     **cluster:** Name of the cluster with updates
     
     Incremented every time there is an update available in a cluster
     
==`discovery.tls.update`==

:    Counter type  
     No tags
     
     Incremented every time there is an updated available for leaf certificates
     
==`discovery.cluster.cleanup`==

:    Counter type
     No tags
     
     Incremented every time a cluster is removed from configuration
     
==`discovery.cluster.flush`==

:    Counter type  
     No tags
     
     Incremented on periodic configuration flush
     
==`discovery.cluster.batch_size`==

:    Gauge type  
     No tags
     
     Represents the number of clusters pushed down to the XDS server in an update
     
==`discovery.cluster.error.flush`==

:    Counter type
     No tags
     
     Incremented every time an error is encountered trying to push updates to the XDS server
     
==`discovery.cache.put_ns`==

:    Gauge type  
     No tags
     
     Number of nanoseconds taken to push the updated to XDS server
     
==`discovery.cache.put.clusters`==

:    Gauge type  
     No tags
     
     Number of cluster entries pushed to XDS server


==`discovery.cache.put.endpoints`==

:    Gauge type  
     No tags
     
     Number of endpoint entries pushed to XDS server


==`discovery.cache.put.routes`==

:    Gauge type  
     No tags
     
     Number of route entries pushed to XDS server

==`discovery.cache.put.listener`==

:    Gauge type  
     No tags
     
     Number of listener entries pushed to XDS server


### Runtime Metrics


==`runtime.goroutines`==

:    Gauge type
     No tags

     Number of active goroutines


==`runtime.mem.alloc`==

:    Gauge type
     No tags

     Bytes of allocated heap objects


==`runtime.mem.total`==

:    Gauge type
     No tags

     Cumulative bytes allocated for heap objects. It only increases when new heap objects are allocated but does not decrease when objects are freed


==`runtime.mem.sys`==

:    Gauge type
     No tags

     Total bytes of memory obtained from the OS


==`runtime.mem.lookups`==

:    Gauge type
     No tags

     Number of pointer lookups performed by the runtime


==`runtime.mem.malloc`==

:    Gauge type
     No tags

     Cumulative count of heap objects allocated


==`runtime.mem.frees`==

:    Gauge type
     No tags

     Cumulative count of heap objects freed


==`runtime.mem.heap.alloc`==

:    Gauge type
     No tags

     Bytes of allocated heap objects including the unreachable objects that haven't been garbage collected


==`runtime.mem.heap.sys`==

:    Gauge type
     No tags

     Bytes of heap memory obtained from the OS


==`runtime.mem.heap.idle`==

:    Gauge type
     No tags

     Bytes in unused memory spans


==`runtime.mem.heap.inuse`==

:    Gauge type
     No tags

     Bytes in in-use spans


==`runtime.mem.heap.released`==

:    Gauge type
     No tags

     Bytes of physical memory returned to the OS


==`runtime.mem.heap.objects`==

:    Gauge type
     No tags

     Number of allocated heap objects


==`runtime.mem.stack.inuse`==

:    Gauge type
     No tags

     Bytes in stack span


==`runtime.mem.stack.sys`==

:    Gauge type
     No tags

     Bytes of stack memory obtained from the OS


==`runtime.mem.stack.mcache_inuse`==

:    Gauge type
     No tags

     Bytes of allocated mcache structures


==`runtime.mem.stack.mcache_sys`==

:    Gauge type
     No tags

     Bytes of memory obtained from the OS for mcache structures


==`runtime.mem.othersys`==

:    Gauge type
     No tags

     Bytes of memory in off-heap runtime allocations


==`runtime.gc.sys`==

:    Gauge type
     No tags

     Bytes of memory in garbage collection metadata


==`runtime.gc.next`==

:    Gauge type
     No tags

     Target heap size of next GC cycle


==`runtime.gc.last`==

:    Gauge type
     No tags

     Time when last garbage collection finished, represented as nanoseconds since UNIX epoch


==`runtime.gc.pause_total_ns`==

:    Gauge type
     No tags

     Cumulative nanoseconds in GC stop-the-workd pauses since the program started


==`runtime.gc.pause`==

:    Gauge type
     No tags

     Time in nanoseconds of recent GC stop-the-world pause


==`runtime.gc.count`==

:    Gauge type
     No tags

     Number of completed GC cycles



[Configuration]: ./configuration.md