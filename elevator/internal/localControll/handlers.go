package localControl

import (
	"elevator/internal/config"
	"elevator/internal/lights"
	"elevator/internal/localControll/hardware"
	"elevator/internal/types"
)

func elevatorInit() types.Elevator {
	return types.Elevator{
		CurrentFloor:   -1,
		MotorDirection: types.Down,
		LocalOrders:    [config.NFloors][config.NButtons]bool{},
		Behaviour:      types.ElevatorMoving,
		ActiveStatus:   true,
	}
}

func handleFloorArrival(e *types.Elevator, doorOpenChan chan<- bool, localLightsChan chan<- lights.LocalLightUpdate, arrivalDir types.MotorDirection) [config.NFloors][config.NButtons]bool {
	hardware.SetMotorDirection(types.Stop)
	e.MotorDirection = types.Stop
	e.Behaviour = types.ElevatorDoorOpen
	doorOpenChan <- true

	completed, needsExtraDoorTime := clearOrdersAtFloor(e, e.CurrentFloor, arrivalDir)

	if needsExtraDoorTime {
		doorOpenChan <- true // Extra 3 sec for direction change announcement
	}

	sendLightUpdate(localLightsChan, *e, true)
	return completed
}

func handleDoorClosed(e *types.Elevator, motorActiveChan chan<- bool, localLightsChan chan<- lights.LocalLightUpdate) {
	newDir := chooseDirection(*e)
	e.MotorDirection = newDir

	if newDir == types.Stop {
		e.Behaviour = types.ElevatorIdle
	} else {
		e.Behaviour = types.ElevatorMoving
		motorActiveChan <- true
	}
	hardware.SetMotorDirection(newDir)
	sendLightUpdate(localLightsChan, *e, false)
}

func sendElevatorUpdate(ch chan<- types.FromLocalToDM, e types.Elevator, obstructed bool, completed [config.NFloors][config.NButtons]bool, btn *types.ButtonEvent) {
	ch <- types.FromLocalToDM{
		Elevator:       e,
		CompletedOrder: completed,
		NewButtonPress: btn,
		Obstructed:     obstructed,
	}
}

func sendLightUpdate(ch chan<- lights.LocalLightUpdate, e types.Elevator, doorOpen bool) {
	var cabLights [config.NFloors]bool
	for f := 0; f < config.NFloors; f++ {
		cabLights[f] = e.LocalOrders[f][int(types.BTCab)]
	}
	ch <- lights.LocalLightUpdate{
		CabLights:    cabLights,
		DoorOpen:     doorOpen,
		CurrentFloor: e.CurrentFloor,
	}
}

func updateFloorIndicator(ch chan<- lights.LocalLightUpdate, e types.Elevator) {
	sendLightUpdate(ch, e, e.Behaviour == types.ElevatorDoorOpen)
}
