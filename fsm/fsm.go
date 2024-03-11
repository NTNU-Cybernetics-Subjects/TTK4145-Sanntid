package fsm

import (
	"Driver-go/elevio"
	"elevator/config"
	"time"
)

/*
Finite state machine for operating a single elevator.

Inputs from:
	Elevio:
		- FloorSensor
		- ObstructionSwitch
		- StopButton
	Distibutor:
		- DecisionCommand - Not Defined Yet

Outputs to:
	Distibutor:
		State: - Not Finalised
			- Floor
			- Obstruction
			- Direction
			- ServiceQueue
*/

/*
Input:
	- Orders from decider
	- State from elevio or syncronizer?
Output:
	-
*/

var elevator ElevatorState
var numFloors int = config.NumberFloors
var DoorOpenTime float64 = float64(3 * time.Second.Nanoseconds())

func Fsm(
	buttonsChan <-chan elevio.ButtonEvent,
	hallRequestChan <-chan [][2]bool,
	floorSensorChan <-chan int,
	obstructionChan <-chan bool,
	stateUpdateChan chan<- ElevatorState) {

	elevator = InitializeElevator(<-floorSensorChan)
	doorTimerChan := make(chan bool)
	go PollTimer(doorTimerChan)

	select {
	case obstruction := <-obstructionChan:
		if obstruction {
			StopMotor()
			OpenDoor()
		} else {
			StartMotor()
		}

	case buttonPress := <-buttonsChan:
		onButtonPress(buttonPress.Floor, buttonPress.Button)

	case newFloor := <-floorSensorChan:
		elevator.Floor = newFloor

	case doorTimeOut := <-doorTimerChan:

	}
}

func onInitBetweenFloors() {
	elevio.SetMotorDirection(MD_Down)
	elevator.Direction = MD_Down
	elevator.Behavior = EB_Moving
}

func onButtonPress(buttonFloor int, buttonType elevio.ButtonType) {
	switch elevator.Behavior {
	case EB_DoorOpen:
		if shouldClearImmediately(buttonFloor, buttonType) {
			StartTimer(DoorOpenTime)
		} else {
			distributeButtonPress(buttonFloor, buttonType)
		}

	default:
		distributeButtonPress(buttonFloor, buttonType)
	}
}

func onNewFloor(floor int) {
	elevator.Floor = floor
	elevio.SetFloorIndicator(elevator.Floor)

}

func distributeButtonPress(buttonFloor int, buttonType elevio.ButtonType) {
	if buttonType == elevio.BT_Cab {
		elevator.Requests[buttonFloor][buttonType] = true
		// TODO: Send updated state
	} else {
		// TODO: Send hall request to syncronizer
	}
}
