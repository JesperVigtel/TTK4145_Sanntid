package hardware

// Driver interface for the simulator / elevator server TCP protocol.

import (
	"elevator/internal/config"
	"elevator/internal/types"
	"fmt"
	"net"
	"sync"
	"time"
)

var _initialized bool = false
var _numFloors int = config.NFloors
var _mtx sync.Mutex
var _conn net.Conn

func Init(addr string, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_numFloors = numFloors
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

func SetMotorDirection(dir types.MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button types.ButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

func PollButtons(receiver chan<- types.ButtonEvent) {
	prev := make([][config.NButtons]bool, _numFloors)
	for {
		time.Sleep(config.PollRateMS)
		for f := 0; f < _numFloors; f++ {
			for button := types.ButtonType(0); button < types.ButtonType(config.NButtons); button++ {
				pressed := GetButton(button, f)
				if pressed && pressed != prev[f][button] {
					receiver <- types.ButtonEvent{Floor: f, Button: button}
				}
				prev[f][button] = pressed
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(config.PollRateMS)
		v := GetFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(config.PollRateMS)
		v := GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func GetButton(button types.ButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

func read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
