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

type Config struct {
	FocusDuration time.Duration
	RestDuration  time.Duration
	MaxLoop       int
}

func main() {
	config := parseConfig()
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

// ParseConfig read what user pass throw command flags and wrap that input in Config.
// If input empty use a default constant values
func parseConfig() Config {
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

	return Config{focusDuration, restDuration, loopCount}
}

// Send user notification about loop passed, then handel user answer and update user config
func sendMessage(state string, loopCount int, config *Config) bool {
	cmd := createCmd(state, loopCount, config)
	bytes, _ := cmd.Output()
	output := string(bytes)
	return handleCmdResult(cmd.ProcessState.Success(), output, state, config)
}

// Create command with params that display GTK dialogs
func createCmd(state string, loopCount int, config *Config) *exec.Cmd {
	messageMap := map[string]string{
		StateFocus:  fmt.Sprintf("%d focus loop passed.", loopCount),
		StateRest:   fmt.Sprintf("%d rest loop passed.", loopCount),
		StateFinish: fmt.Sprintf("All %d focus loops finished. Congrats!", loopCount),
	}
	okLabelMap := map[string]string{
		StateFocus:  "Take a break",
		StateRest:   "Focus",
		StateFinish: "Finish",
	}
	title := "--title=Pomodoro " + messageMap[state]
	text := "--text=Next loop will be (minutes)"
	okLabel := "--ok-label=" + okLabelMap[state] + " ✅"
	stopLabel := "--cancel-label= Off"
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

func handleCmdResult(isSuccess bool, output string, state string, config *Config) bool {
	if isSuccess {
		minutesStr := regexp.MustCompile(`\d+`).FindString(output)
		minutes, _ := strconv.Atoi(minutesStr)
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
func restart(config Config) {
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
