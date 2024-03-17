package config

// Networking
const (
	PEER_PORT     int    = 12348
	BCAST_PORT    int    = 4875
	ELEVATOR_HOST string = "localhost"

	NumberFloors             int   = 4
	DoorOpenTimeMs           int64 = 3000 //ms
	LightUpdateTimeMs        int   = 100
	CheckClearedOrdersTimeMs int   = 50

	ElevatorMalfunctionTimeMs int = 10000
	ElevatorObstructionTimeMs int = 6000

	BroadcastStateIntervalMs int64 = 200  // ms
	RequestOrderTimeOutMS    int64 = 15000 // ms


)

var HallRequestAssignerExecutable string = "bin/hall_request_assigner"

var ElevatorId string
