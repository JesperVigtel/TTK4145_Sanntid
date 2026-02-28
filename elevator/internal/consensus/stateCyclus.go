package sync

import (
	."elevator/internal/types"
	."elevator/interal"
)


func stateCyclus(
	distributedDecisionBasis 	DecisionBasisFromNetwork,
	agreedDecisionBasis 		DecisionBasisFromNetwork,
	localDecisionBasis 			DecisionBasisFromAssigner,
) {



	for{
		select{

		case OrderStandby:


		case OrderPending:


		case OrderAssigned:


		case OrderComplete:
		}
	}
}
