package consensus

import (
	"elevator/internal/config"
	"elevator/internal/types"
)

// -----------------------------------------------------------------------------
// Maintains a converged distributed view of elevator, hall, and cab order state.
// Alive peers continuously exchange snapshots, and order states advance through
// the shared cyclic state machine until a consistent system view can be published.
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
		hallOrders     [config.NElevators]types.HallOrderTable
		elevStates     [config.NElevators]types.HRAElevState
		lastPeerStates [config.NElevators]types.HRAElevState
		peerCabViews   [config.NElevators][config.NElevators]types.CabOrderTable
		peerIsAlive    [config.NElevators]bool
		peerConsistent [config.NElevators]bool
	)

	for {
		select {
		case registry := <-peerEvents:
			peerIsAlive, hallOrders = updatePeerAvailability(registry, peerIsAlive, hallOrders, selfID)
			lastPeerStates, peerCabViews = resetPeerSnapshots(registry, lastPeerStates, peerCabViews, selfID)
			peerConsistent = [config.NElevators]bool{}

		case msg := <-peerMsg:
			if !isRemotePeerID(msg.SenderID, selfID) {
				continue
			}

			peerConsistent[msg.SenderID] = matchesLastPeerState(msg, hallOrders, lastPeerStates)
			lastPeerStates[msg.SenderID] = msg.ElevatorList[msg.SenderID]
			peerCabViews = recordPeerCabViews(msg, peerCabViews)
			elevStates = applyPeerState(msg, elevStates)
			hallOrders[msg.SenderID] = msg.HallOrderTable
			peerIsAlive[msg.SenderID] = msg.AliveStatus

		case state := <-localState:
			elevStates[selfID] = state.ElevatorState
			hallOrders[selfID] = state.HallRequests
			peerIsAlive[selfID] = state.AliveStatus
		}

		hallOrders = advanceHallOrderStates(hallOrders, selfID, peerIsAlive)
		elevStates = advanceCabOrderStates(elevStates, peerCabViews, selfID, peerIsAlive)
		broadcast <- buildBroadcastState(selfID, peerIsAlive, elevStates, hallOrders)
		if alivePeersConsistent(peerConsistent, peerIsAlive, selfID) {
			peerConsistent = [config.NElevators]bool{}
			trySend(converged, buildConvergedState(peerIsAlive, elevStates, hallOrders))
		}
	}
}
