package distribitor

import (
	// "Driver-go/elevio"
	"Network-go/network/bcast"
	"Network-go/network/peers" // "bytes"
	"fmt"
)

/*
Syncronizer part of distributor. The puprose is to keep track of
all active peers in the network, and syncronize the states between them.
It also takes inn requests, and acknowledge them. The requests gets
acknowledged if the nummber of acknowledgement recived is the same as the
number of active peers. The request is then broadcasted, and we consider
the request active.
*/

// type HRAElevState struct {
//     Behavior    string      `json:"behaviour"`
//     Floor       int         `json:"floor"`
//     Direction   string      `json:"direction"`
//     CabRequests []bool      `json:"cabRequests"`
// }
//
// type HRAInput struct {
//     HallRequests    [][2]bool                   `json:"hallRequests"`
//     States          map[string]HRAElevState     `json:"states"`
// }

// Config variabes, FIXME: should be defined in config file?
const (
	PEER_PORT  int    = 12348
	HOST       string = "localhost"
	BCAST_PORT int    = 4875
)

const (
	n_elevators int = 3
	n_floors    int = 4
)

type State struct {
	CabRequests []bool
	Floor       int
	Direction   int
	Behavior    int
}

type State_msg_bcast struct {
	Id              string
	HallRequests    [][2]bool
	State           State
	Sequence        int
	Checksum        [160]byte
	RequestToUpdate bool
}

/* Global clollection of all states. This variable contains the state off
* all elevators that are connected to the peer to peer network. */
var peerStates map[string]State = make(map[string]State)

/*
This function will watch which peers that are connected to the network,
it will update the global map peerStates accordingly by adding/removing
the state in the map.
*/
func PeerWatcher(id string) {
	// Channel for peer updates
	peerUpdateCh := make(chan peers.PeerUpdate)

	// Turn peer on/off (default on)
	peerEnablCh := make(chan bool)

	go peers.Transmitter(PEER_PORT, id, peerEnablCh)
	go peers.Receiver(PEER_PORT, peerUpdateCh)

	for p := range peerUpdateCh {

		// Add new peers to peerStates
		if p.New != "" {
			fmt.Print("adding elevator")
			add_elevator(p.New)
		}
		// Remove all lost peers from peerStates
		for i := 0; i < len(p.Lost); i++ {
			fmt.Println("removing elevator: ", p.Lost[i])
			remove_elevator(p.Lost[i])
		}
	}
}

/* Add a new peer to the global map peerStates. This should be called when new peers join the
* network.
* TODO:: add_elevator should syncronize state with network state*/
func add_elevator(id string) {
	// defalut value in bool array is false
	state := State{
		CabRequests: make([]bool, n_floors),
		Floor:       0,
		Direction:   0,
		Behavior:    0,
	}
	// TODO: add mutex around peerStates?
	peerStates[id] = state
}

/* Remove peer from the global peerStates map. This should be called on all peers disconecting
* from the newtwork. */
func remove_elevator(id string) {
	_, ok := peerStates[id]
	if !ok {
		panic("[remove_elevator]: Tried removing state that is not in peerStates")
	}
	// TODO: add mutex around peerStates?
	delete(peerStates, id)
}

func Syncronizer() {
	// broadcast channel  TODO: channels should be parameters
	bcastChTx := make(chan State_msg_bcast)
	bcastChRx := make(chan State_msg_bcast)
	go bcast.Receiver(BCAST_PORT, bcastChRx)
	go bcast.Transmitter(BCAST_PORT, bcastChTx)

	for peer_msg := range bcastChRx {
		if !peer_msg.RequestToUpdate {
			peerStates[peer_msg.Id] = peer_msg.State
		}
	}
}
