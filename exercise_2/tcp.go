
package main

import (
    . "fmt"
    "net"
    "time"
    // "bufio"
)

const HOST string = "localhost"
const PORT string = "34933"




func tcp_listen(){

    listen, err := net.Listen("tcp", ":" + "6969")
    if err != nil {
        Println(err.Error())
    }
    defer listen.Close()
    conn, err := listen.Accept()
    if err != nil{
        Println(err.Error())
    }
    for {

        response := make([]byte, 1024)
        _, error := conn.Read(response)
        if error != nil {
            Println(err.Error())
        }
        Println(string(response))
        time.Sleep(2 * time.Second)
        
    }

}

func send(conn net.Conn){
    conn, err := net.Dial("tcp", HOST + ":" + PORT)
    if err != nil{
        Println(err.Error())
    }
    defer conn.Close()
    _, write_error := conn.Write([]byte("Connect to: 127.0.0.1:6969\x00"))
    if write_error != nil{
        Println(write_error.Error())
    }
    time.Sleep(2 * time.Second)
    for {
        _, writewrite_error := conn.Write([]byte("newTest\x00"))
        if writewrite_error != nil {
            Println(writewrite_error.Error())
        }
    time.Sleep(2 * time.Second)
    }
}

func recive(conn net.Conn){
    response := make([]byte, 1024)
    _, read_error := conn.Read(response)
    if read_error != nil {
        Println(read_error.Error())
    }
    for {
        _, read_error := conn.Read(response)
        if read_error != nil{
            Println(read_error)
        }
        // time.Sleep(2 * time.Second)
    }

}


func main() {
    conn, err := net.Dial("tcp", HOST + ":" + PORT)
    if err != nil{
        Println(err.Error())
    }

    go send(conn)
    // go recive(conn)
    go tcp_listen()

    // adress := HOST + ":" + PORT
    //
    // conn, err := net.Dial("tcp", adress)
    // if err != nil{
    //     Println(err.Error())
    // }
    // defer conn.Close()
    //
    // _, write_error := conn.Write([]byte("Connect to: 127.0.0.1:34933\x00"))
    // if write_error != nil {
    //     Println(write_error.Error())
    // }
    //
    // response := make([]byte, 1024)
    // _,read_error := conn.Read(response)
    // if read_error != nil {
    //     Println(read_error.Error())
    // }
    // Println(string(response))
    //
    time.Sleep(10 * time.Second)


}
