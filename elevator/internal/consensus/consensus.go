package consensus

import (
	"elevator/internal/config"
	"elevator/internal/types"
)

// -----------------------------------------------------------------------------
// Enforces distributed consensus over hall order state by requiring all alive
// peers to report a consistent view before publishing a converged state.
// Uses a cyclic order-state counter (Standby→Pending→Assigned→Complete→Standby)
// so that state transitions are self-synchronising without a central coordinator.
// -----------------------------------------------------------------------------

func Run(
	peerMsg <-chan types.Message,
	broadcast chan<- types.Message,
	peerEvents <-chan types.GlobalNodeRegistry,
	localState <-chan types.LocalSystemState,
	converged chan<- types.ConvergedSystemState,
	selfID int,
) {
	var (
		systemHallOrders   [config.NElevators]types.HallOrderTable
		systemElevStates   [config.NElevators]types.HRAElevState
		peerIsAlive        [config.NElevators]bool
		peerIsConsistent   [config.NElevators]bool
		recoveryActive     = true
		peerDiscoveryReady bool
	)

	systemHallOrders = newSystemHallOrders()

	for {
		select {

		case registry := <-peerEvents:
			peerDiscoveryReady = true
			peerIsAlive, systemHallOrders = updatePeerAvailability(registry, peerIsAlive, systemHallOrders, selfID)
			peerIsConsistent = [config.NElevators]bool{}
			recoveryActive, peerIsConsistent = publishIfConsistent(
				recoveryActive,
				peerDiscoveryReady,
				peerIsConsistent,
				peerIsAlive,
				systemElevStates,
				systemHallOrders,
				selfID,
				broadcast,
				converged,
			)

		case msg := <-peerMsg:
			if msg.SenderID < 0 || msg.SenderID >= config.NElevators || msg.SenderID == selfID {
				continue
			}

			peerDiscoveryReady = true
			peerIsConsistent[msg.SenderID] = peerStateMatchesRecorded(msg, systemHallOrders, systemElevStates)
			systemElevStates = mergePeerElevatorStates(msg, systemElevStates, selfID, recoveryActive)
			systemHallOrders[msg.SenderID] = msg.HallOrderTable
			peerIsAlive[msg.SenderID] = msg.AliveStatus

			systemHallOrders = advanceLocalOrderStates(systemHallOrders, selfID, peerIsAlive)
			recoveryActive, peerIsConsistent = publishIfConsistent(
				recoveryActive,
				peerDiscoveryReady,
				peerIsConsistent,
				peerIsAlive,
				systemElevStates,
				systemHallOrders,
				selfID,
				broadcast,
				converged,
			)

		case state := <-localState:
			if recoveryActive {
				state.ElevatorState.CabRequests = types.MergeCabRequests(
					state.ElevatorState.CabRequests,
					systemElevStates[selfID].CabRequests,
				)
			}
			systemElevStates[selfID] = state.ElevatorState

			systemHallOrders[selfID] = state.HallRequests
			peerIsAlive[selfID] = state.AliveStatus

			systemHallOrders = advanceLocalOrderStates(systemHallOrders, selfID, peerIsAlive)
			sendStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders, recoveryActive)
			recoveryActive, peerIsConsistent = publishIfConsistent(
				recoveryActive,
				peerDiscoveryReady,
				peerIsConsistent,
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
