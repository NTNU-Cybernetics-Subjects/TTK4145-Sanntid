
package main

import (
    . "fmt"
    "net"
)


func main() {

    res := make([]byte, 1024)

    server, err := net.ResolveTCPAddr("tcp", "localhost:33546")
    if err != nil {
        Println(err.Error())
    }

    conn, err := net.DialTCP("tcp", nil, server)
    if err != nil {
        Println(err.Error())
    }
    defer conn.Close()

    _, err = conn.Read(res)
    if err != nil {
        Println(err.Error())
    }
    Println(string(res))


    // _, err = conn.Write([]byte("test"))
    // if err != nil {
    //     Println(err.Error())
    // }

    // // res = make([]byte, 1024)
    // _, err = conn.Read(res)
    // if err != nil {
    //     Println(err.Error())
    // }
    // Println(string(res))


}
