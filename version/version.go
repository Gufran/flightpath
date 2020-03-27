package version

import (
	"fmt"
	"runtime"
)

var (
	Version   string = "0.0.3"
	Commit    string
	BuildTime string
)

func FullString() string {
	return fmt.Sprintf("v%s on commit (%s) built on %s with runtime %s", Version, Commit, BuildTime, runtime.Version())
}
