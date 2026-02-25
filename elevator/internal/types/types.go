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

type ButtonState int	//deklarert lengre ned som OrderState (jesper)

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
//	enum types for decisionMaker
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



// // bruke vanlig elevator isteden eller?? har en nesten lik struct over (Andreas)
// type LocalElevatorState struct {
// 	Behaviour   ElevatorBehaviour
// 	Floor       int
// 	Direction   MotorDirection
// 	CabRequests []bool
// }

type Req struct {
	Active     bool
	AssignedTo string
}

type State struct {
	ID    string
	State Elevator
	Time  int64
}

// type Dirn int

// const (
// 	DirnDown Dirn = -1
// 	DirnStop Dirn = 0
// 	DirnUp   Dirn = 1
// )
//Erstattes med types.motorDirection


type OrderState int

const (
	OrderStandby OrderState = iota
	OrderPending
	OrderAssigned
	OrderComplete
)

type HallOrderTable [NFloors][NButtons]OrderState


type HRAElevState struct {
    Behavior    string      `json:"behaviour"`
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"`
    CabRequests []bool      `json:"cabRequests"`
}

type HRAInput struct {
    HallRequests    [NFloors][2]bool           	`json:"hallRequests"`
    States          map[string]HRAElevState     `json:"states"`
}


type DecisionBasisFromNetwork struct {
	AliveList    [NElevators]bool
	ElevatorList  [NElevators]HRAElevState
	HallOrderTable [NElevators][NFloors][NButtons]ButtonState
}

type LocalElevatorFromDriver struct{
	Elevator Elevator
	ExecutedOrders [NFloors][NButtons]bool
}


type DecisionBasisFromAssigner struct{	//Placeholder
	AliveList    [NElevators]bool
	ElevatorList  [NElevators]HRAElevState
	HallOrderTable [NElevators][NFloors][NButtons]ButtonState
}