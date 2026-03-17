package localControl

import (
	"elevator/internal/config"
	"elevator/internal/localControl/hardware"
	"elevator/internal/types"
	"fmt"
)

// Floor -1 is unknown position, moves down to nearest floor
func newElevator() types.Elevator {
	return types.Elevator{
		CurrentFloor:           -1,
		CurrentTravelDirection: types.Down,
		PhysicalMotorDirection: types.Down,
		LocalOrders:            [config.NFloors][config.NButtons]bool{},
		Behaviour:              types.ElevatorMoving,
		ActiveStatus:           true,
	}
}

func stopAndServeFloor(
	elevator *types.Elevator,
	doorOpenChan chan<- bool,
	motorActiveChan chan<- bool,
	travelDir types.MotorDirection,
) (types.CompletedOrderTable, bool) {

	hardware.SetMotorDirection(types.Stop)
	elevator.PhysicalMotorDirection = types.Stop
	elevator.Behaviour = types.ElevatorDoorOpen
	doorOpenChan <- true
	motorActiveChan <- false

	completed, needsExtraDoorTime := clearOrdersAtFloor(elevator, elevator.CurrentFloor, travelDir)

	applyLocalLights(*elevator, true)
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
		fmt.Printf("[Recovery] chooseDir=%v, CurrentTravelDir=%v, floor=%v, orders=%v\n",
			newDir, elevator.CurrentTravelDirection, elevator.CurrentFloor, elevator.LocalOrders)
		fmt.Printf("travelDirection is %v\n", elevator.CurrentTravelDirection)

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

// Applies lamp outputs that are purely local to this elevator.
// Cab lamps follow the assigned local orders, while floor and door lamps
// follow the current physical elevator state.
func applyLocalLights(elevator types.Elevator, doorOpen bool) {
	for floor := range config.NFloors {
		hardware.SetButtonLamp(types.BtnCab, floor, elevator.LocalOrders[floor][types.BtnCab])
	}
	if elevator.CurrentFloor >= 0 {
		hardware.SetFloorIndicator(elevator.CurrentFloor)
	}
	hardware.SetDoorOpenLamp(doorOpen)
}

// Applies hall lamps from the converged distributed hall-order view.
// This is kept separate from applyLocalLights so hall lamps only reflect
// consensus state, not transient local assignment state.
func applyHallLights(hallLights types.HallOrderTable) {
	for floor := range config.NFloors {
		hardware.SetButtonLamp(types.BtnHallUp, floor, hallLights[floor][types.BtnHallUp] == types.OrderAssigned)
		hardware.SetButtonLamp(types.BtnHallDown, floor, hallLights[floor][types.BtnHallDown] == types.OrderAssigned)
	}
}
