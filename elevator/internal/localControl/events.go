package localControl

import (
	"elevator/internal/config"
	"elevator/internal/types"
)

func sendElevatorUpdate(
	channel chan<- types.ElevatorEvents,
	elevator types.Elevator,
	obstructed bool,
	completed types.CompletedOrderTable,
	btn *types.ButtonEvent,
) {
	channel <- types.ElevatorEvents{
		Elevator:       elevator,
		CompletedOrder: completed,
		NewButtonPress: btn,
		Obstructed:     obstructed,
	}
}

func sendLightUpdate(channel chan<- types.LocalLightUpdate, elevator types.Elevator, doorOpen bool) {
	var cabLights [config.NFloors]bool
	for floor := range config.NFloors {
		cabLights[floor] = elevator.LocalOrders[floor][int(types.BtnCab)]
	}

	channel <- types.LocalLightUpdate{
		CabLights:    cabLights,
		DoorOpen:     doorOpen,
		CurrentFloor: elevator.CurrentFloor,
	}
}

func updateFloorIndicator(channel chan<- types.LocalLightUpdate, elevator types.Elevator) {
	sendLightUpdate(channel, elevator, elevator.Behaviour == types.ElevatorDoorOpen)
}
