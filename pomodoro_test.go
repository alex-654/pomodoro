package pomodoro

import (
	"testing"
)

func TestPomodoro(t *testing.T) {
	pomodoroData := []struct {
		m    spyMessenger
		want int
	}{
		{
			spyMessenger{responses: success()},
			7},
		{
			spyMessenger{responses: off()},
			1,
		},
		{
			spyMessenger{responses: reset()},
			2,
		},
	}

	for _, test := range pomodoroData {
		p := pomodoro{
			conf:      conf(),
			messenger: &test.m,
		}
		p.run()

		if test.m.sendCount != test.want {
			t.Errorf("got %d, expected %d", test.m.sendCount, test.want)
		}
	}

}

func success() []response {
	return []response{{
		off:      false,
		reset:    false,
		nextLoop: 1,
	}}
}

func off() []response {
	return []response{
		{
			off:   true,
			reset: false,
		},
	}
}

func reset() []response {
	return []response{
		{
			off:   true,
			reset: false,
		},
		{
			off:   false,
			reset: true,
		},
	}
}

func conf() config {
	return config{1, 1, 3}
}

type spyMessenger struct {
	sendCount int
	responses []response
}

func (m *spyMessenger) send(state string, loopCount int, conf config) response {
	m.sendCount++

	length := len(m.responses)
	res := m.responses[length-1]
	if length > 1 {
		m.responses = m.responses[:length-1]
	}

	return res
}
