package types

import (
	. "elevator/internal/config"
)

// ------------------------------------------------------------------------------------
//	enum types for intermodule contracts
// ------------------------------------------------------------------------------------

type ConvergedSystemState struct {
	AliveList      [NElevators]bool
	ElevatorList   [NElevators]HRAElevState
	HallOrderTable [NElevators]HallOrderTable
}

type LocalSystemState struct {
	ElevatorID    int
	AliveStatus   bool
	ElevatorState HRAElevState
	HallRequests  HallOrderTable
}

type Message struct {
	SenderID      int
	ElevatorList  [NElevators]HRAElevState
	HallOrderTable HallOrderTable
	AliveStatus   bool
	AliveList     [NElevators]bool
}

//Elevator types START:

type MotorDirection int

const (
	Down MotorDirection = iota - 1
	Stop
	Up
)

type Elevator struct {
	CurrentFloor   int
	MotorDirection MotorDirection
	LocalOrders    LocalOrderTable
	Behaviour      ElevatorBehaviour
	ActiveStatus   bool
}

type ButtonType int

const (
	BTHallUp ButtonType = iota
	BTHallDown
	BTCab
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

type FromLocalToDM struct {
	Elevator       Elevator
	CompletedOrder LocalOrderTable
	NewButtonPress *ButtonEvent
	Obstructed     bool
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
//	enum types for Local Control
// ------------------------------------------------------------------------------------

type CallType int

const (
	CallTypeHallUp CallType = iota
	CallTypeHallDown
	CallTypeCab
)

type ClearRequestType int

const (
	ClearAll ClearRequestType = iota
	ClearInDirn
)

type Req struct {
	Active     bool
	AssignedTo string
}

type State struct {
	ID    string
	State Elevator
	Time  int64
}

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

type HallOrderTable [NFloors][NButtons]OrderState
type LocalOrderTable [NFloors][NButtons]bool

// ------------------------------------------------------------------------------------
//	enum types for assigner
// ------------------------------------------------------------------------------------

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [NFloors][2]bool        `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}
