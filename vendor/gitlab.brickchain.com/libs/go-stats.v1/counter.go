package stats

import (
	"strings"

	metrics "github.com/armon/go-metrics"
)

// Increment increments a counter
func Increment(path string, val float32) {
	metrics.IncrCounter(strings.Split(path, "."), val)
}

// Gauge stores a value
func Gauge(path string, val float32) {
	metrics.SetGauge(strings.Split(path, "."), val)
}
