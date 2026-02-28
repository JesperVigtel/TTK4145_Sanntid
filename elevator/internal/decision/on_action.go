package decisionMaker

import( 
	."elevator/internal/types"
)


// onButtonEvent processes a button press by updating the worldview and publishing it to the network.
func onButtonEvent(
	localDecisionBasis 				*DecisionBasisFromAssigner,
	elevatorID 						int,
	buttonEvent 					ButtonEvent,
	decisionBasisUpdates	chan<- 	DecisionBasisFromAssigner,
) {
	*localDecisionBasis = handleButtonPressed(localDecisionBasis, elevatorID, buttonEvent)
	decisionBasisUpdates 	<- 	*localDecisionBasis
}

// onElevatorHardwareUpdate synchronizes the local worldview with current hardware state and notifies the network.
func onElevatorHardwareUpdate(
	localDecisionBasis 				*DecisionBasisFromAssigner,
	elevatorID 						int,
	elevatorUpdate 					LocalElevatorFromDriver,
	decisionBasisUpdates	chan<- 	DecisionBasisFromAssigner,
) {
	*localDecisionBasis = syncElevatorState(elevatorUpdate, localDecisionBasis, elevatorID)
	decisionBasisUpdates 	<- 	*localDecisionBasis
}

// onNetworkConsensus merges network-wide order data, assigns new local orders if needed, and updates button lights.
func onNetworkConsensus(
	localDecisionBasis 				*DecisionBasisFromAssigner,
	networkConsensusBasis 			DecisionBasisFromNetwork,
	elevatorID 						int,
	newLocalOrders 			chan<- 	CabOrderTable,
	previousLocalOrders 			*CabOrderTable,
	lightUpdateRequests 	chan<- 	HallOrderTable,
) {
	*localDecisionBasis 	= 	mergeNetworkHallOrders(localDecisionBasis, networkConsensusBasis, elevatorID)
	localAssignedOrders 	:= 	assignLocalOrders(networkConsensusBasis, elevatorID)
	if localAssignedOrders 	!= 	*previousLocalOrders {
		newLocalOrders 		<- 	localAssignedOrders
		*previousLocalOrders = 	localAssignedOrders
	}
	lightUpdateRequests 	<- 	updateLightStates(networkConsensusBasis, elevatorID)
}
