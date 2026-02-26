package decisionMaker

import (
	."elevator/internal/types"
)

func handleButtonPressed(
	localDecisionBasis *DecisionBasisFromAssigner, 
	elevatorID int, 
	buttonEvent ButtonEvent) DecisionBasisFromAssigner {
		
	return *localDecisionBasis
}