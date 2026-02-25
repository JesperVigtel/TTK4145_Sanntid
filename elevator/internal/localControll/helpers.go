package localControl

import (
	"elevator/internal/config"
	"elevator/internal/types"
)

func elevatorInit() types.Elevator {
	return types.Elevator{
		CurrentFloor:   -1,
		MotorDirection: types.Down,
		Request:        [config.NFloors][config.NButtons]bool{},
		Behaviour:      types.ElevatorMoving,
		ActiveStatus:   true,
	}
}

func requestsAbove(e types.Elevator) bool {
	for floor := e.CurrentFloor + 1; floor < config.NFloors; floor++ {
		for btn := 0; btn < config.NButtons; btn++ {
			if e.Request[floor][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e types.Elevator) bool {
	for floor := e.CurrentFloor - 1; floor >= 0; floor-- {
		for btn := 0; btn < config.NButtons; btn++ {
			if e.Request[floor][btn] {
				return true
			}
		}
	}
	return false
}

func hasRequestAtFloor(e types.Elevator, floor int) bool {
	if floor < 0 || floor >= config.NFloors {
		return false
	}
	return e.Request[floor][int(types.BTHallUp)] ||
		e.Request[floor][int(types.BTHallDown)] ||
		e.Request[floor][int(types.BTCab)]
}

func switchDirection(e types.Elevator) types.MotorDirection {
	switch e.MotorDirection {

	case types.Up:
		if requestsAbove(e) {
			return types.Up
		}
		if requestsBelow(e) {
			return types.Down
		}
		return types.Stop

	case types.Down:
		if requestsBelow(e) {
			return types.Down
		}
		if requestsAbove(e) {
			return types.Up
		}
		return types.Stop

	default:
		switch {
		case requestsAbove(e):
			return types.Up
		case requestsBelow(e):
			return types.Down
		default:
			return types.Stop
		}
	}
}



func localClearHallOrder(e *types.Elevator, floor int, dir types.MotorDirection) bool {
	if floor < 0 || floor >= config.NFloors {
		return false
	}

	switch dir {
	case types.Up:
		if e.Request[floor][int(types.BTHallUp)] {
			e.Request[floor][int(types.BTHallUp)] = false
			return true
		}
	case types.Down:
		if e.Request[floor][int(types.BTHallDown)] {
			e.Request[floor][int(types.BTHallDown)] = false
			return true
		}
	}
	return false
}

// usikker på om denne fungerer, men tror det, må se litt på DecisionMaker koden. altså den over

func localClearCabOrder(
	e *types.Elevator,
	floor int,
) bool {

	if floor < 0 || floor >= config.NFloors {
		return false
	}

	if e.Request[floor][int(types.BTCab)] {
		e.Request[floor][int(types.BTCab)] = false
		return true
	}

	return false
}

//Fjerner local caborders lokalt, hvordan få dette inn i DecisionMaker

func shouldStopAtFloor(e types.Elevator, floor int) bool {
	if floor < 0 || floor >= config.NFloors {
		return false
	}

	// Cab always stops
	if e.Request[floor][int(types.BTCab)] {
		return true
	}

	// stopp for Hall only if direction matches announced direction
	switch e.MotorDirection {
	case types.Up:
		return e.Request[floor][int(types.BTHallUp)]
	case types.Down:
		return e.Request[floor][int(types.BTHallDown)]
	default:
		return false
	}
}
