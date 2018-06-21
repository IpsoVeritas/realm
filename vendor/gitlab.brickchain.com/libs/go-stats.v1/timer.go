package stats

import (
	"strings"
	"time"

	logger "github.com/Brickchain/go-logger.v1"

	metrics "github.com/armon/go-metrics"
	"github.com/spf13/viper"
)

// Timer struct for the timings
type Timer struct {
	start time.Time
	path  string
}

// StartTimer starts the timer
func StartTimer(path string) *Timer {
	if viper.GetBool("stats_debug") {
		logger.Debugf("Timer %s started", path)
	}
	return &Timer{
		start: time.Now(),
		path:  path,
	}
}

// Stop stops the timer
func (t *Timer) Stop() {
	if viper.GetBool("stats_debug") {
		logger.Debugf("Timer %s stopped: %v", t.path, time.Since(t.start))
	}
	metrics.MeasureSince(strings.Split(t.path, "."), t.start)
}
