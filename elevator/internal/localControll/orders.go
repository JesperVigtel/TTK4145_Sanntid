package localControl

import (
	"elevator/internal/config"
	"elevator/internal/types"
)

func hasLocalOrderAbove(e types.Elevator) bool {
	for floor := e.CurrentFloor + 1; floor < config.NFloors; floor++ {
		for btn := 0; btn < config.NButtons; btn++ {
			if e.LocalOrders[floor][btn] {
				return true
			}
		}
	}
	return false
}

func hasLocalOrderBelow(e types.Elevator) bool {
	for floor := e.CurrentFloor - 1; floor >= 0; floor-- {
		for btn := 0; btn < config.NButtons; btn++ {
			if e.LocalOrders[floor][btn] {
				return true
			}
		}
	}
	return false
}

func hasAnyOrderAtFloor(e types.Elevator, floor int) bool {
	return e.LocalOrders[floor][int(types.BTCab)] ||
		e.LocalOrders[floor][int(types.BTHallUp)] ||
		e.LocalOrders[floor][int(types.BTHallDown)]
}



func chooseDirection(e types.Elevator) types.MotorDirection {
	switch e.MotorDirection {
	case types.Up:
		if hasLocalOrderAbove(e) {
			return types.Up
		}
		if hasLocalOrderBelow(e) {
			return types.Down
		}
	case types.Down:
		if hasLocalOrderBelow(e) {
			return types.Down
		}
		if hasLocalOrderAbove(e) {
			return types.Up
		}
	default:
		if hasLocalOrderAbove(e) {
			return types.Up
		}
		if hasLocalOrderBelow(e) {
			return types.Down
		}
	}
	return types.Stop
}

func shouldStopAtFloor(e types.Elevator, floor int) bool {
	if e.LocalOrders[floor][int(types.BTCab)] {
		return true
	}

	hallUp := e.LocalOrders[floor][int(types.BTHallUp)]
	hallDown := e.LocalOrders[floor][int(types.BTHallDown)]

	switch e.MotorDirection {
	case types.Up:
		return hallUp || (!hasLocalOrderAbove(e) && hallDown)
	case types.Down:
		return hallDown || (!hasLocalOrderBelow(e) && hallUp)
	default:
		return hallUp || hallDown
	}
}


func clearCabOrder(e *types.Elevator, floor int) bool {
	if e.LocalOrders[floor][int(types.BTCab)] {
		e.LocalOrders[floor][int(types.BTCab)] = false
		return true
	}
	return false
}

func clearHallOrder(e *types.Elevator, floor int, dir types.MotorDirection) bool {
	btn := types.BTHallUp
	if dir == types.Down {
		btn = types.BTHallDown
	}
	if e.LocalOrders[floor][int(btn)] {
		e.LocalOrders[floor][int(btn)] = false
		return true
	}
	return false
}

// clearOrdersAtFloor clears appropriate orders and returns what was cleared.
// Returns true if direction change announcement is needed (extra door time).

func clearOrdersAtFloor(
	e *types.Elevator,
	floor int,
	arrivalDir types.MotorDirection,
) (completed [config.NFloors][config.NButtons]bool, needsExtraDoorTime bool) {
	if clearCabOrder(e, floor) {
		completed[floor][int(types.BTCab)] = true
	}

	switch arrivalDir {
	case types.Up:
		if clearHallOrder(e, floor, types.Up) {
			completed[floor][int(types.BTHallUp)] = true
		}
		// Direction change: clear down if no more orders above
		if !hasLocalOrderAbove(*e) && clearHallOrder(e, floor, types.Down) {
			completed[floor][int(types.BTHallDown)] = true
			needsExtraDoorTime = true
		}
	case types.Down:
		if clearHallOrder(e, floor, types.Down) {
			completed[floor][int(types.BTHallDown)] = true
		}
		// Direction change: clear up if no more orders below
		if !hasLocalOrderBelow(*e) && clearHallOrder(e, floor, types.Up) {
			completed[floor][int(types.BTHallUp)] = true
			needsExtraDoorTime = true
		}
	default:
		// Idle: clear both
		if clearHallOrder(e, floor, types.Up) {
			completed[floor][int(types.BTHallUp)] = true
		}
		if clearHallOrder(e, floor, types.Down) {
			completed[floor][int(types.BTHallDown)] = true
		}
	}
	return
}
