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

func DecideMotorDirection() elevio.MotorDirection {
	switch elevator.Direction {
	case elevio.MD_Stop:
		if RequestsAbove() {
			return elevio.MD_Up
		}
		if RequestsBelow() {
			return elevio.MD_Down
		}

	case elevio.MD_Up:

	case elevio.MD_Down:

	}
}
