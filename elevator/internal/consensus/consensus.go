package consensus

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

func RunConsensusManager(
	incomingMessages 	<-chan Message,
	nodeRegistryEvents 	<-chan GlobalNodeRegistry,
	localSystemState 	<-chan LocalSystemState,
	agreedSystemState 	chan<- AgreedSystemState,
	selfID 				int,
) {
	var (
		systemHallOrders [NElevators]HallOrderTable
		systemElevStates [NElevators]HRAElevState
		peerIsAlive      [NElevators]bool
		peerHasConverged [NElevators]bool
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

			peerHasConverged[msg.SenderID] 	= peerStateMatchesRecorded(msg, systemHallOrders, systemElevStates)
			systemElevStates[msg.SenderID] 	= msg.ElevatorList[msg.SenderID]
			systemHallOrders[msg.SenderID] 	= msg.HallOrderList
			peerIsAlive[msg.SenderID] 		= msg.AliveStatus

			systemHallOrders = advanceLocalOrderStates(systemHallOrders, selfID, peerIsAlive)

			if allAlivePeersConverged(peerHasConverged, peerIsAlive, selfID) {
				peerHasConverged = [NElevators]bool{}
				publishAgreedState(agreedSystemState, peerIsAlive, systemElevStates, systemHallOrders)
			}

		case state := <-localSystemState:
			systemHallOrders[selfID] = state.HallRequests
			systemElevStates[selfID] = state.ElevatorState
			peerIsAlive[selfID] = state.AliveStatus

			systemHallOrders = advanceLocalOrderStates(systemHallOrders, selfID, peerIsAlive)
			if allAlivePeersConverged(peerHasConverged, peerIsAlive, selfID) {
				peerHasConverged = [NElevators]bool{}
			publishAgreedState(agreedSystemState, peerIsAlive, systemElevStates, systemHallOrders)
			}
		}
	}
}
