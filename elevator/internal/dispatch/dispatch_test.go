package dispatch

// Black-box / unit tests for the dispatch module.
//
// Two layers are tested here:
//
//  1. Pure state-transformation helpers (applyButtonPress, applyHardwareUpdate,
//     mergeConvergedHallOrders).  These have no side-effects and are the easiest
//     to cover thoroughly with table-driven tests.
//
//  2. RunDispatch itself – the goroutine that is the module's public boundary.
//     We treat it as a true black box: we create all channels, start it, feed
//     inputs, and assert on the outputs. No internal state is inspected.
//
// Run with:
//   go test ./internal/dispatch/...

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	. "elevator/internal/config"
	. "elevator/internal/types"
)

// TestMain runs before all tests and adds the directory containing the
// hall_request_assigner binary to $PATH, so exec.Command can find it on
// any machine that has the repo checked out (including Linux Codespaces).
func TestMain(m *testing.M) {
	absDir, _ := filepath.Abs(".")
	os.Setenv("PATH", absDir+":"+os.Getenv("PATH"))
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func baseElevator() Elevator {
	return Elevator{
		CurrentFloor:   0,
		MotorDirection: Stop,
		ActiveStatus:   true,
	}
}

func baseLocalState(id int) LocalSystemState {
	return LocalSystemState{
		ElevatorID:    id,
		AliveStatus:   true,
		ElevatorState: HRAElevState{Behavior: "idle", Floor: 0, Direction: "stop", CabRequests: make([]bool, NFloors)},
		HallRequests:  HallOrderTable{},
	}
}

// ---------------------------------------------------------------------------
// 1. applyButtonPress
// ---------------------------------------------------------------------------

func TestApplyButtonPress_HallUp_SetsOrderPending(t *testing.T) {
	state := baseLocalState(0)
	event := ButtonEvent{Floor: 2, Button: BTHallUp}

	next := applyButtonPress(state, event)

	if next.HallRequests[2][BTHallUp] != OrderPending {
		t.Errorf("expected OrderPending, got %v", next.HallRequests[2][BTHallUp])
	}
}

func TestApplyButtonPress_HallUp_DoesNotOverrideAssigned(t *testing.T) {
	state := baseLocalState(0)
	state.HallRequests[2][BTHallUp] = OrderAssigned
	event := ButtonEvent{Floor: 2, Button: BTHallUp}

	next := applyButtonPress(state, event)

	if next.HallRequests[2][BTHallUp] != OrderAssigned {
		t.Errorf("expected OrderAssigned to be preserved, got %v", next.HallRequests[2][BTHallUp])
	}
}

func TestApplyButtonPress_Cab_SetsCabRequest(t *testing.T) {
	state := baseLocalState(0)
	event := ButtonEvent{Floor: 1, Button: BTCab}

	next := applyButtonPress(state, event)

	if !next.ElevatorState.CabRequests[1] {
		t.Errorf("expected CabRequest[1] == true")
	}
}

// ---------------------------------------------------------------------------
// 2. applyHardwareUpdate
// ---------------------------------------------------------------------------

func makeHWEvent(floor int, dir MotorDirection) FromLocalToDM {
	e := baseElevator()
	e.CurrentFloor   = floor
	e.MotorDirection = dir
	return FromLocalToDM{Elevator: e}
}

func TestApplyHardwareUpdate_UpdatesFloor(t *testing.T) {
	state := baseLocalState(0)
	hw    := makeHWEvent(3, Up)

	next := applyHardwareUpdate(state, hw)

	if next.ElevatorState.Floor != 3 {
		t.Errorf("expected floor 3, got %d", next.ElevatorState.Floor)
	}
}

func TestApplyHardwareUpdate_CompletedCabOrder_ClearsRequest(t *testing.T) {
	state := baseLocalState(0)
	state.ElevatorState.CabRequests[2] = true

	hw := makeHWEvent(2, Stop)
	hw.CompletedOrder[2][BTCab] = true

	next := applyHardwareUpdate(state, hw)

	if next.ElevatorState.CabRequests[2] {
		t.Errorf("expected CabRequests[2] to be cleared")
	}
}

func TestApplyHardwareUpdate_CompletedHallUp_SetsOrderComplete(t *testing.T) {
	state := baseLocalState(0)
	state.HallRequests[1][BTHallUp] = OrderAssigned

	hw := makeHWEvent(1, Stop)
	hw.CompletedOrder[1][BTHallUp] = true

	next := applyHardwareUpdate(state, hw)

	if next.HallRequests[1][BTHallUp] != OrderComplete {
		t.Errorf("expected OrderComplete, got %v", next.HallRequests[1][BTHallUp])
	}
}

// ---------------------------------------------------------------------------
// 3. mergeConvergedHallOrders
// ---------------------------------------------------------------------------

func TestMergeConvergedHallOrders_Standby_BecomesAssigned(t *testing.T) {
	const id = 0
	localState := baseLocalState(id)

	var converged ConvergedSystemState
	converged.AliveList[id] = true
	converged.HallOrderTable[id][0][BTHallUp] = OrderAssigned

	next := mergeConvergedHallOrders(localState, converged, id)

	if next.HallRequests[0][BTHallUp] != OrderAssigned {
		t.Errorf("expected OrderAssigned, got %v", next.HallRequests[0][BTHallUp])
	}
}

func TestMergeConvergedHallOrders_LocalComplete_NotRegressedByAssigned(t *testing.T) {
	// Once we locally marked an order as Complete, the consensus layer may
	// still say Assigned (it hasn't caught up yet). We must NOT go back.
	const id = 0
	localState := baseLocalState(id)
	localState.HallRequests[0][BTHallUp] = OrderComplete

	var converged ConvergedSystemState
	converged.AliveList[id] = true
	converged.HallOrderTable[id][0][BTHallUp] = OrderAssigned

	next := mergeConvergedHallOrders(localState, converged, id)

	if next.HallRequests[0][BTHallUp] != OrderComplete {
		t.Errorf("regression: OrderComplete was overwritten by stale OrderAssigned")
	}
}

// ---------------------------------------------------------------------------
// 4. RunDispatch – black-box channel test
//
// Strategy: start RunDispatch in a goroutine, prime it with an initial
// hardware event, then send a button press and verify that the updated
// LocalSystemState is forwarded on the localSystemCh output channel.
//
// We skip assertions that require the external hall_request_assigner binary
// (i.e. we do not send an AgreedSystemState), so the test is self-contained.
// ---------------------------------------------------------------------------

func makeInitialHWEvent() FromLocalToDM {
	e := baseElevator()
	e.CurrentFloor = 0
	return FromLocalToDM{Elevator: e}
}

func TestRunDispatch_ButtonPress_PublishesUpdatedLocalState(t *testing.T) {
	newLocalOrders      := make(chan CabOrderTable, 10)
	localSystemCh       := make(chan LocalSystemState, 10)
	lightUpdateRequests := make(chan HallOrderTable, 10)
	localControlEvents  := make(chan FromLocalToDM, 10)
	agreedSystemState   := make(chan ConvergedSystemState, 10)

	// The goroutine blocks on the first event to initialise; feed it now.
	localControlEvents <- makeInitialHWEvent()

	go RunDispatch(
		newLocalOrders,
		localSystemCh,
		lightUpdateRequests,
		localControlEvents,
		agreedSystemState,
		0,
	)

	// Drain the init publish.
	select {
	case <-localSystemCh:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for initial LocalSystemState")
	}

	// Now simulate a cab button press.
	btn := ButtonEvent{Floor: 2, Button: BTCab}
	localControlEvents <- FromLocalToDM{
		Elevator:       baseElevator(),
		NewButtonPress: &btn,
	}

	// Expect an updated state on localSystemCh.
	select {
	case state := <-localSystemCh:
		if !state.ElevatorState.CabRequests[2] {
			t.Errorf("expected CabRequests[2] = true after cab button press")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for LocalSystemState after button press")
	}

	// newLocalOrders should NOT have been written (no ConvergedSystemState sent).
	select {
	case <-newLocalOrders:
		t.Error("unexpected write to newLocalOrders without ConvergedSystemState")
	default:
	}
}
