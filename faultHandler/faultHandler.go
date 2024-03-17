package faulthandler

import (
	"elevator/config"
	"elevator/fsm"
	"time"
)

func restartSystem() {

}

func disconnectFromNetwork() {

}

func reconnectToNetwork() {

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

func CheckObstruction(elevatorObstructionChan <-chan bool) {
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
					reconnectToNetwork()
					onlineStatus = true
				}
			}

		case <-timer.C:
			if alreadyObstructed {
				disconnectFromNetwork()
				onlineStatus = false
			}
		}
	}
}