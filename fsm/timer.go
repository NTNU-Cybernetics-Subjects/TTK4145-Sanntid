package fsm

import "time"

var endTime float64
var timerActive bool

func GetWallTime() float64 {
	return float64(time.Now().UnixNano()) / float64(time.Second.Nanoseconds())
}

func StartTimer(timerDuration float64) {
	timerActive = true
	endTime = GetWallTime() + timerDuration
}

func StopTimer() {
	timerActive = false
}

func TimerTimedOut() bool {
	return timerActive && float64(time.Now().UnixNano()) > endTime
}

// Maybe move this to another location?
func PollTimer(receiver chan<- bool) {
	for {
		time.Sleep(20 * time.Millisecond) // Maybe use configurable constant?
		if TimerTimedOut() {
			receiver <- true
		}
	}
}
