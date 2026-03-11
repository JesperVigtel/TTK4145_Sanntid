package dispatch

import (
	. "elevator/internal/types"
	"fmt"
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
		cabsRestored   bool 
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
				var restored bool
				localState, restored = restoreOwnCabsFromNetwork(localState, globalState)
				cabsRestored = restored
				if restored {
					localStateCh <- localState
					fmt.Println("[dispatch] Cabs restored")
				}
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
