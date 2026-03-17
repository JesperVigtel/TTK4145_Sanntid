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
	hallLights chan<- HallOrderTable,
	elevEvents <-chan ElevatorEvents,
	convergedSystem <-chan ConvergedSystemState,
	elevatorID int,
) {
	var (
		localState     LocalSystemState
		previousOrders LocalOrderTable
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
			localState = mergeConvergedOrders(localState, globalState)

			assignedOrders := computeAssignedOrders(globalState, localState, elevatorID)
			hallLightUpdate := globalState.HallOrderTable[elevatorID]

			if assignedOrders != previousOrders {
				localOrders <- assignedOrders
				previousOrders = assignedOrders
			}
			hallLights <- hallLightUpdate
		}
	}
}
