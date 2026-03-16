package consensus

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

//Peer event

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
		systemHallOrders[newPeerID] = newStandbyHallOrders() //Possibly removem, test. Can improve contnuity
	}
	return peerIsAlive, systemHallOrders
}

//Peer messages

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

func mergeCabs(self, peer HRAElevState) HRAElevState {
	if len(self.CabRequests) != NFloors {
		return self
	}
	if len(peer.CabRequests) != NFloors {
		return self
	}
	for floor := range NFloors {
		if peer.CabRequests[floor] == true {
			self.CabRequests[floor] = true
		}
	}
	return self
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

func mergeCabKnowledge(base, incoming HRAElevState) HRAElevState {
	if len(incoming.CabRequests) != NFloors {
		return base
	}
	if len(base.CabRequests) != NFloors {
		base.CabRequests = make([]bool, NFloors)
	}
	return mergeCabs(base, incoming)
}

func adoptPeerStates(
	msg Message,
	systemStates [NElevators]HRAElevState,
	selfID int,
	recoveryMode bool,
) [NElevators]HRAElevState {
	senderID := msg.SenderID
	if senderID < 0 || senderID >= NElevators {
		return systemStates
	}

	if recoveryMode {
		systemStates[selfID] = mergeCabKnowledge(systemStates[selfID], msg.ElevatorList[selfID])
	}

	senderState := msg.ElevatorList[senderID]
	if len(senderState.CabRequests) != NFloors {
		return systemStates
	}

	if msg.Recovering {
		systemStates[senderID] = mergeCabKnowledge(senderState, systemStates[senderID])
		return systemStates
	}

	systemStates[senderID] = senderState
	return systemStates
}
