package distributor

import (
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
		slog.Info("[Distributor]: json.Marshal error: ", err)
		return nil
	}

	ret, err := exec.Command(HRAExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		// fmt.Println("exec.Command error: ", err)
		// fmt.Println(string(ret))
		slog.Info("[Distributor]: exec.Command error: ", err, slog.String("returned", string(ret)))
		return nil
	}

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		// fmt.Println("json.Unmarshal error: ", err)
		slog.Info("[distribitor]: json.Unmarshal error: ", err)
		return nil
	}

	return *output
}


func ConstructHRAState(input fsm.ElevatorState) ElevatorStateHRA {

    // slog.Info("[HRA constructor]: ", "fsm.requests", input.Requests)

    cabIndex := 2 // FIXME: make sure this is right index
    var cabRequests [config.NumberFloors]bool
    for floor := 0; floor < config.NumberFloors; floor++{
        cabRequests[floor] = input.Requests[floor][cabIndex]
    }
    // slog.Info("[HRA constructor]: maped cab requests ", "cabRequests", cabRequests)

	HRAState := ElevatorStateHRA{
		Behavior:    "",
		Floor:       input.Floor,
		Direction:   "",
        CabRequests: cabRequests[:],
	}

	switch input.Behavior {
	case fsm.EB_Idle:
		HRAState.Behavior = "idle"
	case fsm.EB_Moving:
		HRAState.Behavior = "moving"
	case fsm.EB_DoorOpen:
		HRAState.Behavior = "doorOpen" // FIXME: is this a valid input in HRA?
	}
    // slog.Info("[HRA constructor]: ", "Behavior", HRAState.Behavior)

	switch input.Direction {
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
		HallRequests: getHallReqeusts(),
	}
	for i := range activeElevators {
        stateMessage := getLocalElevatorStateMessage(activeElevators[i])
        slog.Info("[HRA]: collected state message", "stateMessage", stateMessage)
        HRAState := ConstructHRAState(stateMessage.State)
        HRAInput.States[activeElevators[i]] = HRAState
        slog.Info("[HRA]: Constructed state", "id", activeElevators[i], "state", HRAState)
	}

	return HRAInput
}

func GetHallOrder(mainID string) [config.NumberFloors][2]bool{
    currentActivePeers := getActivePeers()
    slog.Info("[GetOrder]", "acitvePeers", currentActivePeers)
    HRAInput := ConstructHRAInput(currentActivePeers)
    orders := CalulateOrders(HRAInput)
    slog.Info("[GetOrder]", "allOrders", orders)

    ourOrder := orders[mainID]
    slog.Info("[GetOrder]", "ourOrder", ourOrder)

    return [config.NumberFloors][2]bool(ourOrder)
}

func Distributor(
    mainID string,
    distributeSignal <-chan bool,
    sendHallReqeustsFsm chan <- [config.NumberFloors][2]bool,
) {

    var currentHallRequests [config.NumberFloors][2]bool
    for range distributeSignal{
        currentActivePeers := getActivePeers()
        slog.Info("[distributor]: Got distribute signal", "activePeers", activePeers)

        HRAInput := ConstructHRAInput(currentActivePeers)
        slog.Info("[distributor]: HRA input succsefully created")

        output := CalulateOrders(HRAInput)
        slog.Info("[distribitor]: HRA caluclated", "HRA_output" ,output)

        currentHallRequests = [4][2]bool(output[mainID])
        slog.Info("[distribitor]: our elevators", "hallRequests", currentHallRequests)

        sendHallReqeustsFsm <- [config.NumberFloors][2]bool(output[mainID])
        slog.Info("[distributor]: Sending to FSM", "hallrequest", [config.NumberFloors][2]bool(output[mainID]))
    }
}
