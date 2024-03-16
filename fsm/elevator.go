package fsm

import (
	"Driver-go/elevio"
	"elevator/config"
)

type ElevatorBehavior int

const (
	EB_Idle ElevatorBehavior = iota
	EB_Moving
	EB_DoorOpen
)

// Elevator state
type ElevatorState struct {
	Behavior     ElevatorBehavior
	Floor        int
	Direction    elevio.MotorDirection
	ButtonsState [config.NumberFloors][3]bool // Used for lights
	Orders       [config.NumberFloors][3]bool // Orders to service
	Obstructed   bool
}

type DirectionBehaviorPair struct {
	Direction elevio.MotorDirection
	Behavior  ElevatorBehavior
}

func InitializeElevator() ElevatorState {
	var empty_state [config.NumberFloors][3]bool
	for i := 0; i < config.NumberFloors; i++ {
		for j := 0; j < 3; j++ {
			empty_state[i][j] = false
		}
	}
	currentFloor := elevio.GetFloor()
	return ElevatorState{EB_Idle, currentFloor, elevio.MD_Stop, empty_state, empty_state, false}
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
	for i := 0; i < config.NumberFloors; i++ {
		for j := 0; j < 3; j++ {
			elevio.SetButtonLamp(elevio.ButtonType(j), i, elevator.Orders[i][j])
		}
	}
}
