package peerNetwork

import (
	"Driver-go/elevio"
	"elevator/config"
	"elevator/orders"
	"log/slog"
	"time"
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

func getLastRequestMessage(id string) RequestMessage {
	return lastRequestMessage[id]
}

func saveLastRequestMessage(id string, message RequestMessage) {
	lastRequestMessage[id] = message
}

func makeNewRequestMessage(newOrder orders.Order) RequestMessage {
	newRequestMessage := RequestMessage{
		Id:                config.ElevatorId,
		Requestor:         config.ElevatorId,
		Checksum:          nil,
		Order:             newOrder,
		Sequence:          requestMessageSequence,
		ProposeUpdateFlag: true,
	}
	newRequestMessage.Checksum, _ = Checksum(newRequestMessage)
	saveLastRequestMessage(config.ElevatorId, newRequestMessage)
	requestMessageSequence += 1
	return newRequestMessage
}

func makeAcknowledgeMessage(incommingRequest RequestMessage) RequestMessage {
	incommingRequest.Id = config.ElevatorId
	incommingRequest.Sequence = requestMessageSequence
	incommingRequest.Checksum, _ = Checksum(incommingRequest)
	requestMessageSequence += 1
	return incommingRequest
}

// TODO:
func validAck(message RequestMessage) bool {
	// checksum, _ := Checksum(message)
	// slog.Info("checksum", "our", checksum, "incomming", incommingChecksum)
	return true
}

func Handler(
	requestBcast RequestChan,
	buttonEvent <-chan elevio.ButtonEvent,
	clearOrderChan <-chan elevio.ButtonEvent,
	signalAssignChan chan<- bool,
) {
	slog.Info("[Handler]: starting")
	acknowlegeToTransaction := make(chan RequestMessage)
	startTransaction := make(chan orders.Order)

	// NOTE: we can use same transmitter because there is no problem with two senders.
	go Transaction(startTransaction, acknowlegeToTransaction, requestBcast.Transmitt, signalAssignChan)

	for {
		select {
		case incommingRequest := <-requestBcast.Receive:

			// Ignore our own messages.
			if incommingRequest.Id == config.ElevatorId {
				continue
			}

			// Ack to us.
			if incommingRequest.Requestor == config.ElevatorId {
				saveLastRequestMessage(incommingRequest.Id, incommingRequest)
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
					orders.CommitOrder(incommingRequest.Id, incommingRequest.Order) // NOTE: orders.commitOrder
					slog.Info("[Handler]: commiting",
						"from", incommingRequest.Requestor,
						"order", incommingRequest.Order,
						"sequence", incommingRequest.Sequence,
						"hallorders", orders.GetHallOrders())
					signalAssignChan <- true
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
			if orders.OrderAllredyActive(newOrder) {
				continue
			}
			slog.Info("[Handler] proposing", "order", newOrder, "sequence", requestMessageSequence)
			startTransaction <- newOrder

		case clearOrder := <-clearOrderChan:
			newClearOrder := orders.Order{
				Floor:      clearOrder.Floor,
				ButtonType: clearOrder.Button,
				Operation:  orders.RH_CLEAR,
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
	transmittRequest chan<- RequestMessage,
	activePeers []string,
	acknowledgeGranted <-chan RequestMessage,
) bool {
	startTime := time.Now().UnixMilli()
	lastSendt := startTime
	peersToAck := make([]string, len(activePeers))
	copy(peersToAck, activePeers)

	for len(acknowledgeGranted) > 0 {
		slog.Info("[waitForConfirmation] flushing old acks")
		<-acknowledgeGranted
	}

	for {
		select {

		case ackMessage := <-acknowledgeGranted:
			if requestMessage.ProposeUpdateFlag != ackMessage.ProposeUpdateFlag {
				slog.Info("[transaction] propose do not mactch in ack")
				continue
			}
			peersToAck = registrerAck(peersToAck, ackMessage.Id)
			slog.Info("[Transaction] F waitForConfirmation: ack registred", "from", ackMessage.Id, "remaning", peersToAck)

		default:
			// we do not require ack from ourself
			if len(peersToAck) <= 1 {
				return false
			}
			if time.Now().UnixMilli() >= startTime+config.RequestOrderTimeOutMS {
				return true
			}
			// spam request until we get through
			if time.Now().UnixMilli() >= lastSendt + 200 {
				transmittRequest <- requestMessage
				lastSendt = time.Now().UnixMilli()
			}
		}
	}
}

func Transaction(
	newTransaction <-chan orders.Order,
	acknowledgeGranted <-chan RequestMessage,
	transactionBcast chan<- RequestMessage,
	singalAssignChan chan<- bool,
) {
	// slog.Info("[Transaction] starting")
	for requestedOrder := range newTransaction {
		activePeers := GetActivePeers() // NOTE: from syncronizer
		requestMessage := makeNewRequestMessage(requestedOrder)

		// request to update
		slog.Info("[Transaction]: proposing update, require acks", "from", activePeers)
		abort := waitForConfirmation(requestMessage, transactionBcast, activePeers, acknowledgeGranted)
		if abort {
			slog.Info("[transaction]: aborting proposing", "sequence", requestMessage.Sequence)
			continue
		}

        commitMessage := makeNewRequestMessage(requestedOrder)
		commitMessage.ProposeUpdateFlag = false

		slog.Info("[Transaction]: proceeding with commit request, require ack", "from", activePeers)
        activePeers = GetActivePeers()
		abort = waitForConfirmation(commitMessage, transactionBcast, activePeers, acknowledgeGranted)
		if abort {
			slog.Info("[transaction]: abotring commit phase 2")
			continue
		}

		orders.CommitOrder(config.ElevatorId, requestedOrder) // NOTE: orders.CommitOrder
		singalAssignChan <- true
		slog.Info("[transaction]: order went through, commiting", "order", requestedOrder, "hallrequests", orders.GetHallOrders())

	}
	slog.Info("[transaction] exited")
}
