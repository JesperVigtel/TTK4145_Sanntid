package dispatch

import (
	"elevator/internal/types"
	"path/filepath"
	"runtime"
)

func newReplicatedElevatorState(event types.ElevatorEvents) types.HRAElevState {
	return types.NewHRAElevState(event.Elevator, isAvailableForAssignment(event))
}

func isAvailableForAssignment(event types.ElevatorEvents) bool {
	return event.Elevator.ActiveStatus && !event.Obstructed
}

func getHRAPath() string {
	_, currentFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(currentFile)

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(dir, "hall_request_assigner_mac")
	default: // linux and others
		return filepath.Join(dir, "hall_request_assigner")
	}
}
