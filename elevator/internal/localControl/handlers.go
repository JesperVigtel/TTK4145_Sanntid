package localControl

import (
	"elevator/internal/config"
	"elevator/internal/localControl/hardware"
	"elevator/internal/types"
)

func elevatorInit() types.Elevator {
	return types.Elevator{
		CurrentFloor:           -1,
		CurrentTravelDirection: types.Down,
		PhysicalMotorDirection: types.Down,
		LocalOrders:            [config.NFloors][config.NButtons]bool{},
		Behaviour:              types.ElevatorMoving,
		ActiveStatus:           true,
	}
}

func handleFloorArrival(
	elevator *types.Elevator,
	doorOpenChan chan<- bool,
	motorActiveChan chan<- bool,
	localLightsChan chan<- types.LocalLightUpdate,
	arrivalDir types.MotorDirection,
) (types.CompletedOrderTable, bool) {

	hardware.SetMotorDirection(types.Stop)
	elevator.PhysicalMotorDirection = types.Stop
	elevator.Behaviour = types.ElevatorDoorOpen
	doorOpenChan <- true
	motorActiveChan <- false

	completed, needsExtraDoorTime := clearOrdersAtFloor(elevator, elevator.CurrentFloor, arrivalDir, false)

	sendLightUpdate(localLightsChan, *elevator, true)
	return completed, needsExtraDoorTime
}

func handleDoorClosed(
	elevator *types.Elevator,
	motorActiveChan chan<- bool,
	localLightsChan chan<- types.LocalLightUpdate,
) {
	newDir := chooseDirection(*elevator)
	elevator.CurrentTravelDirection = newDir

	if newDir == types.Stop {
		elevator.Behaviour = types.ElevatorIdle
		elevator.PhysicalMotorDirection = types.Stop
		hardware.SetMotorDirection(types.Stop)
	} else {
		elevator.Behaviour = types.ElevatorMoving
		elevator.PhysicalMotorDirection = newDir
		hardware.SetMotorDirection(newDir)
		motorActiveChan <- true
	}

	sendLightUpdate(localLightsChan, *elevator, false)
}

// Denne er good

func tryMoving(elevator *types.Elevator,) {

	if elevator.Behaviour == types.ElevatorIdle && !elevator.ActiveStatus {
		newDir := chooseDirection(*elevator)
		if newDir == types.Stop {
			elevator.Behaviour = types.ElevatorMoving
			elevator.PhysicalMotorDirection = elevator.CurrentTravelDirection
			hardware.SetMotorDirection(elevator.CurrentTravelDirection)
			return
		}
		elevator.CurrentTravelDirection = newDir
		elevator.PhysicalMotorDirection = newDir
		elevator.Behaviour = types.ElevatorMoving
		hardware.SetMotorDirection(newDir)
	}
}

func killElevator(elevator *types.Elevator,){
				if elevator.Behaviour == types.ElevatorMoving {
				elevator.ActiveStatus = false
				elevator.Behaviour = types.ElevatorIdle
				elevator.PhysicalMotorDirection = types.Stop
				hardware.SetMotorDirection(types.Stop)
			}
}


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
