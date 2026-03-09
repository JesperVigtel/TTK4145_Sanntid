package consensus

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
	"slices"
)

// -----------------------------------------------------------------------------
// Advances each hall order through the cyclic state machine based on peer agreement.
// Standby -> Pending -> Assigned -> Complete -> Standby
// States that have diverged beyond adjacent steps are reset to Standby.
// -----------------------------------------------------------------------------

func advanceLocalOrderStates(
	systemHallOrders [NElevators]HallOrderTable,
	selfID int,
	peerIsAlive [NElevators]bool,
) [NElevators]HallOrderTable {
	for floor := range NFloors {
		for btn := BtnHallUp; btn <= BtnHallDown; btn++ {
			systemHallOrders[selfID][floor][btn] = computeNextOrderState(
				systemHallOrders, floor, int(btn), selfID, peerIsAlive,
			)
		}
	}
	return systemHallOrders
}

func computeNextOrderState(
	systemHallOrders [NElevators]HallOrderTable,
	floor int,
	btn int,
	selfID int,
	peerIsAlive [NElevators]bool,
) OrderState {
	selfState := systemHallOrders[selfID][floor][btn]
	peerStates := alivePeerOrderStates(systemHallOrders, floor, btn, selfID, peerIsAlive)

	next, advanced := tryCyclicAdvance(selfState, peerStates)
	if advanced {
		return next
	}

	if peerStatesHaveDiverged(selfState, peerStates) {
		return OrderStandby
	}

	return selfState
}

func tryCyclicAdvance(currentState OrderState, peerStates []OrderState) (OrderState, bool) {
	alone := len(peerStates) == 0

	switch currentState {
	case OrderStandby:
		if alone || (allAreEither(peerStates, OrderStandby, OrderPending) && slices.Contains(peerStates, OrderPending)) {
			return OrderPending, true
		}
	case OrderPending:
		if alone || allAreEither(peerStates, OrderPending, OrderAssigned) {
			return OrderAssigned, true
		}
	case OrderAssigned:
		if alone || (allAreEither(peerStates, OrderAssigned, OrderComplete) && slices.Contains(peerStates, OrderComplete)) {
			return OrderComplete, true
		}
	case OrderComplete:
		if alone || allAreEither(peerStates, OrderComplete, OrderStandby) {
			return OrderStandby, true
		}
	}
	return currentState, false
}

func peerStatesHaveDiverged(selfState OrderState, peerStates []OrderState) bool {
	switch selfState {
	case OrderStandby:
		return !allAreEither(peerStates, OrderStandby, OrderPending) &&
			!allAreEither(peerStates, OrderStandby, OrderComplete)
	case OrderPending:
		return !allAreEither(peerStates, OrderPending, OrderAssigned) &&
			!allAreEither(peerStates, OrderStandby, OrderPending)
	case OrderAssigned:
		return !allAreEither(peerStates, OrderAssigned, OrderComplete) &&
			!allAreEither(peerStates, OrderAssigned, OrderPending)
	case OrderComplete:
		return !allAreEither(peerStates, OrderComplete, OrderStandby) &&
			!allAreEither(peerStates, OrderAssigned, OrderComplete)
	}
	return false
}

func alivePeerOrderStates(
	systemHallOrders [NElevators]HallOrderTable,
	floor int,
	btn int,
	selfID int,
	peerIsAlive [NElevators]bool,
) []OrderState {
	var peerStates []OrderState
	for peerID := range NElevators {
		if peerID != selfID && peerIsAlive[peerID] {
			peerStates = append(peerStates, systemHallOrders[peerID][floor][btn])
		}
	}
	return peerStates
}

func allAreEither(peerStates []OrderState, stateA, stateB OrderState) bool {
	for _, state := range peerStates {
		if state != stateA && state != stateB {
			return false
		}
	}
	return true
}
