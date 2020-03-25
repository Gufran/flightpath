package discovery

import (
	"flag"
	"fmt"
	"github.com/Gufran/flightpath/version"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	consul "github.com/hashicorp/consul/api"
	"os"
)

type Config struct {
	Global *GlobalConfig
	Consul *ConsulConfig
	XDS    *XDS
}

func NewEmptyConfig() *Config {
	return &Config{
		Global: &GlobalConfig{},
		Consul: &ConsulConfig{},
		XDS: &XDS{
			Envoy: &EnvoyConfig{},
			Debug: &DebugConfig{},
		},
	}
}

type GlobalConfig struct {
	LogLevel             string
	LogFormat            string
	MetricsSink          string
	EnableRuntimeMetrics bool
	DogstatsdAddr        string
	DogstatsdPort        int
	DogstatsdNS          string
}

type XDS struct {
	ServiceName string
	ListenPort  int
	Consul      *consul.Client
	Cache       cache.SnapshotCache

	Envoy *EnvoyConfig
	Debug *DebugConfig
}

func (x *XDS) Init(client *consul.Client, sn cache.SnapshotCache) {
	x.Consul = client
	x.Cache = sn
}

type ConsulConfig struct {
	Proto string
	Host  string
	Port  int
	Token string
}

type EnvoyConfig struct {
	NodeName       string
	ListenerPort   int
	AccessLogPath  string
	EnableTracing  bool
	TracingOpName  string
	TracingVerbose bool
}

type DebugConfig struct {
	Enable bool
	Port   int
}

func (c *Config) ParseFlags() {
	var showVersion bool

	flag.StringVar(&c.Global.LogLevel, "log.level", "INFO", "Set log verbosity. Valid options are trace, debug, error, warn, info, fatal and panic")
	flag.StringVar(&c.Global.LogFormat, "log.format", "json", "Format of the log message. Valid options are json and plain")
	flag.StringVar(&c.Global.MetricsSink, "metrics.sink", "", "Set the metrics sink. Valid options are 'dogstatsd' and 'stderr'")
	flag.BoolVar(&c.Global.EnableRuntimeMetrics, "metrics.runtime", true, "Expose runtime stats on memory and CPU")
	flag.StringVar(&c.Global.DogstatsdAddr, "dogstatsd.addr", "127.0.0.1", "Address of the dogstatsd agent")
	flag.IntVar(&c.Global.DogstatsdPort, "dogstatsd.port", 8125, "Port of the dogstatsd agent")
	flag.StringVar(&c.Global.DogstatsdNS, "dogstatsd.namespace", "flightpath", "Metrics namespace for dogstatsd")

	flag.StringVar(&c.Consul.Proto, "consul.proto", "http", "Protocol used to connect with consul agent")
	flag.IntVar(&c.Consul.Port, "consul.port", 8500, "Port on which the consul agent is listening")
	flag.StringVar(&c.Consul.Host, "consul.host", "127.0.0.1", "Network address to a consul agent")
	flag.StringVar(&c.Consul.Token, "consul.token", "", "Consul token to use")

	flag.StringVar(&c.XDS.ServiceName, "name", "flightpath", "Name used to register the flightpath service in Consul Catalog")
	flag.IntVar(&c.XDS.ListenPort, "port", 7171, "Port for XDS listener")

	flag.IntVar(&c.XDS.Envoy.ListenerPort, "envoy.listen.port", 9292, "Port used by Envoy Listener")
	flag.StringVar(&c.XDS.Envoy.AccessLogPath, "envoy.access-logs", "/var/log/envoy/access.log", "Path to the file where envoy will write listener access logs")
	flag.StringVar(&c.XDS.Envoy.NodeName, "node-name", "flightpath-edge", "Named of the Envoy node")
	flag.BoolVar(&c.XDS.Envoy.EnableTracing, "envoy.tracing.enabled", false, "Enable request tracing on envoy")
	flag.StringVar(&c.XDS.Envoy.TracingOpName, "envoy.tracing.op-name", "egress", "Tracing operation name, valid values are 'ingress' or 'egress'")
	flag.BoolVar(&c.XDS.Envoy.TracingVerbose, "envoy.tracing.verbose", false, "Add verbose information to traces")

	flag.BoolVar(&c.XDS.Debug.Enable, "debug", false, "Start debug HTTP server on loopback interface")
	flag.IntVar(&c.XDS.Debug.Port, "debug.port", 7180, "Network port to use for debug HTTP server")

	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.Parse()

	if showVersion {
		fmt.Println(version.FullString())
		os.Exit(0)
	}
}
