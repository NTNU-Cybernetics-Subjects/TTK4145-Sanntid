package assigner

import (
	"elevator/config"
	"encoding/json"
	"fmt"
	"os/exec"
	// "runtime"
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
	States       map[string]ElevatorStateHRA `json:"states"`
	HallRequests [][2]bool                   `json:"hallRequests"`
}

func AssingOrder(HRAInput InputHRA) map[string][][2]bool {
	jsonBytes, err := json.Marshal(HRAInput)
	if err != nil {
		fmt.Println("josn.Marshal error: ", err)
		return nil
	}

	ret, err := exec.Command(HRAExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return nil
	}

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return nil
	}

	return *output
}
