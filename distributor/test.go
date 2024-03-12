package distributor

import (
	"Driver-go/elevio"
	"elevator/config"
	"elevator/fsm"
	"fmt"
)

// import "testing"


func MakeTestElevators() ([]string){

    activePeers := [2]string{"first", "second"}
    addElevator(activePeers[0])
    addElevator(activePeers[1])
    
    firstElevator := ElevatorState{
        Behavior: fsm.EB_Moving,
        Floor: 2,
        Direction: elevio.MD_Up,
        CabRequests: []bool{true, true, true, true},
        Obstructed: false,
    }
    secondElevator := ElevatorState{
        Behavior: fsm.EB_DoorOpen,
        Floor: 2,
        Direction: elevio.MD_Stop,
        CabRequests: []bool{false,false,false,false},
    }

    var testHallRequests [config.NumberFloors][2]bool
    // set two hall orders
    testHallRequests[1][0] = true
    testHallRequests[config.NumberFloors-1][1] = true
    
    fmt.Println("updating first elevator to: ", firstElevator)
    updateElevator(activePeers[0], firstElevator)
    fmt.Println("updating second elevator to: ", secondElevator)
    updateElevator(activePeers[1], secondElevator)

    updateHallRequests([config.NumberFloors][2]bool(testHallRequests))
    return activePeers[:]
}


func CollectDistributorOutput(newHallOrders <-chan [config.NumberFloors][2]bool){

    for HO := range newHallOrders {
        fmt.Println("[distributorCollector]: got hallrequest update", HO)
    }
}
