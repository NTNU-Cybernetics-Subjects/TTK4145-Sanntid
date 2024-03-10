package fsm

import "Driver-go/elevio"

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

func Fsm(
	buttonsChan <-chan elevio.ButtonEvent,
	hallRequestChan <-chan [][2]bool,
	floorSensorChan <-chan int,
	obstructionChan <-chan bool,
	stateUpdateChan chan<- ElevatorState) {

	elevator = InitializeElevator(<-floorSensorChan)

	select {
	case obstruction := <-obstructionChan:
		if obstruction {
			StopMotor()
			OpenDoor()
		} else {
			StartMotor()
		}
	}
}
