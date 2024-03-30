package main

import "testing"

func TestPomodoroSuccess(t *testing.T) {
	tm := &testMessenger{responses: []response{{
		off:      false,
		reset:    false,
		nextLoop: 1,
	}}}
	p := pomodoro{
		conf:      conf(),
		messenger: tm,
	}
	p.run()

	expected := p.conf.maxLoop*2 + 1
	if tm.sendCount != expected {
		t.Errorf("got %d, expected %d", tm.sendCount, expected)
	}
}

func TestPomodoroOff(t *testing.T) {
	tm := &testMessenger{responses: []response{{
		off:   true,
		reset: false,
	}}}
	p := pomodoro{
		conf:      conf(),
		messenger: tm,
	}
	p.run()

	expected := 1
	if tm.sendCount != expected {
		t.Errorf("got %d, expected %d", tm.sendCount, expected)
	}
}

//func TestPomodoroReset(t *testing.T) {
//	tm := &testMessenger{responses: []response{{
//		off:   false,
//		reset: true,
//	}}}
//	p := pomodoro{
//		conf:      conf(),
//		messenger: tm,
//	}
//	p.run()
//
//	expected := 1
//	if tm.sendCount != expected {
//		t.Errorf("got %d, expected %d", tm.sendCount, expected)
//	}
//}

func conf() config {
	return config{1, 1, 3}
}

type testMessenger struct {
	sendCount int
	responses []response
}

func (tm *testMessenger) send(state string, loopCount int, conf config) response {
	tm.sendCount++

	return tm.responses[0]
}
