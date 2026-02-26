package decisionMaker

import (
	."elevator/internal/types"
)


// RunAssignmentController manages elevator order assignment and system decision basis broadcasting.
// It listens for button events, elevator hardware state updates, and network-wide consensus states.
func RunDecisionMaker(
	newLocalOrders 			chan<- 	CabOrderTable,
	decisionBasisUpdates 	chan<- 	DecisionBasisFromAssigner,
	lightUpdateRequests 	chan<- 	HallOrderTable,
	elevatorStateUpdates 	<-chan 	LocalElevatorFromDriver,
	networkConsensusBasis 	<-chan 	DecisionBasisFromNetwork,
	orderEvent 				<-chan	OrderEvent,
	elevatorID 						int,
) {
	var (
		//orderEvents         = make(chan OrderEvent)
		previousLocalOrders  CabOrderTable
	)

	// Perform initial synchronization with hardware and network consensus.
	initialElevatorState 	:= <-elevatorStateUpdates
	initialDecisionBasis 	:= <-networkConsensusBasis
	localDecisionBasis 		:= initializeLocalDecisionBasis(initialElevatorState, initialDecisionBasis, elevatorID)
	decisionBasisUpdates 	<- localDecisionBasis

	//go hardware.PollButtons(orderEvents)

	for {
		select {
		case btnEvent 	:= 	<-orderEvent:
			onButtonEvent(
				&localDecisionBasis, elevatorID, btnEvent, decisionBasisUpdates,
			)

		case elevState 	:= 	<-elevatorStateUpdates:
			onElevatorHardwareUpdate(
				&localDecisionBasis, elevatorID, elevState, decisionBasisUpdates,
			)

		case consensusBasis := <-networkConsensusBasis:
			onNetworkConsensus(
				&localDecisionBasis, consensusBasis, elevatorID,
				newLocalOrders, &previousLocalOrders, lightUpdateRequests,
			)
		}
	}
}


// initializeLocalWorldview creates the initial decision basis for this elevator.
func initializeLocalDecisionBasis(
	elevatorState 		LocalElevatorFromDriver,
	globalDecisionBasis DecisionBasisFromNetwork,
	elevatorID 			int,
		) DecisionBasisFromAssigner {
	return initializeLocalDecisionBasis(elevatorState, globalDecisionBasis, elevatorID)
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
