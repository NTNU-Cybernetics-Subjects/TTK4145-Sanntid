package fsm

import (
	"Driver-go/elevio"
)

// TODO: Currently only accounts for CabRequests -> Change CabReq to just Req, or find different solution
func RequestsAbove() bool {
	for i := elevator.Floor; i < numFloors; i++ {
		if elevator.CabRequests[i] == true {
			return true
		}
	}
	return false
}

func RequestsBelow() bool {
	for i := 0; i < elevator.Floor; i++ {
		if elevator.CabRequests[i] == true {
			return true
		}
	}
	return false
}

func RequestsHere() bool {
	if elevator.CabRequests[elevator.Floor] == true {
		return true
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
		}
	default:
		return DirectionBehaviorPair{elevio.MD_Stop, EB_Idle}
	}
}
