package fsm

import (
	"Driver-go/elevio"
	"elevator/config"
	"fmt"
)

type ElevatorBehavior int

const (
	EB_Idle ElevatorBehavior = iota
	EB_Moving
	EB_DoorOpen
)

// Elevator state
type ElevatorState struct {
	Behavior   ElevatorBehavior
	Floor      int
	Direction  elevio.MotorDirection
	Requests   [][3]bool
	Obstructed bool
}

type DirectionBehaviorPair struct {
	Direction elevio.MotorDirection
	Behavior  ElevatorBehavior
}

func InitializeElevator() ElevatorState {
	req := make([][3]bool, numFloors)
	for i := 0; i < numFloors; i++ {
		for j := 0; j < 3; j++ {
			req[i][j] = false
		}
	}
	currentFloor := elevio.GetFloor()
	return ElevatorState{EB_Idle, currentFloor, elevio.MD_Stop, req, false}
}

// Movement
var directionBehavior DirectionBehaviorPair

func StopMotor() {
	elevator.Direction = elevio.MD_Stop
	elevio.SetMotorDirection(elevator.Direction)
}

func StartMotor() {
	if !elevator.Obstructed {
		directionBehavior = DecideMotorDirection()
		elevio.SetMotorDirection(directionBehavior.Direction)
		elevator.Direction = directionBehavior.Direction
		elevator.Behavior = directionBehavior.Behavior
	}
}

// Door
func OpenDoor() {
	elevator.Behavior = EB_DoorOpen
	elevio.SetDoorOpenLamp(true)
	StartTimer(DoorOpenTime)
}

func CloseDoor() {
	elevio.SetDoorOpenLamp(false)
}

func GetElevatorState() ElevatorState {
	return elevator
}

func UpdateLights() {
	fmt.Println("Updating Lights")
	for i := 0; i < config.NumberFloors; i++ {
		for j := 0; j < 3; j++ {
			elevio.SetButtonLamp(elevio.ButtonType(j), i, elevator.Requests[i][j])
		}
	}
}
