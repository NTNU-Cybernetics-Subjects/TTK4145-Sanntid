package faulthandler

import (
	"elevator/config"
	"elevator/fsm"
	"os"
	"os/exec"
	"time"
)


func restartSystem() {

    command := "go run main -id " + config.ElevatorId + " -port " + config.ElevatorServerPort + " -host " + config.ElevatorServerHost

    cmd := exec.Command("gnome-terminal", "--", "bash", "-c", command)
    // Run the command
    err := cmd.Start()
    if err != nil {
        panic(err)
    }

    // Wait for the command to finish
    err = cmd.Wait()
    if err != nil {
        panic(err)
    }
    os.Exit(-1)

}


func CheckElevatorMotorMalfunction(elevatorBehaviorChan <-chan fsm.ElevatorBehavior) {
	// Initialize
	alreadyMoving := false
	timer := time.NewTimer(time.Hour)
	timer.Stop()
	
	for {
		select {
		case elevatorBehavior := <-elevatorBehaviorChan:
			if elevatorBehavior == fsm.EB_Moving {
				if !alreadyMoving {
					alreadyMoving = true
					timer = time.NewTimer(time.Duration(config.ElevatorMalfunctionTimeMs) * time.Millisecond)
				}
			} else {
				timer.Stop()
				alreadyMoving = false
			}

		case <-timer.C:
			if alreadyMoving {
				restartSystem()
			}
		}
	}
}

func CheckObstruction(elevatorObstructionChan <-chan bool, networkEnabledChan chan <- bool) {
	alreadyObstructed := false
	onlineStatus := true
	timer := time.NewTimer(time.Hour)
	timer.Stop()
	
	for {
		select {
		case obstruction := <-elevatorObstructionChan:
			if obstruction {
				if !alreadyObstructed {
					alreadyObstructed = true
					timer = time.NewTimer(time.Duration(config.ElevatorObstructionTimeMs) * time.Millisecond)
				}
			} else {
				timer.Stop()
				alreadyObstructed = false
				if !onlineStatus {
                    networkEnabledChan <- true
					onlineStatus = true
				}
			}

		case <-timer.C:
			if alreadyObstructed {
                networkEnabledChan <- false
				onlineStatus = false
			}
		}
	}
}
