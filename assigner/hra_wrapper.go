package assigner

import (
	"elevator/config"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)

// input := HRAInput{
//     HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
//     States: map[string]HRAElevState{
//         "one": HRAElevState{
//             Behavior:       "moving",
//             Floor:          2,
//             Direction:      "up",
//             CabRequests:    []bool{false, false, false, true},
//         },
//         "two": HRAElevState{
//             Behavior:       "idle",
//             Floor:          0,
//             Direction:      "stop",
//             CabRequests:    []bool{false, false, false, false},
//         },
//     },
// } 

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
	Floor       int    `json:"floor"`
}

type HRAInput struct {
	States       map[string]HRAElevState
	HallRequests [][2]bool `json:"hallRequests"`
}

func AutoSetHRAExecutable(){
    hraExecutable := ""
    switch runtime.GOOS {
    case "linux":   hraExecutable  = "../bin/hall_request_assigner"
    case "windows": hraExecutable  = "../bin/hall_request_assigner.exe"
    default:        panic("OS not supported")
    }
    config.HallRequestAssignerExecutable = hraExecutable
}



func assingHallOrders(input HRAInput) (map[string][][2]bool){
    jsonBytes, err := json.Marshal(input)
        if err != nil {
            fmt.Println("json.Marshal error: ", err)
            return nil
        }
        
        ret, err := exec.Command(config.HallRequestAssignerExecutable, "-i", string(jsonBytes)).CombinedOutput()
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
            
        fmt.Printf("output: \n")
        for k, v := range *output {
            fmt.Printf("%6v :  %+v\n", k, v)
        }

    return *output
}


