package fsm

import (
	"Driver-go/elevio"
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
var numFloors int = 4
var DoorOpenTime float64 = float64(3 * time.Second.Nanoseconds())

doorTimerChan := make(chan bool)

func Fsm(
	buttonsChan <-chan elevio.ButtonEvent,
	hallRequestChan <-chan [][2]bool,
	floorSensorChan <-chan int,
	obstructionChan <-chan bool,
	stateUpdateChan chan<- ElevatorState) {

	elevator = InitializeElevator(<-floorSensorChan)
	go PollTimer(doorTimerChan)

	select {
	case obstruction := <-obstructionChan:
		if obstruction {
			StopMotor()
			OpenDoor()
		} else {
			StartMotor()
		}
	
	case buttonPress := <-buttonsChan:
		if buttonPress.Button == elevio.BT_Cab {
			elevator.CabRequests[buttonPress.Floor] = true
		} else {
			// TODO: Send button event to syncronizer
		}

	case newFloor := <-floorSensorChan:
		elevator.Floor = newFloor

	case doorTimeOut := <-doorTimerChan:
		
	}
}
