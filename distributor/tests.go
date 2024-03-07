
package distribitor

import (
    "fmt"
    "time"
    "Network-go/network/peers"
    // "Network-go/network/bcast"
)

func Test_running_peerwatcher(id string){

        // Channel for peer updates
        peerUpdateCh := make(chan peers.PeerUpdate)

        // Turn peer on/off (default on)
        peerEnablCh := make(chan bool)

        go peers.Transmitter(PEER_PORT, id, peerEnablCh)
        go peers.Receiver(PEER_PORT, peerUpdateCh)

        go PeerWatcher(peerUpdateCh)
         
        // print out states each second
		for {

            fmt.Println(peerStates)
            time.Sleep(time.Second * 1)
    }
}

