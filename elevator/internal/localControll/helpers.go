package localControl

import {
	"elevator/internal/types"
	"elevator/internal/config"
}

func elevatorInit() Elevator {
	return &Elevator{
		CurrentFloor: -1
		MotorDirection: Down
		Request: [NFloors][NButtons]bool
		Behavior: ElevatorMoving
		ActiveStatus: true
	}
}