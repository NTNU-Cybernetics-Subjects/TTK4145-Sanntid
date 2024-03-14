package main

import (
	// "Driver-go/elevio"
	"Driver-go/elevio"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"elevator/config"
	"elevator/distributor"
	"elevator/fsm"
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

	elevatorServerAddr := host + ":" + port
	elevio.Init(elevatorServerAddr, config.NumberFloors)

	doorTimerChan := make(chan bool)
	buttonsChan := make(chan elevio.ButtonEvent)
	floorSensorChan := make(chan int)
	obstructionChan := make(chan bool)
	hallRequestDistributorChan := make(chan [config.NumberFloors][2]bool)

	buttonEventUpdateChan := make(chan elevio.ButtonEvent)
	stateUpdateChan := make(chan fsm.ElevatorState)

	go fsm.Fsm(
		buttonEventUpdateChan,
		stateUpdateChan,
		obstructionChan,
		buttonsChan,
		floorSensorChan,
		doorTimerChan,
		hallRequestDistributorChan,
	)

	go fsm.PollTimer(doorTimerChan)
	go elevio.PollButtons(buttonsChan)
	go elevio.PollFloorSensor(floorSensorChan)
	go elevio.PollObstructionSwitch(obstructionChan)

	broadcastStateMessageRx := make(chan distributor.StateMessageBroadcast)
	broadcastStateMessageTx := make(chan distributor.StateMessageBroadcast)

	broadcastOrderRx := make(chan distributor.HallRequestUpdate)
	broadcastOrderTx := make(chan distributor.HallRequestUpdate)

	go bcast.Receiver(config.BCAST_PORT,
		broadcastStateMessageRx,
		broadcastOrderRx,
	)
	go bcast.Transmitter(
		config.BCAST_PORT,
		broadcastStateMessageTx,
		broadcastOrderTx,
	)

	peersUpdateRx := make(chan peers.PeerUpdate)
	peersEnable := make(chan bool)
	go peers.Transmitter(config.PEER_PORT, id, peersEnable)
	go peers.Receiver(config.PEER_PORT, peersUpdateRx)

	distributorSignal := make(chan bool)
	go distributor.Distributor(
		id,
		distributorSignal,
		hallRequestDistributorChan,
	)

	go distributor.Syncronizer(
		id,
		broadcastStateMessageTx,
		broadcastStateMessageRx,
		peersUpdateRx,
		distributorSignal,
	)
	go distributor.RequestHandler(
		id,
		broadcastOrderRx,
		broadcastOrderTx,
		buttonEventUpdateChan,
		distributorSignal,
	)

	select {}
}
