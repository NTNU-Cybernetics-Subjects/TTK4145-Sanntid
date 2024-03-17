package peerNetwork

import (
    "elevator/orders"
	"Driver-go/elevio"
	"elevator/config"
	"elevator/fsm"
	"encoding/json"
	"log/slog"
	"os/exec"
)

// Path needs to be relative to the executing script, or use full path
var HRAExecutable string = config.HallRequestAssignerExecutable

type ElevatorStateHRA struct {
	Behavior    string `json:"behaviour"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
	Floor       int    `json:"floor"`
}

type InputHRA struct {
	States       map[string]ElevatorStateHRA  `json:"states"`
	HallRequests [config.NumberFloors][2]bool `json:"hallRequests"`
}

func CalulateOrders(HRAInput InputHRA) map[string][][2]bool {
	jsonBytes, err := json.Marshal(HRAInput)
	if err != nil {
		// fmt.Println("josn.Marshal error: ", err)
		slog.Info("[HRA]: json.Marshal error: ", err)
		return nil
	}

	ret, err := exec.Command(HRAExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		// fmt.Println("exec.Command error: ", err)
		// fmt.Println(string(ret))
		slog.Info("[HRA]: exec.Command error: ", err, slog.String("returned", string(ret)))
		return nil
	}

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		// fmt.Println("json.Unmarshal error: ", err)
		slog.Info("[HRA]: json.Unmarshal error: ", err)
		return nil
	}

	return *output
}

func ConstructHRAState(state fsm.ElevatorState, cabOrders [config.NumberFloors]bool) ElevatorStateHRA {
	HRAState := ElevatorStateHRA{
		Behavior:    "",
		Floor:       state.Floor,
		Direction:   "",
		CabRequests: cabOrders[:],
	}

	switch state.Behavior {
	case fsm.EB_Idle:
		HRAState.Behavior = "idle"
	case fsm.EB_Moving:
		HRAState.Behavior = "moving"
	case fsm.EB_DoorOpen:
		HRAState.Behavior = "doorOpen" // FIXME: is this a valid input in HRA?
	}
	// slog.Info("[HRA constructor]: ", "Behavior", HRAState.Behavior)

	switch state.Direction {
	case elevio.MD_Up:
		HRAState.Direction = "up"
	case elevio.MD_Down:
		HRAState.Direction = "down"
	case elevio.MD_Stop:
		HRAState.Direction = "stop"
	}
	// slog.Info("[HRA constructor]: ", "Direction", HRAState.Direction)
	return HRAState
}

func ConstructHRAInput(activeElevators []string) InputHRA {
	HRAInput := InputHRA{
		States:       make(map[string]ElevatorStateHRA),
		HallRequests: orders.GetHallOrders(),
	}
	for i := range activeElevators {
		stateMessage := getLastStateMessage(activeElevators[i])
		// slog.Info("[HRA]: collected state message", "stateMessage", stateMessage)
		cabOrders := orders.GetCabOrders(stateMessage.Id)
		HRAState := ConstructHRAState(stateMessage.State, cabOrders) // FIXME: not sure if we should use cab from orders, or cab from stateMessage
		HRAInput.States[activeElevators[i]] = HRAState
		// slog.Info("[HRA]: Constructed state", "id", activeElevators[i], "state", HRAState)
	}

	return HRAInput
}

func constructFsmOrder(hallOrders [config.NumberFloors][2]bool) [config.NumberFloors][3]bool{
    
    var fsmOrders [config.NumberFloors][3]bool
    cabOrders := orders.GetCabOrders(config.ElevatorId)

    for i := range cabOrders{
        fsmOrders[i][2] = cabOrders[i]
    }
    for i := range hallOrders{
        copy(fsmOrders[i][:], hallOrders[i][:])
    }
    return fsmOrders
}


func Assigner(
    signalAssignChan <-chan  bool,
    newOrdersChan chan <- [config.NumberFloors][3]bool,
) {


	for range signalAssignChan{

            currentActivePeers := GetActivePeers()
            slog.Info("[Assigner]: Got distribute signal", "activePeers", currentActivePeers)

            HRAInput := ConstructHRAInput(currentActivePeers)

            // slog.Info("[Assigner]: HRA input succsefully created")

            output := CalulateOrders(HRAInput)
            fsmOrders := constructFsmOrder([config.NumberFloors][2]bool(output[config.ElevatorId]))
            slog.Info("[Assigner] sending orders to fsm", "orders", fsmOrders)
            newOrdersChan <- fsmOrders

		
	}
}

// func Distributor(
// 	mainID string,
// 	distributeSignal <-chan bool,
// 	sendHallReqeustsFsm chan<- [config.NumberFloors][3]bool,
// ) {
// 	var allReqeusts [config.NumberFloors][3]bool
//
// 	for range distributeSignal {
// 		currentActivePeers := GetActivePeers()
// 		slog.Info("[distributor]: Got distribute signal", "activePeers", currentActivePeers)
//
// 		HRAInput := ConstructHRAInput(currentActivePeers)
// 		slog.Info("[distributor]: HRA input succsefully created")
//
// 		output := CalulateOrders(HRAInput)
// 		slog.Info("[distribitor]: HRA caluclated", "HRA_output", output)
//
// 		currentHallRequests = [config.NumberFloors][2]bool(output[mainID])
// 		slog.Info("[distribitor]: our elevators", "hallRequests", currentHallRequests)
//
// 		// sendHallReqeustsFsm <- [config.NumberFloors][2]bool(output[mainID])
// 		slog.Info("[distributor]: Sending to FSM", "hallrequest", [config.NumberFloors][2]bool(output[mainID]))
// 	}
// }
