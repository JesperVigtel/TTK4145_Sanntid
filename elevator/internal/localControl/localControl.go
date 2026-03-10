package localControl

import (
	"elevator/internal/config"
	"elevator/internal/localControl/hardware"
	"elevator/internal/localControl/timer"
	"elevator/internal/types"
	"fmt"
)

func Run(
	newOrder <-chan types.LocalOrderTable,
	elevatorEvents chan<- types.ElevatorEvents,
	localLightsChan chan<- types.LocalLightUpdate,
) {
	var (
		floorChan                = make(chan int, config.ChannelBufferSize)
		doorOpenChan             = make(chan bool, config.ChannelBufferSize)
		motorActiveChan          = make(chan bool, config.ChannelBufferSize)
		recoveryEnableChan       = make(chan bool, config.ChannelBufferSize)
		doorClosedChan           = make(chan bool, config.ChannelBufferSize)
		motorInactiveChan        = make(chan bool, config.ChannelBufferSize)
		tryRecovery		         = make(chan bool, config.ChannelBufferSize)
		obstructionChan          = make(chan bool, config.ChannelBufferSize)
		buttonPressChan          = make(chan types.ButtonEvent, config.ChannelBufferSize)
		obstruction              bool
		directionChange			 bool
	)
	hardware.Init(config.Addr, config.NFloors)

	go hardware.PollFloorSensor(floorChan)
	go hardware.PollObstructionSwitch(obstructionChan)
	go hardware.PollButtons(buttonPressChan)

	go timer.Timer(doorOpenChan, motorActiveChan, recoveryEnableChan, doorClosedChan, motorInactiveChan, recoveryTickChan)

	elevator := elevatorInit()
	obstruction = false

	hardware.SetMotorDirection(types.Down)
	sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)
	
	for {
		select {

		case floor := <-floorChan:
			fmt.Printf("[LocalControl] Floor sensor triggered: floor=%d\n", floor)
			elevator.CurrentFloor = floor
			elevator.ActiveStatus = true
			recoveryEnableChan <- false
			updateFloorIndicator(localLightsChan, elevator)

			// -- Restart Watchdog --
			if elevator.Behaviour == types.ElevatorMoving {
				motorActiveChan <- true
			}

			if !hasLocalOrderAbove(elevator) && !hasLocalOrderBelow(elevator) &&
				!hasAnyOrderAtFloor(elevator, floor) {
				hardware.SetMotorDirection(types.Stop)
				elevator.PhysicalMotorDirection = types.Stop
				elevator.Behaviour = types.ElevatorIdle
				motorActiveChan <- false
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)
				println("[LocalControl] no orders anywhere -> idle")
				continue
			}

			if shouldStopAtFloor(elevator, floor) {
				completedOrders, needsExtraDoorTime := handleFloorArrival(&elevator, doorOpenChan, motorActiveChan, localLightsChan, elevator.CurrentTravelDirection)
				if needsExtraDoorTime {
					directionChange = true
				}
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, completedOrders, nil)
			} else {
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)
			}

		case orders := <-newOrder:
			fmt.Println("[LocalControl] Received new order table")
			elevator.LocalOrders = orders
			sendLightUpdate(localLightsChan, elevator, elevator.Behaviour == types.ElevatorDoorOpen)

			if elevator.Behaviour == types.ElevatorDoorOpen && hasAnyOrderAtFloor(elevator, elevator.CurrentFloor) {
				completedOrders, _ := clearOrdersAtFloor(&elevator, elevator.CurrentFloor, elevator.CurrentTravelDirection, false)
				doorOpenChan <- true
				sendLightUpdate(localLightsChan, elevator, true)
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, completedOrders, nil)
				continue
			}

			if elevator.Behaviour == types.ElevatorIdle {
				if hasAnyOrderAtFloor(elevator, elevator.CurrentFloor) {

					completedOrders, needsExtraDoorTime := handleFloorArrival(&elevator, doorOpenChan, motorActiveChan, localLightsChan, elevator.CurrentTravelDirection)
					if needsExtraDoorTime {
						directionChange = true
					}
					sendElevatorUpdate(elevatorEvents, elevator, obstruction, completedOrders, nil)
				} else {
					newDir := chooseDirection(elevator)
					if newDir != types.Stop {
						elevator.CurrentTravelDirection = newDir
						elevator.PhysicalMotorDirection = newDir
						elevator.Behaviour = types.ElevatorMoving
						hardware.SetMotorDirection(newDir)
						motorActiveChan <- true
						sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)
					}
				}
			}

		case obstruction = <-obstructionChan:
			fmt.Printf("[LocalControl] Obstruction changed: %v\n", obstruction)
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)

		case buttonEvent := <-buttonPressChan:
			fmt.Printf("[LocalControl] Button pressed: floor=%d button=%d\n",
				buttonEvent.Floor,
				buttonEvent.Button)

			sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, &buttonEvent)

		case <-doorClosedChan:
			if elevator.Behaviour == types.ElevatorDoorOpen {
				if obstruction {
					doorOpenChan <- true
				} else if directionChange {
					fmt.Println("[LocalControl] Announcing direction change - clearing opposite hall order")
					directionChange = false
					completedOrders, _ := clearOrdersAtFloor(&elevator, elevator.CurrentFloor, elevator.CurrentTravelDirection, true)
					sendLightUpdate(localLightsChan, elevator, true)
					sendElevatorUpdate(elevatorEvents, elevator, obstruction, completedOrders, nil)
					doorOpenChan <- true
				} else {
					handleDoorClosed(&elevator, motorActiveChan, localLightsChan)
					sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)
				}
			}

		case <-motorInactiveChan:
			fmt.Println("[LocalControl] Motor inactive; watchdog triggered")
			killElevator(&elevator)
			recoveryEnableChan <- true
			
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)
			

// Under er good
		case <-tryRecovery:
			fmt.Println("[LocalControl] Recovery timer triggered, tries to move again")

			tryMoving(&elevator)
			recoveryEnableChan <- false
			motorActiveChan <-true
		}
	}
}

