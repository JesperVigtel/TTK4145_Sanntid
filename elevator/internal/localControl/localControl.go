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
	hardware.Init(config.Addr, config.NFloors)

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
			fmt.Printf("[EVENT] Floor sensor triggered: floor=%d\n", floor)
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
			fmt.Println("[EVENT] Received new order table")
			elevator.LocalOrders = orders
			sendLightUpdate(localLightsChan, elevator, elevator.Behaviour == types.ElevatorDoorOpen)
			
			if elevator.Behaviour == types.ElevatorDoorOpen {
				if hasAnyOrderAtFloor(elevator, elevator.CurrentFloor) {
					doorOpenChan <- true  
				}
			} else if elevator.Behaviour == types.ElevatorIdle {
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

		case obstruction = <-obstructionChan:
			fmt.Printf("[EVENT] Obstruction changed: %v\n", obstruction)
			if obstruction && elevator.Behaviour == types.ElevatorDoorOpen {
				fmt.Println("[ACTION] Extending door open due to obstruction")
				doorOpenChan <- true
			}
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, [config.NFloors][config.NButtons]bool{}, nil)

		case buttonEvent := <-buttonPressChan:
			fmt.Printf("[EVENT] Button pressed: floor=%d button=%d\n",
				buttonEvent.Floor,
				buttonEvent.Button)

			sendElevatorUpdate(elevatorEvents, elevator, obstruction, [config.NFloors][config.NButtons]bool{}, &buttonEvent)

		case <-doorClosedChan:
			fmt.Println("[EVENT] Door closed timer triggered")

			if elevator.Behaviour == types.ElevatorDoorOpen {
				handleDoorClosed(&elevator, motorActiveChan, localLightsChan)
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, [config.NFloors][config.NButtons]bool{}, nil)
			}

		case <-motorInactiveChan:
			fmt.Println("[EVENT] Motor inactive; watchdog triggered")

			if elevator.Behaviour == types.ElevatorMoving {
				elevator.ActiveStatus = false
				elevator.Behaviour = types.ElevatorIdle
				elevator.MotorDirection = types.Stop
				hardware.SetMotorDirection(types.Stop)
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, [config.NFloors][config.NButtons]bool{}, nil)
				recoveryEnableChan <- true
			}

		case <-recoveryTickChan:
			fmt.Println("[EVENT] Recovery tick triggered, tries to move again")

			if elevator.Behaviour == types.ElevatorIdle && !elevator.ActiveStatus {
				newDir := chooseDirection(elevator)
				if newDir != types.Stop {
					elevator.MotorDirection = newDir
					elevator.Behaviour = types.ElevatorMoving
					hardware.SetMotorDirection(newDir)
					recoveryEnableChan <- false
					motorActiveChan <- true

				} else {
					recoveryEnableChan <- false
				}
			}
		}
	}
}

// endre på navn config [Nfloors][Nbtn] i send elevator update
// i init
