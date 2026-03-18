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
	for _, peerIDs := range [][]int{nodeRegistry.Lost, nodeRegistry.New} {
		for _, peerID := range peerIDs {
			if !isRemotePeerID(peerID, selfID) {
				continue
			}
			lastPeerElevatorStates[peerID] = types.HRAElevState{}
			peerOrderSnapshots[peerID] = [config.NElevators]types.OrderTable{}
		}
	}
	return lastPeerElevatorStates, peerOrderSnapshots
}

func matchesLastPeerSnapshot(
	msg types.Message,
	peerOrderSnapshots [config.NElevators][config.NElevators]types.OrderTable,
	lastPeerElevatorStates [config.NElevators]types.HRAElevState,
) bool {
	return peerOrderSnapshots[msg.SenderID][msg.SenderID] == msg.OrderTables[msg.SenderID] &&
		elevatorStatesEqual(lastPeerElevatorStates[msg.SenderID], msg.ElevatorList[msg.SenderID])
}

func elevatorStatesEqual(a, b types.HRAElevState) bool {
	if a.Behavior != b.Behavior ||
		a.Floor != b.Floor ||
		a.Direction != b.Direction ||
		a.Assignable != b.Assignable {
		return false
	}
	return true
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

func buildBroadcastState(
	selfID int,
	peerIsAlive [config.NElevators]bool,
	elevStates [config.NElevators]types.HRAElevState,
	orderTables [config.NElevators]types.OrderTable,
) types.Message {
	return types.Message{
		SenderID:     selfID,
		ElevatorList: elevStates,
		OrderTables:  orderTables,
		AliveStatus:  peerIsAlive[selfID],
	}
}

func buildConvergedState(
	peerIsAlive [config.NElevators]bool,
	elevStates [config.NElevators]types.HRAElevState,
	orderTables [config.NElevators]types.OrderTable,
) types.ConvergedSystemState {
	return types.ConvergedSystemState{
		AliveList:    peerIsAlive,
		ElevatorList: elevStates,
		OrderTables:  orderTables,
	}
}

func trySend[T any](ch chan<- T, value T) {
	select {
	case ch <- value:
	default:
	}
}

func recordPeerOrderSnapshot(
	msg types.Message,
	peerOrderSnapshots [config.NElevators][config.NElevators]types.OrderTable,
) [config.NElevators][config.NElevators]types.OrderTable {
	peerOrderSnapshots[msg.SenderID] = msg.OrderTables
	return peerOrderSnapshots
}

func applyPeerHallOrders(
	msg types.Message,
	orderTables [config.NElevators]types.OrderTable,
) [config.NElevators]types.OrderTable {
	for floor := range config.NFloors {
		for btn := types.BtnHallUp; btn <= types.BtnHallDown; btn++ {
			orderTables[msg.SenderID][floor][btn] = msg.OrderTables[msg.SenderID][floor][btn]
		}
	}
	return orderTables
}

func applyPeerState(
	msg types.Message,
	elevStates [config.NElevators]types.HRAElevState,
) [config.NElevators]types.HRAElevState {
	senderState := msg.ElevatorList[msg.SenderID]
	elevStates[msg.SenderID].Behavior = senderState.Behavior
	elevStates[msg.SenderID].Floor = senderState.Floor
	elevStates[msg.SenderID].Direction = senderState.Direction
	elevStates[msg.SenderID].Assignable = senderState.Assignable
	return elevStates
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
