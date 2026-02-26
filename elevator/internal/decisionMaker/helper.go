package decisionMaker


import (
	."elevator/internal/types"
)


// initializeLocalWorldview creates the initial decision basis for this elevator.
func initializeLocalDecisionBasis(
	elevatorState 		LocalElevatorFromDriver,
	globalDecisionBasis DecisionBasisFromNetwork,
	elevatorID 			int,
		) DecisionBasisFromAssigner {
	return initializeLocalDecisionBasis(elevatorState, globalDecisionBasis, elevatorID)
}

func updateLightStates(consensusGlobalBasis DecisionBasisFromNetwork, elevatorID int) HallOrderTable{
	return consensusGlobalBasis.HallOrderTable[elevatorID]		//Placeholder
}