package distribitor

import (
	"Driver-go/elevio"
)

// We can only process one hall order at a time, keep it simple
type HallOrderMessage struct {
	Id string

	Checksum []byte
	Sequence int
}

// FIXME: do we ack our own orders? is this a little thin ? need to cheeck more tings?
func acknowledgeOrder(mainID string, hallOrder HallOrderMessage, broadcast chan<- HallOrderMessage) string {
	ackID := hallOrder.Id
	hallOrder.Id = mainID

	hallOrder.Checksum, _ = HashStructSha1(hallOrder)
	broadcast <- hallOrder

	return ackID
}

func requestToUpateOrders(mainID string, broadcastChannel chan<- HallOrderMessage, acknowledgeEvent <-chan string) {
}

func distribitor(mainID string, broadcastTx chan<- HallOrderMessage, broadcastRx <-chan HallOrderMessage, buttonPress <-chan elevio.ButtonEvent) {
	acknowledgeEvent := make(chan string)

	for {
		select {
		case newHallOrder := <-broadcastRx:
			ackID := acknowledgeOrder(mainID, newHallOrder, broadcastTx)
			acknowledgeEvent <- ackID

			// TODO: fsm?
		case buttonEvent := <-buttonPress:
			if buttonEvent.Button == elevio.BT_Cab {
				continue
			}
			go requestToUpateOrders(mainID, broadcastTx, acknowledgeEvent)
		}
	}
}
