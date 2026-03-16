package dispatch

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

func mergedCabRequests(localCabs, networkCabs []bool) []bool {
	merged := make([]bool, NFloors)
	for floor := range NFloors {
		var localCall, networkCall bool
		if floor < len(localCabs) {
			localCall = localCabs[floor]
		}
		if floor < len(networkCabs) {
			networkCall = networkCabs[floor]
		}
		merged[floor] = localCall || networkCall
	}
	return merged
}

func restoreOwnCabsFromNetwork(
	localState LocalSystemState,
	convergedState ConvergedSystemState,
) (LocalSystemState, bool) {
	localElevator := convergedState.ElevatorList[localState.ElevatorID]
	mergedCabs := mergedCabRequests(localState.ElevatorState.CabRequests, localElevator.CabRequests)

	restoredAny := false
	for floor := range NFloors {
		if !localState.ElevatorState.CabRequests[floor] && mergedCabs[floor] {
			restoredAny = true
		}
	}
	localState.ElevatorState.CabRequests = mergedCabs
	return localState, restoredAny
}
