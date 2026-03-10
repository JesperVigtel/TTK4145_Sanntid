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
		systemHallOrders[peerID] = newStandbyHallOrders()
	}

	for _, peerID := range nodeRegistry.New {
		if peerID < 0 || peerID >= NElevators {
			continue
		}
		peerIsAlive[peerID] = true
		systemHallOrders[peerID] = newStandbyHallOrders()
	}
	return peerIsAlive, systemHallOrders
}

func recreateOnLoss(
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

func mergeIncomingHallOrders(		//Possibly remove if unused
	selfOrders HallOrderTable,
	incomingOrders HallOrderTable,
) HallOrderTable {
	var merged HallOrderTable
	for floor := range merged {
		for btn := range merged[floor] {
			selfState := selfOrders[floor][btn]
			incoming := incomingOrders[floor][btn]
			if isActiveOrder(selfState) && incoming == OrderStandby {
				merged[floor][btn] = selfState
			} else {
				merged[floor][btn] = incoming
			}
		}
	}
	return merged
}

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
