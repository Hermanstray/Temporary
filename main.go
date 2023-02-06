package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

func main() {

	var numFloors int = 4

	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Down
	var current_floor int
	var target_floor int

	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_elevator := make(chan elevio.ButtonEvent, 50)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	order_matrix := [4][3]bool{}

	//go controller(channel)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go check_matrix(&order_matrix, drv_elevator, &d)

	for {
		//fmt.Println(order_matrix)
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			order_matrix[a.Floor][a.Button] = true
			fmt.Println(order_matrix)

		case a := <-drv_elevator:
			target_floor = a.Floor
			fmt.Println("Target floor: ", target_floor)
			switch {
			case target_floor > current_floor:
				d = elevio.MD_Up
			case target_floor < current_floor:
				d = elevio.MD_Down
			case target_floor == current_floor:
				d = elevio.MD_Stop
			}
			fmt.Println("Motor direction: ", d)
			elevio.SetMotorDirection(d)

		case a := <-drv_floors:
			current_floor = a
			switch {
			case a == target_floor:
				d = elevio.MD_Stop
				for i := range order_matrix[a] {
					order_matrix[a][i] = false
				}
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

func check_matrix(matrix *[4][3]bool, receiver chan<- elevio.ButtonEvent, dir *elevio.MotorDirection) {
	for {
		for i := range matrix {
			for k := range matrix[i] {
				if matrix[i][k] == true {
					//fmt.Println("Found true at: ", "(", i, " ", k, ")")
					var event elevio.ButtonEvent
					event.Floor = i
					event.Button = int_to_buttontype(k)
					if *dir == elevio.MD_Stop {
						receiver <- event

					}
					time.Sleep(20 * time.Millisecond)
				}
			}
		}
	}
}

func int_to_buttontype(i int) elevio.ButtonType {
	var event elevio.ButtonType
	switch {
	case i == 0:
		event = elevio.BT_HallUp
		return event
	case i == 1:
		event = elevio.BT_HallDown
		return event
	}
	event = elevio.BT_Cab
	return event
}
