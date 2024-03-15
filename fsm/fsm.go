package fsm

import (
	"Driver-go/elevio"
	"elevator/config"
	"fmt"
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
var DoorOpenTime int64 = config.DoorOpenTimeMs

func Fsm(buttonEventOutputChan chan<- elevio.ButtonEvent,
	stateOutputChan chan<- ElevatorState,
	obstructionChan <-chan bool,
	buttonsChan <-chan elevio.ButtonEvent,
	floorSensorChan <-chan int,
	doorTimerChan <-chan bool,
	ordersUpdateChan <-chan [config.NumberFloors][2]bool,
	stateUpdateChan <-chan [config.NumberFloors][3]bool) {

	elevator = InitializeElevator()
	if elevator.Floor == -1 {
		onInitBetweenFloors()
	}

	for {
		select {
		case obstruction := <-obstructionChan:
			fmt.Println()
			fmt.Print("FSM CASE: Obstruction = ")
			onObstruction(obstruction)

		case buttonPress := <-buttonsChan:
			fmt.Println()
			fmt.Println("FSM CASE: Button Press")
			onButtonPress(buttonPress, buttonEventOutputChan)

		case newFloor := <-floorSensorChan:
			fmt.Println()
			fmt.Println("FSM CASE: New Floor")
			onNewFloor(newFloor)
			// fmt.Println(elevator.Orders)

		case <-doorTimerChan:
			if timerActive {
				fmt.Println()
				fmt.Println("FSM CASE: Door Timeout")
				onDoorTimeout()
			}

		case HallOrdersUpdate := <-ordersUpdateChan:
			// Receive and handle orders given by HRA
			fmt.Println()
			fmt.Println("FSM CASE: New Hall Orders")
			onOrdersUpdate(HallOrdersUpdate)

		case validState := <-stateUpdateChan:
			// Receive the valid state for hall and cab
			// buttons as validated by the network module
			// This also validates the Cab calls.
			fmt.Println()
			fmt.Println("FSM CASE: New State Update")
			onStateUpdate(validState)
			UpdateCabOrders(validState)
			UpdateLights()

		}
	}
}

func onInitBetweenFloors() {
	fmt.Println("F: onInitBetweenFloors")
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.Direction = elevio.MD_Down
	elevator.Behavior = EB_Moving
}

func onButtonPress(buttonPress elevio.ButtonEvent, sendToSyncChan chan<- elevio.ButtonEvent) {
	fmt.Println("F: onButtonPress")
	switch elevator.Behavior {
	case EB_DoorOpen:
		if shouldClearImmediately(buttonPress.Floor, buttonPress.Button) {
			StartTimer(DoorOpenTime)
		} else {
			sendToSyncChan <- buttonPress
			// if buttonPress.Button == elevio.BT_Cab {
			// 	elevator.Orders[buttonPress.Floor][buttonPress.Button] = true
			// 	StartMotor() // TODO: This is only here temporary
			// }
		}
	default:
		sendToSyncChan <- buttonPress
		// elevator.Orders[buttonPress.Floor][buttonPress.Button] = true
		// StartMotor() // TODO: This is only here temporary
	}
}

func onNewFloor(floor int) {
	fmt.Println("F: onNewFloor")
	elevator.Floor = floor
	fmt.Print("Current floor is: ")
	fmt.Println(elevator.Floor)
	elevio.SetFloorIndicator(elevator.Floor)
	switch elevator.Behavior {
	case EB_Moving:
		if ShouldStop() {
			fmt.Println("Testing")
			StopMotor()
			OpenDoor()
			ClearRequestAtCurrentFloor()
			UpdateLights() // TODO: Should not be here, should send "Request to clear"
		}
	}
}

func onDoorTimeout() {
	fmt.Println("F: onDoorTimeout")
	timerActive = false
	switch elevator.Behavior {
	case EB_DoorOpen:
		directionBehavior := DecideMotorDirection()
		elevator.Behavior = directionBehavior.Behavior
		elevator.Direction = directionBehavior.Direction
		fmt.Print("New directionBehaviorPair: ")
		fmt.Println(directionBehavior)
		switch elevator.Behavior {
		case EB_DoorOpen:
			fmt.Println("OpenDoor again")
			OpenDoor()
			ClearRequestAtCurrentFloor()
		default:
			fmt.Println("Default, close door and start motor")
			CloseDoor()
			StartMotor()
		}
	default:
		return
	}
}

// TODO: Required for moving the elevator, needs to be turned on and off again.
func onObstruction(obstruction bool) {
	fmt.Print("F: onObstruction ->")
	if obstruction {
		fmt.Println("True")
		StopMotor()
		OpenDoor()
		elevator.Obstructed = true
	} else {
		fmt.Println("False")
		elevator.Obstructed = false
		StartMotor()
	}
}

func onOrdersUpdate(orders [config.NumberFloors][2]bool) {
	fmt.Println("F: onOrdersUpdate")
	for i := 0; i < config.NumberFloors; i++ {
		for j := 0; j < 2; j++ {
			elevator.Orders[i][j] = orders[i][j]
		}
	}
	StartMotor() // TODO: This is only here temporary
}

func onStateUpdate(state [config.NumberFloors][3]bool) {
	fmt.Println("F: onStateUpdate")
	for i := 0; i < config.NumberFloors; i++ {
		for j := 0; j < 3; j++ {
			elevator.ButtonsState[i][j] = state[i][j]
		}
	}
}

func UpdateCabOrders(state [config.NumberFloors][3]bool) {
	fmt.Println("Update Cab orders")
	fmt.Println(state)
	for i := 0; i < config.NumberFloors; i++ {
		elevator.Orders[i][2] = state[i][2]
	}
	fmt.Println(elevator.Orders)
	StartMotor()
}
