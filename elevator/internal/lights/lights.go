package lights

import (
	"elevator/internal/config"
	"elevator/internal/localControll/hardware"
	"elevator/internal/types"
)

type LocalLightUpdate struct {
	CabLights    [config.NFloors]bool
	DoorOpen     bool
	CurrentFloor int
}

type HallLightUpdate struct {
	HallUp   [config.NFloors]bool
	HallDown [config.NFloors]bool
}

func Lights(
	localLights <-chan LocalLightUpdate, // from localControl
	hallLights <-chan HallLightUpdate, // from DecisionMaker
) {
	for {
		select {

		case local := <-localLights:
			for floor := 0; floor < config.NFloors; floor++ {
				hardware.SetButtonLamp(types.BTCab, floor, local.CabLights[floor])
			}
			if local.CurrentFloor >= 0 {
				hardware.SetFloorIndicator(local.CurrentFloor)
			}
			hardware.SetDoorOpenLamp(local.DoorOpen)

		case hall := <-hallLights:

			for floor := 0; floor < config.NFloors; floor++  {
				hardware.SetButtonLamp(types.BTHallUp, floor, hall.HallUp[floor])
				hardware.SetButtonLamp(types.BTHallDown, floor, hall.HallDown[floor])
			}
		}
	}
}