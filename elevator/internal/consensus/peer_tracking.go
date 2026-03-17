package consensus

import (
	"elevator/internal/config"
	"elevator/internal/types"
)

func updatePeerAvailability(
	nodeRegistry types.GlobalNodeRegistry,
	peerIsAlive [config.NElevators]bool,
	systemHallOrders [config.NElevators]types.HallOrderTable,
	selfID int,
) ([config.NElevators]bool, [config.NElevators]types.HallOrderTable) {
	for _, lostPeerID := range nodeRegistry.Lost {
		if lostPeerID == selfID || lostPeerID < 0 || lostPeerID >= config.NElevators {
			continue
		}
		peerIsAlive[lostPeerID] = false
		systemHallOrders[lostPeerID] = newStandbyHallOrders()
	}

	for _, newPeerID := range nodeRegistry.New {
		if newPeerID == selfID || newPeerID < 0 || newPeerID >= config.NElevators {
			continue
		}
		peerIsAlive[newPeerID] = true
		systemHallOrders[newPeerID] = newStandbyHallOrders()
	}
	return peerIsAlive, systemHallOrders
}

func resetPeerObservations(
	nodeRegistry types.GlobalNodeRegistry,
	peerReportedStates [config.NElevators]types.HRAElevState,
	peerCabOrderViews [config.NElevators][config.NElevators]types.CabOrderTable,
	selfID int,
) ([config.NElevators]types.HRAElevState, [config.NElevators][config.NElevators]types.CabOrderTable) {
	for _, peerIDs := range [][]int{nodeRegistry.Lost, nodeRegistry.New} {
		for _, peerID := range peerIDs {
			if peerID == selfID || peerID < 0 || peerID >= config.NElevators {
				continue
			}
			peerReportedStates[peerID] = types.HRAElevState{}
			peerCabOrderViews[peerID] = newPeerCabOrderViews()
		}
	}
	return peerReportedStates, peerCabOrderViews
}

func peerStateMatchesRecorded(
	msg types.Message,
	systemHallOrders [config.NElevators]types.HallOrderTable,
	peerReportedStates [config.NElevators]types.HRAElevState,
) bool {
	return systemHallOrders[msg.SenderID] == msg.HallOrderTable &&
		elevStateEqual(peerReportedStates[msg.SenderID], msg.ElevatorList[msg.SenderID])
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

func publishIfConsistent(
	peerDiscoveryReady bool,
	peerIsConsistent [config.NElevators]bool,
	peerIsAlive [config.NElevators]bool,
	systemElevStates [config.NElevators]types.HRAElevState,
	systemHallOrders [config.NElevators]types.HallOrderTable,
	selfID int,
	broadcast chan<- types.Message,
	converged chan<- types.ConvergedSystemState,
) [config.NElevators]bool {
	snapshot := currentConvergedState(peerIsAlive, systemElevStates, systemHallOrders)

	readyToPublish := peerDiscoveryReady &&
		allAlivePeersConsistent(peerIsConsistent, snapshot.AliveList, selfID)
	if !readyToPublish {
		return peerIsConsistent
	}

	peerIsConsistent = [config.NElevators]bool{}

	select {
	case converged <- snapshot:
	default:
	}

	sendStateUpdate(
		broadcast,
		selfID,
		snapshot.AliveList,
		snapshot.ElevatorList,
		snapshot.HallOrderTable,
	)

	return peerIsConsistent
}

func currentConvergedState(
	peerIsAlive [config.NElevators]bool,
	systemElevStates [config.NElevators]types.HRAElevState,
	systemHallOrders [config.NElevators]types.HallOrderTable,
) types.ConvergedSystemState {
	return types.ConvergedSystemState{
		AliveList:      peerIsAlive,
		ElevatorList:   systemElevStates,
		HallOrderTable: systemHallOrders,
	}
}

func recordPeerCabOrderViews(
	msg types.Message,
	peerCabOrderViews [config.NElevators][config.NElevators]types.CabOrderTable,
) [config.NElevators][config.NElevators]types.CabOrderTable {
	for ownerID := range config.NElevators {
		peerCabOrderViews[msg.SenderID][ownerID] = msg.ElevatorList[ownerID].CabOrders
	}
	return peerCabOrderViews
}

func adoptPeerElevatorState(
	msg types.Message,
	systemStates [config.NElevators]types.HRAElevState,
) [config.NElevators]types.HRAElevState {
	senderState := msg.ElevatorList[msg.SenderID]
	systemStates[msg.SenderID].Behavior = senderState.Behavior
	systemStates[msg.SenderID].Floor = senderState.Floor
	systemStates[msg.SenderID].Direction = senderState.Direction
	return systemStates
}

func newSystemHallOrders() [config.NElevators]types.HallOrderTable {
	var table [config.NElevators]types.HallOrderTable
	for peerID := range config.NElevators {
		table[peerID] = newStandbyHallOrders()
	}
	return table
}

func newStandbyHallOrders() types.HallOrderTable {
	var table types.HallOrderTable
	for floor := range config.NFloors {
		for btn := range config.NButtons {
			table[floor][btn] = types.OrderStandby
		}
	}
	return table
}

func newPeerCabOrderViews() [config.NElevators]types.CabOrderTable {
	var table [config.NElevators]types.CabOrderTable
	for ownerID := range config.NElevators {
		table[ownerID] = newStandbyCabOrders()
	}
	return table
}

func newStandbyCabOrders() types.CabOrderTable {
	var table types.CabOrderTable
	for floor := range config.NFloors {
		table[floor] = types.OrderStandby
	}
	return table
}

func elevStateEqual(a, b types.HRAElevState) bool {
	if a.Behavior != b.Behavior || a.Floor != b.Floor || a.Direction != b.Direction {
		return false
	}
	for floor := range config.NFloors {
		if a.CabOrders[floor] != b.CabOrders[floor] {
			return false
		}
	}
	return true
}
