package main

import (
	"flag"
	"log"
	"log/syslog"
	"os/exec"
	"strconv"
	"time"
)

const (
	FocusLoopMinuteCount = 40
	RestLoopMinuteCount  = 15
	MaxLoop              = 8
)

const (
	StateFocus  = "focus"
	StateRest   = "rest"
	StateFinish = "finish"
)

type UserConfig struct {
	FocusDuration time.Duration
	RestDuration  time.Duration
	MaxLoop       int
}

func main() {

	config := getConfig()
	timeOnLoopStart := time.Now()
	focusLoopCount := 0
	restLoopCount := 0
	state := StateFocus

	for state != StateFinish {
		timeCurrent := time.Now()

		if state == StateFocus {
			if timeCurrent.Sub(timeOnLoopStart) >= config.FocusDuration {
				focusLoopCount++
				timeOnLoopStart = timeCurrent
				if !sendMessage(state, focusLoopCount) {
					break
				}
				state = StateRest
			}

			time.Sleep(config.FocusDuration)

			if focusLoopCount == config.MaxLoop {
				state = StateFinish
				sendMessage(state, focusLoopCount)
			}
		}

		if state == StateRest {
			if timeCurrent.Sub(timeOnLoopStart) >= config.RestDuration {
				restLoopCount++
				timeOnLoopStart = timeCurrent
				if !sendMessage(state, restLoopCount) {
					break
				}
				state = StateFocus
			}
			time.Sleep(config.RestDuration)
		}
	}
}

func getConfig() UserConfig {
	focusPointer := flag.Int("focus", FocusLoopMinuteCount, "focus loop duration in minutes")
	restPointer := flag.Int("rest", RestLoopMinuteCount, "rest loop duration in in minutes")
	loopCountPointer := flag.Int("loopCount", MaxLoop, "max focus loop count")
	flag.Parse()
	focusDuration := time.Duration(*focusPointer) * time.Minute
	restDuration := time.Duration(*restPointer) * time.Minute

	return UserConfig{focusDuration, restDuration, *loopCountPointer}
}

func sendMessage(state string, loopCount int) bool {
	messageMap := map[string]string{
		StateFocus:  strconv.Itoa(loopCount) + " focus loop passed.",
		StateRest:   strconv.Itoa(loopCount) + " rest loop passed.",
		StateFinish: "All (" + strconv.Itoa(loopCount) + ") focus loops done. Congrats!",
	}
	okLabelMap := map[string]string{
		StateFocus:  "Take a break",
		StateRest:   "Focus",
		StateFinish: "Finish",
	}
	text := "--text=" + messageMap[state]
	okLabel := "--ok-label=" + okLabelMap[state]
	cancelLabel := "--cancel-label=Stop"
	title := "--title=Pomodoro"
	cmd := exec.Command("zenity", "--question", cancelLabel, okLabel, text, title)
	err := cmd.Run()
	if err != nil {
		logger, _ := syslog.NewLogger(syslog.LOG_ERR, log.Ldate|log.Lmicroseconds|log.Llongfile)
		logger.Fatal("Error happen when sending message to user. ", err.Error())
	}

	return cmd.ProcessState.Success()
}
