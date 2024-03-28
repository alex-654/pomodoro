package main

import (
	"flag"
	"time"
)

// Default values if user don't pass values
const (
	focusLoopMinuteCount = 40
	restLoopMinuteCount  = 10
	maxLoop              = 8
)

// A parseConfig read what user pass throw command flags and wrap that input to config.
// If input empty use a default constant values
func parseConfig() config {
	var (
		focus     int
		rest      int
		loopCount int
	)
	flag.IntVar(&focus, "focus", focusLoopMinuteCount, "focus loop duration in minutes")
	flag.IntVar(&rest, "rest", restLoopMinuteCount, "rest loop duration in in minutes")
	flag.IntVar(&loopCount, "loopCount", maxLoop, "max focus loop count")
	flag.Parse()

	focusDuration := time.Duration(focus) * time.Millisecond
	restDuration := time.Duration(rest) * time.Millisecond

	return config{focusDuration, restDuration, loopCount}
}
