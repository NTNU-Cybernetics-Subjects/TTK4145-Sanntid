package elevator

import (
	"Driver-go/elevio"
)

type ElevatorBehavior struct {
	idle     bool
	moving   bool
	doorOpen bool
}

type Elevator struct {
	behavior    ElevatorBehavior
	floor       int
	direction   elevio.MotorDirection
	cabRequests []bool
}
