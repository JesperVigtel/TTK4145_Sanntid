package dispatch

import (
	"elevator/internal/config"
	"elevator/internal/types"
)

func initLocalSystemState(
	event types.ElevatorEvents,
	elevatorID int,
) types.LocalSystemState {
	return types.LocalSystemState{
		ElevatorID: elevatorID,
		// A running node should stay in replication even when it is obstructed
		// or temporarily unable to take new assignments.
		AliveStatus:   true,
		ElevatorState: newReplicatedElevatorState(event),
		HallRequests:  types.HallOrderTable{},
	}
}

func applyButtonPress(
	state types.LocalSystemState,
	btn types.ButtonEvent,
) types.LocalSystemState {
	switch btn.Button {
	case types.BtnHallUp, types.BtnHallDown:
		if state.HallRequests[btn.Floor][btn.Button] == types.OrderStandby {
			state.HallRequests[btn.Floor][btn.Button] = types.OrderPending
		}
	case types.BtnCab:
		if state.ElevatorState.CabOrders[btn.Floor] == types.OrderStandby {
			state.ElevatorState.CabOrders[btn.Floor] = types.OrderPending
		}
	}
	return state
}

func applyElevatorEvent(
	state types.LocalSystemState,
	event types.ElevatorEvents,
) types.LocalSystemState {
	if event.NewButtonPress != nil {
		state = applyButtonPress(state, *event.NewButtonPress)
	}
	return applyHardwareUpdate(state, event)
}

func applyHardwareUpdate(
	state types.LocalSystemState,
	event types.ElevatorEvents,
) types.LocalSystemState {
	updatedElevState := newReplicatedElevatorState(event)
	updatedElevState.CabOrders = state.ElevatorState.CabOrders
	state.ElevatorState = updatedElevState

	for floor := range config.NFloors {
		for btn := types.BtnHallUp; btn <= types.BtnCab; btn++ {
			if !event.CompletedOrder[floor][btn] {
				continue
			}
			switch btn {
			case types.BtnHallUp:
				state.HallRequests[floor][types.BtnHallUp] = types.OrderComplete
			case types.BtnHallDown:
				state.HallRequests[floor][types.BtnHallDown] = types.OrderComplete
			case types.BtnCab:
				state.ElevatorState.CabOrders[floor] = types.OrderComplete
			}
		}
	}
	return state
}

func newReplicatedElevatorState(event types.ElevatorEvents) types.HRAElevState {
	return types.NewHRAElevState(event.Elevator, isAvailableForAssignment(event))
}

func isAvailableForAssignment(event types.ElevatorEvents) bool {
	return event.Elevator.ActiveStatus && !event.Obstructed
}

func mergeConvergedOrders(
	state types.LocalSystemState,
	convergedState types.ConvergedSystemState,
) types.LocalSystemState {
	for floor := range config.NFloors {
		for btn := types.BtnHallUp; btn <= types.BtnHallDown; btn++ {
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

func mergeConvergedOrderState(localOrder, convergedOrder types.OrderState) types.OrderState {
	switch {
	case localOrder == types.OrderPending && convergedOrder == types.OrderStandby:
		return localOrder
	case localOrder == types.OrderComplete &&
		(convergedOrder == types.OrderPending || convergedOrder == types.OrderAssigned):
		return localOrder
	default:
		return convergedOrder
	}
}
