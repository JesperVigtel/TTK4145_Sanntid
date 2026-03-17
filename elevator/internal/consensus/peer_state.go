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
	hallOrders [config.NElevators]types.HallOrderTable,
	selfID int,
) ([config.NElevators]bool, [config.NElevators]types.HallOrderTable) {
	for _, lostPeerID := range nodeRegistry.Lost {
		if !isRemotePeerID(lostPeerID, selfID) {
			continue
		}
		peerIsAlive[lostPeerID] = false
		hallOrders[lostPeerID] = types.HallOrderTable{}
	}

	for _, newPeerID := range nodeRegistry.New {
		if !isRemotePeerID(newPeerID, selfID) {
			continue
		}
		peerIsAlive[newPeerID] = true
		hallOrders[newPeerID] = types.HallOrderTable{}
	}
	return peerIsAlive, hallOrders
}

func resetPeerSnapshots(
	nodeRegistry types.GlobalNodeRegistry,
	lastPeerStates [config.NElevators]types.HRAElevState,
	peerCabViews [config.NElevators][config.NElevators]types.CabOrderTable,
	selfID int,
) ([config.NElevators]types.HRAElevState, [config.NElevators][config.NElevators]types.CabOrderTable) {
	for _, peerIDs := range [][]int{nodeRegistry.Lost, nodeRegistry.New} {
		for _, peerID := range peerIDs {
			if !isRemotePeerID(peerID, selfID) {
				continue
			}
			lastPeerStates[peerID] = types.HRAElevState{}
			peerCabViews[peerID] = [config.NElevators]types.CabOrderTable{}
		}
	}
	return lastPeerStates, peerCabViews
}

func matchesLastPeerState(
	msg types.Message,
	hallOrders [config.NElevators]types.HallOrderTable,
	lastPeerStates [config.NElevators]types.HRAElevState,
) bool {
	return hallOrders[msg.SenderID] == msg.HallOrderTable &&
		elevatorStatesEqual(lastPeerStates[msg.SenderID], msg.ElevatorList[msg.SenderID])
}

func elevatorStatesEqual(a, b types.HRAElevState) bool {
	if a.Behavior != b.Behavior ||
		a.Floor != b.Floor ||
		a.Direction != b.Direction ||
		a.Assignable != b.Assignable {
		return false
	}
	for floor := range config.NFloors {
		if a.CabOrders[floor] != b.CabOrders[floor] {
			return false
		}
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
	hallOrders [config.NElevators]types.HallOrderTable,
) types.Message {
	return types.Message{
		SenderID:       selfID,
		ElevatorList:   elevStates,
		HallOrderTable: hallOrders[selfID],
		AliveStatus:    peerIsAlive[selfID],
	}
}

func buildConvergedState(
	peerIsAlive [config.NElevators]bool,
	elevStates [config.NElevators]types.HRAElevState,
	hallOrders [config.NElevators]types.HallOrderTable,
) types.ConvergedSystemState {
	return types.ConvergedSystemState{
		AliveList:      peerIsAlive,
		ElevatorList:   elevStates,
		HallOrderTable: hallOrders,
	}
}

func trySend[T any](ch chan<- T, value T) {
	select {
	case ch <- value:
	default:
	}
}

func recordPeerCabViews(
	msg types.Message,
	peerCabViews [config.NElevators][config.NElevators]types.CabOrderTable,
) [config.NElevators][config.NElevators]types.CabOrderTable {
	for ownerID := range config.NElevators {
		peerCabViews[msg.SenderID][ownerID] = msg.ElevatorList[ownerID].CabOrders
	}
	return peerCabViews
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
