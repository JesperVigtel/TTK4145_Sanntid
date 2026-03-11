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
		selfCabsRestored  bool // one-shot: restore own cabs from network on first peer message only
	)

	systemHallOrders = newSystemHallOrders()

	for {
		select {

		case registry := <-peerEvents:
			peerIsAlive, systemHallOrders = updatePeerAvailability(registry, peerIsAlive, systemHallOrders, selfID)
			//Possibly change updatePeerAvailability name
		
		case msg := <-peerMsg:
			if msg.SenderID < 0 || msg.SenderID >= NElevators || msg.SenderID == selfID {
				continue
			}	//Is this possibly surpulus?
			peerIsConsistent[msg.SenderID] = peerStateMatchesRecorded(msg, systemHallOrders, systemElevStates)
			systemElevStates, selfCabsRestored = adoptPeerElevatorStates(msg.ElevatorList, systemElevStates, selfID, selfCabsRestored)

			systemHallOrders[msg.SenderID] = msg.HallOrderTable
			peerIsAlive[msg.SenderID] = msg.AliveStatus

			systemHallOrders, peerIsConsistent = advanceAndBroadcast(broadcast, converged, selfID, peerIsAlive, peerIsConsistent, systemElevStates, systemHallOrders)

		case state := <-localState:
			systemHallOrders[selfID] = state.HallRequests
			systemElevStates[selfID] = state.ElevatorState
			peerIsAlive[selfID] = state.AliveStatus

			systemHallOrders, peerIsConsistent = advanceAndBroadcast(broadcast, converged, selfID, peerIsAlive, peerIsConsistent, systemElevStates, systemHallOrders)
		}
	}
}
