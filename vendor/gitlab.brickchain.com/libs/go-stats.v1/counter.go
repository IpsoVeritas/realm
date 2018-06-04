package stats

import (
	"strings"

	metrics "github.com/armon/go-metrics"
)

// Increment a metric value.
func Increment(path string, val float32) {
	metrics.IncrCounter(strings.Split(path, "."), val)
}

// Gauge sends a value for a metric.
func Gauge(path string, val float32) {
	metrics.SetGauge(strings.Split(path, "."), val)
}
