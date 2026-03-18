package localControl

import (
	"elevator/internal/config"
	"elevator/internal/localControl/hardware"
	"elevator/internal/types"
)

// -----------------------------------------------------------------------------
// Collects the local control helper routines that mutate elevator state and
// apply it to the hardware interface. These handlers cover initialization,
// motion start/stop, floor serving, timeout recovery, lamp refresh, and
// propagation of local elevator events to the rest of the system.
// -----------------------------------------------------------------------------

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
	hallLamps types.HallLampTable,
) (types.CompletedOrderTable, bool) {

	hardware.SetMotorDirection(types.Stop)
	elevator.PhysicalMotorDirection = types.Stop
	elevator.Behaviour = types.ElevatorDoorOpen
	doorOpenChan <- true
	motorActiveChan <- false

	completed, needsExtraDoorTime := clearOrdersAtFloor(elevator, elevator.CurrentFloor, travelDir)

	refreshLights(*elevator, true, hallLamps)
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

func resumeMovement(elevator *types.Elevator) {
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

func refreshLights(elevator types.Elevator, doorOpen bool, hallLamps types.HallLampTable) {
	for floor := range config.NFloors {
		hardware.SetButtonLamp(types.BtnHallUp, floor, hallLamps[floor][types.BtnHallUp])
		hardware.SetButtonLamp(types.BtnHallDown, floor, hallLamps[floor][types.BtnHallDown])
		hardware.SetButtonLamp(types.BtnCab, floor, elevator.AssignedOrders[floor][types.BtnCab])
	}
	if elevator.CurrentFloor >= 0 {
		hardware.SetFloorIndicator(elevator.CurrentFloor)
	}
	hardware.SetDoorOpenLamp(doorOpen)
}
