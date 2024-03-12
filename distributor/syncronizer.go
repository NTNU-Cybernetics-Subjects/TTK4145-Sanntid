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
    Sequence int64
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
				slog.Info("[peerUpdate]: Adding elevator", slog.String("ID", peerUpdate.New))
				addElevator(peerUpdate.New)
                // TODO: broadcast the current state of the new peer.
			}

			// Send distribute signal each time we get peer update.
			if !networkStateInitialized {
				continue
			}
			signalDistributor <- true

		case incommingStateMessage := <-broadcastStateMessageRx:
            if incommingStateMessage.Id == mainID {
                // TODO: This happens only if we just connected to peer network. If the state is newer than our own sync to incomming.
                continue
            }
			// fmt.Println(mainID, incommingStateMessage.HallRequests)
			slog.Info("[Broadcast<-]: Syncing", "from", incommingStateMessage.Id, "hallrequests", incommingStateMessage.HallRequests)
			updateElevator(incommingStateMessage.Id, incommingStateMessage.State)

			// What happens when someone says on and someone says off?
			updateHallRequests(incommingStateMessage.HallRequests) // FIXME: check if this creates race condition with broadcaster.

		default:
			if time.Now().UnixMilli() < lastStateBroadcast+config.BroadcastStateIntervalMs {
				continue
			}

			stateMessage.Id = mainID
			stateMessage.State = getElevatorState(mainID) // TODO: fsm.GetElevatorState
			stateMessage.HallRequests = getHallReqeusts()
			stateMessage.Checksum, _ = HashStructSha1(stateMessage)
			// fmt.Println("[syncronizer]: Broadcasting, Checksum: ", newStateMessage.Checksum)
			broadcastStateMessageTx <- stateMessage
			stateMessage.Sequence += 1
            stateMessage.State.Sequence = stateMessage.Sequence
			lastStateBroadcast = time.Now().UnixMilli()
		}
	}
}
