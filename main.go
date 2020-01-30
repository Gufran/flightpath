package main

import (
	"context"
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

	if enableStatsd {
		err := metrics.Init(statsdAddr, statsdPort, statsdNS)
		if err != nil {
			log.Global.WithError(err).Errorf("failed to initialize metrics subsystem")
		}
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
