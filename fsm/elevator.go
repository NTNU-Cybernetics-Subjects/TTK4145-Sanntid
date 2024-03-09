package fsm

import (
	"Driver-go/elevio"
)

// Behaviors
//
//	type ElevatorBehavior struct {
//		Idle     bool
//		Moving   bool
//		DoorOpen bool
//	}
const (
	EB_Idle     = 0
	EB_Moving   = 1
	EB_DoorOpen = 2
)

// Elevator state
type ElevatorState struct {
	Behavior    ElevatorBehavior
	Floor       int
	Direction   elevio.MotorDirection
	CabRequests []bool
}

func InitializeElevator(floor int) ElevatorState {
	return ElevatorState{EB_Idle, floor, elevio.MD_Stop} // TODO: fiks
}

func StopMotor() {
	elevio.SetMotorDirection(elevio.MD_Stop)

}
