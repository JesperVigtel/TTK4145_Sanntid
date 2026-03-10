package consensus

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
	"reflect"
)

func newSystemHallOrders() [NElevators]HallOrderTable {
	var table [NElevators]HallOrderTable
	for peerID := range table {
		table[peerID] = newStandbyHallOrders()
	}
	return table
}

func newStandbyHallOrders() HallOrderTable {
	var table HallOrderTable
	for floor := range table {
		for btn := range table[floor] {
			table[floor][btn] = OrderStandby
		}
	}
	return table
}

func updatePeerAvailability(
	nodeRegistry GlobalNodeRegistry,
	peerIsAlive [NElevators]bool,
	systemHallOrders [NElevators]HallOrderTable,
	selfID int,
) ([NElevators]bool, [NElevators]HallOrderTable) {

	for _, peerID := range nodeRegistry.Lost {
		if peerID < 0 || peerID >= NElevators {
			continue
		}
		peerIsAlive[peerID] = false
		// Promote the lost peer's slot to match surviving self-state for active
		// orders so that advanceLocalOrderStates never sees a divergence and
		// silently erases an Assigned/Pending call.
		systemHallOrders[peerID] = redistributeOnLoss(systemHallOrders, selfID)
	}

	for _, peerID := range nodeRegistry.New {
		if peerID < 0 || peerID >= NElevators {
			continue
		}
		peerIsAlive[peerID] = true
	}
	return peerIsAlive, systemHallOrders
}

// redistributeOnLoss builds a replacement HallOrderTable for a lost peer.
// For each (floor, btn) where the surviving self holds an active order
// (OrderPending or OrderAssigned), that value is copied into the peer slot
// so the divergence check in advanceLocalOrderStates never fires.
// All other slots are set to OrderStandby.
func redistributeOnLoss(
	systemHallOrders [NElevators]HallOrderTable,
	selfID int,
) HallOrderTable {
	var table HallOrderTable
	for floor := range table {
		for btn := range table[floor] {
			selfState := systemHallOrders[selfID][floor][btn]
			if isActiveOrder(selfState) {
				table[floor][btn] = selfState
			} else {
				table[floor][btn] = OrderStandby
			}
		}
	}
	return table
}

// mergeIncomingHallOrders prevents a rejoining peer's stale Standby from
// overwriting an active OrderAssigned or OrderPending on the surviving node.
// For every (floor, btn) where self holds an active order and the incoming
// message carries OrderStandby, the current recorded peer value is replaced
// with the self state so consensus can continue without a divergence reset.
func mergeIncomingHallOrders(
	selfOrders HallOrderTable,
	incomingOrders HallOrderTable,
) HallOrderTable {
	var merged HallOrderTable
	for floor := range merged {
		for btn := range merged[floor] {
			selfState := selfOrders[floor][btn]
			incoming := incomingOrders[floor][btn]
			if isActiveOrder(selfState) && incoming == OrderStandby {
				// Block the stale Standby; mirror the self state so the peer
				// slot is consistent and no divergence reset occurs.
				merged[floor][btn] = selfState
			} else {
				merged[floor][btn] = incoming
			}
		}
	}
	return merged
}

// isActiveOrder reports whether an order is in an active (in-progress) state
// that must not be silently reset when a peer is lost or rejoins.
func isActiveOrder(state OrderState) bool {
	return state == OrderPending || state == OrderAssigned
}

func peerStateMatchesRecorded(
	msg Message,
	systemHallOrders [NElevators]HallOrderTable,
	systemElevStates [NElevators]HRAElevState,
) bool {
	return reflect.DeepEqual(systemHallOrders[msg.SenderID], msg.HallOrderTable) &&
		reflect.DeepEqual(systemElevStates[msg.SenderID], msg.ElevatorList[msg.SenderID])
}

func allAlivePeersConsistent(
	peerIsConsistent 	[NElevators]bool,
	peerIsAlive 		[NElevators]bool,
	selfID 				int,
) bool {
	for peerID := range NElevators {
		if peerID == selfID {
			continue
		}
		if peerIsAlive[peerID] && !peerIsConsistent[peerID] {
			return false
		}
	}
	return true
}

func sendStateUpdate(
	outgoingMessages 	chan<- Message,
	selfID 				int,
	peerIsAlive 		[NElevators]bool,
	systemElevStates 	[NElevators]HRAElevState,
	systemHallOrders 	[NElevators]HallOrderTable,
) {
	select {
	case outgoingMessages <- Message{
		SenderID:      selfID,
		ElevatorList:  systemElevStates,
		HallOrderTable: systemHallOrders[selfID],
		AliveStatus:   peerIsAlive[selfID],
		AliveList:     peerIsAlive,
	}:
	default:
	}
}

func publishConsistantState(
	convergedSystemState 	chan<- ConvergedSystemState,
	peerIsAlive 			[NElevators]bool,
	systemElevStates 		[NElevators]HRAElevState,
	systemHallOrders 		[NElevators]HallOrderTable,
) {
	state := ConvergedSystemState{
		AliveList:      peerIsAlive,
		ElevatorList:   systemElevStates,
		HallOrderTable: systemHallOrders,
	}
	select {
	case convergedSystemState <- state:
	default:
	}
}
