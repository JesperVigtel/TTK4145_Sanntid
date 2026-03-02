package consensus

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

func stepAllOrderStates(
	fleetHallOrders [NElevators]HallOrderTable,
	nodeID          int,
) [NElevators]HallOrderTable {
	for floor := 0; floor < NFloors; floor++ {
		for btn := 0; btn < NButtons; btn++ {
			fleetHallOrders[nodeID][floor][btn] = stepOrderState(
				fleetHallOrders, floor, btn, nodeID,
			)
		}
	}
	return fleetHallOrders
}

func stepOrderState(
	fleetHallOrders [NElevators]HallOrderTable,
	floor           int,
	btn             int,
	nodeID          int,
) OrderState {
	current    := fleetHallOrders[nodeID][floor][btn]
	peerStates := collectPeerOrderStates(fleetHallOrders, floor, btn, nodeID)

	next, advanced := tryAdvanceState(current, peerStates)
	if advanced {
		return next
	}

	if statesHaveDiverged(current, peerStates) {
		return OrderStandby
	}

	return current
}

func tryAdvanceState(current OrderState, peerStates []OrderState) (OrderState, bool) {
	switch current {
	case OrderStandby:
		if subsetOf(peerStates, OrderStandby, OrderPending) && contains(peerStates, OrderPending) {
			return OrderPending, true
		}
	case OrderPending:
		if subsetOf(peerStates, OrderPending, OrderAssigned) {
			return OrderAssigned, true
		}
	case OrderAssigned:
		if subsetOf(peerStates, OrderAssigned, OrderComplete) && contains(peerStates, OrderComplete) {
			return OrderComplete, true
		}
	case OrderComplete:
		if subsetOf(peerStates, OrderComplete, OrderStandby) {
			return OrderStandby, true
		}
	}
	return current, false
}

func statesHaveDiverged(current OrderState, peerStates []OrderState) bool {
	switch current {
	case OrderStandby:
		return !subsetOf(peerStates, OrderStandby, OrderPending) &&
			!subsetOf(peerStates, OrderStandby, OrderComplete)
	case OrderPending:
		return !subsetOf(peerStates, OrderPending, OrderAssigned) &&
			!subsetOf(peerStates, OrderStandby, OrderPending)
	case OrderAssigned:
		return !subsetOf(peerStates, OrderAssigned, OrderComplete) &&
			!subsetOf(peerStates, OrderAssigned, OrderPending)
	case OrderComplete:
		return !subsetOf(peerStates, OrderComplete, OrderStandby) &&
			!subsetOf(peerStates, OrderAssigned, OrderComplete)
	}
	return false
}

func collectPeerOrderStates(
	fleetHallOrders [NElevators]HallOrderTable,
	floor           int,
	btn             int,
	nodeID          int,
) []OrderState {
	var peerStates []OrderState
	for id := 0; id < NElevators; id++ {
		if id != nodeID {
			peerStates = append(peerStates, fleetHallOrders[id][floor][btn])
		}
	}
	return peerStates
}

func subsetOf(peerStates []OrderState, a, b OrderState) bool {
	for _, s := range peerStates {
		if s != a && s != b {
			return false
		}
	}
	return true
}

func contains(peerStates []OrderState, target OrderState) bool {
	for _, s := range peerStates {
		if s == target {
			return true
		}
	}
	return false
}