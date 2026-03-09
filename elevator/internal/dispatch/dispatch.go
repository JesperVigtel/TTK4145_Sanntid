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
	)

	localState = initLocalSystemState(<-elevEvents, elevatorID)
	localStateCh <- localState
	fmt.Println("[dispatch] succsessfully initialized")
	

	for {
		select {

		case event := <-elevEvents:
			if event.NewButtonPress != nil {
				localState = applyButtonPress(localState, *event.NewButtonPress)
				fmt.Println("[dispatch] Sucsessfully received button press")
			}
			fmt.Printf("[dispatch] Sucsessfully received hardvare event")
			localState = applyHardwareUpdate(localState, event)
			localStateCh <- localState

		case globalState := <-convergedSystem:
			fmt.Println("[dispatch] Sucsessfully received  coverged system")
			localState = mergeConvergedHallOrders(localState, globalState, localState.ElevatorID)
			assignedOrders, lightUpdate := prepareAssignment(localState, globalState)

			if assignedOrders != previousOrders {
				localOrders <- assignedOrders
				previousOrders = assignedOrders
				fmt.Println("[dispatch] Sucsessfully asigned orders")
			}

			hallLights <- lightUpdate
		}
	}
}
