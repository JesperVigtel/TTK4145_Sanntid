package dispatch

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

func initLocalSystemState(
	event		ElevatorEvents,
	elevatorID 	int,
) LocalSystemState {
	return LocalSystemState{
		ElevatorID:    elevatorID,
		AliveStatus:   event.Elevator.ActiveStatus,
		ElevatorState: toHRAElevState(event.Elevator),
		HallRequests:  HallOrderTable{},
	}
}


func applyButtonPress(
	state LocalSystemState,
	btn   ButtonEvent,
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



func restoreOwnCabsFromNetwork(
	localState     LocalSystemState,
	convergedState ConvergedSystemState,
) LocalSystemState {
	networkCabs := convergedState.ElevatorList[localState.ElevatorID].CabRequests
	for floor := range len(networkCabs) {
		if networkCabs[floor] {
			localState.ElevatorState.CabRequests[floor] = true
		}
	}
	return localState
}

func applyHardwareUpdate(
	state 	LocalSystemState,
	event   ElevatorEvents,
) LocalSystemState {
	updatedElevState             := toHRAElevState(event.Elevator)
	updatedElevState.CabRequests  = state.ElevatorState.CabRequests
	state.ElevatorState           = updatedElevState
	state.AliveStatus             = event.Elevator.ActiveStatus

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