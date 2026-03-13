package lights

import (
	"elevator/internal/config"
	"elevator/internal/localControl/hardware"
	"elevator/internal/types"
	//mt"
)

func Run(
	//localLights <-chan types.LocalLightUpdate, // from localControl
	buttonLights <-chan types.ButtonLightUpdate, // from dispatch
) {
	for {
		select {

		// case local := <-localLights:
		// 	for floor := range config.NFloors {
		// 		hardware.SetButtonLamp(types.BtnCab, floor, local.CabLights[floor])
		// 	}
		// 	//t.Println("[lights] Sucsessfully updated cab lights")
		// 	if local.CurrentFloor >= 0 {
		// 		hardware.SetFloorIndicator(local.CurrentFloor)
		// 		//t.Println("[lights] Sucsessfully updated floor lights")
		// 	}
		// 	hardware.SetDoorOpenLamp(local.DoorOpen)
		// 	//t.Println("[lights] Sucsessfully updated Door lights")

		// case hall := <-hallLights:
		// 	for floor := range config.NFloors {
		// 		hardware.SetButtonLamp(types.BtnHallUp, floor, hall[floor][types.BtnHallUp] == types.OrderAssigned)
		// 		hardware.SetButtonLamp(types.BtnHallDown, floor, hall[floor][types.BtnHallDown] == types.OrderAssigned)
		// 	}
		// 	//fmt.Println("[lights] Sucsessfully updated hall lights")

		case btn := <-buttonLights:
			for floor := range config.NFloors {
				hardware.SetButtonLamp(types.BtnHallUp, floor, btn.HallLights[floor][types.BtnHallUp] == types.OrderAssigned)
				hardware.SetButtonLamp(types.BtnHallDown, floor, btn.HallLights[floor][types.BtnHallDown] == types.OrderAssigned)
				hardware.SetButtonLamp(types.BtnCab, floor, btn.CabLights[floor])
			}
		
			default:
		}
	}
}
