package log

import (
	srvlog "github.com/envoyproxy/go-control-plane/pkg/log"
	"github.com/sirupsen/logrus"
	"strings"
)

var _ srvlog.Logger = &ServerLog{}

var formatter string

// Init sets up the default log level and log output
// format on global logrus instance.
// Subsequent calls to this function will not propagate
// new configuration to already existing loggers.
func Init(level string, fmtname string) {
	formatter = fmtname

	logrus.SetFormatter(formatterFromName(formatter))
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"specified": level,
			"default":   "INFO",
		}).Warnf("invalid log level specified. using default value", )
		lvl = logrus.InfoLevel
	}

	logrus.SetLevel(lvl)
}

// New creates a new logrus instance with preconfigured
// log level and output format and returns the entry
// object with sybsystem name configured on it as a field.
// All output from the entry object will contain the subsystem
// name on key "subsystem"
func New(subsystem string) *logrus.Entry {
	l := logrus.New()
	l.SetLevel(logrus.GetLevel())
	l.SetFormatter(formatterFromName(formatter))
	return l.WithField("subsystem", subsystem)
}

func formatterFromName(f string) logrus.Formatter {
	f = strings.ToLower(f)
	switch f {
	case "json":
		return &logrus.JSONFormatter{}
	case "plain", "text", "cli":
		return &logrus.TextFormatter{}
	default:
		return &logrus.JSONFormatter{}
	}
}

// ServerLog implements the Logger interface for
// go-control-plane/pkg/log.Logger. It wraps an
// instance of *logrus.Entry to facilitate levelled
// and structured logs from go-control-plane server.
type ServerLog struct {
	underlying *logrus.Entry
}

// NewSrvLogger returns a new logger that is compatible
// with go-control-plane server.
func NewSrvLogger() srvlog.Logger {
	return &ServerLog{
		underlying: New("go-control-plane"),
	}
}

func (l *ServerLog) Infof(format string, args ...interface{}) {
	l.underlying.Debugf(format, args...)
}

func (l *ServerLog) Errorf(format string, args ...interface{}) {
	l.underlying.Errorf(format, args...)
}

