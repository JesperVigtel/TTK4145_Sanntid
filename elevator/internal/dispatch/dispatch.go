package dispatch

import (
	. "elevator/internal/types"
)

// ------------------------------------------------------------------------------
// Translates converged distributed state and local hardware events into commands
// for the local elevator: cab order assignments and hall light updates.
// ------------------------------------------------------------------------------

func Run(
	localOrders chan<- LocalOrderTable,
	localStateCh chan<- LocalSystemState,
	buttonLights chan<- ButtonLightUpdate,
	elevEvents <-chan ElevatorEvents,
	convergedSystem <-chan ConvergedSystemState,
	elevatorID int,
) {
	var (
		localState      LocalSystemState
		previousOrders  LocalOrderTable
		cabRecoveryDone bool
	)

	localState = initLocalSystemState(<-elevEvents, elevatorID)
	localStateCh <- localState

	for {
		select {

		case event := <-elevEvents:
			if event.NewButtonPress != nil {
				localState = applyButtonPress(localState, *event.NewButtonPress)
			}
			localState = applyHardwareUpdate(localState, event)
			localStateCh <- localState

		case globalState := <-convergedSystem:
			if !cabRecoveryDone {
				var restored bool
				localState, restored = restoreOwnCabsFromNetwork(localState, globalState)
				cabRecoveryDone = true
				if restored {
					localStateCh <- localState
				}
			}

			localState = mergeConvergedHallOrders(localState, globalState, localState.ElevatorID)
			assignedOrders := computeAssignedOrders(globalState, localState, elevatorID)
			lightUpdate := makeLightUpdate(globalState, elevatorID) //ASsigned orders removed assignedOrders

			if assignedOrders != previousOrders {
				localOrders <- assignedOrders
				previousOrders = assignedOrders
			}
			buttonLights <- lightUpdate
		}
	}
}

func restoreOwnCabsFromNetwork(
	localState LocalSystemState,
	convergedState ConvergedSystemState,
) (LocalSystemState, bool) {
	mergedCabs := MergeCabRequests(
		localState.ElevatorState.CabRequests,
		convergedState.ElevatorList[localState.ElevatorID].CabRequests,
	)

	for floor, active := range mergedCabs {
		if !localState.ElevatorState.CabRequests[floor] && active {
			localState.ElevatorState.CabRequests = mergedCabs
			return localState, true
		}
	}

	localState.ElevatorState.CabRequests = mergedCabs
	return localState, false
}
