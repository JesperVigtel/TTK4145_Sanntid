package decisionMaker

import (
	."elevator/internal/types"
)


func syncElevatorState(
	elevatorUpdate LocalElevatorFromDriver, 
	localDecisionBasis *DecisionBasisFromAssigner, 
	elevatorID int) DecisionBasisFromAssigner {
		
	return *localDecisionBasis //Placeholder
}