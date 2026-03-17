package consensus

import (
	"elevator/internal/config"
	"elevator/internal/types"
	"slices"
)

// -----------------------------------------------------------------------------
// Advances hall and cab orders through the shared cyclic state machine based on
// peer agreement: Standby -> Pending -> Assigned -> Complete -> Standby.
// States that diverge beyond adjacent steps are reset to Standby.
// -----------------------------------------------------------------------------
func advanceHallOrderStates(
	hallOrders [config.NElevators]types.HallOrderTable,
	selfID int,
	peerIsAlive [config.NElevators]bool,
) [config.NElevators]types.HallOrderTable {
	for floor := range config.NFloors {
		for btn := types.BtnHallUp; btn <= types.BtnHallDown; btn++ {
			peerStates := alivePeerHallOrderStates(selfID, peerIsAlive, hallOrders, floor, btn)
			hallOrders[selfID][floor][btn] = nextOrderState(hallOrders[selfID][floor][btn], peerStates)
		}
	}
	return hallOrders
}

func advanceCabOrderStates(
	elevStates [config.NElevators]types.HRAElevState,
	peerCabViews [config.NElevators][config.NElevators]types.CabOrderTable,
	selfID int,
	peerIsAlive [config.NElevators]bool,
) [config.NElevators]types.HRAElevState {
	for ownerID := range config.NElevators {
		for floor := range config.NFloors {
			peerStates := alivePeerCabOrderStates(selfID, peerIsAlive, peerCabViews, ownerID, floor)
			elevStates[ownerID].CabOrders[floor] = nextOrderState(elevStates[ownerID].CabOrders[floor], peerStates)
		}
	}
	return elevStates
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

func alivePeerHallOrderStates(
	selfID int,
	peerIsAlive [config.NElevators]bool,
	hallOrders [config.NElevators]types.HallOrderTable,
	floor int,
	btn types.ButtonType,
) []types.OrderState {
	var peerStates []types.OrderState
	for peerID := range config.NElevators {
		if peerID != selfID && peerIsAlive[peerID] {
			peerStates = append(peerStates, hallOrders[peerID][floor][btn])
		}
	}
	return peerStates
}

func alivePeerCabOrderStates(
	selfID int,
	peerIsAlive [config.NElevators]bool,
	peerCabViews [config.NElevators][config.NElevators]types.CabOrderTable,
	ownerID int,
	floor int,
) []types.OrderState {
	var peerStates []types.OrderState
	for peerID := range config.NElevators {
		if peerID != selfID && peerIsAlive[peerID] {
			peerStates = append(peerStates, peerCabViews[peerID][ownerID][floor])
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
