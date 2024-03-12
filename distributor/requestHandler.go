package distributor

import (
	"Driver-go/elevio"
	"elevator/config"
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
}

// TODO: check sequence, checksum etc
func validAck(message HallRequestUpdate) bool {
	return true
}

func waitForHallOrderConfirmation(
	mainID string,
	buttonEvent elevio.ButtonEvent,
	ackChan <-chan string,
    signalDistributor chan <- bool,
) {
	acknowledgmentsNeeded := len(localPeerStates) - 1
	countAck := 0
	startTime := time.Now().UnixMilli()

	slog.Info("[waitForConfirmation]: Required acks: ", acknowledgmentsNeeded)
	// TODO: check against the ids? do we want to pass the ack if we get mulitple from same elevator?

	for {
		select {
        case ackID := <-ackChan:
			countAck += 1
            slog.Info("[waitForConfirmation] got ack from ", ackID, "ack count: ", countAck) // FIXME:

		default:
			if countAck >= acknowledgmentsNeeded {
				setHallRequest(buttonEvent.Floor, int(buttonEvent.Button), true) // FIXME: RaceCondition between syncronizer update?
				slog.Info("[waitForConfirmation] order got through, updating hallRequests.", getHallReqeusts())

                // send a signal to distributor that hallreqeusts is updated
                signalDistributor <- true
				// TODO: set on light here?
				return
			}
			if time.Now().UnixMilli() >= startTime+config.HallOrderAcknowledgeTimeOut {
				// Timeout, we drop the request.
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
    signalDistributor chan <- bool,
) {
	newOrder := HallRequestUpdate{
		Id:        mainID,
		Checksum:  nil,
		Sequence:  0,
		Floor:     0,
		Direction: 0,
		Requestor: mainID,
	}

	acknowledgeGranted := make(chan string)

	for {
		select {

		case incommingHallRequest := <-broadcastRx:

			// Dont respond to our own packages
			if incommingHallRequest.Id == mainID {
				continue
			}
			// we are the requestor
			if incommingHallRequest.Requestor == mainID {
				slog.Info("[requestHandler] got ack", slog.String("From", incommingHallRequest.Id))
				if !validAck(incommingHallRequest) {
					continue
				}
				acknowledgeGranted <- incommingHallRequest.Id
				continue
			}
			slog.Info("[requestHandler] sending ack", slog.String("from", mainID), slog.String("To", incommingHallRequest.Requestor))
            // ack the request
			incommingHallRequest.Id = mainID
			incommingHallRequest.Checksum, _ = HashStructSha1(incommingHallRequest)
			broadcastTx <- incommingHallRequest

        // TODO: skip if active (do we want to send out new request each time the button is pressed?)
		case button := <-buttonEvent:
			if button.Button == elevio.BT_Cab {
				continue
			}

			slog.Info("[requestHanlder] buttonPress registred", slog.Attr{"floor", slog.StringValue(string(button.Floor))}, slog.Attr{"dir", slog.StringValue(string(button.Button))})
			newOrder.Floor = button.Floor
			newOrder.Direction = int(button.Button)
			newOrder.Checksum, _ = HashStructSha1(newOrder)
			newOrder.Requestor = mainID
			go waitForHallOrderConfirmation(mainID, button, acknowledgeGranted, signalDistributor)
			broadcastTx <- newOrder
			newOrder.Sequence += 1
		}
	}
}
