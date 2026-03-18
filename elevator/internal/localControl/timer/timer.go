package timer

import (
	"elevator/internal/config"
	"time"
)

func Timer(
	doorOpenChan <-chan bool,
	motorActiveChan <-chan bool,
	recoveryEnableChan <-chan bool,
	doorClosedChan chan<- bool,
	motorInactiveChan chan<- bool,
	recoveryTickChan chan<- bool,
) {
	var isDoorOpen, isMotorActive, isRecoveryActive bool

	doorTimer, motorTimer, motorRecoveryTimer := initTimers()

	for {
		select {
		case isDoorOpen = <-doorOpenChan:
			if isDoorOpen {
				drainIfFired(doorTimer)
				doorTimer.Reset(config.DoorOpenTime)
			} else {
				drainIfFired(doorTimer)
			}

		case isMotorActive = <-motorActiveChan:
			if isMotorActive {
				drainIfFired(motorTimer)
				motorTimer.Reset(config.MotorTimeout)
			} else {
				drainIfFired(motorTimer)
			}

		case isRecoveryActive = <-recoveryEnableChan:
			if isRecoveryActive {
				drainIfFired(motorRecoveryTimer)
				motorRecoveryTimer.Reset(config.MotorRecoveryTime)
			} else {
				drainIfFired(motorRecoveryTimer)
			}

		case <-doorTimer.C:
			if isDoorOpen {
				isDoorOpen = false
				doorClosedChan <- true
			}

		case <-motorTimer.C:
			if isMotorActive {
				isMotorActive = false
				motorInactiveChan <- true
			}

		case <-motorRecoveryTimer.C:
			if isRecoveryActive {
				recoveryTickChan <- true
				drainIfFired(motorRecoveryTimer)
				motorRecoveryTimer.Reset(config.MotorRecoveryTime)
			}
		}
	}
}

func initTimers() (*time.Timer, *time.Timer, *time.Timer) {
	doorTimer := time.NewTimer(config.DoorOpenTime)
	motorTimer := time.NewTimer(config.MotorTimeout)
	motorRecoveryTimer := time.NewTimer(config.MotorRecoveryTime)

	drainIfFired(doorTimer)
	drainIfFired(motorTimer)
	drainIfFired(motorRecoveryTimer)

	return doorTimer, motorTimer, motorRecoveryTimer
}

func drainIfFired(timerInstance *time.Timer) {
	if !timerInstance.Stop() {
		select {
		case <-timerInstance.C:
		default:
		}
	}
}
