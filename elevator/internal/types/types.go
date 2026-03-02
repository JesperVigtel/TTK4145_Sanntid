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


type LocalElevatorFromDriver struct{
	Elevator 		Elevator
	ExecutedOrders 	CabOrderTable
}

//Elevator types stop

//Network types START:

type Message struct {
	SenderID      int
	ElevatorList  [NElevators]HRAElevState                   
	HallOrderList HallOrderTable 
	AliveStatus   bool
	AliveList     [NElevators]bool
}


type NetworkNodeRegistry struct {
    Nodes []int // all currently active nodes
    New   []int // nodes that just came online
    Lost  []int // nodes that just went offline
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
type CabOrderTable 	[NFloors][NButtons]bool


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


//Ønsker å endre DecisonBasis til SystemState. Skal visstnok være mer korrekt (Jesper)


// Flows: DecisionMaker → ConsensusManager
type AgreedSystemState struct {
	AliveList    	[NElevators]bool
	ElevatorList  	[NElevators]HRAElevState
	HallOrderTable 	[NElevators]HallOrderTable
}


// Flows: ConsensusManager → DecisionMaker
type LocalSystemState struct{
	AliveStatus    	bool
	ElevatorState  	HRAElevState
	HallRequests 	HallOrderTable
}