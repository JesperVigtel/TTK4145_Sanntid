package dispatch

import "elevator/internal/types"

// ------------------------------------------------------------------------------
// Translates converged distributed state and local hardware events into local
// execution orders and converged lamp snapshots.
// ------------------------------------------------------------------------------

func Run(
	assignedOrderUpdates chan<- types.AssignedOrderTable,
	localStateUpdates chan<- types.LocalSystemState,
	hallLightUpdates chan<- types.HallLampTable,
	elevatorEvents <-chan types.ElevatorEvents,
	convergedStates <-chan types.ConvergedSystemState,
	elevatorID int,
) {
	var (
		localSystemState   types.LocalSystemState
		lastAssignedOrders types.AssignedOrderTable
	)

	localSystemState = initLocalSystemState(<-elevatorEvents, elevatorID)
	localStateUpdates <- localSystemState

	for {
		select {
		case event := <-elevatorEvents:
			localSystemState = applyElevatorEvent(localSystemState, event)
			localStateUpdates <- localSystemState

		case convergedState := <-convergedStates:
			localSystemState = mergeConvergedOrders(localSystemState, convergedState)
			assignedOrders := computeAssignedOrders(convergedState, localSystemState, elevatorID)

			if assignedOrders != lastAssignedOrders {
				assignedOrderUpdates <- assignedOrders
				lastAssignedOrders = assignedOrders
			}
			hallLightUpdates <- buildHallLampTable(convergedState.OrderTables[elevatorID])
		}
	}
}
