package stats

import (
	"strings"
	"time"

	metrics "github.com/armon/go-metrics"
)

// Timer is the struct that holds the information about an ongoing timer.
type Timer struct {
	start time.Time
	path  string
}

// StartTimer starts a new timer and returns a Timer object instance.
func StartTimer(path string) *Timer {
	return &Timer{
		start: time.Now(),
		path:  path,
	}
}

// Stop an ongoing timer.
func (t *Timer) Stop() {
	metrics.MeasureSince(strings.Split(t.path, "."), t.start)
}
