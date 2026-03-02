package decisionMaker

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

func initLocalSystemState(
	hw         LocalElevatorFromDriver,
	elevatorID int,
) LocalSystemState {
	return LocalSystemState{
		AliveStatus:   hw.Elevator.ActiveStatus,
		ElevatorState: toHRAElevState(hw.Elevator),
		HallRequests:  HallOrderTable{},
	}
}


func applyButtonPress(
	state      LocalSystemState,
	elevatorID int,
	btn        ButtonEvent,
) LocalSystemState {

	switch btn.Button {

		case BTHallUp:
			if state.HallRequests[btn.Floor][BTHallUp] == OrderStandby {
				state.HallRequests[btn.Floor][BTHallUp] = OrderPending
			}

		case BTHallDown:
			if state.HallRequests[btn.Floor][BTHallDown] == OrderStandby {
				state.HallRequests[btn.Floor][BTHallDown] = OrderPending
			}

		case BTCab:
			state.ElevatorState.CabRequests[btn.Floor] = true
		}

	return state
}



func applyHardwareUpdate(
	state      LocalSystemState,
	hw         LocalElevatorFromDriver,
	elevatorID int,
) LocalSystemState {
	updatedElevState             := toHRAElevState(hw.Elevator)
	updatedElevState.CabRequests  = state.ElevatorState.CabRequests
	state.ElevatorState           = updatedElevState
	state.AliveStatus             = hw.Elevator.ActiveStatus

	for floor := 0; floor < NFloors; floor++ {
		for btn := BTHallUp; btn <= BTCab; btn++ {
			if !hw.ExecutedOrders[floor][btn] {
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