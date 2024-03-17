package fsm

import (
	"Driver-go/elevio"
	"elevator/config"
)

func RequestsAbove() bool {
	for i := elevator.Floor + 1; i < config.NumberFloors; i++ {
		for j := 0; j < 3; j++ {
			if elevator.Orders[i][j] {
				return true
			}
		}
	}
	return false
}

func RequestsBelow() bool {
	for i := 0; i < elevator.Floor; i++ {
		for j := 0; j < 3; j++ {
			if elevator.Orders[i][j] {
				return true
			}
		}
	}
	return false
}

func RequestsHere() bool {
	for j := 0; j < 3; j++ {
		if elevator.Orders[elevator.Floor][j] {
			return true
		}
	}
	return false
}

func DecideMotorDirection() DirectionBehaviorPair {
	switch elevator.Direction {

	case elevio.MD_Stop:
		if RequestsBelow() {
			return DirectionBehaviorPair{elevio.MD_Down, EB_Moving}
		}
		if RequestsAbove() {
			return DirectionBehaviorPair{elevio.MD_Up, EB_Moving}
		}
		if RequestsHere() {
			return DirectionBehaviorPair{elevio.MD_Stop, EB_DoorOpen}
		} else {
			return DirectionBehaviorPair{elevio.MD_Stop, EB_Idle}
		}
	case elevio.MD_Up:
		if RequestsAbove() {
			return DirectionBehaviorPair{elevio.MD_Up, EB_Moving}
		}
		if RequestsHere() {
			return DirectionBehaviorPair{elevio.MD_Stop, EB_DoorOpen}
		}
		if RequestsBelow() {
			return DirectionBehaviorPair{elevio.MD_Stop, EB_DoorOpen}
		} else {
			return DirectionBehaviorPair{elevio.MD_Stop, EB_Idle}
		}

	case elevio.MD_Down:
		if RequestsBelow() {
			return DirectionBehaviorPair{elevio.MD_Down, EB_Moving}
		}
		if RequestsHere() {
			return DirectionBehaviorPair{elevio.MD_Stop, EB_DoorOpen}
		}
		if RequestsAbove() {
			return DirectionBehaviorPair{elevio.MD_Stop, EB_DoorOpen}
		} else {
			return DirectionBehaviorPair{elevio.MD_Stop, EB_Idle}
		}
	default:
		return DirectionBehaviorPair{elevio.MD_Stop, EB_Idle}
	}
}

func ShouldClearImmediately(buttonFloor int, buttonType elevio.ButtonType) bool {
	return elevator.Floor == buttonFloor && (
		(elevator.Direction == elevio.MD_Up && buttonType == elevio.BT_HallUp) ||
		(elevator.Direction == elevio.MD_Down && buttonType == elevio.BT_HallDown) ||
		elevator.Direction == elevio.MD_Stop ||
		buttonType == elevio.BT_Cab)
}

func ShouldStop() bool {
	switch elevator.Direction {
	case elevio.MD_Down:
		return elevator.Floor == 0 ||
			elevator.Orders[elevator.Floor][elevio.BT_HallDown] ||
			elevator.Orders[elevator.Floor][elevio.BT_Cab] ||
			!RequestsBelow()
	case elevio.MD_Up:
		return elevator.Floor == config.NumberFloors-1 ||
			elevator.Orders[elevator.Floor][elevio.BT_HallUp] ||
			elevator.Orders[elevator.Floor][elevio.BT_Cab] ||
			!RequestsAbove()
	default:
		return true
	}
}

func ClearRequestAtCurrentFloor() {
	elevator.Orders[elevator.Floor][elevio.BT_Cab] = false


	switch elevator.Direction {
	case elevio.MD_Up:
		if !RequestsAbove() && !elevator.Orders[elevator.Floor][elevio.BT_HallUp] {
			elevator.Orders[elevator.Floor][elevio.BT_HallDown] = false
		}
		elevator.Orders[elevator.Floor][elevio.BT_HallUp] = false

	case elevio.MD_Down:
		if !RequestsBelow() && !elevator.Orders[elevator.Floor][elevio.BT_HallDown] {
			elevator.Orders[elevator.Floor][elevio.BT_HallUp] = false
		}
		elevator.Orders[elevator.Floor][elevio.BT_HallDown] = false

	default:
		elevator.Orders[elevator.Floor][elevio.BT_HallUp] = false
		elevator.Orders[elevator.Floor][elevio.BT_HallDown] = false
	}
}
