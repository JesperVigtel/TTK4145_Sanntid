package dispatch

import (
	. "elevator/internal/types"
)

// ------------------------------------------------------------------------------
// Translates converged distributed state and local hardware events into commands
// for the local elevator: cab order assignments and hall light updates.
// ------------------------------------------------------------------------------

func Run(
	localOrders 			chan<- LocalOrderTable,
	localStateCh 			chan<- LocalSystemState,
	hallLights 				chan<- HallOrderTable,
	elevEvents 				<-chan ElevatorEvents,
	convergedSystem 		<-chan ConvergedSystemState,
	elevatorID 				int,
) {
	var (
		localState     LocalSystemState
		previousOrders LocalOrderTable
		cabsRestored   bool // one-shot: restore cabs from network once per session
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

			if !cabsRestored {
				localState = restoreOwnCabsFromNetwork(localState, globalState)
				cabsRestored = true
			}

			localState = mergeConvergedHallOrders(localState, globalState, localState.ElevatorID)
			assignedOrders, lightUpdate := prepareAssignment(localState, globalState)

			if assignedOrders != previousOrders {
				localOrders <- assignedOrders
				previousOrders = assignedOrders
			}
			hallLights <- lightUpdate
		}
	}
}
