package fsm

import (
	"Driver-go/elevio"
	"elevator/config"
	"fmt"
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

func Fsm(buttonEventOutputChan chan<- elevio.ButtonEvent,
	stateOutputChan chan<- ElevatorState,
	obstructionChan <-chan bool,
	buttonsChan <-chan elevio.ButtonEvent,
	floorSensorChan <-chan int,
	doorTimerChan <-chan bool,
	hallRequestChan <-chan [][2]bool) {

	elevator = InitializeElevator()
	if elevator.Floor == -1 {
		onInitBetweenFloors()
	}

	for {
		select {
		case obstruction := <-obstructionChan:
			fmt.Print("FSM CASE: Obstruction = ")
			onObstruction(obstruction)
			fmt.Println(elevator.Obstructed)

		case buttonPress := <-buttonsChan:
			fmt.Println("FSM CASE: Button Press")
			onButtonPress(buttonPress.Floor, buttonPress.Button)

		case newFloor := <-floorSensorChan:
			fmt.Println("FSM CASE: New Floor")
			onNewFloor(newFloor)

		case <-doorTimerChan:
			fmt.Println("FSM CASE: Door Timed Out")
			onDoorTimeout()

		case newHallRequest := <-hallRequestChan:
			fmt.Println("FSM CASE: Hall Request Update")
			onHallRequestUpdate(newHallRequest)
		}
	}
}

func onInitBetweenFloors() {
	fmt.Println("F: onInitBetweenFloors")
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.Direction = elevio.MD_Down
	elevator.Behavior = EB_Moving
}

func onButtonPress(buttonFloor int, buttonType elevio.ButtonType) {
	fmt.Println("F: onButtonPress")
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
	fmt.Println("F: onNewFloor")
	elevator.Floor = floor
	elevio.SetFloorIndicator(elevator.Floor)
	switch elevator.Behavior {
	case EB_Moving:
		if ShouldStop() {
			StopMotor()
			OpenDoor()
			ClearRequestAtCurrentFloor()
		}
	}

}

func distributeButtonPress(buttonFloor int, buttonType elevio.ButtonType) {
	fmt.Println("F: distributeButtonPress")
	if buttonType == elevio.BT_Cab {
		elevator.Requests[buttonFloor][buttonType] = true
		// TODO: Send updated state
	} else {
		// TODO: Send hall request to syncronizer
		return
	}
}

func onDoorTimeout() {
	fmt.Println("F: onDoorTimeout")
	switch elevator.Behavior {
	case EB_DoorOpen:
		StartMotor()
		switch elevator.Behavior {
		case EB_DoorOpen:
			OpenDoor()
		case EB_Idle:
			CloseDoor()
		default:
			return
		}
	default:
		return
	}
}

// TODO: Required for moving the elevator, needs to be turned on and off again.
func onObstruction(obstruction bool) {
	fmt.Println("F: onObstruction")
	if obstruction {
		StopMotor()
		OpenDoor()
		elevator.Obstructed = true
	} else {
		elevator.Obstructed = false
		StartMotor()
	}
}

func onHallRequestUpdate(hallRequest [][2]bool) {
	fmt.Println("F: onHallRequestUpdate")
	for i := 0; i < config.NumberFloors; i++ {
		for j := 0; j < 2; j++ {
			elevator.Requests[i][j] = hallRequest[i][j]
		}
	}
}
