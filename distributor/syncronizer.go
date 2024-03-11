package distribitor

import (
	// "Driver-go/elevio"
	"elevator/config"
	// "Network-go/network/bcast"
	"Network-go/network/peers"
	"fmt"
	"sync"
	"time"
)

/*
Syncronizer part of distributor. The purpose is to keep track of all the active
peers in the network, and syncronize the state among them. This is done by broadcasting
the local state of our elevator to the other peers, and update our local overview of the
other peers's state that are broadcasted.
* */

// TODO: this should be in fsm
type State struct {
	CabRequests []bool
	Floor       int
	Direction   int
	Behavior    int
}

type StateMessageBroadcast struct {
	Id           string
	Checksum     []byte
	State        State
	Sequence     int
	HallRequests [config.NumberFloors][2]bool
	UpdateOrders bool
}

/* Local collection of all states and hall requests. */
var (
	localPeerStates       map[string]State = make(map[string]State)
	localHallRequests     [config.NumberFloors][2]bool
	localPeerStateLock    sync.Mutex
	localHallRequestsLock sync.Mutex
)

func addElevator(id string) {
	// defalut value in bool array is false
	state := State{
		CabRequests: make([]bool, config.NumberFloors),
		Floor:       0,
		Direction:   0,
		Behavior:    0,
	}
	localPeerStateLock.Lock()
	localPeerStates[id] = state
	localPeerStateLock.Unlock()
}

func updateElevator(id string, newState State) {
	localPeerStateLock.Lock()
	localPeerStates[id] = newState
	localPeerStateLock.Unlock()
}

func removeElevator(id string) {
	_, ok := localPeerStates[id]
	if !ok {
		panic("[removeElevator]: Tried removing state that is not in peerStates")
	}
	localPeerStateLock.Lock()
	delete(localPeerStates, id)
	localPeerStateLock.Unlock()
}

func getElevatorState(id string) State {
	localPeerStateLock.Lock()
	localElevatorState := localPeerStates[id]
	localPeerStateLock.Unlock()
	return localElevatorState
}

func updateHallRequests(newHallrequest [config.NumberFloors][2]bool) {
	localHallRequestsLock.Lock()
	localHallRequests = newHallrequest
	localHallRequestsLock.Unlock()
}

func getHallReqeusts() [config.NumberFloors][2]bool {
	localHallRequestsLock.Lock()
	HallRequests := localHallRequests
	localHallRequestsLock.Unlock()
	return HallRequests
}

// TODO: update our own elevator state aswell.
func Syncronizer(
	mainID string,
	broadcastStateMessageTx chan<- StateMessageBroadcast,
	broadcastStateMessageRx <-chan StateMessageBroadcast,
	peerUpdatesRx <-chan peers.PeerUpdate,
) {
	localStateMessage := StateMessageBroadcast{
		Id:           mainID,
		HallRequests: getHallReqeusts(),
		State:        getElevatorState(mainID),
		Sequence:     0,
		Checksum:     nil,
		UpdateOrders: false,
	}

	lastStateBroadcast := time.Now().UnixMilli() - config.BroadcastStateIntervalMs // To broadcast imideatly

	for {
		select {
		case peerUpdate := <-peerUpdatesRx:
			if peerUpdate.New != "" {
				fmt.Println("[Syncronizer]: Adding elevator: ", peerUpdate.New)
				addElevator(peerUpdate.New)
			}

			for i := 0; i < len(peerUpdate.Lost); i++ {
				fmt.Println("[syncronizer]: Removing elevator: ", peerUpdate.Lost[i])
				removeElevator(peerUpdate.Lost[i])
			}

		case stateMessage := <-broadcastStateMessageRx:
			if !stateMessage.UpdateOrders {
				updateElevator(stateMessage.Id, stateMessage.State)
				// If HallRequests is broadscasted without Update orders flag it is valid.
				updateHallRequests(stateMessage.HallRequests)
			}

			// TODO: check that this executes with the right interval
		default:
			if time.Now().UnixMilli() < lastStateBroadcast+config.BroadcastStateIntervalMs {
				continue
			}

			//  TODO: make this as function?
			localStateMessage.State = getElevatorState(mainID) // TODO: state from map or state from fsm?
			localStateMessage.HallRequests = getHallReqeusts()
			localStateMessage.UpdateOrders = false
			localStateMessage.Checksum, _ = HashStructSha1(localStateMessage)
			fmt.Println("[syncronizer]: Broadcasting, Checksum: ", localStateMessage.Checksum)
			broadcastStateMessageTx <- localStateMessage
			lastStateBroadcast = time.Now().UnixMilli()
			localStateMessage.Sequence += 1
		}
	}
}