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
	peerMsgChan <-chan types.Message,
	broadcastChan chan<- types.Message,
	peerEventsChan <-chan types.GlobalNodeRegistry,
	localSystemStateChan <-chan types.LocalSystemState,
	convergedChan chan<- types.ConvergedSystemState,
	selfID int,
) {
	var (
		orderTables        [config.NElevators]types.OrderTable
		elevStates         [config.NElevators]types.HRAElevState
		peerOrderSnapshots [config.NElevators][config.NElevators]types.OrderTable
		peerIsAlive        [config.NElevators]bool
		peerConsistent     [config.NElevators]bool
	)
	peerIsAlive[selfID] = true

	for {
		select {
		case registry := <-peerEventsChan:
			peerIsAlive, orderTables = updatePeerAvailability(registry, peerIsAlive, orderTables, selfID)
			elevStates, peerOrderSnapshots = resetPeerObservations(registry, elevStates, peerOrderSnapshots, selfID)
			peerConsistent = [config.NElevators]bool{}

		case msg := <-peerMsgChan:
			if !isRemotePeerID(msg.SenderID, selfID) {
				continue
			}

			peerConsistent[msg.SenderID] = matchesLastPeerSnapshot(msg, peerOrderSnapshots, elevStates)
			peerOrderSnapshots[msg.SenderID] = msg.OrderTables
			elevStates[msg.SenderID] = msg.ElevatorState
			orderTables = applyPeerHallRow(msg, orderTables)

		case system := <-localSystemStateChan:
			elevStates[selfID] = system.ElevatorState
			orderTables[selfID] = system.OrderStates
		}

		orderTables = advanceOrderStates(orderTables, peerOrderSnapshots, selfID, peerIsAlive)
		broadcastChan <- types.Message{
			SenderID:      selfID,
			ElevatorState: elevStates[selfID],
			OrderTables:   orderTables,
		}
		if alivePeersConsistent(peerConsistent, peerIsAlive, selfID) {
			peerConsistent = [config.NElevators]bool{}
			trySend(convergedChan, types.ConvergedSystemState{
				AliveList:    peerIsAlive,
				ElevatorList: elevStates,
				OrderTables:  orderTables,
			})
		}
	}
}
