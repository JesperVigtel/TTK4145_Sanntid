package dispatch

import (
	. "elevator/internal/types"
)

// ------------------------------------------------
// Module based on the LocalSystemState and agreedSystemState data to assign orders
// ------------------------------------------------

func RunDecisionMaker(
	newLocalOrders chan<- CabOrderTable,
	localSystemCh chan<- LocalSystemState,
	lightUpdateRequests chan<- HallOrderTable,
	localControlEvents <-chan FromLocalToDM,
	agreedSystemState <-chan AgreedSystemState,
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

		case agreedState := <-agreedSystemState:
			localState = mergeAgreedHallOrders(localState, agreedState, localState.ElevatorID)
			assignedOrders, lightUpdate := prepareAssignment(agreedState, localState)

			if assignedOrders != previousOrders {
				newLocalOrders <- assignedOrders
				previousOrders = assignedOrders
			}

			lightUpdateRequests <- lightUpdate
		}
	}
}
