
package distribitor

import (
    "fmt"
    "time"
)

func Tester_rutine_peerwatcher(){
    for {
        fmt.Println(peerStates)
        time.Sleep(time.Second * 1)
    }
}
