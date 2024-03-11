package config

// Networking
const (
	PEER_PORT  int    = 12348
	BCAST_PORT int    = 4875
    ELEVATOR_HOST       string = "localhost"

    NumberElevators int = 3
    NumberFloors    int = 4

    BroadcastStateIntervalMs int64 = 2000 // ms
    HallOrderAcknowledgeTimeOut int64 = 3000 // ms

)

var HallRequestAssignerExecutable string = "bin/hall_request_assigner"
