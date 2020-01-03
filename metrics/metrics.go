package metrics

import (
	"github.com/DataDog/datadog-go/statsd"
	"log"
	"strings"
)

var client *statsd.Client

func Init(ns string) (err error) {
	ns = strings.TrimRight(ns, ".") + "."
	client, err = statsd.New("127.0.0.1:8125", statsd.WithNamespace(ns))
	return err
}

func Gauge(name string, value float64, tags []string, rate float64) {
	err := client.Gauge(name, value, tags, rate)
	if err != nil {
		log.Printf("failed to report Gauge metrics. %s", err)
	}
}

func Incr(name string, tags []string, rate float64) {
	err := client.Incr(name, tags, rate)
	if err != nil {
		log.Printf("failed to report Increment metrics. %s", err)
	}
}