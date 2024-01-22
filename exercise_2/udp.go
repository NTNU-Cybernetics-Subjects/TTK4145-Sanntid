package main

import (
	. "fmt"
	"time"
	// "runtime"
	// "bufio"
	"net"
)

const HOST string = "192.168.0.187"
const PORT string = "20000"

var exit chan bool

func send_udp_test(msg string){
    address := HOST + ":" + PORT
    Println(address)

    conn, err := net.Dial("udp", address)
    if err != nil {
        Println(err.Error())
    }
    payload := []byte(msg)
    conn.Write(payload)


    defer conn.Close()
    exit <- true
}


func main(){
    exit := make(chan bool)

    go send_udp_test("halla")

    time.Sleep(5 * time.Second)
}
