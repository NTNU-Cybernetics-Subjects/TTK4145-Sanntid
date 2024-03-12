package fsm

import (
	"time"
)

var endTime int64
var timerActive bool

func StartTimer(dur int64) {
	endTime = time.Now().UnixMilli() + dur
	timerActive = true
}

func TimerTimedOut() bool {
	if timerActive {
		if time.Now().UnixMilli() > endTime {
			return true
		}
	}
	return false
}

func PollTimer(receiver chan<- bool) {
	for {
		time.Sleep(20 * time.Millisecond)
		if TimerTimedOut() {
			receiver <- true
		}
	}
}
