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
	Obstructed  bool
}

type DirectionBehaviorPair struct {
	Direction elevio.MotorDirection
	Behavior  ElevatorBehavior
}

func InitializeElevator(currentFloor int) ElevatorState {
	cabReq := make([]bool, numFloors)
	for i := 0; i < numFloors; i++ {
		cabReq[i] = false
	}

	return ElevatorState{EB_Idle, currentFloor, elevio.MD_Stop, cabReq, false}
}

// Movement
var directionBehavior DirectionBehaviorPair

func StopMotor() {
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevator.Direction = elevio.MD_Stop
}

func StartMotor() {
	directionBehavior = DecideMotorDirection()
	elevio.SetMotorDirection(directionBehavior.Direction)
	elevator.Direction = directionBehavior.Direction
	elevator.Behavior = directionBehavior.Behavior
}

// Door
func OpenDoor() {
	elevator.Behavior = EB_DoorOpen
	elevio.SetDoorOpenLamp(true)
	StartTimer(DoorOpenTime)
}
