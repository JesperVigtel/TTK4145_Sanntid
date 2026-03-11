package localControl

import (
	"elevator/internal/config"
	"elevator/internal/types"
)

func newElevator() types.Elevator {
	return types.Elevator{
		CurrentFloor:           -1,
		CurrentTravelDirection: types.Down,
		PhysicalMotorDirection: types.Down,
		LocalOrders:            [config.NFloors][config.NButtons]bool{},
		Behaviour:              types.ElevatorMoving,
		ActiveStatus:           true,
	}
}

func updateActiveStatus(elevator *types.Elevator, obstruction bool) {
	if elevator.Behaviour == types.ElevatorDoorOpen && obstruction {
		elevator.ActiveStatus = false
	} else {
		elevator.ActiveStatus = true
	}
}
