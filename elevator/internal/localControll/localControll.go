package localControl

// Tilstandsmaskinen (FSM). Håndterer heisens fysiske bevegelse og dørlogikk.
// Inneholder logikk for OnFloorArrival, OnOrderRequest og OnTimerTimeout.
// Snakker med hardware for å styre motor og dør.

import (
	"elevator/internal/config"
	"elevator/internal/localControll/hardware"
	"elevator/internal/localControll/timer"
	"elevator/internal/types"
	"fmt"
)

func localControl(
	newOrder <-chan [config.NFloors][config.NButtons]bool,
	elevatorEvents chan<- types.FromLocalToDM,
) {
	var (
		floorChan          = make(chan int, config.ChannelBufferSize)
		doorOpenChan       = make(chan bool, 1)
		motorActiveChan    = make(chan bool, 1)
		recoveryEnableChan = make(chan bool, 1)
		doorClosedChan    = make(chan bool, 1) 
		motorInactiveChan = make(chan bool, 1) 
		recoveryTickChan  = make(chan bool, 1) 
		obstructionChan   = make(chan bool, config.ChannelBufferSize)
		buttonPressChan   = make(chan types.OrderEvent, config.ChannelBufferSize)
		obstruction  bool
	)

	go hardware.PollFloorSensor(floorChan)
	go hardware.PollObstructionSwitch(obstructionChan)
	go hardware.PollButtons(buttonPressChan)
	go hardware.PollStopButton(StopPressChan)

	go timer.Timer(doorOpenChan, motorActiveChan, recoveryEnableChan, doorClosedChan, motorInactiveChan, recoveryTickChan)

	elevator := elevatorInit()
	obstruction = false

	hardware.SetMotorDirection(types.Down)

	for {
		select {

		case floor := <-floorChan:
			elevator.CurrentFloor = floor
			elevator.ActiveStatus = true
			hardware.SetFloorIndicator(floor)

			if shouldStopAtFloor(elevator, floor) {
				completedOrders := handleFloorArrival(&elevator, doorOpenChan)

				elevatorEvents <- types.FromLocalToDM{
					Elevator:       elevator,
					CompletedOrder: completedOrders,
					NewButtonPress: nil,
					Obstructed:     obstruction,
				}
			} else {
				sendElevatorUpdate(elevatorEvents, elevator, obstruction)
			}

		case orders := <-newOrder:
			elevator.Request = orders

			if elevator.Behaviour == types.ElevatorIdle {
				newDir := switchDirection(elevator)
				if newDir != types.Stop {
					elevator.MotorDirection = newDir
					elevator.Behaviour = types.ElevatorMoving
					hardware.SetMotorDirection(newDir)
					motorActiveChan <- true
					sendElevatorUpdate(elevatorEvents, elevator, obstruction)
				}
			}

		case <-doorClosedChan:
			if elevator.Behaviour == types.ElevatorDoorOpen {
				handleDoorClosed(&elevator, motorActiveChan)
				sendElevatorUpdate(elevatorEvents, elevator, obstruction)
			}

		case <-motorInactiveChan:
			if elevator.Behaviour == types.ElevatorMoving {
				fmt.Println("Motor timeout - elevator may be stuck!")
				elevator.ActiveStatus = false
				elevator.Behaviour = types.ElevatorIdle // Allow recovery when new orders come
				hardware.SetMotorDirection(types.Stop)
				sendElevatorUpdate(elevatorEvents, elevator, obstruction)

				// Enable recovery timer to try again
				recoveryEnableChan <- true
			}

		case obstruction = <-obstructionChan:
			if obstruction && elevator.Behaviour == types.ElevatorDoorOpen {
				doorOpenChan <- true
			}
			sendElevatorUpdate(elevatorEvents, elevator, obstruction)

		case buttonEvent := <-buttonPressChan:
			elevatorEvents <- types.FromLocalToDM{
				Elevator:       elevator,
				CompletedOrder: [config.NFloors][config.NButtons]bool{},
				NewButtonPress: &buttonEvent,
				Obstructed:     obstruction,
			}

		case <-StopPressChan:
			handleStopButton(&elevator)
			sendElevatorUpdate(elevatorEvents, elevator, obstruction)

		case <-recoveryTickChan:
			if elevator.Behaviour == types.ElevatorIdle && !elevator.ActiveStatus {
				newDir := switchDirection(elevator)
				if newDir != types.Stop {
					fmt.Println("Attempting recovery - trying to move again")
					elevator.MotorDirection = newDir
					elevator.Behaviour = types.ElevatorMoving
					hardware.SetMotorDirection(newDir)
					motorActiveChan <- true // Re-enable watchdog
				}
				// her må jeg sende noe på recoverychan for a stoppe denne timeren
			}
		}
	}
}

//kan hende det enkleste er å fjerne recoveryTimeren. Er usikker på om den er nødvendig.
