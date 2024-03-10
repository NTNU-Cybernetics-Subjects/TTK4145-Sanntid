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

type ElevatorBehavior int

const (
	EB_Idle ElevatorBehavior = iota
	EB_Moving
	EB_DoorOpen
)

// Elevator state
type ElevatorState struct {
	Behavior    ElevatorBehavior
	Floor       int
	Direction   elevio.MotorDirection
	CabRequests []bool
}

func InitializeElevator(currentFloor int) ElevatorState {
	cabReq := make([]bool, numFloors)
	for i := 0; i < numFloors; i++ {
		cabReq[i] = false
	}

	return ElevatorState{EB_Idle, currentFloor, elevio.MD_Stop, cabReq}
}

// Movement
var dirn elevio.MotorDirection

func StopMotor() {
	dirn = elevio.MD_Stop
	elevio.SetMotorDirection(dirn)
	elevator.Direction = dirn
}

func StartMotor() {
	dirn = DecideMotorDirection()
	elevio.SetMotorDirection(dirn)
	elevator.Direction = dirn
	elevator.Behavior = EB_Moving
}

// Door
func OpenDoor() {

}
