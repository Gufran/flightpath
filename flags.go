package main

import (
	"flag"
	"fmt"
	"github.com/Gufran/flightpath/discovery"
	"github.com/Gufran/flightpath/version"
	"os"
)

var config = &discovery.Config{}

var (
	logLevel  string
	logFormat string

	enableStatsd bool
	statsdAddr   string
	statsdPort   int
	statsdNS     string

	showVersion bool
)

func init() {
	flag.StringVar(&config.SelfName, "name", "flightpath", "Name used to register the flightpath service in Consul Catalog")
	flag.StringVar(&config.NodeName, "node-name", "flightpath-edge", "Named of the Envoy node")
	flag.IntVar(&config.ListenPort, "port", 7171, "Port for XDS listener")
	flag.StringVar(&config.ConsulProto, "consul.proto", "http", "Protocol used to connect with consul agent")
	flag.IntVar(&config.ConsulPort, "consul.port", 8500, "Port on which the consul agent is listening")
	flag.StringVar(&config.ConsulHost, "consul.host", "127.0.0.1", "Network address to a consul agent")
	flag.StringVar(&config.ConsulToken, "consul.token", "", "Consul token to use")
	flag.IntVar(&config.EnvoyListenPort, "envoy.listen.port", 9292, "Port used by Envoy Listener")
	flag.StringVar(&config.EnvoyAccessLogPath, "envoy.access-logs", "/var/log/envoy/access.log", "Path to the file where envoy will write listener access logs")
	flag.StringVar(&logLevel, "log.level", "INFO", "Set log verbosity. Valid options are trace, debug, error, warn, info, fatal and panic")
	flag.StringVar(&logFormat, "log.format", "json", "Format of the log message. Valid options are json and plain")
	flag.BoolVar(&enableStatsd, "dogstatsd", false, "Enable publishing metrics to dogstatsd agent")
	flag.StringVar(&statsdAddr, "dogstatsd.addr", "127.0.0.1", "Address of the dogstatsd agent")
	flag.IntVar(&statsdPort, "dogstatsd.port", 8125, "Port of the dogstatsd agent")
	flag.StringVar(&statsdNS, "dogstatsd.namespace", "flightpath", "Metrics namespace for dogstatsd")
	flag.BoolVar(&config.StartDebugServer, "debug", true, "Start debug HTTP server on loopback interface")
	flag.IntVar(&config.DebugServerPort, "debug.port", 7180, "Network port to use for debug HTTP server")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.Parse()

	if showVersion {
		fmt.Println(version.FullString())
		os.Exit(0)
	}
}
