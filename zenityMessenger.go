package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type zenityMessenger struct{}

// send func send notification to user about loop passed, then wrap user answer to response structure
func (z zenityMessenger) send(state string, loopCount int, conf config) response {
	cmd := z.createCmd(state, loopCount, conf)
	bytes, _ := cmd.Output()
	output := string(bytes)

	if cmd.ProcessState.Success() {
		minutesStr := regexp.MustCompile(`\d+`).FindString(output)
		minutes, _ := strconv.Atoi(minutesStr)
		return response{nextLoopMinutes: minutes}
	}
	if strings.Contains(output, "reset") {
		return response{reset: true}
	}

	return response{off: true}
}

// createCmd create command with params that display GTK dialogs
func (z zenityMessenger) createCmd(state string, loopCount int, conf config) *exec.Cmd {
	messageMap := map[string]string{
		stateFocus:  fmt.Sprintf("%d focus loop passed.", loopCount),
		stateRest:   fmt.Sprintf("%d rest loop passed.", loopCount),
		stateFinish: fmt.Sprintf("All %d focus loops finished. Congrats!", loopCount),
	}
	okLabelMap := map[string]string{
		stateFocus:  "Take a break",
		stateRest:   "Focus",
		stateFinish: "Finish",
	}
	title := "--title=Pomodoro " + messageMap[state]
	text := "--text=Next loop will be (minutes)"
	okLabel := "--ok-label=" + okLabelMap[state] + " ✅"
	stopLabel := "--cancel-label=Off"
	resetLabel := "--extra-button=Reset 🔄"
	currentDuration := 0.0
	if state == stateFocus {
		currentDuration = conf.restDuration.Minutes()
	} else {
		currentDuration = conf.focusDuration.Minutes()
	}
	form := fmt.Sprintf("--entry-text=%.f", currentDuration)
	return exec.Command("zenity", "--entry", title, text, okLabel, form, stopLabel, resetLabel)
}
