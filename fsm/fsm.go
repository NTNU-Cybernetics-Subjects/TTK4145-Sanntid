package fsm

import (
	"Driver-go/elevio"
	"elevator/config"
	"log/slog"
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
var DoorOpenTime int64 = config.DoorOpenTimeMs

func Fsm(
	buttonEventOutputChan chan<- elevio.ButtonEvent,
	stateOutputChan chan<- ElevatorState,
	obstructionChan <-chan bool,
	buttonsChan <-chan elevio.ButtonEvent,
	floorSensorChan <-chan int,
	doorTimerChan <-chan bool,
	ordersUpdateChan <-chan [config.NumberFloors][3]bool) {

	elevator = InitializeElevator()
	if elevator.Floor == -1 {
		onInitBetweenFloors()
	}
	go lightsHandler()

	for {
		// fmt.Println()
		// slog.Info("Overview State", "", elevator)
		select {
		case obstruction := <-obstructionChan:
			slog.Info("FSM CASE: Obstruction", "value", obstruction)
			onObstruction(obstruction)

		case buttonPress := <-buttonsChan:
			slog.Info("FSM CASE: Button Press", "floor", buttonPress.Floor, "button", buttonPress.Button)
			onButtonPress(buttonPress, buttonEventOutputChan)

		case newFloor := <-floorSensorChan:
			slog.Info("FSM CASE: New Floor", "floor", newFloor)
			onNewFloor(newFloor)

		case <-doorTimerChan:
			if timerActive {
				slog.Info("FSM CASE: Door Timeout")
				onDoorTimeout()
			}

		case ordersUpdate := <-ordersUpdateChan:
			slog.Info("FSM CASE: New Orders\n", "orders", ordersUpdate)
			onOrdersUpdate(ordersUpdate)
		}
	}
}

func lightsHandler() {
	for {
		time.Sleep(20 * time.Millisecond)
		for i := 0; i < config.NumberFloors; i++ {
			for j := 0; j < 3; j++ {
				elevio.SetButtonLamp(elevio.ButtonType(j), i, elevator.Orders[i][j])
			}
		}
	}
}

func onInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.Direction = elevio.MD_Down
	elevator.Behavior = EB_Moving
}

func onButtonPress(buttonPress elevio.ButtonEvent, sendToSyncChan chan<- elevio.ButtonEvent) {
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
	elevator.Floor = floor
	elevio.SetFloorIndicator(elevator.Floor)
	switch elevator.Behavior {
	case EB_Moving:
		if ShouldStop() {
			StopMotor()
			OpenDoor()
			ClearRequestAtCurrentFloor()
			UpdateLights() // TODO: Should not be here, should send "Request to clear"
		}
	}
}

func onDoorTimeout() {
	timerActive = false
	switch elevator.Behavior {
	case EB_DoorOpen:
		directionBehavior := DecideMotorDirection()
		elevator.Behavior = directionBehavior.Behavior
		elevator.Direction = directionBehavior.Direction
		switch elevator.Behavior {
		case EB_DoorOpen:
			OpenDoor()
			ClearRequestAtCurrentFloor()
		default:
			CloseDoor()
			StartMotor()
		}
	default:
		return
	}
}

// TODO: Required for moving the elevator, needs to be turned on and off again.
func onObstruction(obstruction bool) {
	if obstruction {
		StopMotor()
		OpenDoor()
		elevator.Obstructed = true
	} else {
		elevator.Obstructed = false
		StartMotor()
	}
}

func onOrdersUpdate(orders [config.NumberFloors][3]bool) {
	for i := 0; i < config.NumberFloors; i++ {
		for j := 0; j < 3; j++ {
			elevator.Orders[i][j] = orders[i][j]
		}
	}
	StartMotor() // TODO: This is only here temporary
}
