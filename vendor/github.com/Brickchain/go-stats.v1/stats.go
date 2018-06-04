// Package stats is a helper lib for collecting statistics.
//
// Example:
//	package main
//
//	import (
//		"os"
//		"time"
//		"path"
//
//		stats "github.com/Brickchain/go-stats.v1"
//	)
//
//	func main() {
//		// start an inmem metrics sink that will print metrics once per minute.
//		// set the instance name to the name of our binary.
// 		stats.Setup("inmem", path.Base(os.Args[0]))
//
//		someFunc()
//
//		// Wait a bit more than a minute in order to see the metrics being printed
//		time.Sleep(time.Second * 62)
//	}
//
//	func someFunc() {
//		t := stats.StartTimer("someFunc.total")
//		defer t.Stop()
//	}
package stats

import (
	"net/http"
	"os"
	"syscall"
	"time"

	metrics "github.com/armon/go-metrics"
	"github.com/armon/go-metrics/datadog"
	"github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Setup the metrics sink. Takes sink type (datadog, prometheus or inmem) and a name to prepend the metrics keys with.
func Setup(sinkType, name string) error {
	var sink metrics.MetricSink
	var err error
	switch sinkType {
	case "datadog":
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		sink, err = datadog.NewDogStatsdSink(os.Getenv("DOGSTATSD"), hostname)
		if err != nil {
			return err
		}
	case "prometheus":
		sink, err = prometheus.NewPrometheusSink()
		if err != nil {
			return err
		}
		http.Handle("/metrics", promhttp.Handler())
	case "inmem":
		inmem := metrics.NewInmemSink(time.Second*1, time.Minute*5)
		_ = metrics.NewInmemSignal(inmem, syscall.SIGUSR1, os.Stdout)
		go func() {
			for {
				time.Sleep(time.Second * 60)
				syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
			}
		}()
		sink = inmem
	default:
	}

	if sink != nil {
		metrics.NewGlobal(metrics.DefaultConfig(name), sink)
	}

	return nil
}
