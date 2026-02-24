package timer

import (
	"elevator/internal/config"
	"time"
)

type TimerType int

const (
	DoorTimer TimerType = iota
	WatchDogTimer
)

func Timer(
	doorOpenChan <-chan bool,
	motorActiveChan <-chan bool,
	doorClosedChan chan<- bool,
	motorInactiveChan chan<- bool,
) {
	var doorActive, watchDogActive bool

	doorTimer := time.NewTimer(config.DoorOpenTime)
	doorTimer.Stop()
	motorTimer := time.NewTimer(config.MotorTimeout)
	motorTimer.Stop()


	select { 
		case <-doorTimer.C: 
		default: 
	}
	select { 
		case <-motorTimer.C: 
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

		case  isMotorActive := <-motorActiveChan:
			watchDogActive = isMotorActive
			if watchDogActive {
				stopAndDrain(motorTimer)
				motorTimer.Reset(config.MotorTimeout)
			} else {
				stopAndDrain(motorTimer)
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