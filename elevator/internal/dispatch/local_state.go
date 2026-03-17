package dispatch

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

func initLocalSystemState(
	event ElevatorEvents,
	elevatorID int,
) LocalSystemState {
	return LocalSystemState{
		ElevatorID:    elevatorID,
		AliveStatus:   event.Elevator.ActiveStatus,
		ElevatorState: NewHRAElevState(event.Elevator),
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
	updatedElevState.CabRequests = state.ElevatorState.CabRequests
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

func mergeRecoveredCabRequests(
	state LocalSystemState,
	convergedState ConvergedSystemState,
) (LocalSystemState, bool) {
	if !convergedState.Recovering {
		return state, false
	}

	recoveredCabRequests := MergeCabRequests(
		state.ElevatorState.CabRequests,
		convergedState.ElevatorList[state.ElevatorID].CabRequests,
	)
	if cabRequestsEqual(state.ElevatorState.CabRequests, recoveredCabRequests) {
		return state, false
	}

	state.ElevatorState.CabRequests = recoveredCabRequests
	return state, true
}

func mergeConvergedHallOrders(
	state LocalSystemState,
	convergedState ConvergedSystemState,
	elevatorID int,
) LocalSystemState {
	for floor := range NFloors {
		for btn := range NButtons {
			convergedOrder := convergedState.HallOrderTable[elevatorID][floor][btn]
			localOrder := state.HallRequests[floor][btn]
			if localOrder == OrderComplete && convergedOrder == OrderAssigned {
				continue
			}
			state.HallRequests[floor][btn] = convergedOrder
		}
	}
	return state
}

func makeLightUpdate(
	localState LocalSystemState,
	convergedState ConvergedSystemState,
	elevatorID int,
) ButtonLightUpdate {
	var cabLights [NFloors]bool
	for floor := range NFloors {
		cabLights[floor] = localState.ElevatorState.CabRequests[floor]
	}
	hallLightUpdate := convergedState.HallOrderTable[elevatorID]

	return ButtonLightUpdate{
		HallLights: hallLightUpdate,
		CabLights:  cabLights,
	}
}

func cabRequestsEqual(a, b []bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
