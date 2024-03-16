package peerNetwork

import (
	"Driver-go/elevio"
	"elevator/config"
	"log/slog"
	"sync"
	"time"
)

type Order struct {
	Floor      int
	ButtonType elevio.ButtonType
	Operation  Operation
}

type Operation int

const (
	RH_NONE  Operation = 0
	RH_SET   Operation = 1
	RH_CLEAR Operation = 2
)

var (
	globalCabOrders map[string][config.NumberFloors]bool = make(map[string][config.NumberFloors]bool)
	cabOrderLock    sync.Mutex
)

var (
	globalHallOrders [config.NumberFloors][2]bool
	HallOrderLock    sync.Mutex
)

func GetHallOrders() [config.NumberFloors][2]bool {
	return globalHallOrders
}

func GetCabOrders(id string) [config.NumberFloors]bool {
	return globalCabOrders[id]
}

// Use with casion
func OverWrideHallOrders(newHallOrders [config.NumberFloors][2]bool) {
	globalHallOrders = newHallOrders
}

func CommitOrder(id string, order Order) {
	active := false

	switch order.Operation {
	case RH_NONE:
		return
	case RH_SET:
		active = true
	case RH_CLEAR:
		active = false

	}

	// Commit cab
	if order.ButtonType == elevio.BT_Cab {
		slog.Info("[orderCommit] order is cab", "active", active, "floor", order.Floor)
		currentCabOrders := GetCabOrders(id)
		currentCabOrders[order.Floor] = active
		globalCabOrders[id] = currentCabOrders
		return
	}
	slog.Info("[orderCommit] otder is hall", "active", active)
	globalHallOrders[order.Floor][order.ButtonType] = active
}

func mergeHallOrders(newHallOrders [config.NumberFloors][2]bool, operation Operation) {
	if operation == RH_SET {

		for floor := range newHallOrders {
			for dir := range newHallOrders[floor] {
				globalHallOrders[floor][dir] = newHallOrders[floor][dir] || globalHallOrders[floor][dir]
			}
		}
		return
	}

	// Clear
	if operation == RH_CLEAR {
		for floor := range newHallOrders {
			for dir := range newHallOrders[floor] {
				globalHallOrders[floor][dir] = globalHallOrders[floor][dir] && newHallOrders[floor][dir]
			}
		}
	}
}

func OrderPrinter() {
	lastPrint := time.Now().UnixMilli()

	for {
		if time.Now().UnixMilli() < lastPrint+1000 {
			continue
		}
		// slog.Info("[orders]: ", "hallorders", GetHallOrders(), config.ElevatorId, GetCabOrders(config.ElevatorId))
		slog.Info("[orders]", "cabOrders", globalCabOrders)
		lastPrint = time.Now().UnixMilli()
	}
}

func AssingerSpoofer(
	sendOrdersChan chan [config.NumberFloors][3]bool,
) {
    var allOrders [config.NumberFloors][3]bool
    time.Sleep(time.Second * 5)
	for {
        time.Sleep(time.Second)
		hallOrders := GetHallOrders()
		for i := 0; i < config.NumberFloors; i++ {
			for j := 0; j < 2; j++ {
				allOrders[i][j] = hallOrders[i][j]
			}
		}
        myCabOrders := GetCabOrders(config.ElevatorId)
        for i := 0; i < config.NumberFloors; i++{
            allOrders[i][2] = myCabOrders[i]
        }
        slog.Info("[assinge] trying to send")
        sendOrdersChan <- allOrders
        slog.Info("[assinger] sending request to fsm", "orders", allOrders)
	}
}
