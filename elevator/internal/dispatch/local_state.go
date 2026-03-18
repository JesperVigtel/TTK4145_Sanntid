package dispatch

import (
	"elevator/internal/config"
	"elevator/internal/types"
)

func newReplicatedElevatorState(event types.ElevatorEvents) types.HRAElevState {
	return types.NewHRAElevState(event.Elevator, isAvailableForAssignment(event))
}

func isAvailableForAssignment(event types.ElevatorEvents) bool {
	return event.Elevator.ActiveStatus && !event.Obstructed
}

func initLocalSystemState(
	event types.ElevatorEvents,
	elevatorID int,
) types.LocalSystemState {
	return types.LocalSystemState{
		ElevatorID:    elevatorID,
		ElevatorState: newReplicatedElevatorState(event),
		OrderStates:   types.OrderTable{},
	}
}

func applyElevatorEvent(
	state types.LocalSystemState,
	event types.ElevatorEvents,
) types.LocalSystemState {
	if event.NewButtonPress != nil {
		button := *event.NewButtonPress
		if state.OrderStates[button.Floor][button.Button] == types.OrderStandby {
			state.OrderStates[button.Floor][button.Button] = types.OrderPending
		}
	}
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
