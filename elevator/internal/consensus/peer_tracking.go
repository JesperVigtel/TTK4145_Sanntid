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

func peerStateMatchesRecorded(
	msg types.Message,
	systemHallOrders [config.NElevators]types.HallOrderTable,
	systemElevStates [config.NElevators]types.HRAElevState,
) bool {
	if msg.Recovering {
		return false
	}
	return systemHallOrders[msg.SenderID] == msg.HallOrderTable &&
		elevStateEqual(systemElevStates[msg.SenderID], msg.ElevatorList[msg.SenderID])
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
	recoveryMode bool,
) {
	select {
	case broadcast <- types.Message{
		SenderID:       selfID,
		ElevatorList:   systemElevStates,
		HallOrderTable: systemHallOrders[selfID],
		AliveStatus:    peerIsAlive[selfID],
		Recovering:     recoveryMode,
	}:
	default:
	}
}

func publishIfConsistent(
	recoveryActive bool,
	peerDiscoveryReady bool,
	peerIsConsistent [config.NElevators]bool,
	peerIsAlive [config.NElevators]bool,
	systemElevStates [config.NElevators]types.HRAElevState,
	systemHallOrders [config.NElevators]types.HallOrderTable,
	selfID int,
	broadcast chan<- types.Message,
	converged chan<- types.ConvergedSystemState,
) (bool, [config.NElevators]bool) {
	snapshot := currentConvergedState(peerIsAlive, systemElevStates, systemHallOrders, recoveryActive)

	readyToPublish := peerDiscoveryReady &&
		allAlivePeersConsistent(peerIsConsistent, snapshot.AliveList, selfID)
	if !readyToPublish {
		return recoveryActive, peerIsConsistent
	}

	peerIsConsistent = [config.NElevators]bool{}

	select {
	case converged <- snapshot:
	default:
	}

	if recoveryActive {
		recoveryActive = false
		sendStateUpdate(
			broadcast,
			selfID,
			snapshot.AliveList,
			snapshot.ElevatorList,
			snapshot.HallOrderTable,
			false,
		)
	}

	return recoveryActive, peerIsConsistent
}

func currentConvergedState(
	peerIsAlive [config.NElevators]bool,
	systemElevStates [config.NElevators]types.HRAElevState,
	systemHallOrders [config.NElevators]types.HallOrderTable,
	recoveryActive bool,
) types.ConvergedSystemState {
	return types.ConvergedSystemState{
		AliveList:      peerIsAlive,
		ElevatorList:   systemElevStates,
		HallOrderTable: systemHallOrders,
		Recovering:     recoveryActive,
	}
}

func mergePeerElevatorStates(
	msg types.Message,
	systemStates [config.NElevators]types.HRAElevState,
	selfID int,
	recoveryActive bool,
) [config.NElevators]types.HRAElevState {
	if recoveryActive {
		systemStates[selfID].CabRequests = types.MergeCabRequests(
			systemStates[selfID].CabRequests,
			msg.ElevatorList[selfID].CabRequests,
		)
	}

	senderState := msg.ElevatorList[msg.SenderID]
	if len(senderState.CabRequests) != config.NFloors {
		return systemStates
	}
	if msg.Recovering {
		senderState.CabRequests = types.MergeCabRequests(senderState.CabRequests, systemStates[msg.SenderID].CabRequests)
	}
	systemStates[msg.SenderID] = senderState
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

func elevStateEqual(a, b types.HRAElevState) bool {
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
