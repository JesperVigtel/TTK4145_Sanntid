package consensus

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
	"reflect"
)

func newSystemHallOrders() [NElevators]HallOrderTable {
	var table [NElevators]HallOrderTable
	for id := range table {
		table[id] = newStandbyHallOrders()
	}
	return table
}

func newStandbyHallOrders() HallOrderTable {
	var table HallOrderTable
	for floor := range table {
		for btn := range table[floor] {
			table[floor][btn] = OrderStandby
		}
	}
	return table
}

func updatePeerAvailability(
	nodeRegistry GlobalNodeRegistry,
	peerIsAlive [NElevators]bool,
	systemHallOrders [NElevators]HallOrderTable,
) ([NElevators]bool, [NElevators]HallOrderTable) {
	for _, id := range nodeRegistry.Lost {
		if id < 0 || id >= NElevators {
			continue
		}
		peerIsAlive[id] = false
		systemHallOrders[id] = newStandbyHallOrders()
	}
	for _, id := range nodeRegistry.New {
		if id < 0 || id >= NElevators {
			continue
		}
		peerIsAlive[id] = true
	}
	return peerIsAlive, systemHallOrders
}

func peerStateMatchesRecorded(
	msg Message,
	systemHallOrders [NElevators]HallOrderTable,
	systemElevStates [NElevators]HRAElevState,
) bool {
	// Compare only the sender's slot in our system table against what they broadcast.
	// Comparing the full [NElevators]HallOrderTable against a single HallOrderTable
	// would always return false (type mismatch), so convergence would never be reached.
	return reflect.DeepEqual(systemHallOrders[msg.SenderID], msg.HallOrderList) &&
		reflect.DeepEqual(systemElevStates, msg.ElevatorList)
}

func allAlivePeersConverged(
	peerHasConverged [NElevators]bool,
	peerIsAlive [NElevators]bool,
	selfID int,
) bool {
	for id := 0; id < NElevators; id++ {
		if id == selfID {
			continue
		}
		if peerIsAlive[id] && !peerHasConverged[id] {
			return false
		}
	}
	return true
}

// publishAgreedState sends the agreed state non-blocking. If the downstream
// consumer is busy or the channel is full the publication is dropped and will
// be retried on the next convergence event. This prevents the consensus loop
// from stalling and letting peer messages pile up.
func publishAgreedState(
	agreedSystemState chan<- AgreedSystemState,
	peerIsAlive [NElevators]bool,
	systemElevStates [NElevators]HRAElevState,
	systemHallOrders [NElevators]HallOrderTable,
) {
	state := AgreedSystemState{
		AliveList:      peerIsAlive,
		ElevatorList:   systemElevStates,
		HallOrderTable: systemHallOrders,
	}
	select {
	case agreedSystemState <- state:
	default:
	}
}
