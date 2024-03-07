package distribitor

import (
	// "Driver-go/elevio"
	"Driver-go/elevio"
	// "Network-go/network/bcast"
	"Network-go/network/peers"
	"fmt"
	"sync"
	"time"
)

/* Syncronizer part of distributor. The puprose is to keep track of
all active peers in the network, and syncronize the states between them.
It also takes inn requests, and acknowledge them. The requests gets
acknowledged if the nummber of acknowledgement recived is the same as the
number of active peers. The request is then broadcasted, and we consider
the request active. */


type State struct {
	CabRequests []bool
	Floor       int
	Direction   int
	Behavior    int
}

type StateMessageBroadcast struct {
	Id              string
	HallRequests    [][2]bool
	State           State
	Sequence        int
	Checksum        []byte
	RequestToUpdate bool
}

/* Local clollection of all states. This variable contains the state off
* all elevators that are connected to the peer to peer network. */
var (
	peerStates       map[string]State = make(map[string]State)
	HallRequests     [n_floors][2]bool
	peerStateLock    sync.Mutex
	HallRequestsLock sync.Mutex
)

/*
This function will watch which peers that are connected to the network,
it will update the local peerStates map accordingly by adding/removing
the state in the map.
*/

func PeerWatcher(peerUpdateCh <-chan peers.PeerUpdate) {
	for p := range peerUpdateCh {

		// Add new peers to peerStates
		if p.New != "" {
			fmt.Print("Adding elevator")
			addElevator(p.New)
		}
		// Remove all lost peers from peerStates
		for i := 0; i < len(p.Lost); i++ {
			fmt.Println("Removing elevator: ", p.Lost[i])
			removeElevator(p.Lost[i])
		}
	}
}

/* Add a new peer to the local map peerStates. This should be called when new peers join the
* network.
*  TODO:: addElevator should syncronize state with network state? */
func addElevator(id string) {
	// defalut value in bool array is false
	state := State{
		CabRequests: make([]bool, n_floors),
		Floor:       0,
		Direction:   0,
		Behavior:    0,
	}
	peerStateLock.Lock()
	peerStates[id] = state
	peerStateLock.Unlock()
}

/*Update a peer in the local peerStates map. This is used by the syncronizer to update the states
* off all elevators. */
func updateElevator(id string, newState State) {
	peerStateLock.Lock()
	peerStates[id] = newState
	peerStateLock.Unlock()
}

/* Remove peer from the local peerStates map. This should be called on all peers disconecting
*  from the newtwork. */
func removeElevator(id string) {
	_, ok := peerStates[id]
	if !ok {
		panic("[removeElevator]: Tried removing state that is not in peerStates")
	}
	peerStateLock.Lock()
	delete(peerStates, id)
	peerStateLock.Unlock()
}

func updateHallRequests(newHallrequest [n_floors][2]bool) {
	HallRequestsLock.Lock()
	HallRequests = newHallrequest
	HallRequestsLock.Unlock()
}

// // TODO: This should broadcast state, update its own peer, and handle state updates from fsm
// func BroadcastState(mainId string, stateBcast chan<- StateMsgBcast, stateUpdateFsm <-chan State) {
// 	stateMsg := StateMsgBcast{
// 		Id:              mainId,
// 		HallRequests:    HallRequests[:],
// 		State:           peerStates[mainId],
// 		Sequence:        0,
// 		RequestToUpdate: false,
// 	}
//
// 	var sendFrequence int64 = 800 // ms
// 	lastTimeSendt := time.Now().UnixMilli()
//
// 	for {
// 		select {
// 		// TODO: updates from fsm
// 		case fsmUpdate := <-stateUpdateFsm:
// 			stateMsg.State = fsmUpdate
//
// 		// TODO: request granted
// 		default:
// 			if time.Now().UnixMilli() >= lastTimeSendt+sendFrequence {
// 				stateMsg.Checksum, _ = HashStructSha1(stateMsg)
// 				stateBcast <- stateMsg
// 				lastTimeSendt = time.Now().UnixMilli()
// 				stateMsg.Sequence += 1
// 			}
// 		}
// 	}
// }

/*  NOTE: what happens if state updates in the middle of a request? */
func requestToUpdate(
    mainId string,
    button elevio.ButtonEvent,
    acknowledgeGranted <-chan string,
    stateMsg StateMessageBroadcast,
) {

    // subtract buttons and pack them into "request"
    // prepare message
    //
    // for not all ack 
    // take inn updates from peers, so that only active peers need to ack
    // return

}

func acknowledgeRequest(mainId string, stateMessage StateMessageBroadcast, broadcastSTateTx chan <-StateMessageBroadcast) (ackId string) {
 
    // TODO: what to do with checksum, and do we just ack or do some checks?
    acknowledgeId := stateMessage.Id

    stateMessage.Id = mainId

    broadcastSTateTx <- stateMessage

	return acknowledgeId
}

func Syncronizer(
    mainId string,
	broadcastStateChRx <-chan StateMessageBroadcast,
	broadcastStateChTx chan<- StateMessageBroadcast,
	stateUpdateFsm <-chan State,
	buttonsChan <-chan elevio.ButtonEvent,
) {
	stateMsgLocal := StateMessageBroadcast{
		Id:              mainId,
		HallRequests:    HallRequests[:],
		State:           peerStates[mainId],
		Sequence:        0,
		RequestToUpdate: false,
        Checksum: nil,
	}

	const sendIntervalMs int64 = 800 // move to config
	lastTimeSendt := time.Now().UnixMilli()-sendIntervalMs // To send update imediatly
	acknowledgeGranted := make(chan string)

	for {
		select {

		case peerMsg := <-broadcastStateChRx:
			// If only state update we do not need ack, Each elevator "knows" best its own state.
			// In addition if it stateMsg are broadcasted without the requestToUpdate flag it is considered acknowledged.
			// This means that new connecting peers will update their data accoording to the network state.
			if !peerMsg.RequestToUpdate {
				updateElevator(peerMsg.Id, peerMsg.State)
				updateHallRequests([n_floors][2]bool(peerMsg.HallRequests))
				fmt.Println("[Syncronizer]: update Recived ", peerMsg)
			} else {
				// This will execute on all pacakgees that have requests to update flag. This means that we will ack our on packages aswell.
				ackID := acknowledgeRequest(mainId, stateMsgLocal, broadcastStateChTx)
				acknowledgeGranted <- ackID
			}

		case button := <-buttonsChan:
			// start a requestToUpdate process
			go requestToUpdate(mainId, button, acknowledgeGranted, stateMsgLocal)

			// TODO: this comes from fsm, implement when ready
		case fsmUpdate := <-stateUpdateFsm:
			stateMsgLocal.State = fsmUpdate

			// Use defalt to broadcast stateMsg
		default:
            if time.Now().UnixMilli() < lastTimeSendt+sendIntervalMs {
                continue
            }

			stateMsgLocal.Checksum, _ = HashStructSha1(stateMsgLocal)
			broadcastStateChTx <- stateMsgLocal
			lastTimeSendt = time.Now().UnixMilli()
			stateMsgLocal.Sequence += 1
		}
	}
}
