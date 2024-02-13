package main

import (
	"fmt"
	// "time"
	// "Network-go/network/bcast"
	"Network-go/network/peers"
	"flag"
	"os/exec"
    // "sync"
)

type msg struct{
    Count int;
}


func DoShell(command string){
 
    cmd := exec.Command("gnome-terminal", "--", "bash", "-c", command)

    // Run the command
    err := cmd.Start()
    if err != nil {
        panic(err)
    }

    // Wait for the command to finish
    err = cmd.Wait()
    if err != nil {
        panic(err)
    }
}

func Find(word string, in []string) bool{
    for _, element := range in{

        if element == word {
            return true
        }
    }
    return false
}


func Watcher(id string){

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	fmt.Println("Started")    
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

            if (!Find("primary", p.Peers)){
                DoShell("go run /home/hurodor/school/sanntid/exercise_4/peerTest.go -id primary")

            }
        
		}
	}
}


func main(){
    
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

    fmt.Println("new session")

    Watcher(id)
    
}




