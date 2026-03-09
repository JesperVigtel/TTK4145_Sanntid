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
		recoveryTickChan         = make(chan bool, config.ChannelBufferSize)
		obstructionChan          = make(chan bool, config.ChannelBufferSize)
		buttonPressChan          = make(chan types.ButtonEvent, config.ChannelBufferSize)
		obstruction              bool
		directionChangeAnnounced bool
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
			fmt.Printf("[LocalControl] Floor sensor triggered: floor=%d\n", floor)
			elevator.CurrentFloor = floor
			elevator.ActiveStatus = true
			recoveryEnableChan <- false
			updateFloorIndicator(localLightsChan, elevator)
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
				completedOrders, dirChanged := handleFloorArrival(&elevator, doorOpenChan, motorActiveChan, localLightsChan, elevator.CurrentTravelDirection)
				if dirChanged {
					directionChangeAnnounced = true
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
				completedOrders, _ := clearOrdersAtFloor(&elevator, elevator.CurrentFloor, elevator.CurrentTravelDirection)
				doorOpenChan <- true
				sendLightUpdate(localLightsChan, elevator, true)
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, completedOrders, nil)
				continue
			}

			if elevator.Behaviour == types.ElevatorIdle {
				if hasAnyOrderAtFloor(elevator, elevator.CurrentFloor) {

					completedOrders, dirChanged := handleFloorArrival(&elevator, doorOpenChan, motorActiveChan, localLightsChan, elevator.CurrentTravelDirection)
					if dirChanged {
						directionChangeAnnounced = true
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
			if obstruction && elevator.Behaviour == types.ElevatorDoorOpen {
				fmt.Println("[LocalControl] Extending door open due to obstruction")
				doorOpenChan <- true
			}
			sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)

		case buttonEvent := <-buttonPressChan:
			fmt.Printf("[LocalControl] Button pressed: floor=%d button=%d\n",
				buttonEvent.Floor,
				buttonEvent.Button)

			sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, &buttonEvent)

		case <-doorClosedChan:
			fmt.Println("[LocalControl] Door closed timer triggered")
			if elevator.Behaviour == types.ElevatorDoorOpen {
				if obstruction {
					doorOpenChan <- true
				} else if directionChangeAnnounced {
					fmt.Println("[LocalControl] Announcing direction change - extending door time")
					directionChangeAnnounced = false
					doorOpenChan <- true
				} else {
					handleDoorClosed(&elevator, motorActiveChan, localLightsChan)
					sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)
				}
			}

		case <-motorInactiveChan:
			fmt.Println("[LocalControl] Motor inactive; watchdog triggered")

			if elevator.Behaviour == types.ElevatorMoving {
				elevator.ActiveStatus = false
				elevator.Behaviour = types.ElevatorIdle
				elevator.PhysicalMotorDirection = types.Stop
				hardware.SetMotorDirection(types.Stop)
				sendElevatorUpdate(elevatorEvents, elevator, obstruction, types.CompletedOrderTable{}, nil)
				recoveryEnableChan <- true
			}

		case <-recoveryTickChan:
			fmt.Println("[LocalControl] Recovery tick triggered, tries to move again")

			if elevator.Behaviour == types.ElevatorIdle && !elevator.ActiveStatus {
				newDir := chooseDirection(elevator)
				if newDir != types.Stop {
					elevator.CurrentTravelDirection = newDir
					elevator.PhysicalMotorDirection = newDir
					elevator.Behaviour = types.ElevatorMoving
					hardware.SetMotorDirection(newDir)
					recoveryEnableChan <- false
					motorActiveChan <- true
				}
			}
		}
	}
}



// feil med recoverytick når den lokale heisen er stuck og en annen heis fullfører orderen den skulle gjøre, da skjer det ingen ting
// dersom man da trykker på en hallorder på heisen som er stoppet, skjer det ingen ting før man trykker på en caborder. da betjenes også hallorderen i henhold til HRA

// directionchange blir annonsert ved øverste etasje og nederste etasje og døren holdes åpen i forlenga tid. det skal ikke skje, den skal bare annonsere retningsbytte dersom begge knappene er trykket i en etasje
// og det ikke er noen ordre i annkommstretningen, men det er det i motsatt retning