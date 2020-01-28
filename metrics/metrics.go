package metrics

import (
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"github.com/Gufran/flightpath/log"
	"strings"
	"time"
)

var (
	client Sink = NewNoOpSink()
    logger = log.New("metrics")
)

// Init initializes the metrics subsystem.
// addr is the host address where the statsd agent can be
// reached. port defines the network port to use for
// communication.
// ns is used as a prefix for all metrics that are published
// by flightpath.
func Init(addr string, port int, ns string) (err error) {
	if ns != "" {
		ns = strings.TrimRight(ns, ".") + "."
	}

	client, err = statsd.New(fmt.Sprintf("%s:%d", addr, port), statsd.WithNamespace(ns))
	return err
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