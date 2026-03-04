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

func RunConsensusManager(
	incomingMessages 	<-chan Message,
	outgoingMessages	chan<- Message,
	nodeRegistryEvents 	<-chan GlobalNodeRegistry,
	localSystemState 	<-chan LocalSystemState,
	convergedSystemState chan<- ConvergedSystemState,
	selfID 				int,
) {
	var (
		systemHallOrders [NElevators]HallOrderTable
		systemElevStates [NElevators]HRAElevState
		peerIsAlive      [NElevators]bool
		peerIsConsistent [NElevators]bool
	)

	systemHallOrders = newSystemHallOrders()

	for {
		select {

		case registry := <-nodeRegistryEvents:
			peerIsAlive, systemHallOrders = updatePeerAvailability(registry, peerIsAlive, systemHallOrders)

		case msg := <-incomingMessages:
			if msg.SenderID < 0 || msg.SenderID >= NElevators || msg.SenderID == selfID {
				continue
			}

			peerIsConsistent[msg.SenderID] = peerStateMatchesRecorded(msg, systemHallOrders, systemElevStates)
			systemElevStates[msg.SenderID] = msg.ElevatorList[msg.SenderID]
			systemHallOrders[msg.SenderID] = msg.HallOrderList
			peerIsAlive[msg.SenderID] = msg.AliveStatus

			systemHallOrders = advanceLocalOrderStates(systemHallOrders, selfID, peerIsAlive)

			if allAlivePeersConsistent(peerIsConsistent, peerIsAlive, selfID) {
				peerIsConsistent = [NElevators]bool{}
				publishConsistantState(convergedSystemState, peerIsAlive, systemElevStates, systemHallOrders)
			}

		case state := <-localSystemState:
			systemHallOrders[selfID] = state.HallRequests
			systemElevStates[selfID] = state.ElevatorState
			peerIsAlive[selfID] = state.AliveStatus

			systemHallOrders = advanceLocalOrderStates(systemHallOrders, selfID, peerIsAlive)
			broadcastLocalState(outgoingMessages, selfID, peerIsAlive, systemElevStates, systemHallOrders)
			if allAlivePeersConsistent(peerIsConsistent, peerIsAlive, selfID) {
				peerIsConsistent = [NElevators]bool{}
				publishConsistantState(convergedSystemState, peerIsAlive, systemElevStates, systemHallOrders)
			}
		}
	}
}
