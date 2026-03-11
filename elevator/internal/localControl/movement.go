package localControl

import (
	"elevator/internal/localControl/hardware"
	"elevator/internal/types"
)

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

	completed, needsExtraDoorTime := clearOrdersAtFloor(elevator, elevator.CurrentFloor, arrivalDir)

	sendLightUpdate(localLightsChan, *elevator, true)
	return completed, needsExtraDoorTime
}

func startNextMovement(
	elevator *types.Elevator,
	motorActiveChan chan<- bool,
) {
	newDir := chooseDirection(*elevator)
	elevator.CurrentTravelDirection = newDir

	if elevator.CurrentTravelDirection == types.Stop {
		elevator.Behaviour = types.ElevatorIdle
		elevator.PhysicalMotorDirection = types.Stop
		hardware.SetMotorDirection(types.Stop)
	} else {
		elevator.Behaviour = types.ElevatorMoving
		elevator.PhysicalMotorDirection = newDir
		hardware.SetMotorDirection(newDir)
		motorActiveChan <- true
	}
}

func resumeMovement(elevator *types.Elevator) {
	if elevator.Behaviour == types.ElevatorIdle && !elevator.ActiveStatus {
		newDir := chooseDirection(*elevator)
		if newDir == types.Stop {
			// No orders exist, but the elevator may be stuck between floors.
			// Resume movement in the last known direction to reach a floor sensor.
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

func stopOnMotorTimeout(elevator *types.Elevator) {
	if elevator.Behaviour == types.ElevatorMoving {
		elevator.ActiveStatus = false
		elevator.Behaviour = types.ElevatorIdle
		elevator.PhysicalMotorDirection = types.Stop
		hardware.SetMotorDirection(types.Stop)
	}
}
