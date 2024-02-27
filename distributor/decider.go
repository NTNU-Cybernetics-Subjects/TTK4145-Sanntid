package distribitor

// "Network-go/network/peers"
// "Network-go/network/bcast"
// "Driver-go/elevio"
// "fmt"

/*
Decition part of the distributor. The purpose is to decide which peer should,
execute the active request.

The active order should fall on the peer that is closesd to the floor and going
in right direction from which the cab is called. If there are any conflicts
between peers, if should fall on the node with the lowest id.

The decition should happen in the following order:
The node that feels responsible for the order, should broadcast that it want to
take the order, if all nodes acknowledge then the peer broadcast that it is taking
the order. The order is then considered taken.
*/



