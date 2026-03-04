package localControl

import (
	"elevator/internal/config"
	"elevator/internal/lights"
	"elevator/internal/localControll/hardware"
	"elevator/internal/localControll/timer"
	"elevator/internal/types"
)

func localControl(
	newOrder <-chan [config.NFloors][config.NButtons]bool,
	elevatorEvents chan<- types.FromLocalToDM,
	localLightsChan chan<- lights.LocalLightUpdate,
) {
	var (
		floorChan          = make(chan int, config.ChannelBufferSize)
		doorOpenChan       = make(chan bool, config.ChannelBufferSize)
		motorActiveChan    = make(chan bool, config.ChannelBufferSize)
		recoveryEnableChan = make(chan bool, config.ChannelBufferSize)
		doorClosedChan     = make(chan bool, config.ChannelBufferSize)
		motorInactiveChan  = make(chan bool, config.ChannelBufferSize)
		recoveryTickChan   = make(chan bool, config.ChannelBufferSize)
		obstructionChan    = make(chan bool, config.ChannelBufferSize)
		buttonPressChan    = make(chan types.ButtonEvent, config.ChannelBufferSize)
		obstruction        bool
	)

	go hardware.PollFloorSensor(floorChan)
	go hardware.PollObstructionSwitch(obstructionChan)
	go hardware.PollButtons(buttonPressChan)

	go timer.Timer(doorOpenChan, motorActiveChan, recoveryEnableChan, doorClosedChan, motorInactiveChan, recoveryTickChan)

	elevator := elevatorInit()
	obstruction = false

	hardware.SetMotorDirection(types.Down)

	for {
		select {

		case floor := <-floorChan:
			elevator.CurrentFloor = floor
			elevator.ActiveStatus = true
			recoveryEnableChan <- false
			updateFloorIndicator(localLightsChan, elevator)

			if !hasLocalOrderAbove(elevator) && !hasLocalOrderBelow(elevator) &&
				!hasAnyOrderAtFloor(elevator, floor) {
				hardware.SetMotorDirection(types.Stop)
				elevator.MotorDirection = types.Stop
				elevator.Behaviour = types.ElevatorIdle
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, [config.NFloors][config.NButtons]bool{}, nil)
				continue
			}

			if shouldStopAtFloor(elevator, floor) {
				completedOrders := handleFloorArrival(&elevator, doorOpenChan, localLightsChan, elevator.MotorDirection)
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, completedOrders, nil)
			} else {
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, [config.NFloors][config.NButtons]bool{}, nil)
			}

		case orders := <-newOrder:
			elevator.LocalOrders = orders
			if elevator.Behaviour == types.ElevatorIdle {
				if hasAnyOrderAtFloor(elevator, elevator.CurrentFloor) {
					completedOrders := handleFloorArrival(&elevator, doorOpenChan, localLightsChan, elevator.MotorDirection)
					sendElevatorUpdate(elevatorEvents, elevator, obstruction, completedOrders, nil)
				} else {
					newDir := chooseDirection(elevator)
					if newDir != types.Stop {
						elevator.MotorDirection = newDir
						elevator.Behaviour = types.ElevatorMoving
						hardware.SetMotorDirection(newDir)
						motorActiveChan <- true
						sendElevatorUpdate(elevatorEvents, elevator, obstruction, [config.NFloors][config.NButtons]bool{}, nil)
					}
				}
			}

		case <-doorClosedChan:
			if elevator.Behaviour == types.ElevatorDoorOpen {
				handleDoorClosed(&elevator, motorActiveChan, localLightsChan)
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, [config.NFloors][config.NButtons]bool{}, nil)
			}

		case <-motorInactiveChan:
			if elevator.Behaviour == types.ElevatorMoving {
				elevator.ActiveStatus = false
				elevator.Behaviour = types.ElevatorIdle
				hardware.SetMotorDirection(types.Stop)
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, [config.NFloors][config.NButtons]bool{}, nil)
				recoveryEnableChan <- true
			}

		case obstruction = <-obstructionChan:
			if obstruction && elevator.Behaviour == types.ElevatorDoorOpen {
				doorOpenChan <- true
			}
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, [config.NFloors][config.NButtons]bool{}, nil)

		case buttonEvent := <-buttonPressChan:
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, [config.NFloors][config.NButtons]bool{}, &buttonEvent)

		case <-recoveryTickChan:
			if elevator.Behaviour == types.ElevatorIdle && !elevator.ActiveStatus {
				newDir := chooseDirection(elevator)
				if newDir != types.Stop {
					elevator.MotorDirection = newDir
					elevator.Behaviour = types.ElevatorMoving
					hardware.SetMotorDirection(newDir)
					motorActiveChan <- true
				}
			}
		}
	}
}
