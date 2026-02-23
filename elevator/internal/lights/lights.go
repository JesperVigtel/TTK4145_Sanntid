package lights

// Dedikert modul for synkronisering og oppdatering av knappelys.
// Lytter på den globale tilstanden fra nettverket og oppdaterer hall-lys slik at alle heiser viser det samme bildet.
// Garanti for service - sikrer at knappelys blir håndtert korrekt across alle noder.







// import (
// 	"elevator/internal/config"
// 	"elevator/internal/localControll/hardware"
// 	"elevator/internal/types"
// )

// // LightState inneholder all informasjon modulen trenger for å sette lys
// type LightState struct {
// 	HallLights   [config.NFloors][2]bool // Hall up/down for hver etasje (synkronisert fra nettverk)
// 	CabLights    [config.NFloors]bool    // Cab-lys (lokal heis)
// 	CurrentFloor int                     // For etasjeindikator
// 	DoorOpen     bool                    // Dørlys
// }

// // Lights kjører som egen goroutine og håndterer all lysstyring
// // lightUpdates: mottar oppdateringer når lys skal endres
// func Lights(lightUpdates <-chan LightState) {
// 	var currentState LightState

// 	// Initialiser alle lys til av
// 	clearAllLights()

// 	for state := range lightUpdates {
// 		updateLights(currentState, state)
// 		currentState = state
// 	}
// }

// // updateLights sammenligner gammel og ny state, oppdaterer kun endrede lys
// func updateLights(old, new LightState) {
// 	// Oppdater hall-lys (up/down)
// 	for floor := 0; floor < config.NFloors; floor++ {
// 		// Hall Up (ikke på øverste etasje)
// 		if floor < config.NFloors-1 {
// 			if old.HallLights[floor][0] != new.HallLights[floor][0] {
// 				hardware.SetButtonLamp(types.BTHallUp, floor, new.HallLights[floor][0])
// 			}
// 		}
// 		// Hall Down (ikke på nederste etasje)
// 		if floor > 0 {
// 			if old.HallLights[floor][1] != new.HallLights[floor][1] {
// 				hardware.SetButtonLamp(types.BTHallDown, floor, new.HallLights[floor][1])
// 			}
// 		}
// 		// Cab-lys
// 		if old.CabLights[floor] != new.CabLights[floor] {
// 			hardware.SetButtonLamp(types.BTCab, floor, new.CabLights[floor])
// 		}
// 	}

// 	// Oppdater etasjeindikator
// 	if old.CurrentFloor != new.CurrentFloor && new.CurrentFloor >= 0 {
// 		hardware.SetFloorIndicator(new.CurrentFloor)
// 	}

// 	// Oppdater dørlys
// 	if old.DoorOpen != new.DoorOpen {
// 		hardware.SetDoorOpenLamp(new.DoorOpen)
// 	}
// }

// // clearAllLights slår av alle lys ved oppstart
// func clearAllLights() {
// 	for floor := 0; floor < config.NFloors; floor++ {
// 		hardware.SetButtonLamp(types.BTHallUp, floor, false)
// 		hardware.SetButtonLamp(types.BTHallDown, floor, false)
// 		hardware.SetButtonLamp(types.BTCab, floor, false)
// 	}
// 	hardware.SetDoorOpenLamp(false)
// 	hardware.SetStopLamp(false)
// }

// // SetSingleLight er en hjelpefunksjon for å sette ett enkelt lys
// // Kan brukes hvis du vil sende enkelthendelser istedenfor full state
// func SetSingleLight(button types.ButtonType, floor int, value bool) {
// 	hardware.SetButtonLamp(button, floor, value)
// }

// // SetDoorLight setter dørlyset direkte
// func SetDoorLight(value bool) {
// 	hardware.SetDoorOpenLamp(value)
// }

// // SetFloorIndicatorLight setter etasjeindikatoren direkte
// func SetFloorIndicatorLight(floor int) {
// 	if floor >= 0 && floor < config.NFloors {
// 		hardware.SetFloorIndicator(floor)
// 	}
// }
