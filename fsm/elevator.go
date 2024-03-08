package elevator

import (
	"Driver-go/elevio"
)

// States
type ElevatorBehavior struct {
	idle     bool
	moving   bool
	doorOpen bool
}

// Elevator object
type Elevator struct {
	behavior    ElevatorBehavior
	floor       int
	direction   elevio.MotorDirection
	cabRequests []bool
}
