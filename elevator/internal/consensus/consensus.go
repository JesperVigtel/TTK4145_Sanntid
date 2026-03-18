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
	localSystemState <-chan types.LocalSystemState,
	converged chan<- types.ConvergedSystemState,
	selfID int,
) {
	var (
		orderTables            [config.NElevators]types.OrderTable
		elevStates             [config.NElevators]types.HRAElevState
		lastPeerElevatorStates [config.NElevators]types.HRAElevState
		peerOrderSnapshots     [config.NElevators][config.NElevators]types.OrderTable
		peerIsAlive            [config.NElevators]bool
		peerConsistent         [config.NElevators]bool
	)
	peerIsAlive[selfID] = true

	for {
		select {
		case registry := <-peerEvents:
			peerIsAlive, orderTables = updatePeerAvailability(registry, peerIsAlive, orderTables, selfID)
			lastPeerElevatorStates, peerOrderSnapshots = resetPeerSnapshots(registry, lastPeerElevatorStates, peerOrderSnapshots, selfID)
			peerConsistent = [config.NElevators]bool{}

		case msg := <-peerMsg:
			if !isRemotePeerID(msg.SenderID, selfID) {
				continue
			}

			peerConsistent[msg.SenderID] = matchesLastPeerSnapshot(msg, peerOrderSnapshots, lastPeerElevatorStates)
			lastPeerElevatorStates[msg.SenderID] = msg.ElevatorState
			peerOrderSnapshots[msg.SenderID] = msg.OrderTables
			elevStates[msg.SenderID] = msg.ElevatorState
			orderTables = applyPeerHallRow(msg, orderTables)

		case state := <-localSystemState:
			elevStates[selfID] = state.ElevatorState
			orderTables[selfID] = state.OrderStates
		}

		orderTables = advanceOrderStates(orderTables, peerOrderSnapshots, selfID, peerIsAlive)
		broadcast <- types.Message{
			SenderID:      selfID,
			ElevatorState: elevStates[selfID],
			OrderTables:   orderTables,
		}
		if alivePeersConsistent(peerConsistent, peerIsAlive, selfID) {
			peerConsistent = [config.NElevators]bool{}
			trySend(converged, types.ConvergedSystemState{
				AliveList:    peerIsAlive,
				ElevatorList: elevStates,
				OrderTables:  orderTables,
			})
		}
	}
}
