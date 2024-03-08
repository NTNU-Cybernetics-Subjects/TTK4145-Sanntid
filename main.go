package main

import (
	// "Driver-go/elevio"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"elevator/config"
	"elevator/distributor"

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


func main(){
    var id string;
    flag.StringVar(&id, "id", "", "-id ID")
    flag.Parse()

    broadcastStateMessageRx := make(chan distribitor.StateMessageBroadcast)
    broadcastStateMessageTx := make(chan distribitor.StateMessageBroadcast)
    peersUpdateRx := make(chan peers.PeerUpdate)
    peersEnable := make(chan bool)

    go bcast.Receiver(config.BCAST_PORT, broadcastStateMessageRx)
    go bcast.Transmitter(config.BCAST_PORT, broadcastStateMessageTx)
    go peers.Transmitter(config.PEER_PORT, id , peersEnable)
    go peers.Receiver(config.PEER_PORT, peersUpdateRx)

    go distribitor.Syncronizer(id, broadcastStateMessageTx, broadcastStateMessageRx, peersUpdateRx)

    select {}
}

