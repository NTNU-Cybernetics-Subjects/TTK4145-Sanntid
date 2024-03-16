package fsm

import (
	"Driver-go/elevio"
	"elevator/config"
	"log/slog"
)

// Elevator state
type ElevatorState struct {
	Behavior   ElevatorBehavior
	Floor      int
	Direction  elevio.MotorDirection
	Orders     [config.NumberFloors][3]bool
	Obstructed bool
}

type DirectionBehaviorPair struct {
	Direction elevio.MotorDirection
	Behavior  ElevatorBehavior
}

type ElevatorBehavior int

const (
	EB_Idle ElevatorBehavior = iota
	EB_Moving
	EB_DoorOpen
)

func InitializeElevator() ElevatorState {
	var empty_orders [config.NumberFloors][3]bool
	for i := 0; i < config.NumberFloors; i++ {
		for j := 0; j < 3; j++ {
			empty_orders[i][j] = false
		}
	}
	currentFloor := elevio.GetFloor()
	return ElevatorState{EB_Idle, currentFloor, elevio.MD_Stop, empty_orders, false}
}

func GetElevatorState() ElevatorState {
	return elevator
}

// Movement

func StopMotor() {
	elevio.SetMotorDirection(elevio.MD_Stop)
}

func StartMotor() {
	if !elevator.Obstructed {
		directionBehavior := DecideMotorDirection()
		elevio.SetMotorDirection(directionBehavior.Direction)
		elevator.Direction = directionBehavior.Direction
		elevator.Behavior = directionBehavior.Behavior
		slog.Info("\t[FSM StartMotor]Starting motor", "b", elevator.Behavior, "d", elevator.Direction)
	}
}

// Door
func OpenDoor() {
	elevator.Behavior = EB_DoorOpen
	elevio.SetDoorOpenLamp(true)
	StartTimer(config.DoorOpenTimeMs)
}

func CloseDoor() {
	slog.Info("\t\t CLOSING DOOR")
	elevio.SetDoorOpenLamp(false)
}
