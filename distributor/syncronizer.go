package distribitor

import (
	"Driver-go/elevio"
	"Network-go/network/peers"
	"elevator/config"
	"elevator/fsm"
	"log/slog"
	"sync"
	"time"
)

/*
Syncronizer part of distributor. The purpose is to keep track of all the active peers in the network, and syncronize the state among them. This is done by broadcasting
the local state of our elevator to the other peers, and update our local overview of the
other peers's state that are broadcasted.
* */

type ElevatorState struct {
	Behavior    fsm.ElevatorBehavior
	Floor       int
	Direction   elevio.MotorDirection
	CabRequests []bool
	Obstructed  bool
}

// Dont need updateOrders flag
type StateMessageBroadcast struct {
	Id           string
	Checksum     []byte
	State        ElevatorState
	Sequence     int
	HallRequests [config.NumberFloors][2]bool
}

/* Local collection of all states and hall requests. */
var (
	localPeerStates    map[string]ElevatorState = make(map[string]ElevatorState)
	localPeerStateLock sync.Mutex

	localHallRequests     [config.NumberFloors][2]bool
	localHallRequestsLock sync.Mutex

	activePeers     []string
	activePeersLock sync.Mutex
)

func addElevator(id string) {
	// TODO: sync with network state if we reconnect.
	state := ElevatorState{
		CabRequests: make([]bool, config.NumberFloors),
		Floor:       0,
		Direction:   0,
		Behavior:    0,
	}
	localPeerStateLock.Lock()
	localPeerStates[id] = state
	localPeerStateLock.Unlock()
}

func updateElevator(id string, newState ElevatorState) {
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

func getElevatorState(id string) ElevatorState {
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

func setHallRequest(floor int, direction int, active bool) {
	localHallRequestsLock.Lock()
	localHallRequests[floor][direction] = active
	localHallRequestsLock.Unlock()
}

func getHallReqeusts() [config.NumberFloors][2]bool {
	localHallRequestsLock.Lock()
	HallRequests := localHallRequests
	localHallRequestsLock.Unlock()
	return HallRequests
}

func Syncronizer(
	mainID string,
	broadcastStateMessageTx chan<- StateMessageBroadcast,
	broadcastStateMessageRx <-chan StateMessageBroadcast,
	peerUpdatesRx <-chan peers.PeerUpdate,
) {
	lastStateBroadcast := time.Now().UnixMilli() - config.BroadcastStateIntervalMs // To broadcast imideatly
	stateMessage := StateMessageBroadcast{
		Id:           mainID,
		HallRequests: getHallReqeusts(),
		State:        getElevatorState(mainID),
		Sequence:     0,
		Checksum:     nil,
	}

	for {
		select {
		case peerUpdate := <-peerUpdatesRx:

			activePeersLock.Lock()
			activePeers = peerUpdate.Peers
			activePeersLock.Unlock()

			if peerUpdate.New != "" {
				slog.Info("[peerUpdate]: Adding elevator", slog.String("ID", peerUpdate.New))
				addElevator(peerUpdate.New)
				// TODO: distribute
			}

			// FIXME: this should not remove, we want to store the state of other elevators if they reconnect.
			for i := 0; i < len(peerUpdate.Lost); i++ {
				slog.Info("[peerUpdate]: Removing elevator", slog.String("ID", peerUpdate.Lost[i]))
				removeElevator(peerUpdate.Lost[i])
				// TODO: distribute
			}

			// Syncronize incoming state
		case incommingStateMessage := <-broadcastStateMessageRx:
			// fmt.Println(mainID, incommingStateMessage.HallRequests)
			slog.Info("[Broadcast<-]: Syncing hallrequests to: ", incommingStateMessage.HallRequests)
			updateElevator(incommingStateMessage.Id, incommingStateMessage.State)
			// The broadcasted hallRequests are always valid. FIXME: check if this creates race condition when new request go through.
			updateHallRequests(incommingStateMessage.HallRequests)

		default:
			if time.Now().UnixMilli() < lastStateBroadcast+config.BroadcastStateIntervalMs {
				continue
			}

			stateMessage.State = getElevatorState(mainID)
			stateMessage.HallRequests = getHallReqeusts()
			stateMessage.Checksum, _ = HashStructSha1(stateMessage)
			// fmt.Println("[syncronizer]: Broadcasting, Checksum: ", newStateMessage.Checksum)
			broadcastStateMessageTx <- stateMessage
			stateMessage.Sequence += 1
			lastStateBroadcast = time.Now().UnixMilli()
		}
	}
}
