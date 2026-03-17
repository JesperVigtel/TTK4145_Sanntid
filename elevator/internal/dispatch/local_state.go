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
		ElevatorID: elevatorID,
		// A running node should stay in replication even when it is obstructed
		// or temporarily unable to take new assignments.
		AliveStatus:   true,
		ElevatorState: newReplicatedElevatorState(event),
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
		if state.ElevatorState.CabOrders[btn.Floor] == OrderStandby {
			state.ElevatorState.CabOrders[btn.Floor] = OrderPending
		}
	}
	return state
}

func applyElevatorEvent(
	state LocalSystemState,
	event ElevatorEvents,
) LocalSystemState {
	if event.NewButtonPress != nil {
		state = applyButtonPress(state, *event.NewButtonPress)
	}
	return applyHardwareUpdate(state, event)
}

func applyHardwareUpdate(
	state LocalSystemState,
	event ElevatorEvents,
) LocalSystemState {
	updatedElevState := newReplicatedElevatorState(event)
	updatedElevState.CabOrders = state.ElevatorState.CabOrders
	state.ElevatorState = updatedElevState

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
				state.ElevatorState.CabOrders[floor] = OrderComplete
			}
		}
	}
	return state
}

func newReplicatedElevatorState(event ElevatorEvents) HRAElevState {
	return NewHRAElevState(event.Elevator, isAvailableForAssignment(event))
}

func isAvailableForAssignment(event ElevatorEvents) bool {
	return event.Elevator.ActiveStatus && !event.Obstructed
}

func mergeConvergedOrders(
	state LocalSystemState,
	convergedState ConvergedSystemState,
) LocalSystemState {
	for floor := range NFloors {
		for btn := BtnHallUp; btn <= BtnHallDown; btn++ {
			state.HallRequests[floor][btn] = mergeConvergedOrderState(
				state.HallRequests[floor][btn],
				convergedState.HallOrderTable[state.ElevatorID][floor][btn],
			)
		}
		state.ElevatorState.CabOrders[floor] = mergeConvergedOrderState(
			state.ElevatorState.CabOrders[floor],
			convergedState.ElevatorList[state.ElevatorID].CabOrders[floor],
		)
	}
	return state
}

func mergeConvergedOrderState(localOrder, convergedOrder OrderState) OrderState {
	switch {
	case localOrder == OrderPending && convergedOrder == OrderStandby:
		return localOrder
	case localOrder == OrderComplete &&
		(convergedOrder == OrderPending || convergedOrder == OrderAssigned):
		return localOrder
	default:
		return convergedOrder
	}
}
