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
		peerReportedStates [config.NElevators]types.HRAElevState
		peerCabOrderViews  [config.NElevators][config.NElevators]types.CabOrderTable
		peerIsAlive        [config.NElevators]bool
		peerIsConsistent   [config.NElevators]bool
	)

	systemHallOrders = newSystemHallOrders()

	for {
		select {

		case registry := <-peerEvents:
			peerIsAlive, systemHallOrders = updatePeerAvailability(registry, peerIsAlive, systemHallOrders, selfID)
			peerReportedStates, peerCabOrderViews = resetPeerObservations(
				registry,
				peerReportedStates,
				peerCabOrderViews,
				selfID,
			)
			peerIsConsistent = [config.NElevators]bool{}
			systemElevStates, systemHallOrders = advanceAndBroadcast(
				broadcast,
				selfID,
				peerIsAlive,
				systemElevStates,
				systemHallOrders,
				peerCabOrderViews,
			)
			peerIsConsistent = publishIfConsistent(
				peerIsConsistent,
				peerIsAlive,
				systemElevStates,
				systemHallOrders,
				selfID,
				converged,
			)

		case msg := <-peerMsg:
			if msg.SenderID < 0 || msg.SenderID >= config.NElevators || msg.SenderID == selfID {
				continue
			}

			peerIsConsistent[msg.SenderID] = peerStateMatchesRecorded(msg, systemHallOrders, peerReportedStates)
			peerReportedStates[msg.SenderID] = msg.ElevatorList[msg.SenderID]
			peerCabOrderViews = recordPeerCabOrderViews(msg, peerCabOrderViews)
			systemElevStates = adoptPeerElevatorState(msg, systemElevStates)
			systemHallOrders[msg.SenderID] = msg.HallOrderTable
			peerIsAlive[msg.SenderID] = msg.AliveStatus

			systemElevStates, systemHallOrders = advanceAndBroadcast(
				broadcast,
				selfID,
				peerIsAlive,
				systemElevStates,
				systemHallOrders,
				peerCabOrderViews,
			)
			peerIsConsistent = publishIfConsistent(
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

			systemElevStates, systemHallOrders = advanceAndBroadcast(
				broadcast,
				selfID,
				peerIsAlive,
				systemElevStates,
				systemHallOrders,
				peerCabOrderViews,
			)
			peerIsConsistent = publishIfConsistent(
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
