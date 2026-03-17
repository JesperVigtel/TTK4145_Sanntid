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
	systemHallOrders [config.NElevators]types.HallOrderTable,
	selfID int,
) ([config.NElevators]bool, [config.NElevators]types.HallOrderTable) {
	for _, lostPeerID := range nodeRegistry.Lost {
		if !isRemotePeerID(lostPeerID, selfID) {
			continue
		}
		peerIsAlive[lostPeerID] = false
		systemHallOrders[lostPeerID] = types.HallOrderTable{}
	}

	for _, newPeerID := range nodeRegistry.New {
		if !isRemotePeerID(newPeerID, selfID) {
			continue
		}
		peerIsAlive[newPeerID] = true
		systemHallOrders[newPeerID] = types.HallOrderTable{}
	}
	return peerIsAlive, systemHallOrders
}

func resetPeerSnapshots(
	nodeRegistry types.GlobalNodeRegistry,
	peerReportedSelfStates [config.NElevators]types.HRAElevState,
	peerObservedCabOrders [config.NElevators][config.NElevators]types.CabOrderTable,
	selfID int,
) ([config.NElevators]types.HRAElevState, [config.NElevators][config.NElevators]types.CabOrderTable) {
	for _, peerIDs := range [][]int{nodeRegistry.Lost, nodeRegistry.New} {
		for _, peerID := range peerIDs {
			if !isRemotePeerID(peerID, selfID) {
				continue
			}
			peerReportedSelfStates[peerID] = types.HRAElevState{}
			peerObservedCabOrders[peerID] = [config.NElevators]types.CabOrderTable{}
		}
	}
	return peerReportedSelfStates, peerObservedCabOrders
}

func peerStateMatchesRecorded(
	msg types.Message,
	systemHallOrders [config.NElevators]types.HallOrderTable,
	peerReportedSelfStates [config.NElevators]types.HRAElevState,
) bool {
	return systemHallOrders[msg.SenderID] == msg.HallOrderTable &&
		elevatorStatesEqual(peerReportedSelfStates[msg.SenderID], msg.ElevatorList[msg.SenderID])
}

func allAlivePeersConsistent(
	peerIsConsistent [config.NElevators]bool,
	peerIsAlive [config.NElevators]bool,
	selfID int,
) bool {
	for peerID := range config.NElevators {
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
	broadcast chan<- types.Message,
	selfID int,
	peerIsAlive [config.NElevators]bool,
	systemElevStates [config.NElevators]types.HRAElevState,
	systemHallOrders [config.NElevators]types.HallOrderTable,
) {
	select {
	case broadcast <- types.Message{
		SenderID:       selfID,
		ElevatorList:   systemElevStates,
		HallOrderTable: systemHallOrders[selfID],
		AliveStatus:    peerIsAlive[selfID],
	}:
	default:
	}
}

func reconcileAndPublish(
	broadcast chan<- types.Message,
	converged chan<- types.ConvergedSystemState,
	selfID int,
	peerIsAlive [config.NElevators]bool,
	peerIsConsistent [config.NElevators]bool,
	systemElevStates [config.NElevators]types.HRAElevState,
	systemHallOrders [config.NElevators]types.HallOrderTable,
	peerObservedCabOrders [config.NElevators][config.NElevators]types.CabOrderTable,
) ([config.NElevators]types.HRAElevState, [config.NElevators]types.HallOrderTable, [config.NElevators]bool) {
	systemHallOrders = advanceHallOrderStates(systemHallOrders, selfID, peerIsAlive)
	systemElevStates = advanceCabOrderStates(systemElevStates, peerObservedCabOrders, selfID, peerIsAlive)
	sendStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders)
	peerIsConsistent = publishIfConsistent(
		peerIsConsistent,
		peerIsAlive,
		systemElevStates,
		systemHallOrders,
		selfID,
		converged,
	)
	return systemElevStates, systemHallOrders, peerIsConsistent
}

func publishIfConsistent(
	peerIsConsistent [config.NElevators]bool,
	peerIsAlive [config.NElevators]bool,
	systemElevStates [config.NElevators]types.HRAElevState,
	systemHallOrders [config.NElevators]types.HallOrderTable,
	selfID int,
	converged chan<- types.ConvergedSystemState,
) [config.NElevators]bool {
	if !allAlivePeersConsistent(peerIsConsistent, peerIsAlive, selfID) {
		return peerIsConsistent
	}

	peerIsConsistent = [config.NElevators]bool{}

	select {
	case converged <- types.ConvergedSystemState{
		AliveList:      peerIsAlive,
		ElevatorList:   systemElevStates,
		HallOrderTable: systemHallOrders,
	}:
	default:
	}

	return peerIsConsistent
}

func recordPeerObservedCabOrders(
	msg types.Message,
	peerObservedCabOrders [config.NElevators][config.NElevators]types.CabOrderTable,
) [config.NElevators][config.NElevators]types.CabOrderTable {
	for ownerID := range config.NElevators {
		peerObservedCabOrders[msg.SenderID][ownerID] = msg.ElevatorList[ownerID].CabOrders
	}
	return peerObservedCabOrders
}

func adoptPeerElevatorStatus(
	msg types.Message,
	systemStates [config.NElevators]types.HRAElevState,
) [config.NElevators]types.HRAElevState {
	senderState := msg.ElevatorList[msg.SenderID]
	systemStates[msg.SenderID].Behavior = senderState.Behavior
	systemStates[msg.SenderID].Floor = senderState.Floor
	systemStates[msg.SenderID].Direction = senderState.Direction
	systemStates[msg.SenderID].Assignable = senderState.Assignable
	return systemStates
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
