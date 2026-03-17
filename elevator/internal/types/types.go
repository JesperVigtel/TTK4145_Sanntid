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
	SenderID       int
	ElevatorList   [NElevators]HRAElevState
	HallOrderTable HallOrderTable
	AliveStatus    bool
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
	LocalOrders            LocalOrderTable
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
	Elevator       Elevator
	CompletedOrder CompletedOrderTable
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
type CabOrderTable [NFloors]OrderState
type LocalOrderTable [NFloors][NButtons]bool

// er det en ide å lage en completedorders matrise som man kan bruke?? bruk dette til å fjerne matrisehelvete fra
// sendElevatorUpdate()

// ------------------------------------------------------------------------------------
//	enum types for assigner
// ------------------------------------------------------------------------------------

type HRAElevState struct {
	Behavior  string        `json:"behaviour"`
	Floor     int           `json:"floor"`
	Direction string        `json:"direction"`
	CabOrders CabOrderTable `json:"cabOrders"`
}

func NewHRAElevState(elev Elevator) HRAElevState {
	return HRAElevState{
		Behavior:  elevBehaviorToString(elev.Behaviour),
		Floor:     elev.CurrentFloor,
		Direction: elevDirectionToString(elev.PhysicalMotorDirection),
	}
}

type HRAAssignerState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

func NewHRAAssignerState(elevState HRAElevState) HRAAssignerState {
	return HRAAssignerState{
		Behavior:    elevState.Behavior,
		Floor:       elevState.Floor,
		Direction:   elevState.Direction,
		CabRequests: CabRequestsFromOrders(elevState.CabOrders),
	}
}

func CabRequestsFromOrders(cabOrders CabOrderTable) []bool {
	requests := make([]bool, NFloors)
	for floor := range NFloors {
		requests[floor] = IsActiveOrder(cabOrders[floor])
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
