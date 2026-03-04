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
) ([NElevators]bool, [NElevators]HallOrderTable) {

	for _, peerID := range nodeRegistry.Lost {
		if peerID < 0 || peerID >= NElevators {
			continue
		}
		// Reset to Standby on peer loss: we can no longer confirm the state of orders
		// that peer was tracking, so we return to the cycle's safe starting state.
		peerIsAlive[peerID] = false
		systemHallOrders[peerID] = newStandbyHallOrders()
	}

	for _, peerID := range nodeRegistry.New {
		if peerID < 0 || peerID >= NElevators {
			continue
		}
		peerIsAlive[peerID] = true
	}
	return peerIsAlive, systemHallOrders
}

func peerStateMatchesRecorded(
	msg Message,
	systemHallOrders [NElevators]HallOrderTable,
	systemElevStates [NElevators]HRAElevState,
) bool {
	return reflect.DeepEqual(systemHallOrders[msg.SenderID], msg.HallOrderList) &&
		reflect.DeepEqual(systemElevStates, msg.ElevatorList)
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

func broadcastLocalState(
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
		HallOrderList: systemHallOrders[selfID],
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
