package dispatch

import (
	. "elevator/internal/types"
)

// ------------------------------------------------
// Translates converged distributed state and local hardware events into commands
// for the local elevator: cab order assignments and hall light updates.
// ------------------------------------------------

func RunDispatch(
	newLocalOrders chan<- CabOrderTable,
	localSystemCh chan<- LocalSystemState,
	lightUpdateRequests chan<- HallOrderTable,
	localControlEvents <-chan FromLocalToDM,
	convergedSystemState <-chan ConvergedSystemState,
	elevatorID int,
) {
	var (
		localState     LocalSystemState
		previousOrders CabOrderTable
	)

	localState = initLocalSystemState(<-localControlEvents, elevatorID)
	localSystemCh <- localState

	for {
		select {

		case event := <-localControlEvents:
			if event.NewButtonPress != nil {
				localState = applyButtonPress(localState, *event.NewButtonPress)
			}

			localState = applyHardwareUpdate(localState, event)
			localSystemCh <- localState

		case convergedState := <-convergedSystemState:
			localState = mergeConvergedHallOrders(localState, convergedState, localState.ElevatorID)
			assignedOrders, lightUpdate := prepareAssignment(convergedState, localState)

			// Only forward new assignments to avoid re-interrupting local control
			// with an identical order table on every consensus tick.
			if assignedOrders != previousOrders {
				newLocalOrders <- assignedOrders
				previousOrders = assignedOrders
			}

			lightUpdateRequests <- lightUpdate
		}
	}
}
