package fsm

import (
	"Driver-go/elevio"
	"elevator/config"
	"log/slog"
	"time"
)

var elevator ElevatorState
var numFloors int = config.NumberFloors
var DoorOpenTime int64 = config.DoorOpenTimeMs

func Fsm(
	buttonEventOutputChan chan<- elevio.ButtonEvent,
	stateOutputChan 	  chan<- ElevatorState,
	ordersUpdateChan 	  <-chan [config.NumberFloors][3]bool) {
	
	slog.Info("Starting FSM, begin initializing of channels and elevator")

	buttonsChan 	:= make(chan elevio.ButtonEvent)
	floorSensorChan := make(chan int)
	obstructionChan := make(chan bool)
	doorTimerChan := make(chan bool)
	
	go PollTimer(doorTimerChan)
	go lightsHandler()
	go elevio.PollButtons(buttonsChan)
	go elevio.PollFloorSensor(floorSensorChan)
	go elevio.PollObstructionSwitch(obstructionChan)
	slog.Info("Channels initialized")

	elevator = InitializeElevator()
	slog.Info("Elevator initialized")
	if elevator.Floor == -1 {
		onInitBetweenFloors()
	}

	slog.Info("Initialization complete, starting case handling")
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
		time.Sleep(100 * time.Millisecond)
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
		}
	default:
		sendToSyncChan <- buttonPress
	}
}

func onNewFloor(floor int) {
	elevator.Floor = floor
	elevio.SetFloorIndicator(elevator.Floor)
	switch elevator.Behavior {
	case EB_Moving:
		if ShouldStop() {
			ClearRequestAtCurrentFloor()
			StopMotor()
			OpenDoor()
		}
	}
}

func onDoorTimeout() {
	slog.Info("On Door Timeout", "obstructed", elevator.Obstructed)

	if elevator.Obstructed {
		StartTimer(config.DoorOpenTimeMs)
		return
	}

	timerActive = false
	switch elevator.Behavior {
	case EB_DoorOpen:
		directionBehavior := DecideMotorDirection()
		elevator.Behavior = directionBehavior.Behavior
		elevator.Direction = directionBehavior.Direction
		slog.Info("Door open, new directionBehavior", "b", elevator.Behavior, "d", elevator.Direction)
		switch elevator.Behavior {
		case EB_DoorOpen:
			slog.Info("Door open again")
			OpenDoor()
			ClearRequestAtCurrentFloor()
		default:
			slog.Info("Close door and move")
			CloseDoor()
			StartMotor()
		}
	default:
		return
	}
}

func onObstruction(obstruction bool) {
	if obstruction {
		StopMotor()
		OpenDoor()
		elevator.Obstructed = true
	} else {
		elevator.Obstructed = false
		if elevator.Floor == -1 {
			onInitBetweenFloors()
		}
		//StartMotor()
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
