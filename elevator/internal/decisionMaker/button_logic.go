package decisionMaker

import (
	."elevator/internal/types"
)

func handleButtonPressed(localDecisionBasis *DecisionBasisFromAssigner, elevatorID int, orderEvent OrderEvent) DecisionBasisFromAssigner{
	return *localDecisionBasis
}