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
	)

	systemHallOrders = newSystemHallOrders()
	recovery := recoveryState{active: true}

	for {
		select {

		case registry := <-peerEvents:
			recovery.peerSnapshotSeen = true
			peerIsAlive, systemHallOrders = updatePeerAvailability(registry, peerIsAlive, systemHallOrders, selfID)
			peerIsConsistent = [NElevators]bool{}
			recovery.maybePublish(
				&peerIsConsistent,
				peerIsAlive,
				systemElevStates,
				systemHallOrders,
				selfID,
				broadcast,
				converged,
			)

		case msg := <-peerMsg:
			if msg.SenderID < 0 || msg.SenderID >= NElevators || msg.SenderID == selfID {
				continue
			}

			recovery.peerSnapshotSeen = true
			peerIsConsistent[msg.SenderID] = peerStateMatchesRecorded(msg, systemHallOrders, systemElevStates)
			systemElevStates = adoptPeerStates(msg, systemElevStates, selfID, recovery.active)
			systemHallOrders[msg.SenderID] = msg.HallOrderTable
			peerIsAlive[msg.SenderID] = msg.AliveStatus

			systemHallOrders = advanceLocalOrderStates(systemHallOrders, selfID, peerIsAlive)
			sendStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders, recovery.active)
			recovery.maybePublish(
				&peerIsConsistent,
				peerIsAlive,
				systemElevStates,
				systemHallOrders,
				selfID,
				broadcast,
				converged,
			)

		case state := <-localState:
			if recovery.active {
				systemElevStates[selfID] = mergeCabRequestsIntoState(state.ElevatorState, systemElevStates[selfID])
			} else {
				systemElevStates[selfID] = state.ElevatorState
			}

			systemHallOrders[selfID] = state.HallRequests
			peerIsAlive[selfID] = state.AliveStatus

			systemHallOrders = advanceLocalOrderStates(systemHallOrders, selfID, peerIsAlive)
			sendStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders, recovery.active)
			recovery.maybePublish(
				&peerIsConsistent,
				peerIsAlive,
				systemElevStates,
				systemHallOrders,
				selfID,
				broadcast,
				converged,
			)
		}
	}
}
