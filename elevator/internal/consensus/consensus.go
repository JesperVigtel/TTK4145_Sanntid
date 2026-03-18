package consensus

import (
	"elevator/internal/config"
	"elevator/internal/types"
)

// -----------------------------------------------------------------------------
// Maintains a converged distributed view of elevator state and unified order
// tables. Alive peers exchange full order snapshots, while the shared cyclic
// state machine preserves the original hall/cab transition semantics.
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
		orderTables    [config.NElevators]types.OrderTable
		elevStates     [config.NElevators]types.HRAElevState
		lastPeerStates [config.NElevators]types.HRAElevState
		peerOrderViews [config.NElevators][config.NElevators]types.OrderTable
		peerIsAlive    [config.NElevators]bool
		peerConsistent [config.NElevators]bool
	)

	for {
		select {
		case registry := <-peerEvents:
			peerIsAlive, orderTables = updatePeerAvailability(registry, peerIsAlive, orderTables, selfID)
			lastPeerStates, peerOrderViews = resetPeerSnapshots(registry, lastPeerStates, peerOrderViews, selfID)
			peerConsistent = [config.NElevators]bool{}

		case msg := <-peerMsg:
			if !isRemotePeerID(msg.SenderID, selfID) {
				continue
			}

			peerConsistent[msg.SenderID] = matchesLastPeerState(msg, peerOrderViews, lastPeerStates)
			lastPeerStates[msg.SenderID] = msg.ElevatorList[msg.SenderID]
			peerOrderViews = recordPeerOrderViews(msg, peerOrderViews)
			elevStates = applyPeerState(msg, elevStates)
			orderTables = applyPeerHallRow(msg, orderTables)
			peerIsAlive[msg.SenderID] = msg.AliveStatus

		case state := <-localState:
			elevStates[selfID] = state.ElevatorState
			orderTables[selfID] = state.Orders
			peerIsAlive[selfID] = state.AliveStatus
		}

		orderTables = advanceOrderStates(orderTables, peerOrderViews, selfID, peerIsAlive)
		broadcast <- buildBroadcastState(selfID, peerIsAlive, elevStates, orderTables)
		if alivePeersConsistent(peerConsistent, peerIsAlive, selfID) {
			peerConsistent = [config.NElevators]bool{}
			trySend(converged, buildConvergedState(peerIsAlive, elevStates, orderTables))
		}
	}
}
