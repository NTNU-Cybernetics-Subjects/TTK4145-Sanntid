package config

// Networking
const (
	PEER_PORT  int    = 12348
	BCAST_PORT int    = 4875
	HOST       string = "localhost"

    NumberElevators int = 3
    NumberFloors    int = 4

    BroadcastStateIntervalMs int64 = 800 // ms

)

var HallRequestAssignerExecutable string = "bin/hall_request_assigner"
