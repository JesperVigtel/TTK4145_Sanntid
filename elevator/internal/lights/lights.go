package lights

// Dedikert modul for synkronisering og oppdatering av knappelys.
// Lytter på den globale tilstanden fra nettverket og oppdaterer hall-lys slik at alle heiser viser det samme bildet.
// Garanti for service - sikrer at knappelys blir håndtert korrekt across alle noder.







import (
	"elevator/internal/config"
	"elevator/internal/localControll/hardware"
	"elevator/internal/types"
)


type LightState struct {
	HallLights   [config.NFloors][2]bool // Hall up/down for hver etasje (synkronisert fra nettverk)
	CabLights    [config.NFloors]bool    // Cab-lys (lokal heis)
	CurrentFloor int                     // For etasjeindikator
	DoorOpen     bool                    // Dørlys
}


