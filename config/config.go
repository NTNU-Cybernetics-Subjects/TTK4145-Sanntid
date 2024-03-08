
package config

const (
	PEER_PORT  int    = 12348
	HOST       string = "localhost"
	BCAST_PORT int    = 4875
)

const (
	n_elevators int = 3
	NumberFloors    int = 4
)

const (
    BroadcastStateIntervalMs int64 = 800 // ms
)

var HallRequestAssignerExecutable string = "../bin/hall_request_assigner"
