package types

import (
	. "elevator/internal/config"
)

// Felles datatyper, structs og konstanter brukt på tvers av alle pakker.
// Sikrer konsistente data og kontrakter mellom moduler.



// ------------------------------------------------------------------------------------
//	enum types for intermodule contracts
// ------------------------------------------------------------------------------------


// Flows: DecisionMaker → ConsensusManager
type AgreedSystemState struct {
	AliveList    	[NElevators]bool
	ElevatorList  	[NElevators]HRAElevState
	HallOrderTable 	[NElevators]HallOrderTable
}


// Flows: ConsensusManager → DecisionMaker
type LocalSystemState struct{
	ElevatorID		int
	AliveStatus    	bool
	ElevatorState  	HRAElevState
	HallRequests 	HallOrderTable
}


type Message struct {
	SenderID      int
	ElevatorList  [NElevators]HRAElevState                   
	HallOrderList HallOrderTable 
	AliveStatus   bool
	AliveList     [NElevators]bool
}


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
	Request        CabOrderTable
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

type ButtonEvent struct {	//Changed back to ButtonEvent for concistency (Jesper)
	Floor  int
	Button ButtonType
}

type FromLocalToDM struct {
	Elevator       Elevator
	CompletedOrder CabOrderTable
	NewButtonPress *ButtonEvent 
	Obstructed     bool        
}


// type LocalElevatorFromDriver struct{	//Redundant, shoudl be removed
// 	Elevator 		Elevator
// 	ExecutedOrders 	CabOrderTable
// }

//Elevator types stop

//Network types START:



type GlobalNodeRegistry struct {
    Nodes []int 
    New   []int 
    Lost  []int 
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



// ------------------------------------------------------------------------------------
//	enum types for order domain
// ------------------------------------------------------------------------------------
type OrderState int

const (
	OrderStandby OrderState = iota
	OrderPending
	OrderAssigned
	OrderComplete
)

type HallOrderTable [NFloors][NButtons]OrderState
type CabOrderTable 	[NFloors][NButtons]bool




// ------------------------------------------------------------------------------------
//	enum types for assigner.go
// ------------------------------------------------------------------------------------

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


