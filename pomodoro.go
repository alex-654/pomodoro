package main

import (
	"log"
	"log/syslog"
	"os/exec"
	"strconv"
	"time"
)

// States of pomodoro timer
const (
	stateFocus  = "focus"
	stateRest   = "rest"
	stateFinish = "finish"
)

// A config wrap user params in structure
type config struct {
	focusDuration time.Duration
	restDuration  time.Duration
	maxLoop       int
}

type messenger interface {
	send(state string, focusLoopCount int, conf config) response
}

// A response represent user response on loop notification end
type response struct {
	off             bool // turn off timer
	reset           bool // reset/restart timer
	nextLoopMinutes int  // user decide keep going and set next loop to some minutes
}

// A pomodoro func realized logic of timer when to notify user about when need to rest and focus
func pomodoro(conf config, m messenger) {
	timeOnLoopStart := time.Now()
	focusLoopCount := 0
	restLoopCount := 0
	state := stateFocus

	for {
		timeCurrent := time.Now()

		if state == stateFocus {
			if timeCurrent.Sub(timeOnLoopStart) >= conf.focusDuration {
				focusLoopCount++
				r := m.send(state, focusLoopCount, conf)
				if !handleResponse(r, state, &conf) {
					break
				}
				timeOnLoopStart = time.Now()
				timeCurrent = time.Now()
				state = stateRest
			} else {
				time.Sleep(conf.focusDuration)
			}
		}

		if state == stateRest {
			if timeCurrent.Sub(timeOnLoopStart) >= conf.restDuration {
				restLoopCount++
				r := m.send(state, focusLoopCount, conf)
				if !handleResponse(r, state, &conf) {
					break
				}
				timeOnLoopStart = time.Now()
				timeCurrent = time.Now()
				state = stateFocus
			} else {
				time.Sleep(conf.restDuration)
			}
		}

		if focusLoopCount == conf.maxLoop {
			state = stateFinish
			m.send(state, focusLoopCount, conf)
			break
		}
	}
}

func handleResponse(r response, state string, c *config) bool {
	if r.off {
		return false
	}
	if r.reset {
		restart(*c)
		return false
	}
	if r.nextLoopMinutes <= 0 {
		return true
	}

	nextLoopDuration := time.Duration(r.nextLoopMinutes) * time.Minute
	if state == stateFocus {
		c.restDuration = nextLoopDuration
	}
	if state == stateRest {
		c.focusDuration = nextLoopDuration
	}

	return true
}

// restart pomodoro timer with last userConfig params
func restart(config config) {
	focusMinutes := int(config.focusDuration.Minutes())
	restMinutes := int(config.restDuration.Minutes())
	focus := "--focus=" + strconv.Itoa(focusMinutes)
	rest := "--rest=" + strconv.Itoa(restMinutes)
	loopCount := "--loopCount=" + strconv.Itoa(config.maxLoop)
	cmd := exec.Command("pomodoro", focus, rest, loopCount)
	err := cmd.Start()
	if err != nil {
		logger, _ := syslog.NewLogger(syslog.LOG_ERR, log.Ldate|log.Lmicroseconds|log.Llongfile)
		logger.Fatal("Can't create new pomodoro timer. ", err.Error())
	}
}
