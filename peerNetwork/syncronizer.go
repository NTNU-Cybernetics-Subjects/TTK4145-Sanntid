package peerNetwork

import (
	"Network-go/network/peers"
	"elevator/config"
	"elevator/fsm"
	"time"
    "log/slog"
)

type StateMessageBroadcast struct {
	Id           string
	Checksum     []byte
	State        fsm.ElevatorState
	CabOrders    [config.NumberFloors]bool
	Sequence     int64
	HallRequests [config.NumberFloors][2]bool
}

type StateMessagechan struct {
	Transmitt chan StateMessageBroadcast
	Receive   chan StateMessageBroadcast
}

// newtorkOverwiew
var (
	lastStateMessagRecived map[string]StateMessageBroadcast = make(map[string]StateMessageBroadcast)
	StateMessageSequence   int64
	activePeers            []string
)

func getLastStateMessage(id string)StateMessageBroadcast{
    return lastStateMessagRecived[id]
}

func saveStateMessage(id string, stateMessage StateMessageBroadcast){
    lastStateMessagRecived[id] = stateMessage
}


func GetActivePeers() []string {
	return activePeers
}

func updateActivePeers(newActivePeers []string) {
	activePeers = newActivePeers
}

func makeNewStateMessage() StateMessageBroadcast {
	newStateMessage := StateMessageBroadcast{
		Id:           config.ElevatorId,
		Checksum:     nil,
		State:        fsm.GetElevatorState(),
		CabOrders:    GetCabOrders(config.ElevatorId), // NOTE: from storage
		Sequence:     StateMessageSequence,
		HallRequests: GetHallOrders(), // NOTE: from storage
	}
	newStateMessage.Checksum, _ = Checksum(newStateMessage) // NOTE: from cheksum
	StateMessageSequence += 1
	return newStateMessage
}


func Syncronizer(
	stateMessagechan StateMessagechan,
	peerUpdateRx <-chan peers.PeerUpdate,
) {
    // NOTE: Wait with broadcasting the state to give time to listen other elevators overview of our state
    lastStateMessageSendtMs := time.Now().UnixMilli() + config.BroadcastStateIntervalMs * 3

	for {
		select {
		case peerUpdate := <-peerUpdateRx:

			updateActivePeers(peerUpdate.Peers)

			// adding ourself
			if peerUpdate.New == config.ElevatorId {
                continue
			}
			// new elevator is added
			if peerUpdate.New != "" {
				slog.Info("[peerUpdate]: new elevator connected", slog.String("ID", peerUpdate.New))

                if localPeer, exists := lastStateMessagRecived[peerUpdate.New]; exists{
                    slog.Info("[peerUpdate]: found peer in localStorage", "peer", peerUpdate.New, "Sequence", localPeer.Sequence)
                    stateMessagechan.Transmitt <- localPeer
                }
			}

			if len(peerUpdate.Lost) > 0 {
				slog.Info("[peerUpdate]: lost", "elevator(s)", peerUpdate.Lost)
			}


		case incommingStateMessage := <-stateMessagechan.Receive:

			// Message from us, or our state is sent to us from antoher elevator
			if incommingStateMessage.Id == config.ElevatorId {
                // if incomming newer than our state, overwride with new state,
                if incommingStateMessage.Sequence > StateMessageSequence{
                    slog.Info("[Syncronizer]: incomming newer than our, syncing the state", 
                        "incommingSequence", incommingStateMessage.Sequence, 
                        "ourSequence", incommingStateMessage.Sequence)

                    mergeHallOrders(incommingStateMessage.HallRequests, RH_SET) // NOTE: order.mergeHallOrders
                    // TODO: merge cab

                    stateMessage := makeNewStateMessage()
                    stateMessagechan.Transmitt <- stateMessage
                }

                // NOTE: reset broadcast timer
                lastStateMessageSendtMs = time.Now().UnixMilli()
				continue
			}
            saveStateMessage(incommingStateMessage.Id, incommingStateMessage)


		// broadcast
		default:
            if time.Now().UnixMilli() < lastStateMessageSendtMs + config.BroadcastStateIntervalMs{
                continue
            }
			stateMessage := makeNewStateMessage()
			stateMessagechan.Transmitt <- stateMessage
            saveStateMessage(config.ElevatorId, stateMessage)
            lastStateMessageSendtMs = time.Now().UnixMilli()
            // slog.Info("[Syncronizer] broadcast", "sequence", stateMessage.Sequence)

		}
	}
}
