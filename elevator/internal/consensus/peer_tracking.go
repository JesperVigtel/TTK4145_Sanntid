package consensus

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

func updatePeerAvailability(
	nodeRegistry GlobalNodeRegistry,
	peerIsAlive [NElevators]bool,
	systemHallOrders [NElevators]HallOrderTable,
	selfID int,
) ([NElevators]bool, [NElevators]HallOrderTable) {
	for _, lostPeerID := range nodeRegistry.Lost {
		if lostPeerID == selfID || lostPeerID < 0 || lostPeerID >= NElevators {
			continue
		}
		peerIsAlive[lostPeerID] = false
		systemHallOrders[lostPeerID] = newStandbyHallOrders()
	}

	for _, newPeerID := range nodeRegistry.New {
		if newPeerID == selfID || newPeerID < 0 || newPeerID >= NElevators {
			continue
		}
		peerIsAlive[newPeerID] = true
		systemHallOrders[newPeerID] = newStandbyHallOrders()
	}
	return peerIsAlive, systemHallOrders
}

func peerStateMatchesRecorded(
	msg Message,
	systemHallOrders [NElevators]HallOrderTable,
	systemElevStates [NElevators]HRAElevState,
) bool {
	if msg.Recovering {
		return false
	}
	return systemHallOrders[msg.SenderID] == msg.HallOrderTable &&
		elevStateEqual(systemElevStates[msg.SenderID], msg.ElevatorList[msg.SenderID])
}

func allAlivePeersConsistent(
	peerIsConsistent [NElevators]bool,
	peerIsAlive [NElevators]bool,
	selfID int,
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
	broadcast chan<- Message,
	selfID int,
	peerIsAlive [NElevators]bool,
	systemElevStates [NElevators]HRAElevState,
	systemHallOrders [NElevators]HallOrderTable,
	recoveryMode bool,
) {
	select {
	case broadcast <- Message{
		SenderID:       selfID,
		ElevatorList:   systemElevStates,
		HallOrderTable: systemHallOrders[selfID],
		AliveStatus:    peerIsAlive[selfID],
		AliveList:      peerIsAlive,
		Recovering:     recoveryMode,
	}:
	default:
	}
}

func publishConsistentState(
	convergedSystemState chan<- ConvergedSystemState,
	peerIsAlive [NElevators]bool,
	systemElevStates [NElevators]HRAElevState,
	systemHallOrders [NElevators]HallOrderTable,
) {
	select {
	case convergedSystemState <- ConvergedSystemState{
		AliveList:      peerIsAlive,
		ElevatorList:   systemElevStates,
		HallOrderTable: systemHallOrders,
	}:
	default:
	}
}

func newSystemHallOrders() [NElevators]HallOrderTable {
	var table [NElevators]HallOrderTable
	for peerID := range NElevators {
		table[peerID] = newStandbyHallOrders()
	}
	return table
}

func newStandbyHallOrders() HallOrderTable {
	var table HallOrderTable
	for floor := range NFloors {
		for btn := range NButtons {
			table[floor][btn] = OrderStandby
		}
	}
	return table
}

func elevStateEqual(a, b HRAElevState) bool {
	if a.Behavior != b.Behavior || a.Floor != b.Floor || a.Direction != b.Direction {
		return false
	}
	if (a.CabRequests == nil) != (b.CabRequests == nil) {
		return false
	}
	if len(a.CabRequests) != len(b.CabRequests) {
		return false
	}
	for i := range a.CabRequests {
		if a.CabRequests[i] != b.CabRequests[i] {
			return false
		}
	}
	return true
}
