package localControl

import (
	"elevator/internal/config"
	"elevator/internal/types"
)

func hasLocalOrderAbove(elevator types.Elevator) bool {
	for floor := elevator.CurrentFloor + 1; floor < config.NFloors; floor++ {
		for btn := 0; btn < config.NButtons; btn++ {
			if elevator.LocalOrders[floor][btn] {
				return true
			}
		}
	}
	return false
}

func hasLocalOrderBelow(elevator types.Elevator) bool {
	for floor := elevator.CurrentFloor - 1; floor >= 0; floor-- {
		for btn := 0; btn < config.NButtons; btn++ {
			if elevator.LocalOrders[floor][btn] {
				return true
			}
		}
	}
	return false
}

func hasAnyOrderAtFloor(elevator types.Elevator, floor int) bool {
	return elevator.LocalOrders[floor][int(types.BtnCab)] ||
		elevator.LocalOrders[floor][int(types.BtnHallUp)] ||
		elevator.LocalOrders[floor][int(types.BtnHallDown)]
}



func chooseDirection(elevator types.Elevator) types.MotorDirection {
	switch elevator.MotorDirection {
	case types.Up:
		if hasLocalOrderAbove(elevator) {
			return types.Up
		}
		if hasLocalOrderBelow(elevator) {
			return types.Down
		}
	case types.Down:
		if hasLocalOrderBelow(elevator) {
			return types.Down
		}
		if hasLocalOrderAbove(elevator) {
			return types.Up
		}
	default:
		if hasLocalOrderAbove(elevator) {
			return types.Up
		}
		if hasLocalOrderBelow(elevator) {
			return types.Down
		}
	}
	return types.Stop
}

func shouldStopAtFloor(elevator types.Elevator, floor int) bool {
	if elevator.LocalOrders[floor][int(types.BtnCab)] {
		return true
	}

	hallUp := elevator.LocalOrders[floor][int(types.BtnHallUp)]
	hallDown := elevator.LocalOrders[floor][int(types.BtnHallDown)]

	switch elevator.MotorDirection {
	case types.Up:
		return hallUp || (!hasLocalOrderAbove(elevator) && hallDown)
	case types.Down:
		return hallDown || (!hasLocalOrderBelow(elevator) && hallUp)
	default:
		return hallUp || hallDown
	}
}


func clearCabOrder(elevator *types.Elevator, floor int) bool {
	if elevator.LocalOrders[floor][int(types.BtnCab)] {
		elevator.LocalOrders[floor][int(types.BtnCab)] = false
		return true
	}
	return false
}

func clearHallOrder(elevator *types.Elevator, floor int, dir types.MotorDirection) bool {
	btn := types.BtnHallUp
	if dir == types.Down {
		btn = types.BtnHallDown
	}
	if elevator.LocalOrders[floor][int(btn)] {
		elevator.LocalOrders[floor][int(btn)] = false
		return true
	}
	return false
}

func clearOrdersAtFloor(
	elevator *types.Elevator,
	floor int,
	arrivalDir types.MotorDirection,
) (completed [config.NFloors][config.NButtons]bool, needsExtraDoorTime bool) {
	if clearCabOrder(elevator, floor) {
		completed[floor][int(types.BtnCab)] = true
	}

	switch arrivalDir {
	case types.Up:
		if clearHallOrder(elevator, floor, types.Up) {
			completed[floor][int(types.BtnHallUp)] = true
		}
		if !hasLocalOrderAbove(*elevator) && clearHallOrder(elevator, floor, types.Down) {
			completed[floor][int(types.BtnHallDown)] = true
			needsExtraDoorTime = true
		}
	case types.Down:
		if clearHallOrder(elevator, floor, types.Down) {
			completed[floor][int(types.BtnHallDown)] = true
		}
		if !hasLocalOrderBelow(*elevator) && clearHallOrder(elevator, floor, types.Up) {
			completed[floor][int(types.BtnHallUp)] = true
			needsExtraDoorTime = true
		}
	default:
		if clearHallOrder(elevator, floor, types.Up) {
			completed[floor][int(types.BtnHallUp)] = true
		}
		if clearHallOrder(elevator, floor, types.Down) {
			completed[floor][int(types.BtnHallDown)] = true
		}
	}
	return
}
