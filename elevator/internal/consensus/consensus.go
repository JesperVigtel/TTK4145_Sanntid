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
	peerMsg   	<-chan Message,
	broadcast  	chan<- Message,
	peerEvents 	<-chan GlobalNodeRegistry,
	localState 	<-chan LocalSystemState,
	converged 	chan<- ConvergedSystemState,
	selfID 		int,
) {
	var (
		systemHallOrders  [NElevators]HallOrderTable
		systemElevStates  [NElevators]HRAElevState
		peerIsAlive       [NElevators]bool
		peerIsConsistent  [NElevators]bool
	)

	systemHallOrders = newSystemHallOrders()
	recoveryMode := true

	for {
		select {

		case registry := <-peerEvents:
			peerIsAlive, systemHallOrders = updatePeerAvailability(registry, peerIsAlive, systemHallOrders, selfID)
		
		case msg := <-peerMsg:
			peerIsConsistent[msg.SenderID] = peerStateMatchesRecorded(msg, systemHallOrders, systemElevStates)
			systemElevStates = adoptPeerStates(msg.ElevatorList, systemElevStates, selfID, recoveryMode)
			systemHallOrders[msg.SenderID] = msg.HallOrderTable
			peerIsAlive[msg.SenderID] = msg.AliveStatus
		
			systemHallOrders = advanceLocalOrderStates(systemHallOrders, selfID, peerIsAlive)
			sendStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders)
		
			if allAlivePeersConsistent(peerIsConsistent, peerIsAlive, selfID) {
				peerIsConsistent = [NElevators]bool{}
				publishConsistentState(converged, peerIsAlive, systemElevStates, systemHallOrders)
				recoveryMode = false
			}

		
		case state := <-localState:
			if recoveryMode {
				systemElevStates[selfID] = mergeCabs(state.ElevatorState, systemElevStates[selfID])
			} else {
				systemElevStates[selfID] = state.ElevatorState
			}
			
			systemHallOrders[selfID] = state.HallRequests
			peerIsAlive[selfID] = state.AliveStatus
		
			sendStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders)
		}
	}
}
