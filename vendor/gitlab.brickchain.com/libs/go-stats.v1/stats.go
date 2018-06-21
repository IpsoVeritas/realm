package stats

import (
	"net/http"
	"os"
	"syscall"
	"time"

	logger "github.com/Brickchain/go-logger.v1"
	metrics "github.com/armon/go-metrics"
	"github.com/armon/go-metrics/datadog"
	"github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

// FromEnv reads config from environment
func FromEnv(name string) {
	var sink metrics.MetricSink
	var err error
	switch viper.GetString("stats") {
	case "datadog":
		hostname, err := os.Hostname()
		if err != nil {
			logger.Fatal(err)
		}
		sink, err = datadog.NewDogStatsdSink(viper.GetString("dogstatsd"), hostname)
		if err != nil {
			logger.Fatal(err)
		}
	case "prometheus":
		sink, err = prometheus.NewPrometheusSink()
		if err != nil {
			logger.Fatal(err)
		}
		http.Handle("/metrics", promhttp.Handler())
	case "inmem":
		inmem := metrics.NewInmemSink(time.Second*1, time.Minute*5)
		_ = metrics.NewInmemSignal(inmem, metrics.DefaultSignal, os.Stdout)
		go func() {
			for {
				time.Sleep(time.Second * 60)
				proc, err := os.FindProcess(syscall.Getpid())
				if err == nil {
					proc.Signal(metrics.DefaultSignal)
				}
			}
		}()
		sink = inmem
	default:
	}

	if sink != nil {
		metrics.NewGlobal(metrics.DefaultConfig(name), sink)
	}
}
