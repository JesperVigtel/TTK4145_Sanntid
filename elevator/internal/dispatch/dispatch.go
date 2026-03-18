package dispatch

import "elevator/internal/types"

// ------------------------------------------------------------------------------
// Translates converged distributed state and local hardware events into local
// execution orders and converged lamp snapshots.
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
		localState             types.LocalSystemState
		previousAssignedOrders types.AssignedOrderTable
	)

	localState = initLocalSystemState(<-elevatorEvents, elevatorID)
	localStateUpdates <- localState

	for {
		select {
		case event := <-elevatorEvents:
			localState = applyElevatorEvent(localState, event)
			localStateUpdates <- localState

		case convergedState := <-convergedStates:
			localState = mergeConvergedOrderStates(localState, convergedState)

			assignedOrders, lampOrderState := prepareAssignment(convergedState, localState, elevatorID)

			if assignedOrders != previousAssignedOrders {
				assignedOrderUpdates <- assignedOrders
				previousAssignedOrders = assignedOrders
			}
			lampOrderUpdates <- lampOrderState
		}
	}
}
