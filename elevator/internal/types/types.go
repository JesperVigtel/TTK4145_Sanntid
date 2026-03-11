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
	AliveList      [NElevators]bool
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

type LocalLightUpdate struct {
	CabLights    [NFloors]bool
	DoorOpen     bool
	CurrentFloor int
}

type HallLightUpdate struct {
	HallUp   [NFloors]bool
	HallDown [NFloors]bool
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
type LocalOrderTable [NFloors][NButtons]bool

// er det en ide å lage en completedorders matrise som man kan bruke?? bruk dette til å fjerne matrisehelvete fra
// sendElevatorUpdate()

// ------------------------------------------------------------------------------------
//	enum types for assigner
// ------------------------------------------------------------------------------------

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

// NewHRAElevState converts an Elevator to HRAElevState for protocol transmission.
// Centralized here (CC §5.2) for cohesion: conversion logic lives with type definition.
func NewHRAElevState(elev Elevator) HRAElevState {
	return HRAElevState{
		Behavior:    elevBehaviorToString(elev.Behaviour),
		Floor:       elev.CurrentFloor,
		Direction:   elevDirectionToString(elev.CurrentTravelDirection),
		CabRequests: make([]bool, NFloors),
	}
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
	HallRequests [NFloors][2]bool        `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}
