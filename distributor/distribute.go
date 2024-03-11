package distribitor

import (
	"Driver-go/elevio"
	"elevator/config"
	"elevator/fsm"
	"encoding/json"
	"fmt"
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

func AssingOrder(HRAInput InputHRA) map[string][][2]bool {
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

func ConstructHRAState(input ElevatorState) ElevatorStateHRA {
	HRAState := ElevatorStateHRA{
		Behavior:    "",
		Floor:       input.Floor,
		Direction:   "",
		CabRequests: input.CabRequests,
	}

	switch input.Behavior {
	case fsm.EB_Idle:
		HRAState.Behavior = "idle"
	case fsm.EB_Moving:
		HRAState.Behavior = "moving"
	case fsm.EB_DoorOpen:
		HRAState.Behavior = "doorOpen" // FIXME: is this a valid input in HRA?
	}
    fmt.Println("[Constructing state]: choose behaviour: ", HRAState.Behavior)

	switch input.Direction {
	case elevio.MD_Up:
		HRAState.Direction = "up"
	case elevio.MD_Down:
		HRAState.Direction = "down"
	case elevio.MD_Stop:
		HRAState.Direction = "stop"
	}
    fmt.Println("[construction state]: chosing direction", HRAState.Direction)
	return HRAState
}

func ConstructHRAInput(activePeers []string) InputHRA {
	HRAInput := InputHRA{
		States:       make(map[string]ElevatorStateHRA),
		HallRequests: getHallReqeusts(),
	}
	for i := range activePeers {
		HRAInput.States[activePeers[i]] = ConstructHRAState(getElevatorState(activePeers[i]))
	}

	return HRAInput
}

func Distribute() {
}
