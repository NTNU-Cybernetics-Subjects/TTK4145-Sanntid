package main

import (
	"Network-go/network/bcast"
	// "Network-go/network/localip"
	// "Network-go/network/peers"
	"flag"
	"fmt"
	// "os"
	"time"
    "os/exec"
)

const PORT int = 1234

type Message struct {
    Id int;
    Count int;
}


func Backup(id int, terminateValue* int){
    // Set up get channel
	varRx := make(chan Message)
	go bcast.Receiver(PORT, varRx)

    lastRevicedPackage := time.Now().UnixMilli() 
    fmt.Println("Started new backup id:", id)

    var count int;
	for {
		select {
		case a := <-varRx:
            count = a.Count
			// fmt.Printf("Received: %#v\n", count)
            lastRevicedPackage = time.Now().UnixMilli()
        default:
            // TODO: need a faster way of detecting connection loss
            if (time.Now().UnixMilli() - lastRevicedPackage) > 700{
                fmt.Println("No connection")
                *terminateValue = count
                // return to stop backup mode and takeover Primary
                return
                
            }
		}
	} 
}

func Primary(id int, start_count int){


    fmt.Println("takeover as Primary")

    // open new backup
    DoShell("go run ~/school/sanntid/exercise_4/main.go -backup 1")


	varTx := make(chan Message)
	go bcast.Transmitter(PORT, varTx)
   
    count := start_count

    msg := Message{id,0}
    for i:=0; i < 10; i++{
        fmt.Println(count)
        msg.Count = count
        varTx <- msg
        count++
        time.Sleep(500 * time.Millisecond)
    }

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

func main(){
    
    var backup int
    flag.IntVar(&backup, "backup", -1, "start a new backup")
    flag.Parse()

    var terminateValue int
    start_count := 0
    id := 0

    if backup != -1{
        Backup(1, &terminateValue)
        start_count = terminateValue
        Primary(1, start_count+1) // This should start a backup
        return
    }

    // if not backup start Primary (Primary always starts a backup)
    Primary(id, start_count)

}
