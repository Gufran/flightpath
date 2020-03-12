package metrics

import (
	"github.com/Gufran/flightpath/log"
	"time"
)

var (
	client = NewNoOpSink()
	logger = log.New("metrics")
)

// SetupSink initializes the metrics subsystem.
func SetSink(sink Sink) {
	client = sink
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
