package localControl

import (
	"elevator/internal/config"
	"elevator/internal/localControl/hardware"
	"elevator/internal/localControl/timer"
	"elevator/internal/types"
	"fmt"
)

func Run(
	elevAddr string,
	newOrder <-chan types.LocalOrderTable,
	elevatorEvents chan<- types.ElevatorEvents,
	localLightsChan chan<- types.LocalLightUpdate,
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

	// Initialize by driving down until a floor sensor is reached
	hardware.SetMotorDirection(types.Down)
	sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)

	for {
		select {

		case floor := <-floorChan:
			elevator.CurrentFloor = floor
			elevator.ActiveStatus = true
			recoveryEnableChan <- false
			updateFloorIndicator(localLightsChan, elevator)

			if elevator.Behaviour == types.ElevatorMoving {
				motorActiveChan <- true
			}
			// No orders to serve - Elevator idle
			if !hasLocalOrderAbove(elevator) && !hasLocalOrderBelow(elevator) &&
				!hasOrderAtFloor(elevator, floor) {
				hardware.SetMotorDirection(types.Stop)
				elevator.PhysicalMotorDirection = types.Stop
				elevator.Behaviour = types.ElevatorIdle
				motorActiveChan <- false
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)
				continue
			}

			
			if shouldStopAtCurrentFloor(elevator, floor) {
				completedOrders, needsExtraDoorTime := stopAndServeFloor(&elevator, doorOpenChan, motorActiveChan, localLightsChan, elevator.CurrentTravelDirection)
				directionChange = directionChange || needsExtraDoorTime
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, completedOrders, nil)
				continue
			}

			// Passing floor without order, checking if direction should change
			newDir := chooseDirection(elevator)
			if newDir != types.Stop && newDir != elevator.CurrentTravelDirection {
				elevator.CurrentTravelDirection = newDir
				elevator.PhysicalMotorDirection = newDir
				hardware.SetMotorDirection(newDir)
			}
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)

		case orders := <-newOrder:
			elevator.LocalOrders = orders
			sendLightUpdate(localLightsChan, elevator, elevator.Behaviour == types.ElevatorDoorOpen)

			if elevator.Behaviour == types.ElevatorDoorOpen && hasOrderAtFloor(elevator, elevator.CurrentFloor) {
				completedOrders, needsExtraDoorTime := clearOrdersAtFloor(&elevator, elevator.CurrentFloor, elevator.CurrentTravelDirection)
				directionChange = directionChange || needsExtraDoorTime
				doorOpenChan <- true
				sendLightUpdate(localLightsChan, elevator, true)
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
				completedOrders, needsExtraDoorTime := stopAndServeFloor(&elevator, doorOpenChan, motorActiveChan, localLightsChan, elevator.CurrentTravelDirection)
				directionChange = directionChange || needsExtraDoorTime
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, completedOrders, nil)
				continue
			}

			startNextMovement(&elevator, motorActiveChan)
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)

		case obstruction = <-obstructionChan:
			fmt.Printf("[LocalControl] Obstruction changed: %v\n", obstruction)
			updateActiveStatus(&elevator, obstruction)
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)

		case buttonEvent := <-buttonPressChan:
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, &buttonEvent)

		case <-doorClosedChan:
			if obstruction {
				doorOpenChan <- true
				continue
			}
			// Serve opposite hall order before reversing — extra door-open cycle required
			if directionChange {
				directionChange = false
				completedOrders := clearOppositeHallOrder(&elevator, elevator.CurrentFloor, elevator.CurrentTravelDirection)
				elevator.CurrentTravelDirection = -elevator.CurrentTravelDirection
				fmt.Printf("Be advised, the elevator will now change traveldirection")
				doorOpenChan <- true
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, completedOrders, nil)
				continue
			}
			startNextMovement(&elevator, motorActiveChan)
			sendLightUpdate(localLightsChan, elevator, false)
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


// vi har en bug med at 