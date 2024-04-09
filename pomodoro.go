package pomodoro

import (
	"time"
)

func main() {
	p := pomodoro{
		conf:      parseConfig(),
		messenger: zenityMessenger{},
	}
	p.run()
}

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

// A messenger interface send user message about state changes
type messenger interface {
	send(state string, focusLoopCount int, conf config) response
}

// A response represent user response when he gets notification about loop end
type response struct {
	off      bool          // turn off timer
	reset    bool          // reset/restart timer
	nextLoop time.Duration // user decide keep going and set next loop to some minutes
}

// A pomodoro struct keep user settings and message provider
type pomodoro struct {
	conf      config
	messenger messenger
}

// A run func realized logic of timer when to notify user about when need to rest and focus
func (p *pomodoro) run() {
	timeOnLoopStart := time.Now()
	focusLoopCount := 0
	restLoopCount := 0
	state := stateFocus

	for {
		timeCurrent := time.Now()

		if state == stateFocus {
			if timeCurrent.Sub(timeOnLoopStart) >= p.conf.focusDuration {
				focusLoopCount++
				r := p.messenger.send(state, focusLoopCount, p.conf)
				if !p.handleResponse(r, state) {
					break
				}
				timeOnLoopStart = time.Now()
				timeCurrent = time.Now()
				state = stateRest
			} else {
				time.Sleep(p.conf.focusDuration)
			}
		}

		if state == stateRest {
			if timeCurrent.Sub(timeOnLoopStart) >= p.conf.restDuration {
				restLoopCount++
				r := p.messenger.send(state, focusLoopCount, p.conf)
				if !p.handleResponse(r, state) {
					break
				}
				timeOnLoopStart = time.Now()
				timeCurrent = time.Now()
				state = stateFocus
			} else {
				time.Sleep(p.conf.restDuration)
			}
		}

		if focusLoopCount == p.conf.maxLoop {
			state = stateFinish
			p.messenger.send(state, focusLoopCount, p.conf)
			break
		}
	}
}

// A handleResponse func handel user response
func (p *pomodoro) handleResponse(r response, state string) bool {
	if r.off {
		return false
	}
	if r.reset {
		p.run()
		return false
	}
	if r.nextLoop <= 0 {
		return true
	}

	if state == stateFocus {
		p.conf.restDuration = r.nextLoop
	}
	if state == stateRest {
		p.conf.focusDuration = r.nextLoop
	}

	return true
}
