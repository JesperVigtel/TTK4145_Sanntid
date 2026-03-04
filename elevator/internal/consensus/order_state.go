package consensus

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

func advanceLocalOrderStates(
	systemHallOrders [NElevators]HallOrderTable,
	selfID int,
	peerIsAlive [NElevators]bool,
) [NElevators]HallOrderTable {
	for floor := 0; floor < NFloors; floor++ {
		for btn := 0; btn < NButtons; btn++ {
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
	switch current {
	case OrderStandby:
		if allAreEither(peerStates, OrderStandby, OrderPending) && anyIs(peerStates, OrderPending) {
			return OrderPending, true
		}
	case OrderPending:
		if allAreEither(peerStates, OrderPending, OrderAssigned) {
			return OrderAssigned, true
		}
	case OrderAssigned:
		if allAreEither(peerStates, OrderAssigned, OrderComplete) && anyIs(peerStates, OrderComplete) {
			return OrderComplete, true
		}
	case OrderComplete:
		if allAreEither(peerStates, OrderComplete, OrderStandby) {
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
	// Only include alive peers. A dead node frozen mid-cycle would otherwise
	// permanently block cycle progression until the registry event fires.
	var peerStates []OrderState
	for id := 0; id < NElevators; id++ {
		if id != selfID && peerIsAlive[id] {
			peerStates = append(peerStates, systemHallOrders[id][floor][btn])
		}
	}
	return peerStates
}

func allAreEither(peerStates []OrderState, a, b OrderState) bool {
	for _, s := range peerStates {
		if s != a && s != b {
			return false
		}
	}
	return true
}

func anyIs(peerStates []OrderState, target OrderState) bool {
	for _, s := range peerStates {
		if s == target {
			return true
		}
	}
	return false
}
