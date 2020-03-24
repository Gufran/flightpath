package main

import (
	"context"
	"fmt"
	"github.com/Gufran/flightpath/log"
	"github.com/Gufran/flightpath/metrics"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"

	"github.com/Gufran/flightpath/discovery"
)

func main() {
	config := discovery.NewEmptyConfig()
	config.ParseFlags()

	ctx, cancel := context.WithCancel(context.Background())

	log.Init(config.Global.LogLevel, config.Global.LogFormat)

	err := setupMetrics(config.Global)
	if err != nil {
		log.Global.WithError(err).Errorf("failed to initialize metrics subsystem")
	}

	exit := make(chan os.Signal, 1)
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

func setupMetrics(config *discovery.GlobalConfig) error {
	if config.MetricsSink == "" {
		return nil
	}

	if config.MetricsSink == "dogstatsd" {
		sink, err := metrics.NewStatsdSink(config.DogstatsdAddr, config.DogstatsdPort, config.DogstatsdNS)
		if err != nil {
			return err
		}

		metrics.SetSink(sink)
		return nil
	}

	if config.MetricsSink == "stderr" {
		sink := metrics.NewFileSink(os.Stderr, "plain")
		metrics.SetSink(sink)
		return nil
	}

	return fmt.Errorf("unsupported metrics sink: %s", config.MetricsSink)
}
