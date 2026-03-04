package consensus

// Tests for the consensus module.
//
// Two layers:
//
//  1. Cyclic state-machine unit tests (tryCyclicAdvance, peerStatesHaveDiverged).
//     These cover the core invariant of the consensus algorithm.
//
//  2. RunConsensusManager black-box tests – channels in, channels out.
//
// Run with:
//   go test ./internal/consensus/... -v

import (
	"testing"
	"time"

	. "elevator/internal/config"
	. "elevator/internal/types"
)

// ---------------------------------------------------------------------------
// 1. tryCyclicAdvance – the core state machine
// The cycle is:  Standby -> Pending -> Assigned -> Complete -> Standby
// ---------------------------------------------------------------------------

func TestTryCyclicAdvance_StandbyToPending_WhenOnePeerPending(t *testing.T) {
	next, advanced := tryCyclicAdvance(OrderStandby, []OrderState{OrderPending})
	if !advanced || next != OrderPending {
		t.Errorf("expected Standby→Pending, got advanced=%v next=%v", advanced, next)
	}
}

func TestTryCyclicAdvance_StandbyStays_WhenAllPeersStandby(t *testing.T) {
	_, advanced := tryCyclicAdvance(OrderStandby, []OrderState{OrderStandby, OrderStandby})
	if advanced {
		t.Error("should not advance from Standby when all peers are Standby")
	}
}

func TestTryCyclicAdvance_StandbyStays_WhenPeerIsAssigned(t *testing.T) {
	_, advanced := tryCyclicAdvance(OrderStandby, []OrderState{OrderAssigned})
	if advanced {
		t.Error("should not advance from Standby when a peer is Assigned")
	}
}

func TestTryCyclicAdvance_PendingToAssigned_WhenAllPendingOrAssigned(t *testing.T) {
	next, advanced := tryCyclicAdvance(OrderPending, []OrderState{OrderPending, OrderAssigned})
	if !advanced || next != OrderAssigned {
		t.Errorf("expected Pending→Assigned, got advanced=%v next=%v", advanced, next)
	}
}

func TestTryCyclicAdvance_PendingStays_WhenPeerIsStillStandby(t *testing.T) {
	_, advanced := tryCyclicAdvance(OrderPending, []OrderState{OrderStandby})
	if advanced {
		t.Error("should not advance from Pending while a peer is still Standby")
	}
}

func TestTryCyclicAdvance_AssignedToComplete_WhenOnePeerComplete(t *testing.T) {
	next, advanced := tryCyclicAdvance(OrderAssigned, []OrderState{OrderComplete, OrderAssigned})
	if !advanced || next != OrderComplete {
		t.Errorf("expected Assigned→Complete, got advanced=%v next=%v", advanced, next)
	}
}

func TestTryCyclicAdvance_AssignedStays_WhenNoPeerComplete(t *testing.T) {
	_, advanced := tryCyclicAdvance(OrderAssigned, []OrderState{OrderAssigned})
	if advanced {
		t.Error("should not advance from Assigned when no peer is Complete")
	}
}

func TestTryCyclicAdvance_CompleteToStandby_WhenAllCompleteOrStandby(t *testing.T) {
	next, advanced := tryCyclicAdvance(OrderComplete, []OrderState{OrderStandby, OrderComplete})
	if !advanced || next != OrderStandby {
		t.Errorf("expected Complete→Standby, got advanced=%v next=%v", advanced, next)
	}
}

func TestTryCyclicAdvance_CompleteStays_WhenPeerIsAssigned(t *testing.T) {
	_, advanced := tryCyclicAdvance(OrderComplete, []OrderState{OrderAssigned})
	if advanced {
		t.Error("should not advance from Complete while a peer is in Assigned")
	}
}

// ---------------------------------------------------------------------------
// 2. peerStatesHaveDiverged – safety reset logic
// ---------------------------------------------------------------------------

func TestPeerStatesHaveDiverged_PendingWithCompleteIsDiverged(t *testing.T) {
	if !peerStatesHaveDiverged(OrderPending, []OrderState{OrderComplete}) {
		t.Error("Pending with peer=Complete should be flagged as diverged")
	}
}

func TestPeerStatesHaveDiverged_AssignedWithPendingIsNotDiverged(t *testing.T) {
	if peerStatesHaveDiverged(OrderAssigned, []OrderState{OrderPending}) {
		t.Error("Assigned with peer=Pending should not be flagged as diverged")
	}
}

func TestPeerStatesHaveDiverged_StandbyWithCompleteIsNotDiverged(t *testing.T) {
	// Standby diverges only when NOT (Standby|Pending) AND NOT (Standby|Complete).
	// Complete satisfies allAreEither(Complete, Standby|Complete) → no diverge.
	if peerStatesHaveDiverged(OrderStandby, []OrderState{OrderComplete}) {
		t.Error("Standby with peer=Complete should not be flagged as diverged")
	}
}

// ---------------------------------------------------------------------------
// 3. updatePeerAvailability – peer tracking
// ---------------------------------------------------------------------------

func TestUpdatePeerAvailability_LostPeer_ResetsOrdersAndAlive(t *testing.T) {
	var alive [NElevators]bool
	alive[1] = true

	orders := newSystemHallOrders()
	orders[1][0][BTHallUp] = OrderAssigned

	registry := GlobalNodeRegistry{Lost: []int{1}}
	newAlive, newOrders := updatePeerAvailability(registry, alive, orders)

	if newAlive[1] {
		t.Error("peer 1 should be marked not-alive after being lost")
	}
	if newOrders[1][0][BTHallUp] != OrderStandby {
		t.Errorf("peer 1 orders should be reset to Standby, got %v", newOrders[1][0][BTHallUp])
	}
}

func TestUpdatePeerAvailability_NewPeer_MarkedAlive(t *testing.T) {
	var alive [NElevators]bool
	orders := newSystemHallOrders()

	registry := GlobalNodeRegistry{New: []int{2}}
	newAlive, _ := updatePeerAvailability(registry, alive, orders)

	if !newAlive[2] {
		t.Error("peer 2 should be marked alive after joining")
	}
}

// ---------------------------------------------------------------------------
// 4. allAlivePeersConsistent
// ---------------------------------------------------------------------------

func TestAllAlivePeersConsistent_NoPeers_ReturnsTrue(t *testing.T) {
	var consistent [NElevators]bool
	var alive [NElevators]bool
	alive[0] = true // only self

	if !allAlivePeersConsistent(consistent, alive, 0) {
		t.Error("should be consistent immediately when there are no peers")
	}
}

func TestAllAlivePeersConsistent_InconsistentPeer_ReturnsFalse(t *testing.T) {
	var consistent [NElevators]bool
	var alive [NElevators]bool
	alive[0] = true
	alive[1] = true // peer 1 alive but state not yet consistent

	if allAlivePeersConsistent(consistent, alive, 0) {
		t.Error("should not be consistent when an alive peer's state does not match")
	}
}

func TestAllAlivePeersConsistent_DeadPeerNotRequired(t *testing.T) {
	var consistent [NElevators]bool
	var alive [NElevators]bool
	alive[0] = true
	// peer 1 dead – its consistency flag should not matter

	if !allAlivePeersConsistent(consistent, alive, 0) {
		t.Error("dead peer should not block consistency check")
	}
}

// ---------------------------------------------------------------------------
// 5. RunConsensusManager – black-box channel tests
// ---------------------------------------------------------------------------

func TestRunConsensusManager_SingleNode_PublishesOnLocalState(t *testing.T) {
	incomingMessages    := make(chan Message, 10)
	nodeRegistryEvents  := make(chan GlobalNodeRegistry, 10)
	localSystemStateCh  := make(chan LocalSystemState, 10)
	agreedSystemStateCh := make(chan AgreedSystemState, 10)

	go RunConsensusManager(
		incomingMessages,
		nodeRegistryEvents,
		localSystemStateCh,
		agreedSystemStateCh,
		0,
	)

	localSystemStateCh <- LocalSystemState{
		ElevatorID:  0,
		AliveStatus: true,
		ElevatorState: HRAElevState{
			Behavior:    "idle",
			Floor:       0,
			Direction:   "stop",
			CabRequests: make([]bool, NFloors),
		},
	}

	select {
	case agreed := <-agreedSystemStateCh:
		if !agreed.AliveList[0] {
			t.Error("self should be marked alive in AgreedSystemState")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for AgreedSystemState")
	}
}

func TestRunConsensusManager_LostPeer_ResetsItsOrders(t *testing.T) {
	incomingMessages    := make(chan Message, 10)
	nodeRegistryEvents  := make(chan GlobalNodeRegistry, 10)
	localSystemStateCh  := make(chan LocalSystemState, 10)
	agreedSystemStateCh := make(chan AgreedSystemState, 10)

	go RunConsensusManager(
		incomingMessages,
		nodeRegistryEvents,
		localSystemStateCh,
		agreedSystemStateCh,
		0,
	)

	nodeRegistryEvents <- GlobalNodeRegistry{Lost: []int{1}}

	localSystemStateCh <- LocalSystemState{
		ElevatorID:  0,
		AliveStatus: true,
		ElevatorState: HRAElevState{
			Behavior:    "idle",
			Floor:       0,
			Direction:   "stop",
			CabRequests: make([]bool, NFloors),
		},
	}

	select {
	case agreed := <-agreedSystemStateCh:
		if agreed.AliveList[1] {
			t.Error("peer 1 should not be alive after being reported lost")
		}
		for floor := 0; floor < NFloors; floor++ {
			for btn := 0; btn < NButtons; btn++ {
				if agreed.HallOrderTable[1][floor][btn] != OrderStandby {
					t.Errorf("peer 1 orders should be Standby after loss, got %v at [%d][%d]",
						agreed.HallOrderTable[1][floor][btn], floor, btn)
				}
			}
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for AgreedSystemState")
	}
}
