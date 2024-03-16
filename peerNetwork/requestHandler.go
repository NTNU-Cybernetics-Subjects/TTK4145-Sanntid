package peerNetwork

import (
	"Driver-go/elevio"
	"elevator/config"
	"log/slog"
	"time"
    "elevator/orders"
)

type RequestChan struct {
	Transmitt chan RequestMessage
	Receive   chan RequestMessage
}

type RequestMessage struct {
	Id                string
	Requestor         string
	Checksum          []byte
	Order             orders.Order
	Sequence          int64
	ProposeUpdateFlag bool
}



var (
	lastRequestMessage     map[string]RequestMessage = make(map[string]RequestMessage)
	requestMessageSequence int64                     = 0
)

func makeNewRequestMessage(newOrder orders.Order) RequestMessage {
	newStateMessage := RequestMessage{
		Id:                config.ElevatorId,
		Requestor:         config.ElevatorId,
		Checksum:          nil,
		Order:             newOrder,
		Sequence:          requestMessageSequence,
		ProposeUpdateFlag: true,
	}
	newStateMessage.Checksum, _ = Checksum(newStateMessage) // NOTE: from checksum
	requestMessageSequence += 1
	return newStateMessage
}

func makeAcknowledgeMessage(incommingRequest RequestMessage) RequestMessage {
	incommingRequest.Id = config.ElevatorId
	incommingRequest.Sequence = requestMessageSequence
	incommingRequest.Checksum, _ = Checksum(incommingRequest) // NOTE: from checksum
	requestMessageSequence += 1                               // FIXME: not sure if we should increment acknowledgement seq number (mabye its own sequence?)
	return incommingRequest
}

// TODO:
func validAck(message RequestMessage) bool {
	return true
}

func Handler(
	requestBcast RequestChan,
	buttonEvent <-chan elevio.ButtonEvent,
    clearOrderChan <- chan elevio.ButtonEvent,
    singalAssignChan chan <- bool,
) {
	slog.Info("[Handler]: starting")
	acknowlegeToTransaction := make(chan RequestMessage)
	startTransaction := make(chan orders.Order)

	// NOTE: we can use same transmitter because there is no problem with two senders.
	go Transaction(startTransaction, acknowlegeToTransaction, requestBcast.Transmitt, singalAssignChan)

	for {
		select {
		case incommingRequest := <-requestBcast.Receive:

			// FIXME: DEBUG.. remove
			// slog.Info("[Handler] incomming request message",
			//     "from", incommingRequest.Id,
			//     "requestor", incommingRequest.Requestor,
			//     "sequence", incommingRequest.Sequence)

			// Ignore our own messages.
			if incommingRequest.Id == config.ElevatorId {
				continue
			}

			// Ack to us.
			if incommingRequest.Requestor == config.ElevatorId {
				if !validAck(incommingRequest) {
					continue
				}
				acknowlegeToTransaction <- incommingRequest
				// slog.Info("[Handler]: ack received, sendt to transaction", "propose", incommingRequest.ProposeUpdateFlag, "from", incommingRequest.Id)
				continue
			}

			// We need to ack
			if incommingRequest.Id == incommingRequest.Requestor {
				if !incommingRequest.ProposeUpdateFlag {
                    // TODO: send to fsm
					slog.Info("[Handler]: commiting order", "propose", incommingRequest.ProposeUpdateFlag, "from", incommingRequest.Requestor)
                    orders.CommitOrder(incommingRequest.Id, incommingRequest.Order) // NOTE: orders.commitOrder
                    singalAssignChan <- true
                    
				}
				acknowlegeMessage := makeAcknowledgeMessage(incommingRequest)
				slog.Info("[Handler]: broadcasting ack",
					"propose", acknowlegeMessage.ProposeUpdateFlag,
					"from", acknowlegeMessage.Id,
					"to", acknowlegeMessage.Requestor)

				requestBcast.Transmitt <- acknowlegeMessage
			}

		case buttonPress := <-buttonEvent:
			newOrder := orders.Order{
				Floor:      buttonPress.Floor,
				ButtonType: buttonPress.Button,
				Operation:  orders.RH_SET,
			}
			slog.Info("[Handler]: new set request registred", "order", newOrder)
			startTransaction <- newOrder

        case clearOrder := <- clearOrderChan:
            newClearOrder := orders.Order{
                Floor: clearOrder.Floor,
                ButtonType: clearOrder.Button,
                Operation: orders.RH_CLEAR,
            }
			slog.Info("[Handler]: new clear request registred", "order", newClearOrder)
            startTransaction <- newClearOrder
		}
	}
}

func registrerAck(activeElevators []string, id string) []string {
	remaningElevators := activeElevators
	for i := 0; i < len(remaningElevators); i++ {
		if remaningElevators[i] != id {
			continue
		}
		remaningElevators[i] = remaningElevators[len(remaningElevators)-1]
		return remaningElevators[:len(remaningElevators)-1]
	}
	return activeElevators
}

func waitForConfirmation(
	requestMessage RequestMessage,
	activePeers []string,
	acknowledgeGranted <-chan RequestMessage,
) bool {
	startTime := time.Now().UnixMilli()
	peersToAck := make([]string, len(activePeers)) // TODO: should we require ack from the same peers on both start transaction and commit transaction? (or the current peers)
	copy(peersToAck, activePeers)                  // make a fresh copy to not alter the input list

	// TODO: check that we acked on the right order

	for {
		select {
		case ackMessage := <-acknowledgeGranted:
			peersToAck = registrerAck(peersToAck, ackMessage.Id) // FIXME: this breaks if len(peersToAck) = zero
			slog.Info("[Transaction] F waitForConfirmation: ack registred", "from", ackMessage.Id, "remaning", peersToAck)

		default:
			// we do not require ack from ourself
			if len(peersToAck) <= 1 {
				return false
			}
			if time.Now().UnixMilli() >= startTime+config.RequestOrderTimeOutMS {
				return true
			}
		}
	}
}

func Transaction(
	newTransaction <-chan orders.Order,
	acknowledgeGranted <-chan RequestMessage,
	transactionBcast chan<- RequestMessage,
    singalAssignChan chan <- bool,
) {
	slog.Info("[Transaction] starting")
	for requestedOrder := range newTransaction {
		activePeers := GetActivePeers() // NOTE: from syncronizer
		requestMessage := makeNewRequestMessage(requestedOrder)

		// request to update
		transactionBcast <- requestMessage
		slog.Info("[Transaction]: proposing update, require acks", "from", activePeers)
		abort := waitForConfirmation(requestMessage, activePeers, acknowledgeGranted)
		if abort {
			slog.Info("[transaction]: aborting proposing", "sequence", requestMessage.Sequence)
			// TODO: abort Transaction
			continue
		}
		// TODO: notify syncronizer?

		requestMessage.ProposeUpdateFlag = false
		// requestMessage.Checksum, _ = peerNetwork.Checksum(requestMessage)
		transactionBcast <- requestMessage

		slog.Info("[Transaction]: proceeding with commit request, require ack", "from", activePeers)
		abort = waitForConfirmation(requestMessage, activePeers, acknowledgeGranted)
		if abort {
			slog.Info("[transaction]: abotring commit")
			// TODO: abort Transaction
			continue
		}

        // TODO: send to fsm
        orders.CommitOrder(config.ElevatorId, requestedOrder) // NOTE: orders.CommitOrder
        singalAssignChan <- true
		slog.Info("[transaction]: order went through, commiting", "order", requestedOrder)
	}
	slog.Info("[transaction] exited") // TODO: error handling
}
