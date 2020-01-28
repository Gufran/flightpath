package metrics

type Sink interface {
	Gauge(string, float64, []string, float64) error
	Incr(string, []string, float64) error
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

