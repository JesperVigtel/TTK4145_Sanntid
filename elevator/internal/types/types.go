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
	ActiveSatus    bool
}

type ButtonType int

const (
	BTHallUp ButtonType = iota
	BTHallDown
	BTCab
)

type ButtonEvent struct {
	Floor int
	Button ButtonType
}

type ElevatorBehavior int

const (
	idle ElevatorBehavior = iota
	Moving
	DoorOpen
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

type ElevatorBehaviour int

const (
	ElevatorIdle ElevatorBehaviour = iota
	ElevatorMoving
	ElevatorDoorOpen
)

type ClearRequestType int

const (
	ClearAll ClearRequestType = iota
	ClearInDirn
)

type LocalElevatorState struct {
	Behaviour   ElevatorBehaviour
	Floor       int
	Direction   motordirection
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
