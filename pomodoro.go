package main

import (
	"flag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os/exec"
	"os/user"
	"strconv"
	"time"
)

const FocusLoopMinuteCount = 1
const RestLoopMinuteCount = 1
const LoopMax = 10
const FocusState = "focus"
const RestState = "rest"

type userConfig struct {
	focusDuration time.Duration
	restDuration  time.Duration
	loopMax       int
}

func main() {

	config := getConfig()
	userName := getUserName()

	timeOnLoopStart := time.Now()
	focusLoopPassed := 0
	restLoopPassed := 0
	state := FocusState

	for focusLoopPassed < config.loopMax {
		timeCurrent := time.Now()

		if state == FocusState {
			if timeCurrent.Sub(timeOnLoopStart) >= config.focusDuration {
				focusLoopPassed++
				timeOnLoopStart = timeCurrent
				state = RestState

				message := strconv.Itoa(focusLoopPassed) + " focus loop passed. Take a break and start again."
				if !sendMessage(message, userName) {
					break
				}

			}

			time.Sleep(config.focusDuration)

			if focusLoopPassed == config.loopMax {
				message := "you finish all (" + strconv.Itoa(config.loopMax) + ") your focus loops. Congrats!"
				sendMessage(message, userName)
				break
			}
		}

		if state == RestState {
			if timeCurrent.Sub(timeOnLoopStart) >= config.restDuration {
				restLoopPassed++
				timeOnLoopStart = timeCurrent
				state = FocusState

				message := strconv.Itoa(restLoopPassed) + " rest loop passed. Get back to work."
				if !sendMessage(message, userName) {
					break
				}
			}
			time.Sleep(config.restDuration)
		}

	}

}

func getUserName() string {
	userCurrent, _ := user.Current()
	userName := cases.Title(language.English).String(userCurrent.Username)

	return userName
}

func getConfig() userConfig {
	focusPointer := flag.Int("focus", FocusLoopMinuteCount, "focus loop duration in minutes")
	restPointer := flag.Int("rest", RestLoopMinuteCount, "rest loop duration in in minutes")
	loopCountPointer := flag.Int("loopCount", LoopMax, "max focus loop count")

	flag.Parse()
	focusDuration := time.Duration(*focusPointer) * time.Minute
	restDuration := time.Duration(*restPointer) * time.Minute

	return userConfig{focusDuration, restDuration, *loopCountPointer}
}

func sendMessage(message string, userName string) bool {
	text := "--text=" + userName + ", " + message
	cancelLabel := "--cancel-label=Stop Pomodoro"
	okLabel := "--ok-label=Start next Loop"
	title := "--title=Pomodoro"
	cmd := exec.Command("zenity", "--question", cancelLabel, okLabel, text, title)
	err := cmd.Run()
	if err != nil {
		return false
	}

	return cmd.ProcessState.Success()
}
