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

	for _, peerID := range nodeRegistry.Lost {
		if peerID == selfID || peerID < 0 || peerID >= NElevators {
			continue
		} //Spør studass, possibly surpplus and can be removed
		peerIsAlive[peerID] = false
		systemHallOrders[peerID] = newStandbyHallOrders()
	}

	for _, peerID := range nodeRegistry.New {
		if peerID == selfID || peerID < 0 || peerID >= NElevators {
			continue
		} //Spør studass
		peerIsAlive[peerID] = true
		systemHallOrders[peerID] = newStandbyHallOrders() //Possibly removem, test. Can improve contnuity
	}
	return peerIsAlive, systemHallOrders
}

//Peer messages

func peerStateMatchesRecorded(
	msg Message,
	systemHallOrders [NElevators]HallOrderTable,
	systemElevStates [NElevators]HRAElevState,
) bool {
	return systemHallOrders[msg.SenderID] == msg.HallOrderTable &&
		elevStateEquals(systemElevStates[msg.SenderID], msg.ElevatorList[msg.SenderID])
}

func adoptPeerElevatorStates(
	msg 				[NElevators]HRAElevState,
	systemStates 		[NElevators]HRAElevState,
	selfID 				int,
	selfCabsRestored 	bool,
) ([NElevators]HRAElevState, bool) {
	for peerID := range NElevators {
		if len(msg[peerID].CabRequests) != NFloors {
			continue
		}
		if peerID != selfID {
			systemStates[peerID] = msg[peerID]
			continue
		}
		if selfCabsRestored {
			continue
		}
		systemStates[selfID] = mergeCabsOnRecovery(systemStates[selfID], msg[peerID])
		selfCabsRestored = true
	}
	return systemStates, selfCabsRestored
}


func mergeCabsOnRecovery(self, peer HRAElevState) HRAElevState {
	if len(self.CabRequests) != NFloors {
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
	outgoingMessages chan<- Message,
	selfID int,
	peerIsAlive [NElevators]bool,
	systemElevStates [NElevators]HRAElevState,
	systemHallOrders [NElevators]HallOrderTable,
) {
	select {
	case outgoingMessages <- Message{
		SenderID:       selfID,
		ElevatorList:   systemElevStates,
		HallOrderTable: systemHallOrders[selfID],
		AliveStatus:    peerIsAlive[selfID],
		AliveList:      peerIsAlive,
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

func elevStateEquals(a, b HRAElevState) bool {
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
