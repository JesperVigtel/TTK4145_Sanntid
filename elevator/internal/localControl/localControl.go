package localControl

import (
	"elevator/internal/config"
	"elevator/internal/localControl/hardware"
	"elevator/internal/localControl/timer"
	"elevator/internal/types"
)

// -----------------------------------------------------------------------------
// Runs the local elevator controller and applies assigned orders plus hall lamp
// updates to the physical elevator. It combines hardware polling, timer events,
// recovery handling, and obstruction state to drive motion, serve orders,
// refresh lamps, and publish local elevator events.
// -----------------------------------------------------------------------------

func Run(
	elevAddr string,
	assignedOrderChan <-chan types.AssignedOrderTable,
	hallLightChan <-chan types.HallLampTable,
	elevatorEventChan chan<- types.ElevatorEvents,
) {
	var (
		currentFloorChan          = make(chan int, config.ChannelBufferSize)
		doorOpenChan       = make(chan bool, config.ChannelBufferSize)
		motorActiveChan    = make(chan bool, config.ChannelBufferSize)
		recoveryEnableChan = make(chan bool, config.ChannelBufferSize)
		doorClosedChan     = make(chan bool, config.ChannelBufferSize)
		motorInactiveChan  = make(chan bool, config.ChannelBufferSize)
		tryRecoveryChan    = make(chan bool, config.ChannelBufferSize)
		obstructionChan    = make(chan bool, config.ChannelBufferSize)
		buttonPressChan    = make(chan types.ButtonEvent, config.ChannelBufferSize)
		obstruction        bool
		directionChange    bool
		hallLamps          types.HallLampTable
	)
	hardware.Init(elevAddr, config.NFloors)

	go hardware.PollFloorSensor(currentFloorChan)
	go hardware.PollObstructionSwitch(obstructionChan)
	go hardware.PollButtons(buttonPressChan)

	go timer.Timer(doorOpenChan, motorActiveChan, recoveryEnableChan, doorClosedChan, motorInactiveChan, tryRecoveryChan)

	elevator := newElevator()
	obstruction = false
	refreshLights(elevator, false, hallLamps)

	hardware.SetMotorDirection(types.Down)
	sendElevatorUpdate(elevatorEventChan, elevator, obstruction, types.CompletedOrderTable{}, nil)

	for {
		select {

		case floor := <-currentFloorChan:
			elevator.CurrentFloor = floor
			elevator.ActiveStatus = true
			recoveryEnableChan <- false
			refreshLights(elevator, elevator.Behaviour == types.ElevatorDoorOpen, hallLamps)

			if elevator.Behaviour == types.ElevatorMoving {
				motorActiveChan <- true
			}

			if !hasAssignedOrderAbove(elevator) && !hasAssignedOrderBelow(elevator) &&
				!hasOrderAtFloor(elevator, floor) {
				hardware.SetMotorDirection(types.Stop)
				elevator.PhysicalMotorDirection = types.Stop
				elevator.Behaviour = types.ElevatorIdle
				motorActiveChan <- false
				sendElevatorUpdate(elevatorEventChan, elevator, obstruction, types.CompletedOrderTable{}, nil)
				continue
			}

			if shouldStopAtCurrentFloor(elevator, floor) {
				completedOrders, needsExtraDoorTime := stopAndServeFloor(&elevator, doorOpenChan, motorActiveChan, elevator.CurrentTravelDirection, hallLamps)
				directionChange = directionChange || needsExtraDoorTime
				sendElevatorUpdate(elevatorEventChan, elevator, obstruction, completedOrders, nil)
				continue
			}

			newDir := chooseDirection(elevator)
			if newDir != types.Stop {
				elevator.CurrentTravelDirection = newDir
				if elevator.PhysicalMotorDirection != newDir || elevator.Behaviour != types.ElevatorMoving {
					elevator.PhysicalMotorDirection = newDir
					elevator.Behaviour = types.ElevatorMoving
					hardware.SetMotorDirection(newDir)
				}
			}
			sendElevatorUpdate(elevatorEventChan, elevator, obstruction, types.CompletedOrderTable{}, nil)

		case assignedOrders := <-assignedOrderChan:
			elevator.AssignedOrders = assignedOrders
			refreshLights(elevator, elevator.Behaviour == types.ElevatorDoorOpen, hallLamps)

			if elevator.Behaviour == types.ElevatorDoorOpen && hasOrderAtFloor(elevator, elevator.CurrentFloor) {
				completedOrders, needsExtraDoorTime := clearOrdersAtFloor(&elevator, elevator.CurrentFloor, elevator.CurrentTravelDirection)
				directionChange = directionChange || needsExtraDoorTime
				doorOpenChan <- true
				refreshLights(elevator, true, hallLamps)
				sendElevatorUpdate(elevatorEventChan, elevator, obstruction, completedOrders, nil)
				continue
			}

			if elevator.Behaviour != types.ElevatorIdle {
				continue
			}

			if !elevator.ActiveStatus {
				continue
			}

			if hasOrderAtFloor(elevator, elevator.CurrentFloor) {
				completedOrders, needsExtraDoorTime := stopAndServeFloor(&elevator, doorOpenChan, motorActiveChan, elevator.CurrentTravelDirection, hallLamps)
				directionChange = directionChange || needsExtraDoorTime
				sendElevatorUpdate(elevatorEventChan, elevator, obstruction, completedOrders, nil)
				continue
			}

			startNextMovement(&elevator, motorActiveChan)
			sendElevatorUpdate(elevatorEventChan, elevator, obstruction, types.CompletedOrderTable{}, nil)

		case hallLightState := <-hallLightChan:
			hallLamps = hallLightState
			refreshLights(elevator, elevator.Behaviour == types.ElevatorDoorOpen, hallLamps)

		case obstruction = <-obstructionChan:
			updateActiveStatus(&elevator, obstruction)
			sendElevatorUpdate(elevatorEventChan, elevator, obstruction, types.CompletedOrderTable{}, nil)

		case buttonEvent := <-buttonPressChan:
			sendElevatorUpdate(elevatorEventChan, elevator, obstruction, types.CompletedOrderTable{}, &buttonEvent)

		case <-doorClosedChan:
			if obstruction {
				doorOpenChan <- true
				continue
			}
			if directionChange {
				directionChange = false
				completedOrders := clearOppositeHallOrder(&elevator, elevator.CurrentFloor, elevator.CurrentTravelDirection)
				elevator.CurrentTravelDirection = -elevator.CurrentTravelDirection
				doorOpenChan <- true
				sendElevatorUpdate(elevatorEventChan, elevator, obstruction, completedOrders, nil)
				continue
			}

			startNextMovement(&elevator, motorActiveChan)
			refreshLights(elevator, false, hallLamps)
			sendElevatorUpdate(elevatorEventChan, elevator, obstruction, types.CompletedOrderTable{}, nil)

		case <-motorInactiveChan:
			stopOnMotorTimeout(&elevator)
			recoveryEnableChan <- true
			sendElevatorUpdate(elevatorEventChan, elevator, obstruction, types.CompletedOrderTable{}, nil)

		case <-tryRecoveryChan:
			resumeMovement(&elevator)
			recoveryEnableChan <- false
			motorActiveChan <- true
		}
	}
}
