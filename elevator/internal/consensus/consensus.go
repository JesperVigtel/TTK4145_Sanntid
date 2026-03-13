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
		//selfCabsRestored  bool 
	)

	systemHallOrders = newSystemHallOrders()

	for {
		select {

		case registry := <-peerEvents:
			peerIsAlive, systemHallOrders = updatePeerAvailability(registry, peerIsAlive, systemHallOrders, selfID)
		
		case msg := <-peerMsg:
			if msg.SenderID < 0 || msg.SenderID >= NElevators || msg.SenderID == selfID {continue}	//Is this possibly surpulus?

			peerIsConsistent[msg.SenderID] 		= peerStateMatchesRecorded(msg, systemHallOrders, systemElevStates)
			systemElevStates 					= adoptPeerElevatorStates(msg.ElevatorList, systemElevStates, selfID) //return selfCabsRestored
			systemHallOrders[msg.SenderID] 		= msg.HallOrderTable
			peerIsAlive[msg.SenderID] 			= msg.AliveStatus

			systemHallOrders = advanceLocalOrderStates(systemHallOrders, selfID, peerIsAlive)
			sendStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders)

			if allAlivePeersConsistent(peerIsConsistent, peerIsAlive, selfID) {
				peerIsConsistent = [NElevators]bool{}
				publishConsistentState(converged, peerIsAlive, systemElevStates, systemHallOrders)
			}
		case state := <-localState:
			systemHallOrders[selfID] 	= state.HallRequests
			//systemElevStates[selfID] 	= state.ElevatorState
			systemElevStates[selfID] 	= mergeCabs(state.ElevatorState, systemElevStates[selfID])
			peerIsAlive[selfID] 		= state.AliveStatus

			systemHallOrders, peerIsConsistent = advanceAndBroadcast(broadcast, converged, selfID, peerIsAlive, peerIsConsistent, systemElevStates, systemHallOrders)
		}
	}
}



func advanceAndBroadcast(
	broadcast 			chan<- Message,
	converged 			chan<- ConvergedSystemState,
	selfID 				int,
	peerIsAlive 		[NElevators]bool,
	peerIsConsistent 	[NElevators]bool,
	systemElevStates 	[NElevators]HRAElevState,
	systemHallOrders 	[NElevators]HallOrderTable,
) ([NElevators]HallOrderTable, [NElevators]bool) {
	systemHallOrders = advanceLocalOrderStates(systemHallOrders, selfID, peerIsAlive)
	sendStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders)

	if allAlivePeersConsistent(peerIsConsistent, peerIsAlive, selfID) {
		peerIsConsistent = [NElevators]bool{}
		publishConsistentState(converged, peerIsAlive, systemElevStates, systemHallOrders)
	}

	return systemHallOrders, peerIsConsistent
}