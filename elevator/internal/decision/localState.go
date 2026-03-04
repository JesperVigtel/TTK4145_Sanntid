package decisionMaker

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

func initLocalSystemState(
	HWEvent         FromLocalToDM,
	elevatorID int,
) LocalSystemState {
	return LocalSystemState{
		ElevatorID:    elevatorID,
		AliveStatus:   HWEvent.Elevator.ActiveStatus,
		ElevatorState: toHRAElevState(HWEvent.Elevator),
		HallRequests:  HallOrderTable{},
	}
}


func applyButtonPress(
	state LocalSystemState,
	btn   ButtonEvent,
) LocalSystemState {

	switch btn.Button {

		case BTHallUp, BTHallDown:
			if state.HallRequests[btn.Floor][btn.Button] == OrderStandby {
				state.HallRequests[btn.Floor][btn.Button] = OrderPending
			}

		case BTCab:
			state.ElevatorState.CabRequests[btn.Floor] = true
		}

	return state
}



func applyHardwareUpdate(
	state LocalSystemState,
	hw    FromLocalToDM,
) LocalSystemState {
	updatedElevState             := toHRAElevState(hw.Elevator)
	updatedElevState.CabRequests  = state.ElevatorState.CabRequests
	state.ElevatorState           = updatedElevState
	state.AliveStatus             = hw.Elevator.ActiveStatus

	for floor := 0; floor < NFloors; floor++ {
		for btn := BTHallUp; btn <= BTCab; btn++ {
			if !hw.CompletedOrder[floor][btn] {
				continue
			}
			switch ButtonType(btn) {
				
				case BTHallUp:
					state.HallRequests[floor][BTHallUp] = OrderComplete

				case BTHallDown:
					state.HallRequests[floor][BTHallDown] = OrderComplete

				case BTCab:
					state.ElevatorState.CabRequests[floor] = false
			}
		}
	}
	return state
}