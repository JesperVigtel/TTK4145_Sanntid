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

func handleFloorArrival(elevator *types.Elevator, doorOpenChan chan<- bool, localLightsChan chan<- lights.LocalLightUpdate, arrivalDir types.MotorDirection) [config.NFloors][config.NButtons]bool {
	hardware.SetMotorDirection(types.Stop)
	elevator.MotorDirection = types.Stop
	elevator.Behaviour = types.ElevatorDoorOpen
	doorOpenChan <- true

	completed, needsExtraDoorTime := clearOrdersAtFloor(elevator, elevator.CurrentFloor, arrivalDir)
	if needsExtraDoorTime {
		doorOpenChan <- true 
	}

	sendLightUpdate(localLightsChan, *elevator, true)
	return completed
}

func handleDoorClosed(elevator *types.Elevator, motorActiveChan chan<- bool, localLightsChan chan<- lights.LocalLightUpdate) {
	newDir := chooseDirection(*elevator)
	elevator.MotorDirection = newDir

	if newDir == types.Stop {
		elevator.Behaviour = types.ElevatorIdle
	} else {
		elevator.Behaviour = types.ElevatorMoving
		motorActiveChan <- true
	}
	hardware.SetMotorDirection(newDir)
	sendLightUpdate(localLightsChan, *elevator, false)
}

func sendElevatorUpdate(channel chan<- types.FromLocalToDM, elevator types.Elevator, obstructed bool, completed [config.NFloors][config.NButtons]bool, btn *types.ButtonEvent) {
	channel <- types.FromLocalToDM{
		Elevator:       elevator,
		CompletedOrder: completed,
		NewButtonPress: btn,
		Obstructed:     obstructed,
	}
}

func sendLightUpdate(channel chan<- lights.LocalLightUpdate, elevator types.Elevator, doorOpen bool) {
	var cabLights [config.NFloors]bool
	for floor := 0; floor < config.NFloors; floor++ {
		cabLights[floor] = elevator.LocalOrders[floor][int(types.BTCab)]
	}
	channel <- lights.LocalLightUpdate{
		CabLights:    cabLights,
		DoorOpen:     doorOpen,
		CurrentFloor: elevator.CurrentFloor,
	}
}

func updateFloorIndicator(channel chan<- lights.LocalLightUpdate, elevator types.Elevator) {
	sendLightUpdate(channel, elevator, elevator.Behaviour == types.ElevatorDoorOpen)
}
