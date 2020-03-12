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
	ctx, cancel := context.WithCancel(context.Background())

	log.Init(logLevel, logFormat)

	err := setupMetrics()
	if err != nil {
		log.Global.WithError(err).Errorf("failed to initialize metrics subsystem")
	}

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

func setupMetrics() error {
	if metricsSink == "" {
		return nil
	}

	if metricsSink == "dogstatsd" {
		sink, err := metrics.NewStatsdSink(statsdAddr, statsdPort, statsdNS)
		if err != nil {
			return err
		}

		metrics.SetSink(sink)
		return nil
	}

	if metricsSink == "stderr" {
		sink := metrics.NewFileSink(os.Stderr, "plain")
		metrics.SetSink(sink)
		return nil
	}

	return fmt.Errorf("unsupported metrics sink: %s", metricsSink)
}
