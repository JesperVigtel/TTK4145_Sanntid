package localControl

import (
	"elevator/internal/config"
	"elevator/internal/types"
)

func hasLocalOrderAbove(elevator types.Elevator) bool {
	for floor := elevator.CurrentFloor + 1; floor < config.NFloors; floor++ {
		for btn := range config.NButtons {
			if elevator.LocalOrders[floor][btn] {
				return true
			}
		}
	}
	return false
}

func hasLocalOrderBelow(elevator types.Elevator) bool {
	for floor := elevator.CurrentFloor - 1; floor >= 0; floor-- {
		for btn := range config.NButtons {
			if elevator.LocalOrders[floor][btn] {
				return true
			}
		}
	}
	return false
}

func hasOrderAtFloor(elevator types.Elevator, floor int) bool {
	return elevator.LocalOrders[floor][int(types.BtnCab)] ||
		elevator.LocalOrders[floor][int(types.BtnHallUp)] ||
		elevator.LocalOrders[floor][int(types.BtnHallDown)]
}

func chooseDirection(elevator types.Elevator) types.MotorDirection {
	switch elevator.CurrentTravelDirection {
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

func shouldStopAtCurrentFloor(elevator types.Elevator, floor int) bool {
	if elevator.LocalOrders[floor][int(types.BtnCab)] {
		return true
	}

	hallUp := elevator.LocalOrders[floor][int(types.BtnHallUp)]
	hallDown := elevator.LocalOrders[floor][int(types.BtnHallDown)]

	switch elevator.CurrentTravelDirection {
	case types.Up:
		return hallUp || (!hasLocalOrderAbove(elevator) && hallDown)
	case types.Down:
		return hallDown || (!hasLocalOrderBelow(elevator) && hallUp)
	default:
		return hallUp || hallDown
	}
}

func buttonForDirection(dir types.MotorDirection) types.ButtonType {
	if dir == types.Up {
		return types.BtnHallUp
	}
	return types.BtnHallDown
}

func hasOrdersInDirection(elevator types.Elevator, dir types.MotorDirection) bool {
	if dir == types.Up {
		return hasLocalOrderAbove(elevator)
	}
	return hasLocalOrderBelow(elevator)
}

func isEndFloor(floor int, dir types.MotorDirection) bool {
	if dir == types.Up {
		return floor == config.NFloors-1
	}
	return floor == 0
}

func clearCabOrder(elevator *types.Elevator, floor int) bool {
	if elevator.LocalOrders[floor][int(types.BtnCab)] {
		elevator.LocalOrders[floor][int(types.BtnCab)] = false
		return true
	}
	return false
}

func clearHallOrder(elevator *types.Elevator, floor int, dir types.MotorDirection) bool {
	btn := buttonForDirection(dir)
	if elevator.LocalOrders[floor][int(btn)] {
		elevator.LocalOrders[floor][int(btn)] = false
		return true
	}
	return false
}

func clearOrdersAtFloor(
	elevator *types.Elevator,
	floor int,
	travelDir types.MotorDirection,
) (completed types.CompletedOrderTable, needsExtraDoorTime bool) {

	if clearCabOrder(elevator, floor) {
		completed[floor][int(types.BtnCab)] = true
	}

	if travelDir == types.Stop {
		if clearHallOrder(elevator, floor, types.Up) {
			completed[floor][int(types.BtnHallUp)] = true
		}
		if clearHallOrder(elevator, floor, types.Down) {
			completed[floor][int(types.BtnHallDown)] = true
		}
		return
	}

	if clearHallOrder(elevator, floor, travelDir) {
		completed[floor][int(buttonForDirection(travelDir))] = true
	}

	oppositeDir := -travelDir

	if hasOrdersInDirection(*elevator, travelDir) {
		return
	}

	if elevator.LocalOrders[floor][int(buttonForDirection(oppositeDir))] {
		if isEndFloor(floor, travelDir) {
			if clearHallOrder(elevator, floor, oppositeDir) {
				completed[floor][int(buttonForDirection(oppositeDir))] = true
			}
		} else {
			needsExtraDoorTime = true
		}
	} else if hasOrdersInDirection(*elevator, oppositeDir) {
		needsExtraDoorTime = true
	}

	return
}

func clearOppositeHallOrder(
	elevator *types.Elevator,
	floor int,
	travelDir types.MotorDirection,
) types.CompletedOrderTable {
	var completed types.CompletedOrderTable
	oppositeDir := -travelDir
	if clearHallOrder(elevator, floor, oppositeDir) {
		completed[floor][int(buttonForDirection(oppositeDir))] = true
	}
	return completed
}
