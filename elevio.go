package main

import (
	"Driver-go/elevio"
	"fmt"
)

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Up
	elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	var desired_floor int = 0
	var current_floor int = 0

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)

			// Listen for Cab button presses.
			if a.Button == elevio.BT_Cab {
				desired_floor = a.Floor
				fmt.Println("Desired Floor: ", desired_floor)
				fmt.Println("Current Floor: ", current_floor)
				if desired_floor > current_floor {
					d = elevio.MD_Up
					elevio.SetMotorDirection(d)
				} else if desired_floor < current_floor {
					d = elevio.MD_Down
					elevio.SetMotorDirection(d)
				}
			}

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			elevio.SetFloorIndicator(a)
			current_floor = a

			if a == desired_floor {
				d = elevio.MD_Stop
			}
			elevio.SetMotorDirection(d)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}
