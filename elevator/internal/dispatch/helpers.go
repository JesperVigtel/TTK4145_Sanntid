package dispatch

import (
	"elevator/internal/types"
	"path/filepath"
	"runtime"
)

func prepareAssignment(
	convergedState types.ConvergedSystemState,
	localState types.LocalSystemState,
	elevatorID int,
) (types.AssignedOrderTable, types.OrderTable) {
	return computeAssignedOrders(convergedState, localState, elevatorID), convergedState.OrderTables[elevatorID]
}

func replicatedElevatorStateFromEvent(event types.ElevatorEvents) types.HRAElevState {
	return types.NewHRAElevState(event.Elevator, isAssignable(event))
}

func isAssignable(event types.ElevatorEvents) bool {
	return event.Elevator.ActiveStatus && !event.Obstructed
}

func hallRequestAssignerPath() string {
	_, currentFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(currentFile)

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(dir, "hall_request_assigner_mac")
	default: // linux and others
		return filepath.Join(dir, "hall_request_assigner")
	}
}
