package distributor

import (
	// "Driver-go/elevio"
	"Driver-go/elevio"
	"elevator/config"
	"elevator/fsm"

	// "elevator/fsm"
	"fmt"
	// "log/slog"
)

// import "testing"

//	func generateRandomState() fsm.ElevatorState {
//	    ranDir := rand.Intn(2)
//	    if ranDir == 2{
//	        ranDir = -1
//	    }
//	    state := fsm.ElevatorState{
//	        Behavior: rand.Intn(2),
//	        Floor: rand.Intn(config.NumberFloors),
//	        Direction: ranDir,
//	        Requests: [][3]bool,
//	    }
//	}
func MakeTestStateMessages() []string {

	elevatorID := []string{"first", "second"}

	state1 := fsm.ElevatorState{
		Behavior:   fsm.EB_Moving,
		Floor:      2,
		Direction:  elevio.MD_Up,
		Orders:     [][3]bool{{false, false, false}, {false, false, false}, {false, false, false}, {false, false, false}},
		Obstructed: false,
	}
	state2 := fsm.ElevatorState{
		Behavior:   fsm.EB_Idle,
		Floor:      1,
		Direction:  elevio.MD_Stop,
		Orders:     [][3]bool{{false, false, false}, {false, false, false}, {false, false, false}, {false, false, false}},
		Obstructed: false,
	}

	hallRequests := [config.NumberFloors][2]bool{{false, false}, {false, true}, {false, false}, {true, false}}

	stateMessage1 := StateMessageBroadcast{
		Id:           elevatorID[0],
		Checksum:     nil,
		State:        state1,
		Sequence:     0,
		HallRequests: hallRequests,
	}

	stateMessage2 := StateMessageBroadcast{
		Id:           elevatorID[1],
		Checksum:     nil,
		State:        state2,
		Sequence:     0,
		HallRequests: hallRequests,
	}
	storeStateMessage(elevatorID[0], stateMessage1)
	storeStateMessage(elevatorID[1], stateMessage2)
	// fmt.Println("construting test stateMessages")
	// fmt.Println("first", getLocalElevatorStateMessage(elevatorID[0]))
	// fmt.Println("second", getLocalElevatorStateMessage(elevatorID[1]))
	updateHallRequests(hallRequests)
	setActivePeers(elevatorID)
	return elevatorID
}

func TestDistribution() {

	currentActivePeers := MakeTestStateMessages()
	fmt.Println("[distributor]: Got distribute signal", "activePeers", currentActivePeers)

	HRAInput := ConstructHRAInput(currentActivePeers)
	// fmt.Println("[distributor]: HRA input succsefully created")

	output := CalulateOrders(HRAInput)
	fmt.Println("[distribitor]: HRA caluclated", "HRA_output", output)

	// currentHallRequests = [4][2]bool(output[mainID])
	// fmt.Println("[distribitor]: our elevators", "hallRequests", currentHallRequests)
}

func CollectDistributorOutput(newHallOrders <-chan [config.NumberFloors][2]bool) {

	for HO := range newHallOrders {
		fmt.Println("[distributorCollector]: got hallrequest update", HO)
	}
}
