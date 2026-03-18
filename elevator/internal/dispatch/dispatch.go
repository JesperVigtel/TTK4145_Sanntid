package dispatch

import "elevator/internal/types"

// ------------------------------------------------------------------------------
// Translates converged distributed state and local hardware events into local
// execution orders and converged lamp snapshots.
// ------------------------------------------------------------------------------

func Run(
	assignedOrderChan chan<- types.AssignedOrderTable,
	localSystemStateChan chan<- types.LocalSystemState,
	hallLightChan chan<- types.HallLampTable,
	elevatorEventsChan <-chan types.ElevatorEvents,
	convergedStatesChan <-chan types.ConvergedSystemState,
	elevatorID int,
) {
	var (
		localSystemState   types.LocalSystemState
		lastAssignedOrders types.AssignedOrderTable
	)

	localSystemState = initLocalSystemState(<-elevatorEventsChan, elevatorID)
	localSystemStateChan <- localSystemState

	for {
		select {
		case event := <-elevatorEventsChan:
			localSystemState = applyElevatorEvent(localSystemState, event)
			localSystemStateChan <- localSystemState

		case convergedState := <-convergedStatesChan:
			localSystemState = mergeConvergedOrders(localSystemState, convergedState)
			assignedOrders := computeAssignedOrders(convergedState, localSystemState, elevatorID)

			if assignedOrders != lastAssignedOrders {
				assignedOrderChan <- assignedOrders
				lastAssignedOrders = assignedOrders
			}
			hallLightChan <- buildHallLampTable(convergedState.OrderTables[elevatorID])
		}
	}
}
