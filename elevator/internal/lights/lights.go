package lights

import (
	"elevator/internal/config"
	"elevator/internal/localControl/hardware"
	"elevator/internal/types"
)

func Lights(
	localLights <-chan types.LocalLightUpdate, // from localControl
	hallLights <-chan types.HallOrderTable, // from dispatch
) {
	for {
		select {

		case local := <-localLights:
			for floor := range config.NFloors {
				hardware.SetButtonLamp(types.BtnCab, floor, local.CabLights[floor])
			}
			if local.CurrentFloor >= 0 {
				hardware.SetFloorIndicator(local.CurrentFloor)
			}
			hardware.SetDoorOpenLamp(local.DoorOpen)

		case hall := <-hallLights:
			for floor := range config.NFloors {

				hardware.SetButtonLamp(types.BtnHallUp, floor, hall[floor][types.BtnHallUp] == types.OrderAssigned)
				hardware.SetButtonLamp(types.BtnHallDown, floor, hall[floor][types.BtnHallDown] == types.OrderAssigned)
			}
		}
	}
}
