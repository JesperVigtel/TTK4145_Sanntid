package consensus

import (
	"elevator/internal/config"
	"elevator/internal/types"
)

func isRemotePeerID(peerID int, selfID int) bool {
	return peerID >= 0 && peerID < config.NElevators && peerID != selfID
}

// func updatePeerAvailability(
// 	nodeRegistry types.GlobalNodeRegistry,
// 	peerIsAlive [config.NElevators]bool,
// 	orderTables [config.NElevators]types.OrderTable,
// 	selfID int,
// ) ([config.NElevators]bool, [config.NElevators]types.OrderTable) {
// 	for _, lostPeerID := range nodeRegistry.Lost {
// 		if !isRemotePeerID(lostPeerID, selfID) {
// 			continue
// 		}
// 		peerIsAlive[lostPeerID] = false
// 		orderTables = clearPeerHallOrders(orderTables, lostPeerID)
// 	}

// 	for _, newPeerID := range nodeRegistry.New {
// 		if !isRemotePeerID(newPeerID, selfID) {
// 			continue
// 		}
// 		peerIsAlive[newPeerID] = true
// 		orderTables = clearPeerHallOrders(orderTables, newPeerID)
// 	}
// 	return peerIsAlive, orderTables
// }

func updatePeerAvailability(
	nodeRegistry types.GlobalNodeRegistry,
	peerIsAlive [config.NElevators]bool,
	orderTables [config.NElevators]types.OrderTable,
	selfID int,
) ([config.NElevators]bool, [config.NElevators]types.OrderTable) {
	for _, lostPeerID := range nodeRegistry.Lost {
		if !isRemotePeerID(lostPeerID, selfID) {
			continue
		}
		peerIsAlive[lostPeerID] = false
		orderTables = clearAllOrdersForPeer(orderTables, lostPeerID)
	}

	for _, newPeerID := range nodeRegistry.New {
		if !isRemotePeerID(newPeerID, selfID) {
			continue
		}
		peerIsAlive[newPeerID] = true
		orderTables = clearAllOrdersForPeer(orderTables, newPeerID)
	}
	return peerIsAlive, orderTables
}

// func resetPeerObservations(
// 	nodeRegistry types.GlobalNodeRegistry,
// 	recordedElevatorStates [config.NElevators]types.HRAElevState,
// 	peerOrderSnapshots [config.NElevators][config.NElevators]types.OrderTable,
// 	selfID int,
// ) ([config.NElevators]types.HRAElevState, [config.NElevators][config.NElevators]types.OrderTable) {
// 	for _, lostPeerID := range nodeRegistry.Lost {
// 		if !isRemotePeerID(lostPeerID, selfID) {
// 			continue
// 		}
// 		recordedElevatorStates[lostPeerID] = types.HRAElevState{}
// 		peerOrderSnapshots[lostPeerID] = [config.NElevators]types.OrderTable{}
// 	}

// 	for _, newPeerID := range nodeRegistry.New {
// 		if !isRemotePeerID(newPeerID, selfID) {
// 			continue
// 		}
// 		recordedElevatorStates[newPeerID] = types.HRAElevState{}
// 		peerOrderSnapshots[newPeerID] = [config.NElevators]types.OrderTable{}
// 	}
// 	return recordedElevatorStates, peerOrderSnapshots
// }

func resetPeerObservations(
	nodeRegistry types.GlobalNodeRegistry,
	lastPeerElevatorStates [config.NElevators]types.HRAElevState,
	peerOrderSnapshots [config.NElevators][config.NElevators]types.OrderTable,
	selfID int,
) ([config.NElevators]types.HRAElevState, [config.NElevators][config.NElevators]types.OrderTable) {
	for _, lostPeerID := range nodeRegistry.Lost {
		if !isRemotePeerID(lostPeerID, selfID) {
			continue
		}
		lastPeerElevatorStates[lostPeerID] = types.HRAElevState{}

		// Nullstill alt sendt FRA denne peeren
		peerOrderSnapshots[lostPeerID] = [config.NElevators]types.OrderTable{}

		// Nullstill alt andre peers husker OM denne peeren
		peerOrderSnapshots = clearPeerSnapshotColumn(peerOrderSnapshots, lostPeerID)
	}

	for _, newPeerID := range nodeRegistry.New {
		if !isRemotePeerID(newPeerID, selfID) {
			continue
		}
		lastPeerElevatorStates[newPeerID] = types.HRAElevState{}

		// Behandle restart/rejoin som en helt fersk instans
		peerOrderSnapshots[newPeerID] = [config.NElevators]types.OrderTable{}
		peerOrderSnapshots = clearPeerSnapshotColumn(peerOrderSnapshots, newPeerID)
	}
	return lastPeerElevatorStates, peerOrderSnapshots
}

func matchesLastPeerSnapshot(
	msg types.Message,
	peerOrderSnapshots [config.NElevators][config.NElevators]types.OrderTable,
	recordedElevatorStates [config.NElevators]types.HRAElevState,
) bool {
	return peerOrderSnapshots[msg.SenderID][msg.SenderID] == msg.OrderTables[msg.SenderID] &&
		recordedElevatorStates[msg.SenderID] == msg.ElevatorState
}

func alivePeersConsistent(
	peerConsistent [config.NElevators]bool,
	peerIsAlive [config.NElevators]bool,
	selfID int,
) bool {
	for peerID := range config.NElevators {
		if peerID == selfID {
			continue
		}
		if peerIsAlive[peerID] && !peerConsistent[peerID] {
			return false
		}
	}
	return true
}

func trySend[T any](ch chan<- T, value T) {
	select {
	case ch <- value:
	default:
	}
}

func applyPeerHallRow(
	msg types.Message,
	orderTables [config.NElevators]types.OrderTable,
) [config.NElevators]types.OrderTable {
	peerID := msg.SenderID
	peerHallRow := msg.OrderTables[peerID]

	for floor := range config.NFloors {
		for btn := types.BtnHallUp; btn <= types.BtnHallDown; btn++ {
			orderTables[peerID][floor][btn] = peerHallRow[floor][btn]
		}
	}
	return orderTables
}

func clearPeerHallOrders(
	orderTables [config.NElevators]types.OrderTable,
	peerID int,
) [config.NElevators]types.OrderTable {
	for floor := range config.NFloors {
		orderTables[peerID][floor][types.BtnHallUp] = types.OrderStandby
		orderTables[peerID][floor][types.BtnHallDown] = types.OrderStandby
	}
	return orderTables
}

//

func clearPeerCabOrders(
	orderTables [config.NElevators]types.OrderTable,
	peerID int,
) [config.NElevators]types.OrderTable {
	for floor := range config.NFloors {
		orderTables[peerID][floor][types.BtnCab] = types.OrderStandby
	}
	return orderTables
}

func clearAllOrdersForPeer(
	orderTables [config.NElevators]types.OrderTable,
	peerID int,
) [config.NElevators]types.OrderTable {
	for floor := range config.NFloors {
		orderTables[peerID][floor][types.BtnHallUp] = types.OrderStandby
		orderTables[peerID][floor][types.BtnHallDown] = types.OrderStandby
		orderTables[peerID][floor][types.BtnCab] = types.OrderStandby
	}
	return orderTables
}

func clearPeerSnapshotColumn(
	peerOrderSnapshots [config.NElevators][config.NElevators]types.OrderTable,
	ownerID int,
) [config.NElevators][config.NElevators]types.OrderTable {
	for peerID := range config.NElevators {
		peerOrderSnapshots[peerID][ownerID] = types.OrderTable{}
	}
	return peerOrderSnapshots
}
