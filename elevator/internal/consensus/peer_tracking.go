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
		peerIsAlive[lostPeerID] = false
		systemHallOrders[lostPeerID] = newStandbyHallOrders()
	}

	for _, newPeerID := range nodeRegistry.New {
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
	return systemHallOrders[msg.SenderID] == msg.HallOrderTable &&
		elevStateEqual(systemElevStates[msg.SenderID], msg.ElevatorList[msg.SenderID])
}


func mergeCabs(self, peer HRAElevState) HRAElevState {
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
	broadcast 		chan<- Message,
	selfID 			int,
	peerIsAlive 	[NElevators]bool,
	systemElevStates [NElevators]HRAElevState,
	systemHallOrders [NElevators]HallOrderTable,
) {
	select {
	case broadcast <- Message{
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
func adoptPeerStates(
	peerStates [NElevators]HRAElevState,
	systemStates [NElevators]HRAElevState,
	selfID int,
	recoveryMode bool,
) [NElevators]HRAElevState {
	for peerID := 0; peerID < NElevators; peerID++ {
		if peerID == selfID {
			if recoveryMode {
				systemStates[selfID] = mergeCabs(systemStates[selfID], peerStates[selfID])
			} else {
				systemStates[selfID] = peerStates[selfID]
			}
			continue
		}
		if len(peerStates[peerID].CabRequests) == NFloors {
			systemStates[peerID] = peerStates[peerID]
		}
	}
	return systemStates
}
