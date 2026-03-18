package dispatch

import "elevator/internal/types"

// ------------------------------------------------------------------------------
// Translates converged distributed state and local hardware events into commands
// for the local elevator: execution-order assignments and converged hall-lamp snapshots.
// ------------------------------------------------------------------------------

func Run(
	assignedOrderUpdates chan<- types.AssignedOrderTable,
	localStateUpdates chan<- types.LocalSystemState,
	lampOrderUpdates chan<- types.OrderTable,
	elevatorEvents <-chan types.ElevatorEvents,
	convergedStates <-chan types.ConvergedSystemState,
	elevatorID int,
) {
	var (
		localState         types.LocalSystemState
		lastAssignedOrders types.AssignedOrderTable
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
				assignedOrderUpdates <- assignedOrders
				lastAssignedOrders = assignedOrders
			}
			lampOrderUpdates <- convergedState.OrderTables[elevatorID]
		}
	}
}
