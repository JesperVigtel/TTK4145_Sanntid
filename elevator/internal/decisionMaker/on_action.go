package decisionMaker

import( 
	."elevator/internal/types"
)


// onButtonEvent processes a button press by updating the worldview and publishing it to the network.
func onButtonEvent(
	localDecisionBasis 				*DecisionBasisFromAssigner,
	elevatorID 						int,
	orderEvent 						OrderEvent,
	decisionBasisUpdates	chan<- 	DecisionBasisFromAssigner,
) {
	*localDecisionBasis = handleButtonPressed(localDecisionBasis, elevatorID, orderEvent)
	decisionBasisUpdates 	<- 	*localDecisionBasis
}

// onElevatorHardwareUpdate synchronizes the local worldview with current hardware state and notifies the network.
func onElevatorHardwareUpdate(
	localWorldview 				*DecisionBasisFromAssigner,
	elevatorID 					int,
	elevatorUpdate 				LocalElevatorFromDriver,
	worldviewUpdates	chan<- 	DecisionBasisFromAssigner,
) {
	*localWorldview = syncElevatorState(elevatorUpdate, localWorldview, elevatorID)
	worldviewUpdates 	<- 	*localWorldview
}

// onNetworkConsensus merges network-wide order data, assigns new local orders if needed, and updates button lights.
func onNetworkConsensus(
	localDecisionBasis 				*DecisionBasisFromAssigner,
	consensusGlobalBasis 			DecisionBasisFromNetwork,
	elevatorID 						int,
	newLocalOrders 			chan<- 	CabOrderTable,
	previousLocalOrders 			*CabOrderTable,
	lightUpdateRequests 	chan<- 	HallOrderTable,
) {
	*localDecisionBasis 	= 	mergeNetworkHallOrders(localDecisionBasis, consensusGlobalBasis, elevatorID)
	localAssignedOrders 	:= 	assignLocalOrders(consensusGlobalBasis, elevatorID)
	if localAssignedOrders 	!= 	*previousLocalOrders {
		newLocalOrders 		<- 	localAssignedOrders
		*previousLocalOrders = 	localAssignedOrders
	}
	lightUpdateRequests 	<- 	updateLightStates(consensusGlobalBasis, elevatorID)
}


func updateLightStates(consensusGlobalBasis DecisionBasisFromNetwork, elevatorID int) HallOrderTable{
	return consensusGlobalBasis.HallOrderTable[elevatorID]		//Placeholder
}