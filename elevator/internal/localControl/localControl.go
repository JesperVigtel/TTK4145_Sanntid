package localControl

import (
	"elevator/internal/config"
	"elevator/internal/localControl/hardware"
	"elevator/internal/localControl/timer"
	"elevator/internal/types"
)

func Run(
	elevAddr string,
	assignedOrderUpdates <-chan types.AssignedOrderTable,
	hallLampUpdates <-chan types.HallLampTable,
	elevatorEvents chan<- types.ElevatorEvents,
) {
	var (
		floorChan          = make(chan int, config.ChannelBufferSize)
		doorOpenChan       = make(chan bool, config.ChannelBufferSize)
		motorActiveChan    = make(chan bool, config.ChannelBufferSize)
		recoveryEnableChan = make(chan bool, config.ChannelBufferSize)
		doorClosedChan     = make(chan bool, config.ChannelBufferSize)
		motorInactiveChan  = make(chan bool, config.ChannelBufferSize)
		tryRecovery        = make(chan bool, config.ChannelBufferSize)
		obstructionChan    = make(chan bool, config.ChannelBufferSize)
		buttonPressChan    = make(chan types.ButtonEvent, config.ChannelBufferSize)
		obstruction        bool
		hallLamps          types.HallLampTable

		// directionChange is set across select iterations — set by clearOrdersAtFloor, used by doorClosedChan
		directionChange bool
	)
	hardware.Init(elevAddr, config.NFloors)

	go hardware.PollFloorSensor(floorChan)
	go hardware.PollObstructionSwitch(obstructionChan)
	go hardware.PollButtons(buttonPressChan)

	go timer.Timer(doorOpenChan, motorActiveChan, recoveryEnableChan, doorClosedChan, motorInactiveChan, tryRecovery)

	elevator := newElevator()
	obstruction = false
	refreshLights(elevator, false, hallLamps)

	// Initialize by driving down until a floor sensor is reached
	hardware.SetMotorDirection(types.Down)
	sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)

	for {
		select {

		case floor := <-floorChan:
			elevator.CurrentFloor = floor
			elevator.ActiveStatus = true
			recoveryEnableChan <- false
			refreshLights(elevator, elevator.Behaviour == types.ElevatorDoorOpen, hallLamps)

			if elevator.Behaviour == types.ElevatorMoving {
				motorActiveChan <- true
			}
			// No orders to serve - Elevator idle
			if !hasAssignedOrderAbove(elevator) && !hasAssignedOrderBelow(elevator) &&
				!hasOrderAtFloor(elevator, floor) {
				hardware.SetMotorDirection(types.Stop)
				elevator.PhysicalMotorDirection = types.Stop
				elevator.Behaviour = types.ElevatorIdle
				motorActiveChan <- false
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)
				continue
			}

			if shouldStopAtCurrentFloor(elevator, floor) {
				completedOrders, needsExtraDoorTime := stopAndServeFloor(&elevator, doorOpenChan, motorActiveChan, elevator.CurrentTravelDirection, hallLamps)
				directionChange = directionChange || needsExtraDoorTime
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, completedOrders, nil)
				continue
			}

			// Passing floor without order, checking if direction should change
			newDir := chooseDirection(elevator)
			if newDir != types.Stop {
				elevator.CurrentTravelDirection = newDir
				if elevator.PhysicalMotorDirection != newDir || elevator.Behaviour != types.ElevatorMoving {
					elevator.PhysicalMotorDirection = newDir
					elevator.Behaviour = types.ElevatorMoving
					hardware.SetMotorDirection(newDir)
				}
			}
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)

		case assignedOrders := <-assignedOrderUpdates:
			elevator.AssignedOrders = assignedOrders
			refreshLights(elevator, elevator.Behaviour == types.ElevatorDoorOpen, hallLamps)

			if elevator.Behaviour == types.ElevatorDoorOpen && hasOrderAtFloor(elevator, elevator.CurrentFloor) {
				completedOrders, needsExtraDoorTime := clearOrdersAtFloor(&elevator, elevator.CurrentFloor, elevator.CurrentTravelDirection)
				directionChange = directionChange || needsExtraDoorTime
				doorOpenChan <- true
				refreshLights(elevator, true, hallLamps)
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, completedOrders, nil)
				continue
			}

			if elevator.Behaviour != types.ElevatorIdle {
				continue
			}

			// Don't start new movement when stuck between floors — recovery handles that
			if !elevator.ActiveStatus {
				continue
			}

			if hasOrderAtFloor(elevator, elevator.CurrentFloor) {
				completedOrders, needsExtraDoorTime := stopAndServeFloor(&elevator, doorOpenChan, motorActiveChan, elevator.CurrentTravelDirection, hallLamps)
				directionChange = directionChange || needsExtraDoorTime
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, completedOrders, nil)
				continue
			}

			startNextMovement(&elevator, motorActiveChan)
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)

		case hallLampState := <-hallLampUpdates:
			hallLamps = hallLampState
			refreshLights(elevator, elevator.Behaviour == types.ElevatorDoorOpen, hallLamps)

		case obstruction = <-obstructionChan:
			updateActiveStatus(&elevator, obstruction)
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)

		case buttonEvent := <-buttonPressChan:
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, &buttonEvent)

		case <-doorClosedChan:
			if obstruction {
				doorOpenChan <- true
				continue
			}
			// Serve opposite hall order before changing direction — extra door-open cycle required
			if directionChange {
				directionChange = false
				completedOrders := clearOppositeHallOrder(&elevator, elevator.CurrentFloor, elevator.CurrentTravelDirection)
				elevator.CurrentTravelDirection = -elevator.CurrentTravelDirection
				doorOpenChan <- true
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, completedOrders, nil)
				continue
			}
			startNextMovement(&elevator, motorActiveChan)
			refreshLights(elevator, false, hallLamps)
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)

		case <-motorInactiveChan:
			stopOnMotorTimeout(&elevator)
			recoveryEnableChan <- true
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)

		case <-tryRecovery:
			resumeMovement(&elevator)
			recoveryEnableChan <- false
			motorActiveChan <- true
		}
	}
}
