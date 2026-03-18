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
		OrderStates:   types.OrderTable{},
	}
}

func applyButtonPress(
	state types.LocalSystemState,
	btn types.ButtonEvent,
) types.LocalSystemState {
	if state.OrderStates[btn.Floor][btn.Button] == types.OrderStandby {
		state.OrderStates[btn.Floor][btn.Button] = types.OrderPending
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
	state.ElevatorState = newReplicatedElevatorState(event)

	for floor := range config.NFloors {
		for btn := types.BtnHallUp; btn <= types.BtnCab; btn++ {
			if !event.CompletedOrders[floor][btn] {
				continue
			}
			state.OrderStates[floor][btn] = types.OrderComplete
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
		for btn := types.BtnHallUp; btn <= types.BtnCab; btn++ {
			state.OrderStates[floor][btn] = mergeConvergedOrderState(
				state.OrderStates[floor][btn],
				convergedState.OrderTables[state.ElevatorID][floor][btn],
			)
		}
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
