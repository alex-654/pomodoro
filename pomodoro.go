package main

import (
	"flag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"log"
	"log/syslog"
	"os/exec"
	"os/user"
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
	userName := getUserName()

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
				state = StateRest

				message := strconv.Itoa(focusLoopCount) + " focus loop passed. Take a break and start again."
				if !sendMessage(message, userName) {
					break
				}
			}

			time.Sleep(config.FocusDuration)

			if focusLoopCount == config.MaxLoop {
				state = StateFinish
			}
		}

		if state == StateRest {
			if timeCurrent.Sub(timeOnLoopStart) >= config.RestDuration {
				restLoopCount++
				timeOnLoopStart = timeCurrent
				state = StateFocus

				message := strconv.Itoa(restLoopCount) + " rest loop passed. Get back to work."
				if !sendMessage(message, userName) {
					break
				}
			}
			time.Sleep(config.RestDuration)
		}

		if state == StateFinish {
			message := "you finish all (" + strconv.Itoa(config.MaxLoop) + ") your focus loops. Congrats!"
			sendMessage(message, userName)
		}
	}
}

func getUserName() string {
	userCurrent, _ := user.Current()

	return cases.Title(language.English).String(userCurrent.Username)
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

func sendMessage(message string, userName string) bool {
	text := "--text=" + userName + ", " + message
	cancelLabel := "--cancel-label=Stop Pomodoro"
	okLabel := "--ok-label=Start next Loop"
	title := "--title=Pomodoro"
	cmd := exec.Command("zenity", "--question", cancelLabel, okLabel, text, title)
	err := cmd.Run()
	if err != nil {
		logger, _ := syslog.NewLogger(syslog.LOG_ERR, log.Ldate|log.Lmicroseconds|log.Llongfile)
		logger.Fatal("Error happen when sending message to user. ", err.Error())
	}

	return cmd.ProcessState.Success()
}
