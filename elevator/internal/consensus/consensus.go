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
		recoveryActive   = true
		peerSnapshotSeen bool
	)

	systemHallOrders = newSystemHallOrders()

	for {
		select {

		case registry := <-peerEvents:
			peerSnapshotSeen = true
			peerIsAlive, systemHallOrders = updatePeerAvailability(registry, peerIsAlive, systemHallOrders, selfID)
			peerIsConsistent = [NElevators]bool{}
			currentState := ConvergedSystemState{
				AliveList:      peerIsAlive,
				ElevatorList:   systemElevStates,
				HallOrderTable: systemHallOrders,
			}
			tryPublishConvergedState(
				&recoveryActive,
				peerSnapshotSeen,
				&peerIsConsistent,
				currentState,
				selfID,
				broadcast,
				converged,
			)

		case msg := <-peerMsg:
			if msg.SenderID < 0 || msg.SenderID >= NElevators || msg.SenderID == selfID {
				continue
			}

			peerSnapshotSeen = true
			peerIsConsistent[msg.SenderID] = peerStateMatchesRecorded(msg, systemHallOrders, systemElevStates)
			systemElevStates = mergePeerElevatorStates(msg, systemElevStates, selfID, recoveryActive)
			systemHallOrders[msg.SenderID] = msg.HallOrderTable
			peerIsAlive[msg.SenderID] = msg.AliveStatus

			systemHallOrders = advanceLocalOrderStates(systemHallOrders, selfID, peerIsAlive)
			sendStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders, recoveryActive)
			currentState := ConvergedSystemState{
				AliveList:      peerIsAlive,
				ElevatorList:   systemElevStates,
				HallOrderTable: systemHallOrders,
			}
			tryPublishConvergedState(
				&recoveryActive,
				peerSnapshotSeen,
				&peerIsConsistent,
				currentState,
				selfID,
				broadcast,
				converged,
			)

		case state := <-localState:
			if recoveryActive {
				state.ElevatorState.CabRequests = MergeCabRequests(
					state.ElevatorState.CabRequests,
					systemElevStates[selfID].CabRequests,
				)
			}
			systemElevStates[selfID] = state.ElevatorState

			systemHallOrders[selfID] = state.HallRequests
			peerIsAlive[selfID] = state.AliveStatus

			systemHallOrders = advanceLocalOrderStates(systemHallOrders, selfID, peerIsAlive)
			sendStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders, recoveryActive)
			currentState := ConvergedSystemState{
				AliveList:      peerIsAlive,
				ElevatorList:   systemElevStates,
				HallOrderTable: systemHallOrders,
			}
			tryPublishConvergedState(
				&recoveryActive,
				peerSnapshotSeen,
				&peerIsConsistent,
				currentState,
				selfID,
				broadcast,
				converged,
			)
		}
	}
}
