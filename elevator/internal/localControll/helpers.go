package localControl

import (
	"elevator/internal/config"
	"elevator/internal/types"
)

func elevatorInit() types.Elevator {
	return types.Elevator{
		CurrentFloor:   -1,
		MotorDirection: types.Down,
		Request:        [config.NFloors][config.NButtons]bool{},
		Behaviour:      types.ElevatorMoving,
		ActiveStatus:   true,
	}
}