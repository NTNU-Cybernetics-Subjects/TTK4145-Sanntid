package distributor

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
	CabRequests []bool
	Behavior    fsm.ElevatorBehavior
	Floor       int
	Direction   elevio.MotorDirection
	Obstructed  bool
	Sequence    int64
}

// Dont need updateOrders flag
type StateMessageBroadcast struct {
	Id           string
	Checksum     []byte
	State        ElevatorState
	Sequence     int64
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

func getActivePeers() []string {
	activePeersLock.Lock()
	currentActivePeers := activePeers
	activePeersLock.Unlock()
	return currentActivePeers
}

func setActivePeers(newActivePeers []string) {
	activePeersLock.Lock()
	activePeers = newActivePeers
	activePeersLock.Unlock()
}

// NOTE: this should mabye be initElevator?
func addElevator(id string) {
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

func syncHallRequest(newHallrequest [config.NumberFloors][2]bool, operation HallOperation) {
	// Set
	if operation == HRU_SET {

		localHallRequestsLock.Lock()
		for floor := range newHallrequest {
			for dir := range newHallrequest[floor] {
				localHallRequests[floor][dir] = newHallrequest[floor][dir] || localHallRequests[floor][dir]
			}
		}
		localHallRequestsLock.Unlock()
		hallReq := getHallReqeusts() // FIXME: debug variable
		slog.Info("[syncHallRequest SET]: hallrequests synced SET", "Localhallrequests", hallReq)
		return
	}

	// Clear
	if operation == HRU_CLEAR {
		localHallRequestsLock.Lock()
		for floor := range newHallrequest {
			for dir := range newHallrequest[floor] {
				localHallRequests[floor][dir] = newHallrequest[floor][dir] && localHallRequests[floor][dir]
			}
		}
		localHallRequestsLock.Unlock()
		hallReq := getHallReqeusts() // FIXME: debug variable
		slog.Info("[syncHallRequest CLEAR]: hallrequests synced SET", "Localhallrequests", hallReq)
	}
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
	signalDistributor chan<- bool,
) {
	lastStateBroadcast := time.Now().UnixMilli() - config.BroadcastStateIntervalMs // To broadcast imideatly
	networkStateInitialized := false
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

			// update active peers list.
			setActivePeers(peerUpdate.Peers)

			slog.Info("[peerUpdate]: current active peers", "peers", getActivePeers())

			// We are adding ourself
			if peerUpdate.New == mainID {
				addElevator(mainID)
				networkStateInitialized = true
				continue
			}

			// Another elevator is added
			if peerUpdate.New != "" {
				slog.Info("[peerUpdate]: Intialize elevator", slog.String("ID", peerUpdate.New))
				addElevator(peerUpdate.New)
				// TODO: broadcast the current state of the new peer.
			}
			if len(peerUpdate.Lost) > 0 {
				slog.Info("[peerUpdate]: lost elevator")
			}

			// Send distribute signal each time we get peer update.
			if !networkStateInitialized {
				continue
			}
			// want to redistribute when new peers connect
            // signalDistributor <- true // FIX:

		case incommingStateMessage := <-broadcastStateMessageRx:

			if incommingStateMessage.Id == mainID {
				// TODO: This happens only if we just connected to peer network. If the state is newer than our own sync to incomming.
				continue
			}

			updateElevator(incommingStateMessage.Id, incommingStateMessage.State)

			lastHallRequestUpdateMessage := getHallRequestUpdateMessage(incommingStateMessage.Id)
			if lastHallRequestUpdateMessage.Operation == HRU_NONE {
				continue
			}

			syncHallRequest(incommingStateMessage.HallRequests, lastHallRequestUpdateMessage.Operation)
			clearHallRequestUpdateOperationFlag(incommingStateMessage.Id)
			slog.Info("[broadcaster<-]: request is set unactive", "from", incommingStateMessage.Id)
            // signalDistributor <- true FIX: 

		default:
			if time.Now().UnixMilli() < lastStateBroadcast+config.BroadcastStateIntervalMs {
				continue
			}

			stateMessage.Id = mainID
			stateMessage.State = getElevatorState(mainID) // TODO: fsm.GetElevatorState
			stateMessage.HallRequests = getHallReqeusts()
			stateMessage.Checksum, _ = HashStructSha1(stateMessage)
			broadcastStateMessageTx <- stateMessage
			stateMessage.Sequence += 1
			stateMessage.State.Sequence = stateMessage.Sequence
			lastStateBroadcast = time.Now().UnixMilli()
		}
	}
}
