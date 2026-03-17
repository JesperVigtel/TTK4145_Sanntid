package consensus

import (
	"elevator/internal/config"
	"elevator/internal/types"
	"slices"
)

// -----------------------------------------------------------------------------
// Advances each hall order through the cyclic state machine based on peer agreement.
// Standby -> Pending -> Assigned -> Complete -> Standby
// States that have diverged beyond adjacent steps are reset to Standby.
// -----------------------------------------------------------------------------

func advanceHallOrderStates(
	systemHallOrders [config.NElevators]types.HallOrderTable,
	selfID int,
	peerIsAlive [config.NElevators]bool,
) [config.NElevators]types.HallOrderTable {
	for floor := range config.NFloors {
		for btn := types.BtnHallUp; btn <= types.BtnHallDown; btn++ {
			systemHallOrders[selfID][floor][btn] = nextOrderState(
				systemHallOrders[selfID][floor][btn],
				alivePeerOrderStates(selfID, peerIsAlive, func(peerID int) types.OrderState {
					return systemHallOrders[peerID][floor][btn]
				}),
			)
		}
	}
	return systemHallOrders
}

func advanceCabOrderStates(
	systemElevStates [config.NElevators]types.HRAElevState,
	peerObservedCabOrders [config.NElevators][config.NElevators]types.CabOrderTable,
	selfID int,
	peerIsAlive [config.NElevators]bool,
) [config.NElevators]types.HRAElevState {
	for ownerID := range config.NElevators {
		for floor := range config.NFloors {
			systemElevStates[ownerID].CabOrders[floor] = nextOrderState(
				systemElevStates[ownerID].CabOrders[floor],
				alivePeerOrderStates(selfID, peerIsAlive, func(peerID int) types.OrderState {
					return peerObservedCabOrders[peerID][ownerID][floor]
				}),
			)
		}
	}
	return systemElevStates
}

func nextOrderState(
	selfState types.OrderState,
	peerStates []types.OrderState,
) types.OrderState {
	nextState, advanced := tryCyclicAdvance(selfState, peerStates)
	if advanced {
		return nextState
	}
	if peerStatesHaveDiverged(selfState, peerStates) {
		return types.OrderStandby
	}
	return selfState
}

func tryCyclicAdvance(currentState types.OrderState, peerStates []types.OrderState) (types.OrderState, bool) {
	alone := len(peerStates) == 0

	switch currentState {
	case types.OrderStandby:
		if allPeerStatesAreEither(peerStates, types.OrderStandby, types.OrderPending) && slices.Contains(peerStates, types.OrderPending) {
			return types.OrderPending, true
		}
		if slices.Contains(peerStates, types.OrderAssigned) {
			return types.OrderPending, true
		}
	case types.OrderPending:
		if alone || allPeerStatesAreEither(peerStates, types.OrderPending, types.OrderAssigned) {
			return types.OrderAssigned, true
		}
	case types.OrderAssigned:
		if allPeerStatesAreEither(peerStates, types.OrderAssigned, types.OrderComplete) &&
			slices.Contains(peerStates, types.OrderComplete) {
			return types.OrderComplete, true
		}
	case types.OrderComplete:
		if alone || allPeerStatesAreEither(peerStates, types.OrderComplete, types.OrderStandby) {
			return types.OrderStandby, true
		}
	}
	return currentState, false
}

func peerStatesHaveDiverged(selfState types.OrderState, peerStates []types.OrderState) bool {
	switch selfState {
	case types.OrderStandby:
		return !allPeerStatesAreEither(peerStates, types.OrderStandby, types.OrderPending) &&
			!allPeerStatesAreEither(peerStates, types.OrderStandby, types.OrderComplete)
	case types.OrderPending:
		return !allPeerStatesAreEither(peerStates, types.OrderPending, types.OrderAssigned) &&
			!allPeerStatesAreEither(peerStates, types.OrderStandby, types.OrderPending)
	case types.OrderAssigned:
		return !allPeerStatesAreEither(peerStates, types.OrderAssigned, types.OrderComplete) &&
			!allPeerStatesAreEither(peerStates, types.OrderAssigned, types.OrderPending) &&
			!allPeerStatesAreEither(peerStates, types.OrderAssigned, types.OrderStandby)
	case types.OrderComplete:
		return !allPeerStatesAreEither(peerStates, types.OrderComplete, types.OrderStandby) &&
			!allPeerStatesAreEither(peerStates, types.OrderAssigned, types.OrderComplete)
	}
	return false
}

func alivePeerOrderStates(
	selfID int,
	peerIsAlive [config.NElevators]bool,
	peerStateAt func(peerID int) types.OrderState,
) []types.OrderState {
	var peerStates []types.OrderState
	for peerID := range config.NElevators {
		if peerID != selfID && peerIsAlive[peerID] {
			peerStates = append(peerStates, peerStateAt(peerID))
		}
	}
	return peerStates
}

func allPeerStatesAreEither(peerStates []types.OrderState, stateA, stateB types.OrderState) bool {
	for _, state := range peerStates {
		if state != stateA && state != stateB {
			return false
		}
	}
	return true
}
