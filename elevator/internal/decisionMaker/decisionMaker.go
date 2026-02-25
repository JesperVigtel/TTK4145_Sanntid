package decisionMaker

import (
	"elevator/internal/config"
	"elevator/internal/localControll/hardware"
	"elevator/internal/localControll/hardware/"
	. "elevator/internal/types"

	"golang.org/x/tools/go/callgraph/cha"
)

// ------------------------------------------------------------------------------------
//	This module makes decisions for witch elevator to take a hall order
// ------------------------------------------------------------------------------------




type HallOrderTable [NFloors][NButtons]OrderState

func decisionMaker(){

	//Function parameters

	
	buttonEvent = make(chan ButtonEvent)

	//worldView = initWorldView evt. decisionBasis
	go hardware.PollButtons(buttonEvent)


	//Three things can happen: localOrder, gloablOrder -> makeDecision, elevatorStatusUpdate

	for{
		select{

		case newElevatorPosision <- 
		case LocalOrderPending <- buttonEvent:
			localDescionBasis = updateDecisionBasis(LocalOrderPending)
			globalDecisionBasis <- localDescionBasis

		case GlobalOrderPending <- :

		case OrderAssigned:

		case OrderComplete:

		}
	}



}


func handlePendingOrder(table HallOrderTable, floor, button int) HallOrderTable {
	if 
}