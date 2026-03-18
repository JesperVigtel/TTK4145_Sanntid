package localControl

import (
	"elevator/internal/config"
	"elevator/internal/localControl/hardware"
	"elevator/internal/types"
)

// Floor -1 is unknown position, moves down to nearest floor
func newElevator() types.Elevator {
	return types.Elevator{
		CurrentFloor:           -1,
		CurrentTravelDirection: types.Down,
		PhysicalMotorDirection: types.Down,
		AssignedOrders:         [config.NFloors][config.NButtons]bool{},
		Behaviour:              types.ElevatorMoving,
		ActiveStatus:           true,
	}
}

func stopAndServeFloor(
	elevator *types.Elevator,
	doorOpenChan chan<- bool,
	motorActiveChan chan<- bool,
	travelDir types.MotorDirection,
	lampOrderStates types.OrderTable,
) (types.CompletedOrderTable, bool) {

	hardware.SetMotorDirection(types.Stop)
	elevator.PhysicalMotorDirection = types.Stop
	elevator.Behaviour = types.ElevatorDoorOpen
	doorOpenChan <- true
	motorActiveChan <- false

	completed, needsExtraDoorTime := clearOrdersAtFloor(elevator, elevator.CurrentFloor, travelDir)

	refreshLights(*elevator, true, lampOrderStates)
	return completed, needsExtraDoorTime
}

func startNextMovement(
	elevator *types.Elevator,
	motorActiveChan chan<- bool,
) {
	newDir := chooseDirection(*elevator)

	if newDir == types.Stop {
		elevator.Behaviour = types.ElevatorIdle
		elevator.PhysicalMotorDirection = types.Stop
		hardware.SetMotorDirection(types.Stop)
	} else {
		elevator.CurrentTravelDirection = newDir
		elevator.Behaviour = types.ElevatorMoving
		elevator.PhysicalMotorDirection = newDir
		hardware.SetMotorDirection(newDir)
		motorActiveChan <- true
	}
}

// Called periodically after motor timeout — retries movement to escape stuck state
func resumeMovement(elevator *types.Elevator) {
	if elevator.Behaviour == types.ElevatorIdle && !elevator.ActiveStatus {
		newDir := chooseDirection(*elevator)

		// No orders found — keep moving in current direction to reach a floor sensor
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

func stopOnMotorTimeout(elevator *types.Elevator) {
	if elevator.Behaviour == types.ElevatorMoving {
		elevator.ActiveStatus = false
		elevator.Behaviour = types.ElevatorIdle
		elevator.PhysicalMotorDirection = types.Stop
		hardware.SetMotorDirection(types.Stop)
	}
}

func updateActiveStatus(elevator *types.Elevator, obstruction bool) {
	if elevator.Behaviour == types.ElevatorDoorOpen && obstruction {
		elevator.ActiveStatus = false
	} else {
		elevator.ActiveStatus = true
	}
}

func sendElevatorUpdate(
	channel chan<- types.ElevatorEvents,
	elevator types.Elevator,
	obstructed bool,
	completedOrders types.CompletedOrderTable,
	newButtonPress *types.ButtonEvent,
) {
	channel <- types.ElevatorEvents{
		Elevator:        elevator,
		CompletedOrders: completedOrders,
		NewButtonPress:  newButtonPress,
		Obstructed:      obstructed,
	}
}

// refreshLights merges the converged hall-order snapshot with the elevator's
// current assigned orders, while floor and door lamps follow local state.
func refreshLights(elevator types.Elevator, doorOpen bool, lampOrderStates types.OrderTable) {
	for floor := range config.NFloors {
		hardware.SetButtonLamp(types.BtnHallUp, floor, lampOrderStates[floor][types.BtnHallUp] == types.OrderAssigned)
		hardware.SetButtonLamp(types.BtnHallDown, floor, lampOrderStates[floor][types.BtnHallDown] == types.OrderAssigned)
		hardware.SetButtonLamp(types.BtnCab, floor, elevator.AssignedOrders[floor][types.BtnCab])
	}
	if elevator.CurrentFloor >= 0 {
		hardware.SetFloorIndicator(elevator.CurrentFloor)
	}
	hardware.SetDoorOpenLamp(doorOpen)
}
