
package main

import (
	. "fmt"
	"time"
	// "runtime"
	// "bufio"
	"net"
)

const HOST string = "localhost"
const PORT string = "20000"

func send_udp(msg string){ 
    Address := HOST + ":" + PORT
    conn, err := net.Dial("udp", Address)
    if err != nil {
        Println(err)
    }
    payload := []byte(msg)
    conn.Write(payload)
    defer conn.Close()

    // Println("trying to send exit")
}

func read_udp(){

    udpServer, err := net.ListenPacket("udp", "localhost:20001")
    if err != nil {
        Println(err.Error())
    }
    defer udpServer.Close()

    for {
        buf := make([]byte, 1024)
        _, addr, err := udpServer.ReadFrom(buf)
        if err != nil {
            continue
        }
        Println("msg from: ", addr, " payload: ", string(buf))
    }

}



func main(){
 
    // Address := HOST + ":" + PORT
    // Println(Address)
    // conn, err := net.Dial("udp4", Address)
    // if err != nil {
    //     Println(err)
    // }
    // payload := []byte("This is  me")
    // conn.Write(payload)

    // Close connection when exeting main
    // defer conn.Close()

    // conn.Write([]byte ("From antoher place"))

    // conn.SetReadDeadline(time.Now().Add(5 + time.Second))
    // b := make([]byte, 1024)
    // n, err := conn.Read(b)
    // Println("n: ", n, " err: ", err)
    // finished := make(chan bool)
    go send_udp("Testing this from new time in the history \n")
    go read_udp()
    
    time.Sleep(3 * time.Second)

}

