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
	hallRequestChan <-chan [config.NumberFloors][2]bool) {

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
			buttonEventOutputChan <- buttonPress
			onButtonPress(buttonPress)
			UpdateLights()

		case newFloor := <-floorSensorChan:
			fmt.Println()
			fmt.Println("FSM CASE: New Floor")
			onNewFloor(newFloor)

		case test := <-doorTimerChan:
			if timerActive {
				fmt.Println()
				fmt.Print("FSM CASE: Door Timed Out - ")
				fmt.Println(test)
				fmt.Println(elevator.Behavior)
				onDoorTimeout()
				fmt.Println(elevator.Behavior)
				UpdateLights()
			}

		case newHallRequest := <-hallRequestChan:
			fmt.Println()
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

func onButtonPress(buttonPress elevio.ButtonEvent) {
	fmt.Println("F: onButtonPress")
	switch elevator.Behavior {
	case EB_DoorOpen:
		if shouldClearImmediately(buttonPress.Floor, buttonPress.Button) {
			StartTimer(DoorOpenTime)
		} else {
			if buttonPress.Button == elevio.BT_Cab {
				elevator.Orders[buttonPress.Floor][buttonPress.Button] = true
				StartMotor() // TODO: This is only here temporary
			}
		}
	default:
		// if buttonPress.Button == elevio.BT_Cab {
		// 	elevator.Requests[buttonPress.Floor][buttonPress.Button] = true
		// 	StartMotor() // TODO: This is only here temporary
		// }
		elevator.Orders[buttonPress.Floor][buttonPress.Button] = true
		StartMotor() // TODO: This is only here temporary
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

func onHallRequestUpdate(hallRequest [config.NumberFloors][2]bool) {
	fmt.Println("F: onHallRequestUpdate")
	for i := 0; i < config.NumberFloors; i++ {
		for j := 0; j < 2; j++ {
			elevator.Orders[i][j] = hallRequest[i][j]
		}
	}
	StartMotor() // TODO: This is only here temporary
}
