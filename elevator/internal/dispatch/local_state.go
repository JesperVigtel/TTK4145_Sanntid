package dispatch

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

func initLocalSystemState(
	event ElevatorEvents,
	elevatorID int,
) LocalSystemState {
	elevatorState := NewHRAElevState(event.Elevator)
	for floor := range NFloors {
		elevatorState.CabRequests[floor] = event.Elevator.LocalOrders[floor][BtnCab]
	}
	return LocalSystemState{
		ElevatorID:    elevatorID,
		AliveStatus:   event.Elevator.ActiveStatus,
		ElevatorState: elevatorState,
		HallRequests:  HallOrderTable{},
	}
}

func applyButtonPress(
	state LocalSystemState,
	btn ButtonEvent,
) LocalSystemState {

	switch btn.Button {

	case BtnHallUp, BtnHallDown:
		if state.HallRequests[btn.Floor][btn.Button] == OrderStandby {
			state.HallRequests[btn.Floor][btn.Button] = OrderPending
		}

	case BtnCab:
		state.ElevatorState.CabRequests[btn.Floor] = true
	}

	return state
}

func applyHardwareUpdate(
	state LocalSystemState,
	event ElevatorEvents,
) LocalSystemState {
	updatedElevState := NewHRAElevState(event.Elevator)
	for floor := range NFloors {
		updatedElevState.CabRequests[floor] = event.Elevator.LocalOrders[floor][BtnCab]
		if state.ElevatorState.CabRequests[floor] {
			updatedElevState.CabRequests[floor] = true
		}
	}
	state.ElevatorState = updatedElevState
	state.AliveStatus = event.Elevator.ActiveStatus

	for floor := range NFloors {
		for btn := BtnHallUp; btn <= BtnCab; btn++ {
			if !event.CompletedOrder[floor][btn] {
				continue
			}
			switch ButtonType(btn) {

			case BtnHallUp:
				state.HallRequests[floor][BtnHallUp] = OrderComplete

			case BtnHallDown:
				state.HallRequests[floor][BtnHallDown] = OrderComplete

			case BtnCab:
				state.ElevatorState.CabRequests[floor] = false
			}
		}
	}
	return state
}
