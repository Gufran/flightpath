package metrics

import (
	"context"
	"github.com/Gufran/flightpath/log"
	"runtime"
	"time"
)

var (
	client                = NewNoOpSink()
	logger                = log.New("metrics")
	runtimeMetricsEnabled = false
)

// SetupSink initializes the metrics subsystem.
func SetSink(sink Sink) {
	client = sink
}

func EnableRuntimeMetrics(ctx context.Context) {
	if runtimeMetricsEnabled {
		return
	}

	runtimeMetricsEnabled = true
	ticker := time.NewTicker(30 * time.Second)

	mstats := new(runtime.MemStats)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runtime.ReadMemStats(mstats)

			// CPU Stats
			GaugeI("runtime.goroutines", runtime.NumGoroutine(), nil)

			// Memory Stats
			Gauge("runtime.mem.alloc", float64(mstats.Alloc), nil)
			Gauge("runtime.mem.total", float64(mstats.TotalAlloc), nil)
			Gauge("runtime.mem.sys", float64(mstats.Sys), nil)
			Gauge("runtime.mem.lookups", float64(mstats.Lookups), nil)
			Gauge("runtime.mem.malloc", float64(mstats.Mallocs), nil)
			Gauge("runtime.mem.frees", float64(mstats.Frees), nil)

			Gauge("runtime.mem.heap.alloc", float64(mstats.HeapAlloc), nil)
			Gauge("runtime.mem.heap.sys", float64(mstats.HeapSys), nil)
			Gauge("runtime.mem.heap.idle", float64(mstats.HeapIdle), nil)
			Gauge("runtime.mem.heap.inuse", float64(mstats.HeapInuse), nil)
			Gauge("runtime.mem.heap.released", float64(mstats.HeapReleased), nil)
			Gauge("runtime.mem.heap.objects", float64(mstats.HeapObjects), nil)

			Gauge("runtime.mem.stack.inuse", float64(mstats.StackInuse), nil)
			Gauge("runtime.mem.stack.sys", float64(mstats.StackSys), nil)
			Gauge("runtime.mem.stack.mcache_inuse", float64(mstats.MCacheInuse), nil)
			Gauge("runtime.mem.stack.mcache_sys", float64(mstats.MCacheSys), nil)
			Gauge("runtime.mem.othersys", float64(mstats.OtherSys), nil)

			Gauge("runtime.gc.sys", float64(mstats.GCSys), nil)
			Gauge("runtime.gc.next", float64(mstats.NextGC), nil)
			Gauge("runtime.gc.last", float64(mstats.LastGC), nil)
			Gauge("runtime.gc.pause_total_ns", float64(mstats.PauseTotalNs), nil)
			Gauge("runtime.gc.pause", float64(mstats.PauseNs[(mstats.NumGC+255)%256]), nil)
			Gauge("runtime.gc.count", float64(mstats.NumGC), nil)
		}
	}
}

// Gauge publishes the gauge type metrics
func Gauge(name string, value float64, tags []string) {
	err := client.Gauge(name, value, tags, 1)
	if err != nil {
		logger.WithError(err).Error("failed to report Gauge metrics")
	}
}

// GaugeI is same as Gauge but it accepts the
// metric value as an integer
func GaugeI(name string, value int, tags []string) {
	Gauge(name, float64(value), tags)
}

// Incr publishes the counter type metrics
func Incr(name string, tags []string) {
	err := client.Incr(name, tags, 1)
	if err != nil {
		logger.WithError(err).Errorf("failed to report Increment metrics")
	}
}

// Timed publishes gauge type metrics with their value
// set to the nanosecond difference in current time and
// the value of `start`.
func Timed(name string, start time.Time, tags []string) {
	diff := time.Now().Sub(start).Nanoseconds()
	Gauge(name, float64(diff), tags)
}
