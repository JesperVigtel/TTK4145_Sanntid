package dispatch

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

func toHRAElevState(elev Elevator) HRAElevState {
	return HRAElevState{
		Behavior:    behaviourToString(elev.Behaviour),
		Floor:       elev.CurrentFloor,
		Direction:   directionToString(elev.MotorDirection),
		CabRequests: make([]bool, NFloors),
	}
}

func behaviourToString(behaviour ElevatorBehaviour) string {
	switch behaviour {
		case ElevatorIdle:     return "idle"
		case ElevatorMoving:   return "moving"
		case ElevatorDoorOpen: return "doorOpen"
		default:               return "idle"
	}
}

func directionToString(dir MotorDirection) string {
	switch dir {
		case Up:   return "up"
		case Down: return "down"
		default:   return "stop"
	}
}