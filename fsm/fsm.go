package fsm

import (
	"Driver-go/elevio"
	"elevator/config"
	"elevator/orders"
	"log/slog"
	"time"
)

var elevator ElevatorState

func Fsm(
	buttonEventOutputChan 	chan<- elevio.ButtonEvent,
	clearOrdersChan			chan<- elevio.ButtonEvent,
	stateOutputChan 		chan<- ElevatorState,
	newOrdersChan 			<-chan [config.NumberFloors][3]bool) {

	slog.Info("\t[FSM SETUP]: Starting FSM, begin initializing of channels and elevator")

	buttonsChan 	:= make(chan elevio.ButtonEvent)
	floorSensorChan := make(chan int)
	obstructionChan := make(chan bool)
	doorTimerChan 	:= make(chan bool)

	go lightsHandler()
	go PollTimer(doorTimerChan)
	go checkClearedOrders(clearOrdersChan)
	go elevio.PollButtons(buttonsChan)
	go elevio.PollFloorSensor(floorSensorChan)
	go elevio.PollObstructionSwitch(obstructionChan)
	slog.Info("\t[FSM SETUP]: Channels initialized")

	elevator = InitializeElevator()
	slog.Info("\t[FSM SETUP]: Elevator initialized")
	if elevator.Floor == -1 {
		onInitBetweenFloors()
	}

	slog.Info("\t[FSM SETUP]: Initialization complete, starting case handling")
	for {
		select {
		case obstruction := <-obstructionChan:
			slog.Info("\t[FSM Case]: Obstruction")
			onObstruction(obstruction)

		case buttonPress := <-buttonsChan:
			slog.Info("\t[FSM Case]: Button Press")
			onButtonPress(buttonPress, buttonEventOutputChan)

		case newFloor := <-floorSensorChan:
			slog.Info("\t[FSM Case]: New Floor", "floor", newFloor)
			onNewFloor(newFloor)

		case <-doorTimerChan:
			slog.Info("\t[FSM Case]: Door Timeout")
			onDoorTimeout()

		case ordersUpdate := <-newOrdersChan:
			slog.Info("\t[FSM Case]: New Orders")
			onOrdersUpdate(ordersUpdate)
			slog.Info("\t[FSM NEW ORDERS]", "order", elevator.Orders)
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
		// This should handle clearing local buttonevents instantly, without passing through
		// the request handler.
		if ShouldClearImmediately(buttonPress.Floor, buttonPress.Button) {
			slog.Info("[FSM ButtonPress]: Door is open, should clear request immediately")
			OpenDoor()
		} else {
			slog.Info("[FSM ButtonPress]: Door is open, send request to sync")
			sendToSyncChan <- buttonPress
		}
	
	case EB_Idle:
		slog.Info("[FSM ButtonPress]: Elevator is Idle, Start motor")
		StartMotor()
		slog.Info("[FSM ButtonPress]: Started motor", "b", elevator.Behavior, "d", elevator.Direction)
		switch elevator.Behavior{
		case EB_DoorOpen:
			slog.Info("[FSM ButtonPress]: Open Door")
			OpenDoor()
		default:
			slog.Info("[FSM ButtonPress]: Send request to sync")
			sendToSyncChan <- buttonPress
		}

	default:
		slog.Info("[FSM ButtonPress]: Default case, send request to sync")
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
	if elevator.Obstructed {
		StartTimer(config.DoorOpenTimeMs)
		return
	}

	timerActive = false
	slog.Info("Door timeout", "b", elevator.Behavior)
	switch elevator.Behavior {
	case EB_DoorOpen:
		directionBehavior := DecideMotorDirection()
		elevator.Behavior = directionBehavior.Behavior
		elevator.Direction = directionBehavior.Direction
		slog.Info("New directionBehavior", "b", elevator.Behavior, "d", elevator.Direction)
		switch elevator.Behavior {
		case EB_DoorOpen:
			slog.Info("Door open")
			OpenDoor()
			ClearRequestAtCurrentFloor()
		default:
			slog.Info("Door should close")
			CloseDoor()
			StartMotor()
		}
	default:
		slog.Info("Door timeout default statement")	// TODO: This might only be a bandage
		CloseDoor()
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
	}
}

func onOrdersUpdate(orders [config.NumberFloors][3]bool) {
	for i := 0; i < config.NumberFloors; i++ {
		for j := 0; j < 3; j++ {
			elevator.Orders[i][j] = orders[i][j]
		}
	}
	slog.Info("\t[FSM ORDERS UPDATE]\n", "orders", elevator.Orders)

	StartMotor()
	slog.Info("\t[FSM ORDERS UPDATE]: Started motor", "b", elevator.Behavior, "d", elevator.Direction)
	
	// Should handle requests at current location.
	if elevator.Behavior == EB_DoorOpen{
		slog.Info("\t[FSM ORDERS UPDATE]: Opening door")
		OpenDoor()
		slog.Info("\t[FSM ORDERS UPDATE]: Attempting to clear request at current floor")
		ClearRequestAtCurrentFloor()
	}
}

func lightsHandler() {
	for {
		time.Sleep(time.Duration(config.LightUpdateTimeMs) * time.Millisecond)
        hallOrders := orders.GetHallOrders()
        cabOrders := orders.GetCabOrders(config.ElevatorId)
		for i := 0; i < config.NumberFloors; i++ {
            elevio.SetButtonLamp(elevio.BT_Cab, i, cabOrders[i])
			for j := 0; j < 2; j++ {
				elevio.SetButtonLamp(elevio.ButtonType(j), i, hallOrders[i][j])
			}
		}
	}
}

func checkClearedOrders(outputChan chan<- elevio.ButtonEvent){
	previousOrders := elevator.Orders
	for {
		time.Sleep(time.Duration(config.CheckClearedOrdersTimeMs) * time.Millisecond)
		currentOrders := elevator.Orders

		for i := 0; i < config.NumberFloors; i++ {
			for j := 0; j < 3; j++ {
				if !currentOrders[i][j] && previousOrders[i][j]{
					outputChan <- elevio.ButtonEvent{Floor:i, Button:elevio.ButtonType(j)}
					slog.Info("\t[FSM BACKGROUND]: Cleared order", "floor", i, "button", j)
				}
			}
		}
		previousOrders = currentOrders
	}
}
