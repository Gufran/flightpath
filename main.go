package main

import (
	"context"
	"flag"
	"github.com/Gufran/flightpath/log"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"

	"github.com/Gufran/flightpath/discovery"
)

var config = &discovery.Config{}

var (
	logLevel string
	logFormat string
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
	flag.StringVar(&logLevel, "log.level", "INFO", "Set log verbosity. Valid options are trace, debug, error, warn, info, fatal and panic")
	flag.StringVar(&logFormat, "log.format", "json", "Format of the log message. Valid options are json and plain")
	flag.Parse()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	log.Init(logLevel, logFormat)

	exit := make(chan os.Signal)
	signal.Notify(exit, os.Interrupt)

	shutdown, err := discovery.Start(ctx, config)
	if err != nil {
		logrus.Printf("failed to start service discovery server. %s", err)
		return
	}

	<-exit

	shutdown()
	cancel()
}
