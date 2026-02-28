package decisionMaker

import (
	."elevator/internal/types"
)


// RunAssignmentController manages elevator order assignment and system decision basis broadcasting.
// It listens for button events, elevator hardware state updates, and network-wide consensus states.
func RunDecisionMaker(
	newLocalOrders 				chan<- 	CabOrderTable,
	distributedDecisionBasis	chan<- 	DecisionBasisFromAssigner,
	lightUpdateRequests 		chan<- 	HallOrderTable,
	elevatorStateUpdates 		<-chan 	LocalElevatorFromDriver,
	networkConsensusBasis 		<-chan 	DecisionBasisFromNetwork,
	buttonEvent 				<-chan	ButtonEvent,
	elevatorID 							int,
) {
	var (
		//orderEvents         = make(chan OrderEvent)
		previousLocalOrders  CabOrderTable
	)

	initialElevatorState 		:= 	<-	elevatorStateUpdates
	initialDecisionBasis 		:= 	<-	networkConsensusBasis
	localDecisionBasis 			:= 		initializeLocalDecisionBasis(initialElevatorState, initialDecisionBasis, elevatorID)
	distributedDecisionBasis 		<- 	localDecisionBasis

	//go hardware.PollButtons(orderEvents)

	for {
		select {
		case newButtonEvent := 	<-buttonEvent:
			onButtonEvent(
				&localDecisionBasis, elevatorID, newButtonEvent, distributedDecisionBasis,
			)

		case newElevState 	:= 	<-elevatorStateUpdates:
			onElevatorHardwareUpdate(
				&localDecisionBasis, elevatorID, newElevState, distributedDecisionBasis,
			)

		case newConsensusBasis := <-networkConsensusBasis:
			onNetworkConsensus(
				&localDecisionBasis, newConsensusBasis, elevatorID,
				newLocalOrders, &previousLocalOrders, lightUpdateRequests,
			)
		}
	}
}





// //Denne modulen får inn fysiske elevator-states, og synkronisert verdenssyn, og sørger for å ta korrekt HRA-algoritmen korrekt

// // ------------------------------------------------------------------------------------
// //	This module makes decisions for witch elevator to take a hall order
// // ------------------------------------------------------------------------------------


// //initLocalDecisionBasis?



// type HallOrderTable [NFloors][NButtons]OrderState

// func decisionMaker(){
// 	//Function parameters
// 	localElevatorEvent <-chan localElevatorFromDriver
// 	elevatorState := <-localElevatorEvent

// 	syncronizedDecisionBasis <-chan globalDecisionBasis



// 	var (
// 		buttonEvent = make(chan ButtonEvent)
// 	)
	

// 	//worldView = initWorldView evt. decisionBasis
// 	go hardware.PollButtons(buttonEvent)

// 	//Three things can happen: localOrder, gloablOrder -> makeDecision, elevatorStatusUpdate

// 	for{
// 		select{
			
		

// 		}
// 	}



// }
