package util

import (
	"time"
)

// GetCurrentTimeNanos returns the current time in nanoseconds with wall clock time precision
// This is similar to the Java implementation that uses Instant.now().toEpochMilli() * 1_000_000 + (System.nanoTime() % 1_000_000)
func GetCurrentTimeNanos() int64 {
	// Get the current wall clock time in nanoseconds
	wallTime := time.Now().UnixNano()

	// For maximum precision, we could use a hybrid approach similar to the Java implementation
	// but Go's time.Now().UnixNano() already provides nanosecond precision with wall clock time
	// This is sufficient for our latency measurements

	return wallTime
}

// FormatDuration formats a duration in nanoseconds to a human-readable string
func FormatDuration(nanos int64) string {
	d := time.Duration(nanos) * time.Nanosecond

	if d < time.Microsecond {
		return d.String()
	} else if d < time.Millisecond {
		return d.String()
	} else if d < time.Second {
		return d.String()
	} else {
		return d.String()
	}
}
