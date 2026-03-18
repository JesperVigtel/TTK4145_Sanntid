package consensus

import (
	"elevator/internal/config"
	"elevator/internal/types"
)

func isRemotePeerID(peerID int, selfID int) bool {
	return peerID >= 0 && peerID < config.NElevators && peerID != selfID
}

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
		orderTables = clearPeerHallOrders(orderTables, lostPeerID)
	}

	for _, newPeerID := range nodeRegistry.New {
		if !isRemotePeerID(newPeerID, selfID) {
			continue
		}
		peerIsAlive[newPeerID] = true
		orderTables = clearPeerHallOrders(orderTables, newPeerID)
	}
	return peerIsAlive, orderTables
}

func resetPeerSnapshots(
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
		peerOrderSnapshots[lostPeerID] = [config.NElevators]types.OrderTable{}
	}

	for _, newPeerID := range nodeRegistry.New {
		if !isRemotePeerID(newPeerID, selfID) {
			continue
		}
		lastPeerElevatorStates[newPeerID] = types.HRAElevState{}
		peerOrderSnapshots[newPeerID] = [config.NElevators]types.OrderTable{}
	}
	return lastPeerElevatorStates, peerOrderSnapshots
}

func matchesLastPeerSnapshot(
	msg types.Message,
	peerOrderSnapshots [config.NElevators][config.NElevators]types.OrderTable,
	lastPeerElevatorStates [config.NElevators]types.HRAElevState,
) bool {
	return peerOrderSnapshots[msg.SenderID][msg.SenderID] == msg.OrderTables[msg.SenderID] &&
		lastPeerElevatorStates[msg.SenderID] == msg.ElevatorState
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
