package discovery

import (
	"github.com/envoyproxy/go-control-plane/pkg/log"
	stdlog "log"
)

var _ log.Logger = &ServerLog{}

type ServerLog struct {}

func (l *ServerLog) Infof(format string, args ...interface{}) {
	stdlog.Printf(format, args...)
}

func (l *ServerLog) Errorf(format string, args ...interface{}) {
	stdlog.Printf(format, args...)
}
