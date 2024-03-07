package distribitor

// "Network-go/network/peers"
// "Network-go/network/bcast"
// "Driver-go/elevio"
// "fmt"

import "os/exec"
import "fmt"
import "encoding/json"
import "runtime"

/*
Decition part of the distributor. The purpose is to decide which peer should,
execute the active request.

The active order should fall on the peer that is closesd to the floor and going
in right direction from which the cab is called. If there are any conflicts
between peers, if should fall on the node with the lowest id.

The decition should happen in the following order:
The node that feels responsible for the order, should broadcast that it want to
take the order, if all nodes acknowledge then the peer broadcast that it is taking
the order. The order is then considered taken.
*/

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

func makeInputHRA(){

}

func main(){

    hraExecutable := ""
    switch runtime.GOOS {
        case "linux":   hraExecutable  = "hall_request_assigner"
        case "windows": hraExecutable  = "hall_request_assigner.exe"
        default:        panic("OS not supported")
    }

    input := HRAInput{
        HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
        States: map[string]HRAElevState{
            "one": HRAElevState{
                Behavior:       "moving",
                Floor:          2,
                Direction:      "up",
                CabRequests:    []bool{false, false, false, true},
            },
            "two": HRAElevState{
                Behavior:       "idle",
                Floor:          0,
                Direction:      "stop",
                CabRequests:    []bool{false, false, false, false},
            },
        },
    }

    jsonBytes, err := json.Marshal(input)
    if err != nil {
        fmt.Println("json.Marshal error: ", err)
        return
    }
    
    ret, err := exec.Command("../hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
    if err != nil {
        fmt.Println("exec.Command error: ", err)
        fmt.Println(string(ret))
        return
    }
    
    output := new(map[string][][2]bool)
    err = json.Unmarshal(ret, &output)
    if err != nil {
        fmt.Println("json.Unmarshal error: ", err)
        return
    }
        
    fmt.Printf("output: \n")
    for k, v := range *output {
        fmt.Printf("%6v :  %+v\n", k, v)
    }
}
