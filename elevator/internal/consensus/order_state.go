package consensus

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
	"slices"
)

// -----------------------------------------------------------------------------
// Order state based on a cyclic counter:
// Standby -> Pending -> Assigned -> Complete -> Standby
// -----------------------------------------------------------------------------

func advanceLocalOrderStates(
	systemHallOrders [NElevators]HallOrderTable,
	selfID int,
	peerIsAlive [NElevators]bool,
) [NElevators]HallOrderTable {
	for floor := range NFloors {
		for btn := range NButtons {
			systemHallOrders[selfID][floor][btn] = computeNextOrderState(
				systemHallOrders, floor, btn, selfID, peerIsAlive,
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
	current := systemHallOrders[selfID][floor][btn]
	peerStates := alivePeerOrderStates(systemHallOrders, floor, btn, selfID, peerIsAlive)

	next, advanced := tryCyclicAdvance(current, peerStates)
	if advanced {
		return next
	}

	if peerStatesHaveDiverged(current, peerStates) {
		return OrderStandby
	}

	return current
}

func tryCyclicAdvance(current OrderState, peerStates []OrderState) (OrderState, bool) {
	// With no alive peers, advance unconditionally: there is no one to wait for.
	alone := len(peerStates) == 0

	switch current {
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
	return current, false
}

func peerStatesHaveDiverged(current OrderState, peerStates []OrderState) bool {
	switch current {
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

func allAreEither(peerStates []OrderState, allowedA, allowedB OrderState) bool {
	for _, state := range peerStates {
		if state != allowedA && state != allowedB {
			return false
		}
	}
	return true
}
