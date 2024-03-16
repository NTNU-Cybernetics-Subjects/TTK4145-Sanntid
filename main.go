package main

import (
	"Driver-go/elevio"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"elevator/config"
	"elevator/fsm"
	"elevator/peerNetwork"
	"flag"
)

/*
Setup:
	1. Initialize elevator (Should make a function for this
		 where it settles on the nearest floor and updates
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

func main() {
	var id string
	var port string
	var host string
	flag.StringVar(&id, "id", "", "-id ID")
	flag.StringVar(&port, "port", "15657", "-port PORT")
	flag.StringVar(&host, "host", config.ELEVATOR_HOST, "-host HOST")
	flag.Parse()
    config.ElevatorId = id

	elevatorServerAddr := host + ":" + port
	elevio.Init(elevatorServerAddr, config.NumberFloors)
    // fsm.InitializeElevator() // TODO: 
    // slog.Info("elevator", "is", fsm.GetElevatorState())

    // requestHandler
    requestBcast := peerNetwork.RequestChan{
        Transmitt: make(chan peerNetwork.RequestMessage),
        Receive: make(chan peerNetwork.RequestMessage),
    }

    buttonEventOutputChan := make(chan elevio.ButtonEvent)
    clearOrdersChan := make(chan elevio.ButtonEvent)
    stateOutputChan := make(chan fsm.ElevatorState) // FIXME: this is proboly not used 
    newOrdersChan := make(chan [config.NumberFloors][3]bool)

    go bcast.Receiver(config.BCAST_PORT, requestBcast.Receive)
    go bcast.Transmitter(config.BCAST_PORT, requestBcast.Transmitt)

    go peerNetwork.Handler(requestBcast, buttonEventOutputChan, clearOrdersChan)

    // Syncronizer
    peerUpdateRx := make(chan peers.PeerUpdate)
    peerEnable := make(chan bool)

    go peers.Receiver(config.PEER_PORT, peerUpdateRx)
    go peers.Transmitter(config.PEER_PORT, config.ElevatorId, peerEnable)

    stateMessagechan := peerNetwork.StateMessagechan{
        Transmitt: make(chan peerNetwork.StateMessageBroadcast),
        Receive: make(chan peerNetwork.StateMessageBroadcast),
    }
    go bcast.Receiver(config.BCAST_PORT, stateMessagechan.Receive)
    go bcast.Transmitter(config.BCAST_PORT, stateMessagechan.Transmitt)

    go peerNetwork.Syncronizer(stateMessagechan, peerUpdateRx)

    // go peerNetwork.OrderPrinter()

    // fsm
    go fsm.Fsm(buttonEventOutputChan, clearOrdersChan, stateOutputChan, newOrdersChan)

    go peerNetwork.Assinger(newOrdersChan)

    select {}

}
