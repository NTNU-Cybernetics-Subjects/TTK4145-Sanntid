package distribitor

import (
	// "Driver-go/elevio"
	"Driver-go/elevio"
	"elevator/config"

	// "Network-go/network/bcast"
	"Network-go/network/peers"
	"fmt"
	"log/slog"
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

// Dont need updateOrders flag
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
	localPeerStates   map[string]State = make(map[string]State)
	localHallRequests [config.NumberFloors][2]bool

	localPeerStateLock    sync.Mutex
	localHallRequestsLock sync.Mutex

	messageSequnceNumber int
	messageSequenceLock  sync.Mutex
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

func waitForHallOrderConfirmation(
	mainID string,
	buttonEvent elevio.ButtonEvent,
	activePeers []string,
	ackChan <-chan string,
) {
	// FIXME: can use lenght of peerMap here since we dont bother about id.
	acknowledgmentsNeeded := len(activePeers)
	countAck := 0
	startTime := time.Now().UnixMilli()

	fmt.Println("[waitForConfirmation]: Required acks: ", acknowledgmentsNeeded)

	for {
		select {
		case <-ackChan:
			countAck += 1
			fmt.Println("ack count: ", countAck) // FIXME:
			if countAck >= acknowledgmentsNeeded {
				// TODO: check that the direction is correct.
				setHallRequest(buttonEvent.Floor, int(buttonEvent.Button), true)
				fmt.Println("HallRequest granted: ", getHallReqeusts())
				return

			}
		default:
			if time.Now().UnixMilli() >= startTime+config.HallOrderAcknowledgeTimeOut {
				// Timeout, we drop the request.
				fmt.Println("[waitForConfirmation]: timed out")
				return
			}
		}
	}
}

func makeBroadcastStateMessage(mainID string) StateMessageBroadcast {
	stateMessage := StateMessageBroadcast{
		Id:           mainID,
		HallRequests: getHallReqeusts(),
		State:        getElevatorState(mainID),
		Sequence:     0,
		Checksum:     nil,
		UpdateOrders: false,
	}
	// Automatically increment sequence number when making new stateMessage
	messageSequenceLock.Lock()
	stateMessage.Sequence = messageSequnceNumber
	messageSequnceNumber += 1
	messageSequenceLock.Unlock()

	stateMessage.Checksum, _ = HashStructSha1(stateMessage)
	return stateMessage
}

func validAck(message StateMessageBroadcast) (bool){
    return true
}

func Syncronizer(
	mainID string,
	broadcastStateMessageTx chan<- StateMessageBroadcast,
	broadcastStateMessageRx <-chan StateMessageBroadcast,
	peerUpdatesRx <-chan peers.PeerUpdate,
	buttonEvent <-chan elevio.ButtonEvent,
) {
	lastStateBroadcast := time.Now().UnixMilli() - config.BroadcastStateIntervalMs // To broadcast imideatly
	acknowledgeGranted := make(chan string)
    // HallOrderConfirmationActive := false
	var activePeers []string
    

	for {
		select {
		case peerUpdate := <-peerUpdatesRx:
			activePeers = peerUpdate.Peers
			if peerUpdate.New != "" {
				slog.Info("[peerUpdate]: Adding elevator", slog.String("ID", peerUpdate.New))
				addElevator(peerUpdate.New)
				// TODO: distribute

			}

			for i := 0; i < len(peerUpdate.Lost); i++ {
				slog.Info("[peerUpdate]: Removing elevator", slog.String("ID", peerUpdate.Lost[i]))
				removeElevator(peerUpdate.Lost[i])
			}

			// Syncroniz incoming state, or ack
		case incommingStateMessage := <-broadcastStateMessageRx:
			// fmt.Println("[syncronizer]: updated elevator ", incommingStateMessage.Id, " to ", incommingStateMessage.State)
			if !incommingStateMessage.UpdateOrders {
				slog.Info("[Broadcast<-]: Sync", slog.String("elevator", incommingStateMessage.Id))
				updateElevator(incommingStateMessage.Id, incommingStateMessage.State)
				// If HallRequests is broadscasted without Update orders flag it is valid.
				updateHallRequests(incommingStateMessage.HallRequests)
				continue
			}
			slog.Info("[Broadcast<-]: updateOrders=", incommingStateMessage.UpdateOrders)

            // Recived an ack
			if incommingStateMessage.Id == mainID{
                if !validAck(incommingStateMessage){
                    continue
                }
				slog.Info("[Broadcast<-] Registred ack", slog.String("checksum", string(incommingStateMessage.Checksum)))
                acknowledgeGranted <- mainID // FIXME: This deadlocks on the reciving end. (if waitForHallOrderConfirmation is not running)
				continue
			}
			// we ack other elevators's request
			incommingStateMessage.Id = mainID
			incommingStateMessage.Checksum, _ = HashStructSha1(incommingStateMessage)
			slog.Info("[Broadcast<-]: Acking message from", slog.String("ID", incommingStateMessage.Id), slog.String("Answer checksum", string(incommingStateMessage.Checksum)))
			broadcastStateMessageTx <- incommingStateMessage

			// assuming here we only get hall calls on this channel
		case button := <-buttonEvent:
			if button.Button == elevio.BT_Cab {
				continue
			}
			fmt.Println("[ButtonPress]: floor", button.Floor, "direction", button.Button)
			newStateMessageOrder := makeBroadcastStateMessage(mainID)
			newStateMessageOrder.UpdateOrders = true
			newStateMessageOrder.HallRequests[button.Floor][button.Button] = true
			fmt.Println("[request] sending out hall request, ", newStateMessageOrder)

			go waitForHallOrderConfirmation(mainID, button, activePeers, acknowledgeGranted)
            broadcastStateMessageTx <- newStateMessageOrder

			// broadcast state
		default:
			if time.Now().UnixMilli() < lastStateBroadcast+config.BroadcastStateIntervalMs {
				continue
			}

			newStateMessage := makeBroadcastStateMessage(mainID)
			// fmt.Println("[syncronizer]: Broadcasting, Checksum: ", newStateMessage.Checksum)
			broadcastStateMessageTx <- newStateMessage
			lastStateBroadcast = time.Now().UnixMilli()
		}
	}
}
