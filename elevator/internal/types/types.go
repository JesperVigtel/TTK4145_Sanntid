package types

import (
	. "elevator/internal/config"
)

// Felles datatyper, structs og konstanter brukt på tvers av alle pakker.
// Sikrer konsistente data og kontrakter mellom moduler.

// TODO: Definer structs som Elevator, Order, ElevatorState osv.
// TODO: Definer konstanter for etasjer, motorretning osv.

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
	Request        [NFloors][NButtons]bool
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

type OrderEvent struct {
	Floor  int
	Button ButtonType
}

type FromLocalToDM struct {
	Elevator       Elevator
	CompletedOrder [NFloors][NButtons]bool
	NewButtonPress *OrderEvent 
	Obstructed     bool        
}

type ButtonState int

const (
	initial ButtonState = iota
	standby
	ButtonPressed
	OrderAssigned
	OrderCompleted
)
//Elevator types stop

//Network types START:

type Message struct {
	SenderID      int
	ElevatorList  [NElevators]int                    //ElevatorState??
	HallOrderList [NElevators][NFloors][NButtons]int //ButtonState??
	AliveStatus   bool
	AliveList     [NElevators]bool
}

//Network types END:

// ------------------------------------------------------------------------------------
//	enum types from elevator_algorithm
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

// bruke vanlig elevator isteden eller?? har en nesten lik struct over (Andreas)
type LocalElevatorState struct {
	Behaviour   ElevatorBehaviour
	Floor       int
	Direction   MotorDirection
	CabRequests []bool
}

type Req struct {
	Active     bool
	AssignedTo string
}

type State struct {
	ID    string
	State LocalElevatorState
	Time  int64
}

// type Dirn int

// const (
// 	DirnDown Dirn = -1
// 	DirnStop Dirn = 0
// 	DirnUp   Dirn = 1
// )
//Erstattes med types.motorDirection
