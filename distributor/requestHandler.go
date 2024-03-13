package distributor

import (
	"Driver-go/elevio"
	"elevator/config"
	"sync"

	// "fmt"
	"log/slog"
	"time"
)

type HallRequestUpdate struct {
	Id        string
	Requestor string
	Checksum  []byte
	Floor     int
	Direction int
	Sequence  int
	Operation HallOperation
}

type HallOperation int

const (
    HRU_NONE HallOperation = 0
	HRU_SET   HallOperation = 1
	HRU_CLEAR               = 2
)

var (
	hallRequestUpdateOverview     map[string]HallRequestUpdate = make(map[string]HallRequestUpdate)
	hallRequestUpdateOverviewLock sync.Mutex
)

func getHallRequestUpdateMessage(id string)HallRequestUpdate{
    hallRequestUpdateOverviewLock.Lock()
    hallReq := hallRequestUpdateOverview[id]
    hallRequestUpdateOverviewLock.Unlock()
    return hallReq
}

func storeHallRequestUpdate(id string, request HallRequestUpdate){
    hallRequestUpdateOverviewLock.Lock()
    hallRequestUpdateOverview[id] = request
    hallRequestUpdateOverviewLock.Unlock()
}

func clearHallRequestUpdateOperationFlag(id string){
    // TODO: check sequence here? what if the message is changed?
    hallRequestUpdateOverviewLock.Lock()
    current := hallRequestUpdateOverview[id]
    current.Operation = HRU_NONE
    hallRequestUpdateOverview[id] = current
    hallRequestUpdateOverviewLock.Unlock()
}


// var activeHallChanges map[string]HallRequestUpdate = make(map[string]HallRequestUpdate)

// var ongoingHallUpdate map[string][2]bool = make(map[string][2]bool)
// var ongoingHallUpdateLock sync.Mutex
//
// func GetOngoingRequest(id string)bool{
//     ongoingHallUpdateLock.Lock()
//     active, exists := ongoingHallUpdate[id]
//     ongoingHallUpdateLock.Unlock()
//     if exists {
//         return active[1]
//     }
//     return false
// }
//
// func SetOngoingRequest(id string) {
//     ongoingHallUpdateLock.Lock()
//     ongoingHallUpdate[id] = [2]bool{true, false}
//     ongoingHallUpdateLock.Unlock()
// }

// func ClearOngoginRequest(id string){
//     ongoingHallUpdateLock.Lock()
//     ongoingHallUpdate[id] = [2]bool{false, false}
//     ongoingHallUpdateLock.Unlock()
// }

// TODO: check sequence, checksum etc
func validAck(message HallRequestUpdate) bool {
	return true
}

func waitForHallOrderConfirmation(
	mainID string,
	buttonEvent elevio.ButtonEvent,
	ackChan <-chan string,
	signalDistributor chan<- bool,
) {
	acknowledgmentsNeeded := len(localPeerStates) - 1
	countAck := 0
	startTime := time.Now().UnixMilli()

	slog.Info("[waitForConfirmation]: Required ", "acks", acknowledgmentsNeeded)
	// TODO: check against the ids? do we want to pass the ack if we get mulitple from same elevator?

	for {
		select {
		case ackID := <-ackChan:
			countAck += 1
			slog.Info("[waitForConfirmation]: got ack", "from", ackID, "count", countAck)

		default:
			if countAck >= acknowledgmentsNeeded {
				setHallRequest(buttonEvent.Floor, int(buttonEvent.Button), true) // FIXME: RaceCondition between syncronizer update?
				slog.Info("[waitForConfirmation]: order got through", "hallRequests", getHallReqeusts())

				// send a signal to distributor that hallRequests is updated // TODO: move this?
                // signalDistributor <- true // FIX:
				// TODO: set on light here?
				return
			}
			if time.Now().UnixMilli() >= startTime+config.HallOrderAcknowledgeTimeOut {
				// Timeout, we drop the request. TODO: we could try a few more times if no response
				slog.Info("[waitForConfirmation]: timed out")
				return
			}
		}
	}
}

func RequestHandler(
	mainID string,
	broadcastRx <-chan HallRequestUpdate,
	broadcastTx chan<- HallRequestUpdate,
	buttonEvent <-chan elevio.ButtonEvent,
	signalDistributor chan<- bool,
) {
	newOrder := HallRequestUpdate{
		Id:        mainID,
		Checksum:  nil,
		Sequence:  0,
		Floor:     0,
		Direction: 0,
		Requestor: mainID,
        Operation: HRU_NONE,

	}

	acknowledgeGranted := make(chan string)
	acknowledgedSequenceNumber := make(map[string]int)

	for {
		select {

		case incommingHallRequest := <-broadcastRx:

			// Dont respond to our own packages
			if incommingHallRequest.Id == mainID {
				continue
			}
			// we are the requestor
			if incommingHallRequest.Requestor == mainID {
				if !validAck(incommingHallRequest) {
					continue
				}
                // slog.Info("sending ack")
				acknowledgeGranted <- incommingHallRequest.Id
				slog.Info("[requestHandler] ack got", "From", incommingHallRequest.Id)
				continue
			}

            lastHallRequest := getHallRequestUpdateMessage(incommingHallRequest.Id)
			if incommingHallRequest.Sequence <= lastHallRequest.Sequence {
				// We allready ack the request with that sequence number from that requestor
                slog.Info("continuing")
				continue
			}

			// acknowledge request.
            storeHallRequestUpdate(incommingHallRequest.Id, incommingHallRequest)
			slog.Info("[requestHanlder]: active request", "from", incommingHallRequest.Id)

			incommingHallRequest.Id = mainID
			incommingHallRequest.Checksum, _ = HashStructSha1(incommingHallRequest)
			broadcastTx <- incommingHallRequest

			acknowledgedSequenceNumber[incommingHallRequest.Requestor] = incommingHallRequest.Sequence // store the sequence number of the acknowleged request.
			slog.Info("[requestHandler] sending ack", "from", mainID, "To", incommingHallRequest.Requestor, "Sequence_number", incommingHallRequest.Sequence)

			// TODO: skip if active (do we want to send out new request each time the button is pressed?)
		case button := <-buttonEvent:
			if button.Button == elevio.BT_Cab {
				continue
			}

			slog.Info("[requestHanlder] hallrequest registred", "floor", button.Floor, "dir", button.Button)
            newOrder.Sequence += 1
			newOrder.Floor = button.Floor
			newOrder.Direction = int(button.Button)
			newOrder.Checksum, _ = HashStructSha1(newOrder)
			newOrder.Requestor = mainID
            newOrder.Operation = HRU_SET
			go waitForHallOrderConfirmation(mainID, button, acknowledgeGranted, signalDistributor)
			broadcastTx <- newOrder
		}
	}
}
