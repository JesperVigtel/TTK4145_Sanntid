package decisionMaker

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

func toHRAElevState(e Elevator) HRAElevState {
	return HRAElevState{
		Behavior:    behaviourToString(e.Behaviour),
		Floor:       e.CurrentFloor,
		Direction:   directionToString(e.MotorDirection),
		CabRequests: cabTableToBoolSlice(e.Request),
	}
}

func behaviourToString(b ElevatorBehaviour) string {
	switch b {
		case ElevatorIdle:     return "idle"
		case ElevatorMoving:   return "moving"
		case ElevatorDoorOpen: return "doorOpen"
		default:               return "idle"
	}
}

func directionToString(d MotorDirection) string {
	switch d {
		case Up:   return "up"
		case Down: return "down"
		default:   return "stop"
	}
}

func cabTableToBoolSlice(t CabOrderTable) []bool {
	result := make([]bool, NFloors)
	for floor := range t {
		result[floor] = t[floor][BTCab]
	}
	return result
}