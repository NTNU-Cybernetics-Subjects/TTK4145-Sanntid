package main

import (
	"Driver-go/elevio"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	distribitor "elevator/distributor"
	"elevator/fsm"

	"flag"
)

/*
Setup:
	1. Initialize elevator (Should make a function for this
		 where it settles on the nearest floor and updates
		 its state?)
	2. Initialize peer to network
	3. Broadcast and synchronize all peer states

Main loop:
	1. Wait for button event
	2. On button event:
		If Hall call:
			1. Broadcast hall call to all peers - wait for ACK from all.
			2. Most relevant elevator "claims" service call, requests ACK from peers.
			3. Once ACK is received, send service call to fsm and activate lights.
		If Cab call:
			1. Distributor receives service call, sends it to fsm and activate lights.

	    fsm services the call, updates state and sends the updated state to distributor
		Distributor synchronizes all peers.
	3. When arrived at floor:
		If serviced call:
			1. Broadcast service as completed, request ACK.
			2. Once ACK is received, distributor marks service as complete and deactivates the light.
	4. If Obstruction:
		Wait.
	5. If StopButton:
		Wait.

*/

type test struct {
	Id     string
	Number int
}

var elevator fsm.ElevatorState

func main() {
	var id string
	flag.StringVar(&id, "id", "", "-id ID")
	flag.Parse()

	elevio.Init("localhost:15657", 4)
	buttonsChan := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(buttonsChan)

	// Channel for peer updates
	peerUpdateCh := make(chan peers.PeerUpdate)
	// Turn peer on/off (default on)
	peerEnableCh := make(chan bool)
	go peers.Transmitter(distribitor.PEER_PORT, id, peerEnableCh)
	go peers.Receiver(distribitor.PEER_PORT, peerUpdateCh)
	go distribitor.PeerWatcher(peerUpdateCh)

	broadcasdStateChRx := make(chan distribitor.StateMessageBroadcast)
	broadcastStateChTx := make(chan distribitor.StateMessageBroadcast)
	go bcast.Receiver(distribitor.BCAST_PORT, broadcasdStateChRx)
	go bcast.Transmitter(distribitor.BCAST_PORT, broadcastStateChTx)

	stateUpdateFsm := make(chan distribitor.State)

	distribitor.Syncronizer(id, broadcasdStateChRx, broadcastStateChTx, stateUpdateFsm, buttonsChan)
}
