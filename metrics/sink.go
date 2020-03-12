package metrics

import (
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"io"
	"strings"
)

type Sink interface {
	Gauge(string, float64, []string, float64) error
	Incr(string, []string, float64) error
}

// addr is the host address where the statsd agent can be
// reached. port defines the network port to use for
// communication.
// ns is used as a prefix for all metrics that are published
// by flightpath.
func NewStatsdSink(addr string, port int, ns string) (Sink, error) {
	if ns != "" {
		ns = strings.TrimRight(ns, ".") + "."
	}

	return statsd.New(fmt.Sprintf("%s:%d", addr, port), statsd.WithNamespace(ns))
}

var _ Sink = &NoOpSink{}

func NewNoOpSink() Sink {
	return &NoOpSink{}
}

type NoOpSink struct{}

func (s *NoOpSink) Gauge(string, float64, []string, float64) error {
	return nil
}

func (s *NoOpSink) Incr(string, []string, float64) error {
	return nil
}

var _ Sink = (*FileSink)(nil)

func NewFileSink(out io.Writer, format string) Sink {
	return &FileSink{
		out:    out,
		format: format,
	}
}

type FileSink struct {
	out    io.Writer
	format string
}

func (s *FileSink) write(kind string, name string, value float64, tags []string, rate float64) error {
	var item string
	if s.format == "json" {
		b, err := json.Marshal(map[string]interface{}{
			"type":  kind,
			"name":  name,
			"value": value,
			"tags":  strings.Join(tags, ", "),
			"rate":  rate,
		})

		if err != nil {
			return err
		}

		item = string(b[:])
	} else {
		item = fmt.Sprintf("%-40s %-12s %-4f  %s", name, kind, value, strings.Join(tags, ","))
	}

	_, err := fmt.Fprint(s.out, strings.TrimSpace(item)+"\n")
	return err
}

func (s *FileSink) Gauge(name string, value float64, tags []string, rate float64) error {
	return s.write("gauge", name, value, tags, rate)
}

func (s *FileSink) Incr(name string, tags []string, rate float64) error {
	return s.write("increment", name, 1, tags, rate)
}
