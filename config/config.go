package config

// Networking
const (
	PEER_PORT  int    = 12348
	BCAST_PORT int    = 4875
    ELEVATOR_ADDR       string = "localhost:15657"

    NumberElevators int = 3
    NumberFloors    int = 4

    BroadcastStateIntervalMs int64 = 2000 // ms

)

var HallRequestAssignerExecutable string = "bin/hall_request_assigner"
