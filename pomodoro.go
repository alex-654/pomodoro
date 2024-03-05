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

const FocusLoopSecondCount = 30 * 60
const RestLoopSecondCount = 10 * 60
const LoopMax = 5
const FocusState = "focus"
const RestState = "rest"

type userConfig struct {
	focusLoopSecondCount int
	restLoopSecondCount  int
	loopMax              int
}

func main() {

	config := getConfig()
	userName := getUserName()

	unixTimeOnLoopStart := time.Now().Unix()
	focusLoopPassed := 0
	restLoopPassed := 0
	state := FocusState

	for focusLoopPassed < config.loopMax {
		unixTimeCurrent := time.Now().Unix()

		if state == FocusState {
			if unixTimeCurrent-unixTimeOnLoopStart >= int64(config.focusLoopSecondCount) {
				focusLoopPassed++
				unixTimeOnLoopStart = unixTimeCurrent
				state = RestState

				message := strconv.Itoa(focusLoopPassed) + " focus loop passed. Take a break and start again."
				if !sendMessage(message, userName) {
					break
				}

			}

			time.Sleep(time.Duration(int64(config.focusLoopSecondCount)) * time.Second)

			if focusLoopPassed == config.loopMax {
				message := "you finish all (" + strconv.Itoa(config.loopMax) + ") your focus loops. Congrats!"
				sendMessage(message, userName)
				break
			}
		}

		if state == RestState {
			if unixTimeCurrent-unixTimeOnLoopStart >= int64(config.restLoopSecondCount) {
				restLoopPassed++
				unixTimeOnLoopStart = unixTimeCurrent
				state = FocusState

				message := strconv.Itoa(restLoopPassed) + " rest loop passed. Get back to work."
				if !sendMessage(message, userName) {
					break
				}
			}
			time.Sleep(time.Duration(int64(config.restLoopSecondCount)) * time.Second)
		}

	}

}

func getUserName() string {
	userCurrent, _ := user.Current()
	userName := cases.Title(language.English).String(userCurrent.Username)

	return userName
}

func getConfig() userConfig {
	focusPointer := flag.Int("focus", FocusLoopSecondCount, "focus loop duration in seconds")
	restPointer := flag.Int("rest", RestLoopSecondCount, "rest loop duration in seconds")
	loopCountPointer := flag.Int("loopCount", LoopMax, "max focus loop count")

	flag.Parse()
	config := userConfig{*focusPointer, *restPointer, *loopCountPointer}
	return config
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
