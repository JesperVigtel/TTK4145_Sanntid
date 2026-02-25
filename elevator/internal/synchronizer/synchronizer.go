package synchronizer

import (
	"elevator/internal/config"
	"elevator/internal/localControll/hardware"
	"elevator/internal/localControll/hardware/"
	. "elevator/internal/types"

)


func synchronizer(){

	buttonEvent = 	make(chan ButtonEvent)
	elevatorEvent = make(chan ElevatorEvent)

	//worldView = initWorldView evt. decisionBasis
	go hardware.PollButtons(buttonEvent)
	go hardware.


	//Three things can happen: localOrder, gloablOrder -> makeDecision, elevatorStatusUpdate

	for{
		select{
			
	}
}
