package decisionMaker

import (
	. "elevator/internal/types"
)

func RunDecisionMaker(
	newLocalOrders       chan<- CabOrderTable,
	localSystemState     chan<- LocalSystemState,
	lightUpdateRequests  chan<- HallOrderTable,
	elevatorStateUpdates <-chan LocalElevatorFromDriver,
	agreedSystemState    <-chan AgreedSystemState,
	buttonEvents         <-chan ButtonEvent,
	elevatorID           int,
) {
	var (
		localState     LocalSystemState
		previousOrders CabOrderTable
	)

	localState = initLocalSystemState(<-elevatorStateUpdates, elevatorID)
	localSystemState <- localState

	for {
		select {
		case hw := <-elevatorStateUpdates:
			localState = applyHardwareUpdate(localState, hw, elevatorID)
			localSystemState <- localState
		
		case btn := <-buttonEvents:
			localState = applyButtonPress(localState, elevatorID, btn)
			localSystemState <- localState
			

		case agreedState := <-agreedSystemState:
			localState, previousOrders = assignOrders(
				agreedState, localState, previousOrders,
				newLocalOrders, lightUpdateRequests, elevatorID,
			)
		}
	}
}


// package decisionMaker

// import (
// 	."elevator/internal/types"
// )


// func RunDecisionMaker(
// 	newLocalOrders 				chan<- 	CabOrderTable,
// 	distributedDecisionBasis	chan<- 	LocalSystemState,
// 	lightUpdateRequests 		chan<- 	HallOrderTable,
// 	elevatorStateUpdates 		<-chan 	LocalElevatorFromDriver,
// 	networkConsensusBasis 		<-chan 	AgreedSystemState,
// 	buttonEvent 				<-chan	ButtonEvent,
// 	elevatorID 							int,
// ) {
// 	var (
// 		//orderEvents         = make(chan OrderEvent)
// 		previousLocalOrders  CabOrderTable
// 	)

// 	initialElevatorState 		:= 	<-	elevatorStateUpdates
// 	initialDecisionBasis 		:= 	<-	networkConsensusBasis
// 	localDecisionBasis 			:= 		initializeLocalDecisionBasis(initialElevatorState, initialDecisionBasis, elevatorID)
// 	distributedDecisionBasis 		<- 	localDecisionBasis

// 	//go hardware.PollButtons(orderEvents)

// 	for {
// 		select {
// 		case newButtonEvent := 	<-buttonEvent:
// 			onButtonEvent(
// 				&localDecisionBasis, elevatorID, newButtonEvent, distributedDecisionBasis,
// 			)

// 		case newElevState 	:= 	<-elevatorStateUpdates:
// 			onElevatorHardwareUpdate(
// 				&localDecisionBasis, elevatorID, newElevState, distributedDecisionBasis,
// 			)

// 		case newConsensusBasis := <-networkConsensusBasis:
// 			onNetworkConsensus(
// 				&localDecisionBasis, newConsensusBasis, elevatorID,
// 				newLocalOrders, &previousLocalOrders, lightUpdateRequests,
// 			)
// 		}
// 	}
// }
