package main

import (
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	FocusLoopMinuteCount = 30
	RestLoopMinuteCount  = 10
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

	for {
		timeCurrent := time.Now()

		if state == StateFocus {
			if timeCurrent.Sub(timeOnLoopStart) >= config.FocusDuration {
				focusLoopCount++
				if !sendMessage(state, focusLoopCount, &config) {
					break
				}
				timeOnLoopStart = time.Now()
				timeCurrent = time.Now()
				state = StateRest
			} else {
				time.Sleep(config.FocusDuration)
			}
		}

		if state == StateRest {
			if timeCurrent.Sub(timeOnLoopStart) >= config.RestDuration {
				restLoopCount++
				if !sendMessage(state, restLoopCount, &config) {
					break
				}
				timeOnLoopStart = time.Now()
				timeCurrent = time.Now()
				state = StateFocus
			} else {
				time.Sleep(config.RestDuration)
			}
		}

		if focusLoopCount == config.MaxLoop {
			state = StateFinish
			sendMessage(state, focusLoopCount, &config)
			break
		}
	}
}

// Get user config. User can pass rest and focus duration throw command flags
func getConfig() UserConfig {
	var (
		focus     int
		rest      int
		loopCount int
	)
	flag.IntVar(&focus, "focus", FocusLoopMinuteCount, "focus loop duration in minutes")
	flag.IntVar(&rest, "rest", RestLoopMinuteCount, "rest loop duration in in minutes")
	flag.IntVar(&loopCount, "loopCount", MaxLoop, "max focus loop count")
	flag.Parse()

	focusDuration := time.Duration(focus) * time.Minute
	restDuration := time.Duration(rest) * time.Minute

	return UserConfig{focusDuration, restDuration, loopCount}
}

// Send user notification about loop passed, then handel user answer and update user config
func sendMessage(state string, loopCount int, config *UserConfig) bool {
	cmd := createCommand(state, loopCount, config)
	bytes, _ := cmd.Output()
	output := string(bytes)
	return handleCmdResult(cmd.ProcessState.Success(), output, state, config)
}

// Create command with params that display GTK+ dialogs
func createCommand(state string, loopCount int, config *UserConfig) *exec.Cmd {
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
	title := "--title=Pomodoro"
	text := "--text=" + messageMap[state]
	okLabel := "--ok-label=" + okLabelMap[state] + " ✅"
	stopLabel := "--cancel-label= off 🙅"
	resetLabel := "--extra-button=Reset 🔄"
	currentDuration := 0.0
	if state == StateFocus {
		currentDuration = config.RestDuration.Minutes()
	} else {
		currentDuration = config.FocusDuration.Minutes()
	}
	form := fmt.Sprintf("--entry-text=%.f", currentDuration)
	return exec.Command("zenity", "--entry", title, text, okLabel, form, stopLabel, resetLabel)
}

func handleCmdResult(isSuccess bool, output string, state string, config *UserConfig) bool {
	if isSuccess {
		d := regexp.MustCompile(`\d+`).FindString(output)
		minutes, _ := strconv.Atoi(d)
		if minutes <= 0 {
			restart(*config)
			return false
		}
		nextLoopDuration := time.Duration(minutes) * time.Minute
		if state == StateFocus {
			config.RestDuration = nextLoopDuration
		}
		if state == StateRest {
			config.FocusDuration = nextLoopDuration
		}
		return true
	}

	if strings.Contains(output, "Reset") {
		restart(*config)
		return false
	}

	return false
}

// Restart pomodoro timer with last userConfig params
func restart(config UserConfig) {
	focusMinutes := int(config.FocusDuration.Minutes())
	restMinutes := int(config.RestDuration.Minutes())
	focus := "--focus=" + strconv.Itoa(focusMinutes)
	rest := "--rest=" + strconv.Itoa(restMinutes)
	loopCount := "--loopCount=" + strconv.Itoa(config.MaxLoop)
	cmd := exec.Command("pomodoro", focus, rest, loopCount)
	err := cmd.Start()
	if err != nil {
		logger, _ := syslog.NewLogger(syslog.LOG_ERR, log.Ldate|log.Lmicroseconds|log.Llongfile)
		logger.Fatal("Can't create new pomodoro timer. ", err.Error())
	}
}
