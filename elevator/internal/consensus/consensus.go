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
		systemHallOrders [config.NElevators]types.HallOrderTable
		systemElevStates [config.NElevators]types.HRAElevState
		// [peer] last self-reported row from that peer, used for consistency checks.
		peerReportedSelfStates [config.NElevators]types.HRAElevState
		// [observer][owner] cab-order row that one peer reports for another.
		peerObservedCabOrders [config.NElevators][config.NElevators]types.CabOrderTable
		peerIsAlive           [config.NElevators]bool
		peerIsConsistent      [config.NElevators]bool
	)

	systemHallOrders = newSystemHallOrders()

	for {
		select {
		case registry := <-peerEvents:
			peerIsAlive, systemHallOrders = updatePeerAvailability(registry, peerIsAlive, systemHallOrders, selfID)
			peerReportedSelfStates, peerObservedCabOrders = resetPeerSnapshots(
				registry,
				peerReportedSelfStates,
				peerObservedCabOrders,
				selfID,
			)
			peerIsConsistent = [config.NElevators]bool{}
			systemElevStates, systemHallOrders, peerIsConsistent = reconcileAndPublish(
				broadcast,
				converged,
				selfID,
				peerIsAlive,
				peerIsConsistent,
				systemElevStates,
				systemHallOrders,
				peerObservedCabOrders,
			)

		case msg := <-peerMsg:
			if !isRemotePeerID(msg.SenderID, selfID) {
				continue
			}

			peerIsConsistent[msg.SenderID] = peerStateMatchesRecorded(msg, systemHallOrders, peerReportedSelfStates)
			peerReportedSelfStates[msg.SenderID] = msg.ElevatorList[msg.SenderID]
			peerObservedCabOrders = recordPeerObservedCabOrders(msg, peerObservedCabOrders)
			systemElevStates = adoptPeerElevatorStatus(msg, systemElevStates)
			systemHallOrders[msg.SenderID] = msg.HallOrderTable
			peerIsAlive[msg.SenderID] = msg.AliveStatus

			systemElevStates, systemHallOrders, peerIsConsistent = reconcileAndPublish(
				broadcast,
				converged,
				selfID,
				peerIsAlive,
				peerIsConsistent,
				systemElevStates,
				systemHallOrders,
				peerObservedCabOrders,
			)

		case state := <-localState:
			systemElevStates[selfID] = state.ElevatorState
			systemHallOrders[selfID] = state.HallRequests
			peerIsAlive[selfID] = state.AliveStatus

			systemElevStates, systemHallOrders, peerIsConsistent = reconcileAndPublish(
				broadcast,
				converged,
				selfID,
				peerIsAlive,
				peerIsConsistent,
				systemElevStates,
				systemHallOrders,
				peerObservedCabOrders,
			)
		}
	}
}
