package dispatch

import (
	"elevator/internal/types"
)

// ------------------------------------------------------------------------------
// Translates converged distributed state and local hardware events into commands
// for the local elevator: cab order assignments and hall light updates.
// ------------------------------------------------------------------------------

func Run(
	localOrderUpdates chan<- types.LocalOrderTable,
	localStateUpdates chan<- types.LocalSystemState,
	hallLightUpdates chan<- types.HallOrderTable,
	elevatorEvents <-chan types.ElevatorEvents,
	convergedStates <-chan types.ConvergedSystemState,
	elevatorID int,
) {
	var (
		localState         types.LocalSystemState
		lastAssignedOrders types.LocalOrderTable
	)

	localState = initLocalSystemState(<-elevatorEvents, elevatorID)
	localStateUpdates <- localState

	for {
		select {
		case event := <-elevatorEvents:
			localState = applyElevatorEvent(localState, event)
			localStateUpdates <- localState

		case convergedState := <-convergedStates:
			localState = mergeConvergedOrders(localState, convergedState)

			assignedOrders := computeAssignedOrders(convergedState, localState, elevatorID)

			if assignedOrders != lastAssignedOrders {
				localOrderUpdates <- assignedOrders
				lastAssignedOrders = assignedOrders
			}
			hallLightUpdates <- convergedState.HallOrderTable[elevatorID]
		}
	}
}
