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
		systemHallOrders       [config.NElevators]types.HallOrderTable
		systemElevStates       [config.NElevators]types.HRAElevState
		peerReportedSelfStates [config.NElevators]types.HRAElevState
		peerObservedCabOrders  [config.NElevators][config.NElevators]types.CabOrderTable
		peerIsAlive            [config.NElevators]bool
		peerIsConsistent       [config.NElevators]bool
	)

	for {
		select {
		case registry := <-peerEvents:
			peerIsAlive, systemHallOrders = updatePeerAvailability(registry, peerIsAlive, systemHallOrders, selfID)
			peerReportedSelfStates, peerObservedCabOrders = resetPeerSnapshots(registry, peerReportedSelfStates, peerObservedCabOrders, selfID)
			peerIsConsistent = [config.NElevators]bool{}

			systemHallOrders = advanceHallOrderStates(systemHallOrders, selfID, peerIsAlive)
			systemElevStates = advanceCabOrderStates(systemElevStates, peerObservedCabOrders, selfID, peerIsAlive)
			broadcastStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders)
			peerIsConsistent = publishConvergedStateIfConsistent(
				peerIsConsistent,
				peerIsAlive,
				systemElevStates,
				systemHallOrders,
				selfID,
				converged,
			)

		case msg := <-peerMsg:
			if !isRemotePeerID(msg.SenderID, selfID) {
				continue
			}

			peerIsConsistent[msg.SenderID] = peerSelfStateMatchesRecorded(msg, systemHallOrders, peerReportedSelfStates)
			peerReportedSelfStates[msg.SenderID] = msg.ElevatorList[msg.SenderID]
			peerObservedCabOrders = recordPeerObservedCabOrders(msg, peerObservedCabOrders)
			systemElevStates = applyPeerReportedElevatorState(msg, systemElevStates)
			systemHallOrders[msg.SenderID] = msg.HallOrderTable
			peerIsAlive[msg.SenderID] = msg.AliveStatus

			systemHallOrders = advanceHallOrderStates(systemHallOrders, selfID, peerIsAlive)
			systemElevStates = advanceCabOrderStates(systemElevStates, peerObservedCabOrders, selfID, peerIsAlive)
			broadcastStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders)
			peerIsConsistent = publishConvergedStateIfConsistent(
				peerIsConsistent,
				peerIsAlive,
				systemElevStates,
				systemHallOrders,
				selfID,
				converged,
			)

		case state := <-localState:
			systemElevStates[selfID] = state.ElevatorState
			systemHallOrders[selfID] = state.HallRequests
			peerIsAlive[selfID] = state.AliveStatus

			systemHallOrders = advanceHallOrderStates(systemHallOrders, selfID, peerIsAlive)
			systemElevStates = advanceCabOrderStates(systemElevStates, peerObservedCabOrders, selfID, peerIsAlive)
			broadcastStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders)
			peerIsConsistent = publishConvergedStateIfConsistent(
				peerIsConsistent,
				peerIsAlive,
				systemElevStates,
				systemHallOrders,
				selfID,
				converged,
			)
		}
	}
}
