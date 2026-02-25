package decisionMaker


// onButtonEvent processes a button press by updating the worldview and publishing it to the network.
func onButtonEvent(
	localDecisionBasis *decisionBasisFromAssigner,
	elevatorID int,
	btnEvent ButtonEvent,
	decisionBasisUpdates chan<- decisionBasisFromNetwork,
) {
	*&localDecisionBasis = handleButtonPressed(*&localDecisionBasis, elevatorID, btnEvent)
	decisionBasisUpdates <- *localDecisionBasis
}

// onElevatorHardwareUpdate synchronizes the local worldview with current hardware state and notifies the network.
func onElevatorHardwareUpdate(
	localWorldview *FromAssignerToNetwork,
	thisNodeID int,
	elevatorUpdate FromDriverToAssigner,
	worldviewUpdates chan<- FromAssignerToNetwork,
) {
	*localWorldview = syncElevatorState(elevatorUpdate, *localWorldview, thisNodeID)
	worldviewUpdates <- *localWorldview
}

// onNetworkConsensus merges network-wide order data, assigns new local orders if needed, and updates button lights.
func onNetworkConsensus(
	localDecisionBasis *decisionBasisFromAssigner,
	consensusGlobalBasis decisionBasisFromNetwork,
	elevatorID int,
	newLocalOrders chan<- [NFloors][NButtons]bool,
	previousLocalOrders *[NFloors][NButtons]bool,
	lightUpdateRequests chan<- [NFloors][NButtons]ButtonState,
) {
	*&localDecisionBasis = mergeNetworkHallOrders(*&localDecisionBasis, consensusGlobalBasis, elevatorID)
	localAssignedOrders := assignOrders(consensusGlobalBasis, elevatorID)
	if localAssignedOrders != *previousLocalOrders {
		newLocalOrders <- localAssignedOrders
		*previousLocalOrders = localAssignedOrders
	}
	lightUpdateRequests <- updateLightStates(consensusGlobalBasis, elevatorID)
}