# Elevator system - TTK4145 Real Time Programming
This project aims to make a optimized elevator system that can handle n collaborating elevators with m floors.

In order to run this project, update the config.go file with the desired number of floors (4 is default).
Then, spin up the desired number of elevator instances (real or simulated), 
and type the following in a separate terminal window for each elevator. Make sure the host and port match with the elevator instances:
```bash
go run main.go -id "elevatorID" -port "networkPort" -host "networkHost"
```


