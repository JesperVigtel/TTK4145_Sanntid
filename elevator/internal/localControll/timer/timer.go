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
	doorClosedChan chan<- bool,
	motorActiveChan <-chan bool,
	motorInactiveChan chan<- bool,
) {
	var doorActive, watchDogActive bool
	
	doorTimer := time.NewTimer(config.DoorOpenTime)
	doorTimer.Stop()
	motorTimer := time.NewTimer(config.MotorTimeout)
	motorTimer.Stop()

	// We drain in case doortimer and motortimer ticked before 
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
		case doorActive := <-doorOpenChan:
			if doorActive {
				stopAndDrain(doorTimer)
				doorTimer.Reset(config.DoorOpenTime)
			} else {
				stopAndDrain(doorTimer)
			}

		case  watchDogActive := <-motorActiveChan:
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

func stopAndDrain(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}

//forsøk 1 (ble for komplisert timer)

// type DoorTimer struct {
// 	timeout time.Duration
// 	timer *time.Timer
// 	active bool
// 	timerOut chan bool
// }

// // initialiserer ny versjon av structet, derfor returnerer den pointer
// // Vi ønsker ikke flere versjoner av timeren, da er det bedre med en pointer til timeren som endres, hindrer kopiering
// // no timer initilized here to combat
// func NewDoorTimer(duration time.Duration) *DoorTimer {
// 	timer := time.NewTimer(duration) // starter en timer
// 	timer.Stop() //stopper den med en gang
// 	return &DoorTimer{ 
// 		timeout: duration,
// 		timer: timer,
// 		active: false,
// 		timerOut: make(chan bool, 1),
// 	}
// }
//  // starter selvfølgelig doortimer, løk
// func (t *DoorTimer) StartDoorTimer() {
// 	if !t.active {
// 		t.timer.Reset(t.timeout)
// 		t.active = true
// 	}
// }