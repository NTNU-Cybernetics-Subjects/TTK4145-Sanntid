package fsm

import (
	"Driver-go/elevio"
)

func RequestsAbove() bool {
	for i := elevator.Floor; i < numFloors; i++ {
		for j := 0; j < 3; j++ {
			if elevator.Requests[i][j] {
				return true
			}
		}
	}
	return false
}

func RequestsBelow() bool {
	for i := 0; i < elevator.Floor; i++ {
		for j := 0; j < 3; j++ {
			if elevator.Requests[i][j] {
				return true
			}
		}
	}
	return false
}

func RequestsHere() bool {
	for j := 0; j < 3; j++ {
		if elevator.Requests[elevator.Floor][j] {
			return true
		}
	}
	return false
}

// TODO: Unsure if I should use DirectionBehaviorPair or just direction.
func DecideMotorDirection() DirectionBehaviorPair {
	switch elevator.Direction {
	case elevio.MD_Stop:
		if RequestsHere() {
			return DirectionBehaviorPair{elevio.MD_Stop, EB_DoorOpen}
		}
		if RequestsBelow() {
			return DirectionBehaviorPair{elevio.MD_Down, EB_Moving}
		}
		if RequestsAbove() {
			return DirectionBehaviorPair{elevio.MD_Up, EB_Moving}
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
			return DirectionBehaviorPair{elevio.MD_Down, EB_Moving}
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
			return DirectionBehaviorPair{elevio.MD_Up, EB_Moving}
		} else {
			return DirectionBehaviorPair{elevio.MD_Stop, EB_Idle}
		}
	default:
		return DirectionBehaviorPair{elevio.MD_Stop, EB_Idle}
	}
}

func shouldClearImmediately(buttonFloor int, buttonType elevio.ButtonType) bool {
	return elevator.Floor == buttonFloor && ((elevator.Direction == elevio.MD_Up && buttonType == elevio.BT_HallUp) ||
		(elevator.Direction == elevio.MD_Down && buttonType == elevio.BT_HallDown) ||
		elevator.Direction == elevio.MD_Stop ||
		buttonType == elevio.BT_Cab)
}

func ShouldStop() bool {
	switch elevator.Direction {
	case elevio.MD_Down:
		return elevator.Requests[elevator.Floor][elevio.BT_HallDown] ||
			elevator.Requests[elevator.Floor][elevio.BT_Cab] ||
			!RequestsBelow()
	case elevio.MD_Up:
		return elevator.Requests[elevator.Floor][elevio.BT_HallUp] ||
			elevator.Requests[elevator.Floor][elevio.BT_Cab] ||
			!RequestsAbove()
	default:
		return true
	}
}

func ClearRequestAtCurrentFloor() {
	elevator.Requests[elevator.Floor][elevio.BT_Cab] = false
	switch elevator.Direction {
	case elevio.MD_Up:
		if !RequestsAbove() && !elevator.Requests[elevator.Floor][elevio.BT_HallUp] {
			elevator.Requests[elevator.Floor][elevio.BT_HallDown] = false
		}
		elevator.Requests[elevator.Floor][elevio.BT_HallUp] = false

	case elevio.MD_Down:
		if !RequestsBelow() && !elevator.Requests[elevator.Floor][elevio.BT_HallDown] {
			elevator.Requests[elevator.Floor][elevio.BT_HallUp] = false
		}
		elevator.Requests[elevator.Floor][elevio.BT_HallDown] = false

	default:
		elevator.Requests[elevator.Floor][elevio.BT_HallUp] = false
		elevator.Requests[elevator.Floor][elevio.BT_HallDown] = false
	}
}
