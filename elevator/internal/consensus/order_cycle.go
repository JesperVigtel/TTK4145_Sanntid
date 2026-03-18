package consensus

import (
	"elevator/internal/config"
	"elevator/internal/types"
	"slices"
)

// -----------------------------------------------------------------------------
// Advances the unified order tables through the shared cyclic state machine:
// Standby -> Pending -> Assigned -> Complete -> Standby.
// Hall buttons advance in the local peer's self row, while cab buttons advance
// for every owner row from the peer snapshots of that owner.
// -----------------------------------------------------------------------------
func advanceOrderStates(
	orderTables [config.NElevators]types.OrderTable,
	peerOrderSnapshots [config.NElevators][config.NElevators]types.OrderTable,
	selfID int,
	peerIsAlive [config.NElevators]bool,
) [config.NElevators]types.OrderTable {
	for ownerID := range config.NElevators {
		for floor := range config.NFloors {
			for btn := types.BtnHallUp; btn <= types.BtnCab; btn++ {
				if btn != types.BtnCab && ownerID != selfID {
					continue
				}
				peerStates := alivePeerOrderStates(selfID, peerIsAlive, orderTables, peerOrderSnapshots, ownerID, floor, btn)
				orderTables[ownerID][floor][btn] = nextOrderState(orderTables[ownerID][floor][btn], peerStates)
			}
		}
	}
	return orderTables
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
	orderTables [config.NElevators]types.OrderTable,
	peerOrderSnapshots [config.NElevators][config.NElevators]types.OrderTable,
	ownerID int,
	floor int,
	btn types.ButtonType,
) []types.OrderState {
	var peerStates []types.OrderState
	for peerID := range config.NElevators {
		if peerID == selfID || !peerIsAlive[peerID] {
			continue
		}
		if btn == types.BtnCab {
			peerStates = append(peerStates, peerOrderSnapshots[peerID][ownerID][floor][btn])
			continue
		}
		peerStates = append(peerStates, orderTables[peerID][floor][btn])
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
