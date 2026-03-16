package consensus

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

// recoveryState tracks whether startup recovery is still active and whether the
// network layer has produced at least one peer snapshot.
type recoveryState struct {
	active           bool
	peerSnapshotSeen bool
}

func (r *recoveryState) maybePublish(
	peerIsConsistent *[NElevators]bool,
	peerIsAlive [NElevators]bool,
	systemElevStates [NElevators]HRAElevState,
	systemHallOrders [NElevators]HallOrderTable,
	selfID int,
	broadcast chan<- Message,
	converged chan<- ConvergedSystemState,
) {
	if !r.peerSnapshotSeen || !allAlivePeersConsistent(*peerIsConsistent, peerIsAlive, selfID) {
		return
	}

	*peerIsConsistent = [NElevators]bool{}
	if r.active {
		r.active = false
		sendStateUpdate(broadcast, selfID, peerIsAlive, systemElevStates, systemHallOrders, r.active)
	}
	publishConsistentState(converged, peerIsAlive, systemElevStates, systemHallOrders)
}

func mergeCabRequestsIntoState(base, incoming HRAElevState) HRAElevState {
	if len(incoming.CabRequests) != NFloors {
		return base
	}
	if len(base.CabRequests) != NFloors {
		base.CabRequests = make([]bool, NFloors)
	}
	for floor := range NFloors {
		base.CabRequests[floor] = base.CabRequests[floor] || incoming.CabRequests[floor]
	}
	return base
}

func adoptPeerStates(
	msg Message,
	systemStates [NElevators]HRAElevState,
	selfID int,
	recoveryActive bool,
) [NElevators]HRAElevState {
	if recoveryActive {
		systemStates = restoreSelfFromPeerCopies(msg, systemStates, selfID)
	}
	return adoptSenderState(msg, systemStates)
}

func restoreSelfFromPeerCopies(
	msg Message,
	systemStates [NElevators]HRAElevState,
	selfID int,
) [NElevators]HRAElevState {
	systemStates[selfID] = mergeCabRequestsIntoState(systemStates[selfID], msg.ElevatorList[selfID])
	return systemStates
}

func adoptSenderState(
	msg Message,
	systemStates [NElevators]HRAElevState,
) [NElevators]HRAElevState {
	senderID := msg.SenderID
	if senderID < 0 || senderID >= NElevators {
		return systemStates
	}

	senderState := msg.ElevatorList[senderID]
	if len(senderState.CabRequests) != NFloors {
		return systemStates
	}

	if msg.Recovering {
		systemStates[senderID] = mergeCabRequestsIntoState(senderState, systemStates[senderID])
		return systemStates
	}

	systemStates[senderID] = senderState
	return systemStates
}
