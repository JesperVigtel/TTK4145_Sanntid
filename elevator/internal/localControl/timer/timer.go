package timer

import (
	"elevator/internal/config"
	"time"
)

type TimerType int

const (
	DoorTimer TimerType = iota
	WatchDogTimer
	MotorRecoveryTimer
)

func Timer(
	doorOpenChan <-chan bool,
	motorActiveChan <-chan bool,
	recoveryEnableChan <-chan bool,

	doorClosedChan chan<- bool,
	motorInactiveChan chan<- bool,
	recoveryTickChan chan<- bool,
) {
	var doorActive, watchDogActive, recoveryActive bool

	doorTimer := time.NewTimer(config.DoorOpenTime)
	doorTimer.Stop()
	motorTimer := time.NewTimer(config.MotorTimeout)
	motorTimer.Stop()
	motorRecoveryTimer := time.NewTimer(config.MotorRecoveryTime)
	motorRecoveryTimer.Stop()

	select {
	case <-doorTimer.C:
	default:
	}
	select {
	case <-motorTimer.C:
	default:
	}
	select {
	case <-motorRecoveryTimer.C:
	default:
	}

	for {
		select {
		case isDoorOpen := <-doorOpenChan:
			doorActive = isDoorOpen
			if doorActive {
				stopAndDrain(doorTimer)
				doorTimer.Reset(config.DoorOpenTime)
			} else {
				stopAndDrain(doorTimer)
			}

		case isMotorActive := <-motorActiveChan:
			watchDogActive = isMotorActive
			if watchDogActive {
				stopAndDrain(motorTimer)
				motorTimer.Reset(config.MotorTimeout)
			} else {
				stopAndDrain(motorTimer)
			}
		case isRecoveryActive := <-recoveryEnableChan:
			recoveryActive = isRecoveryActive
			if recoveryActive {
				stopAndDrain(motorRecoveryTimer)
				motorRecoveryTimer.Reset(config.MotorRecoveryTime)
			} else {
				stopAndDrain(motorRecoveryTimer)
			}

		case <-doorTimer.C:
			if doorActive {
				doorActive = false
				doorClosedChan <- true
			}

		case <-motorTimer.C:
			if watchDogActive {
				watchDogActive = false
				motorInactiveChan <- true
			}
		case <-motorRecoveryTimer.C:
			if recoveryActive {
				recoveryTickChan <- true
				stopAndDrain(motorRecoveryTimer)
				motorRecoveryTimer.Reset(config.MotorRecoveryTime)
			}
		}
	}
}

func stopAndDrain(timerInstance *time.Timer) {
	if !timerInstance.Stop() {
		select {
		case <-timerInstance.C:
		default:
		}
	}
}
