package distribitor

import (
	// "Driver-go/elevio"
	"Network-go/network/peers"
	"fmt"
	// "Network-go/network/peers"
	// "Network-go/network/bcast"
)

/*
Syncronizer part of distributor. The puprose is to keep track of
all active peers in the network, and syncronize the states between them.
It also takes inn hall requests, and acknowledge them. The requests gets
acknowledged if the nummber of acknowledgement recived is the same as the
number of active peers. The request is then broadcasted, and we consider
the request active.
*/

const PEER_PORT int = 12348
const HOST string = "localhost"
const BCAST_PORT int = 4875

type State struct {
    ServiceQueue []int;
    Id int;
    Floor int;
    Obstruction int;
    Direction int;
}

// global
var peerStates []State = make([]State, 1)

func PeerWatcher(id string){
	// This channel gets peer updates
	peerUpdateCh := make(chan peers.PeerUpdate)

    // Can turn peer on off the peer network, by sending false on this channel
	peerEnablCh := make(chan bool)

	go peers.Transmitter(PEER_PORT, id, peerEnablCh)
	go peers.Receiver(PEER_PORT, peerUpdateCh)

    for p := range peerUpdateCh {
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
    }
}

func Syncronize(){


}

