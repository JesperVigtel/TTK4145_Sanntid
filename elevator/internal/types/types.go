package types

import (
	. "elevator/internal/config"
)

// ------------------------------------------------------------------------------------
//	enum types for intermodule contracts
// ------------------------------------------------------------------------------------

type ConvergedSystemState struct {
	AliveList    [NElevators]bool
	ElevatorList [NElevators]HRAElevState
	OrderTables  [NElevators]OrderTable
}

type LocalSystemState struct {
	ElevatorID    int
	AliveStatus   bool
	ElevatorState HRAElevState
	OrderStates   OrderTable
}

type Message struct {
	SenderID     int
	ElevatorList [NElevators]HRAElevState
	OrderTables  [NElevators]OrderTable
	AliveStatus  bool
}

//Elevator types START:

type MotorDirection int

const (
	Down MotorDirection = iota - 1
	Stop
	Up
)

type Elevator struct {
	CurrentFloor           int
	CurrentTravelDirection MotorDirection
	PhysicalMotorDirection MotorDirection
	AssignedOrders         AssignedOrderTable
	Behaviour              ElevatorBehaviour
	ActiveStatus           bool
}

type CompletedOrderTable [NFloors][NButtons]bool

type ButtonType int

const (
	BtnHallUp ButtonType = iota
	BtnHallDown
	BtnCab
)

type ElevatorBehaviour int

const (
	ElevatorIdle ElevatorBehaviour = iota
	ElevatorMoving
	ElevatorDoorOpen
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type ElevatorEvents struct {
	Elevator        Elevator
	CompletedOrders CompletedOrderTable
	NewButtonPress  *ButtonEvent
	Obstructed      bool
}

//Elevator types stop

//Network types START:

type GlobalNodeRegistry struct {
	Nodes []int
	New   []int
	Lost  []int
}

//Network types END:

// ------------------------------------------------------------------------------------
//
//	enum types for order domain
//
// ------------------------------------------------------------------------------------
type OrderState int

const (
	OrderStandby OrderState = iota
	OrderPending
	OrderAssigned
	OrderComplete
)

type OrderTable [NFloors][NButtons]OrderState
type AssignedOrderTable [NFloors][NButtons]bool

// ------------------------------------------------------------------------------------
//	enum types for assigner
// ------------------------------------------------------------------------------------

type HRAElevState struct {
	Behavior   string `json:"behaviour"`
	Floor      int    `json:"floor"`
	Direction  string `json:"direction"`
	Assignable bool   `json:"assignable"`
}

func NewHRAElevState(elev Elevator, assignable bool) HRAElevState {
	return HRAElevState{
		Behavior:   elevBehaviorToString(elev.Behaviour),
		Floor:      elev.CurrentFloor,
		Direction:  elevDirectionToString(elev.PhysicalMotorDirection),
		Assignable: assignable,
	}
}

type HRAAssignerState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

func NewHRAAssignerState(elevState HRAElevState, orders OrderTable) HRAAssignerState {
	return HRAAssignerState{
		Behavior:    elevState.Behavior,
		Floor:       elevState.Floor,
		Direction:   elevState.Direction,
		CabRequests: CabRequestsFromOrderTable(orders),
	}
}

func CabRequestsFromOrderTable(orderTable OrderTable) []bool {
	requests := make([]bool, NFloors)
	for floor := range NFloors {
		requests[floor] = IsActiveOrder(orderTable[floor][BtnCab])
	}
	return requests
}

func IsActiveOrder(orderState OrderState) bool {
	return orderState == OrderPending || orderState == OrderAssigned
}

func elevBehaviorToString(b ElevatorBehaviour) string {
	switch b {
	case ElevatorIdle:
		return "idle"
	case ElevatorMoving:
		return "moving"
	case ElevatorDoorOpen:
		return "doorOpen"
	default:
		return "idle"
	}
}

func elevDirectionToString(d MotorDirection) string {
	switch d {
	case Up:
		return "up"
	case Down:
		return "down"
	default:
		return "stop"
	}
}

type HRAInput struct {
	HallRequests [NFloors][2]bool            `json:"hallRequests"`
	States       map[string]HRAAssignerState `json:"states"`
}
