package decisionMaker


import (
	"elevator/internal/config"
	."elevator/internal/types"
)

// Kostnadsfunksjon. Bestemmer hvilken heis som skal ta en hall-ordre.

// ------------------------------------------------------------------------------------
//	This module makes decisions for witch elevator to take a hall order
// ------------------------------------------------------------------------------------

type HallOrderTable [NFloors][NButtons]OrderState

func decisionMaker(){



	for{
		switch{
		case PendingOrder:

		case AssignedOrder:


		case CompletedOrder:

		}
	}



}


func handlePendingOrder(table HallOrderTable, floor, button int) HallOrderTable {
	if 
}