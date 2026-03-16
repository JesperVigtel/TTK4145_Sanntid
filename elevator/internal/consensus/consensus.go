package consensus

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

// -----------------------------------------------------------------------------
// Enforces distributed consensus over hall order state by requiring all alive
// peers to report a consistent view before publishing a converged state.
// Uses a cyclic order-state counter (Standby→Pending→Assigned→Complete→Standby)
// so that state transitions are self-synchronising without a central coordinator.
// -----------------------------------------------------------------------------

func Run(
	peerMsg <-chan Message,
	broadcast chan<- Message,
	peerEvents <-chan GlobalNodeRegistry,
	localState <-chan LocalSystemState,
	converged chan<- ConvergedSystemState,
	selfID int,
) {
	var (
		systemHallOrders [NElevators]HallOrderTable
		systemElevStates [NElevators]HRAElevState
		peerIsAlive      [NElevators]bool
		peerIsConsistent [NElevators]bool
		peerSnapshotSeen bool
	)

	systemHallOrders = newSystemHallOrders()
	recoveryMode := true
	publishIfReady := func() {
		if !peerSnapshotSeen || !allAlivePeersConsistent(peerIsConsistent, peerIsAlive, selfID) {
			return
		}
		peerIsConsistent = [NElevators]bool{}
		if recoveryMode {
			recoveryMode = false
			sendStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders, recoveryMode)
		}
		publishConsistentState(converged, peerIsAlive, systemElevStates, systemHallOrders)
	}

	for {
		select {

		case registry := <-peerEvents:
			peerSnapshotSeen = true
			peerIsAlive, systemHallOrders = updatePeerAvailability(registry, peerIsAlive, systemHallOrders, selfID)
			peerIsConsistent = [NElevators]bool{}
			publishIfReady()

		case msg := <-peerMsg:
			if msg.SenderID < 0 || msg.SenderID >= NElevators || msg.SenderID == selfID {
				continue
			}

			peerSnapshotSeen = true
			peerIsConsistent[msg.SenderID] = peerStateMatchesRecorded(msg, systemHallOrders, systemElevStates)
			systemElevStates = adoptPeerStates(msg, systemElevStates, selfID, recoveryMode)
			systemHallOrders[msg.SenderID] = msg.HallOrderTable
			peerIsAlive[msg.SenderID] = msg.AliveStatus

			systemHallOrders = advanceLocalOrderStates(systemHallOrders, selfID, peerIsAlive)
			sendStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders, recoveryMode)
			publishIfReady()

		case state := <-localState:
			if recoveryMode {
				systemElevStates[selfID] = mergeCabKnowledge(state.ElevatorState, systemElevStates[selfID])
			} else {
				systemElevStates[selfID] = state.ElevatorState
			}

			systemHallOrders[selfID] = state.HallRequests
			peerIsAlive[selfID] = state.AliveStatus

			systemHallOrders = advanceLocalOrderStates(systemHallOrders, selfID, peerIsAlive)
			sendStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders, recoveryMode)
			publishIfReady()
		}
	}
}
