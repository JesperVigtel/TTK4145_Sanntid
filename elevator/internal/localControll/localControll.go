package localControl

// Tilstandsmaskinen (FSM). Håndterer heisens fysiske bevegelse og dørlogikk.
// Inneholder logikk for OnFloorArrival, OnOrderRequest og OnTimerTimeout.
// Snakker med hardware for å styre motor og dør.

// TODO: Implementer FSM-logikk

import(
	"elevator/internal/config"
	"elevator/internal/localControll/hardware"
	"elevator/internal/types"
	"elevator/internal/localControll/timer"
	"time"
	"fmt"
)

func localControl(
	newOrder <- chan [NFloor][NButton]bool
	elevatorEvents chan <- fromLocalToDM
	
	var {
		
	}
)


// Sjekker en array av noe slag som får inn ordre fra decisionMaker, 
// Denne skal da vite hvor heisen er og avhengig av orderen si hvilken retning heisen skal gå
// Den må vite kontinuerlig hvilken etasje den er i, motordirection, og sjekke arrayen hele tiden om den har fått en ny ordre som skal til først
// den skal hele tiden komminisere med decisionMaker om den er Alive (kanskje finne et annet ord??)
// watchdogtimer skal nullstilles hver gang den klarer å lese ny sensorinformasjon, med mindre den er idle uten error, da skal ikke watchdog startes, den skal bare startes når den er moving
// oppdatere Elevator structen
// fortelle decisionMaker hvilke knapper som trykkes, enten cab eller hall for at den kan sette lys
// skal åpne relevant dør når den er i etasjen den skal, skal så bruke ElevatorBehavior og DoorTimer for å fortelle om den er obstructa til decisionMaker