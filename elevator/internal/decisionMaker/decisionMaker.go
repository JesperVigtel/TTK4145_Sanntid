package decisionMaker

import (
	"elevator/internal/config"
	"elevator/internal/localControll/hardware"
	"elevator/internal/localControll/hardware/"
	."elevator/internal/types"

	"golang.org/x/tools/go/callgraph/cha"
)


//Denne modulen får inn fysiske elevator-states, og synkronisert verdenssyn, og sørger for å ta korrekt HRA-algoritmen korrekt

// ------------------------------------------------------------------------------------
//	This module makes decisions for witch elevator to take a hall order
// ------------------------------------------------------------------------------------


//initLocalDecisionBasis?



type HallOrderTable [NFloors][NButtons]OrderState

func decisionMaker(){
	//Function parameters
	localElevatorEvent <-chan localElevatorFromDriver
	elevatorState := <-localElevatorEvent

	syncronizedDecisionBasis <-chan globalDecisionBasis



	var (
		buttonEvent = make(chan ButtonEvent)
	)
	

	//worldView = initWorldView evt. decisionBasis
	go hardware.PollButtons(buttonEvent)

	//Three things can happen: localOrder, gloablOrder -> makeDecision, elevatorStatusUpdate

	for{
		select{
			
		

		}
	}



}
