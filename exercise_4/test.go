package main

import (
    "time"
    "fmt"
)

func main() {

    now := time.Now().UnixMilli()
    for {
        
        fmt.Println(time.Now().UnixMilli() - now)
    }
}
