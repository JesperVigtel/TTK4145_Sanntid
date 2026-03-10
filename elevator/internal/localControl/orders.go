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

// Er det ikke mulig å legge til håndtering av edgecase -> recoverytick


func anyOrdersAtCurrentFloor(elevator types.Elevator, floor int) bool {
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

// tror det er en bedre måte å gjøre det her på

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
	hasDirectionChangeAnnouncement bool,
) (completed types.CompletedOrderTable, needsExtraDoorTime bool) {

	lastFloor := config.NFloors - 1
	firstFloor := 0

	if hasDirectionChangeAnnouncement {
		switch arrivalDir {
		case types.Up:
			if clearHallOrder(elevator, floor, types.Down) {
				completed[floor][int(types.BtnHallDown)] = true
			}
		case types.Down:
			if clearHallOrder(elevator, floor, types.Up) {
				completed[floor][int(types.BtnHallUp)] = true
			}
		}
		return
	}

	if clearCabOrder(elevator, floor) {
		completed[floor][int(types.BtnCab)] = true
	}

	switch arrivalDir {
	case types.Up:
		if clearHallOrder(elevator, floor, types.Up) {
			completed[floor][int(types.BtnHallUp)] = true
		}

		if !hasLocalOrderAbove(*elevator) && elevator.LocalOrders[floor][int(types.BtnHallDown)] {
			if elevator.CurrentFloor == lastFloor {
				if clearHallOrder(elevator, floor, types.Down) {
					completed[floor][int(types.BtnHallDown)] = true
				}
			} else {
				needsExtraDoorTime = true
			}
		}
	case types.Down:
		if clearHallOrder(elevator, floor, types.Down) {
			completed[floor][int(types.BtnHallDown)] = true
		}

		if !hasLocalOrderBelow(*elevator) && elevator.LocalOrders[floor][int(types.BtnHallUp)] {
			if elevator.CurrentFloor == firstFloor {
				if clearHallOrder(elevator, floor, types.Up) {
					completed[floor][int(types.BtnHallUp)] = true
				}
			} else {
				needsExtraDoorTime = true
			}
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

// Jeg tror det skal være en bedre måte å gjøre det her på
