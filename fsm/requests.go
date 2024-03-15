package fsm

import (
	"Driver-go/elevio"
	"fmt"
)

func RequestsAbove() bool {
	for i := elevator.Floor; i < numFloors; i++ {
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

// TODO: Unsure if I should use DirectionBehaviorPair or just direction.
func DecideMotorDirection() DirectionBehaviorPair {
	fmt.Print("DecideMotorDirection: ")
	switch elevator.Direction {
	case elevio.MD_Stop:
		fmt.Print("Stop: ")
		if RequestsHere() {
			fmt.Println("Req here")
			return DirectionBehaviorPair{elevio.MD_Stop, EB_DoorOpen}
		}
		if RequestsBelow() {
			fmt.Println("Req below")
			return DirectionBehaviorPair{elevio.MD_Down, EB_Moving}
		}
		if RequestsAbove() {
			fmt.Println("Req above")
			return DirectionBehaviorPair{elevio.MD_Up, EB_Moving}
		} else {
			fmt.Println("Else statement")
			return DirectionBehaviorPair{elevio.MD_Stop, EB_Idle}
		}

	case elevio.MD_Up:
		fmt.Print("Up: ")
		if RequestsAbove() {
			fmt.Println("Req above")
			return DirectionBehaviorPair{elevio.MD_Up, EB_Moving}
		}
		if RequestsHere() {
			fmt.Println("Req here")
			return DirectionBehaviorPair{elevio.MD_Stop, EB_DoorOpen}
		}
		if RequestsBelow() {
			fmt.Println("Req below")
			return DirectionBehaviorPair{elevio.MD_Down, EB_Moving}
		} else {
			fmt.Println("Else statement")
			return DirectionBehaviorPair{elevio.MD_Stop, EB_Idle}
		}

	case elevio.MD_Down:
		fmt.Print("Down: ")
		if RequestsBelow() {
			fmt.Println("Req below")
			return DirectionBehaviorPair{elevio.MD_Down, EB_Moving}
		}
		if RequestsHere() {
			fmt.Println("Req here")
			return DirectionBehaviorPair{elevio.MD_Stop, EB_DoorOpen}
		}
		if RequestsAbove() {
			fmt.Println("Req above")
			return DirectionBehaviorPair{elevio.MD_Up, EB_Moving}
		} else {
			fmt.Println("Else statement")
			return DirectionBehaviorPair{elevio.MD_Stop, EB_Idle}
		}
	default:
		fmt.Println("Default statement")
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
		return elevator.Orders[elevator.Floor][elevio.BT_HallDown] ||
			elevator.Orders[elevator.Floor][elevio.BT_Cab] ||
			!RequestsBelow()
	case elevio.MD_Up:
		return elevator.Orders[elevator.Floor][elevio.BT_HallUp] ||
			elevator.Orders[elevator.Floor][elevio.BT_Cab] ||
			!RequestsAbove()
	default:
		return true
	}
}

func ClearRequestAtCurrentFloor() {
	fmt.Println("Clearing Requests")
	// fmt.Println(elevator.Requests)
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
	// fmt.Println(elevator.Requests)
}
