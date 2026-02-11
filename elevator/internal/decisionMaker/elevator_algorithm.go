package decisionMaker

// ------------------------------------------------------------------------------------
//	This file contains functionality from delivered HRA code transelated to golang
// ------------------------------------------------------------------------------------

import (
	."elevator/internal/config"
	"errors"
	"sort"
)

// ------------------------------------------------------------------------------------
//	Constants
// ------------------------------------------------------------------------------------

var TravelDurationMs int64 = 2500
var IncludeCab bool = false
var ClearRequests ClearRequestType = ClearInDirn

// ------------------------------------------------------------------------------------
//	enum types
// ------------------------------------------------------------------------------------

type CallType int

const (
	CallTypeHallUp CallType = iota
	CallTypeHallDown
	CallTypeCab
)

// type Dirn int

// const (
// 	DirnDown Dirn = -1
// 	DirnStop Dirn = 0
// 	DirnUp   Dirn = 1
// )
//Erstattes med types.motorDirection

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
	Direction   types.motorDirection
	CabRequests []bool
}

type ElevatorState struct {
	Behaviour ElevatorBehaviour
	Floor     int
	Direction Dirn
	Requests  [][]bool
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

func withRequests(e LocalElevatorState, hallReqs [][2]bool) ElevatorState {
	requests := make([][]bool, len(hallReqs))
	for f := 0; f < len(hallReqs); f++ {
		row := make([]bool, 3)
		row[0] = hallReqs[f][0]
		row[1] = hallReqs[f][1]
		if f < len(e.CabRequests) {
			row[2] = e.CabRequests[f]
		}
		requests[f] = row
	}
	return ElevatorState{
		Behaviour: e.Behaviour,
		Floor:     e.Floor,
		Direction: e.Direction,
		Requests:  requests,
	}
}

func (e ElevatorState) requestsAbove() bool {
	if e.Floor+1 >= len(e.Requests) {
		return false
	}
	for f := e.Floor + 1; f < len(e.Requests); f++ {
		if anyInRow(e.Requests[f]) {
			return true
		}
	}
	return false
}

func (e ElevatorState) requestsBelow() bool {
	if e.Floor <= 0 {
		return false
	}
	for f := 0; f < e.Floor; f++ {
		if anyInRow(e.Requests[f]) {
			return true
		}
	}
	return false
}

func (e ElevatorState) anyRequests() bool {
	for f := 0; f < len(e.Requests); f++ {
		if anyInRow(e.Requests[f]) {
			return true
		}
	}
	return false
}

func (e ElevatorState) anyRequestsAtFloor() bool {
	if e.Floor < 0 || e.Floor >= len(e.Requests) {
		return false
	}
	return anyInRow(e.Requests[e.Floor])
}

func (e ElevatorState) shouldStop() bool {
	switch e.Direction {
	case DirnUp:
		return e.Requests[e.Floor][CallTypeHallUp] ||
			e.Requests[e.Floor][CallTypeCab] ||
			!e.requestsAbove() ||
			e.Floor == 0 ||
			e.Floor == len(e.Requests)-1
	case DirnDown:
		return e.Requests[e.Floor][CallTypeHallDown] ||
			e.Requests[e.Floor][CallTypeCab] ||
			!e.requestsBelow() ||
			e.Floor == 0 ||
			e.Floor == len(e.Requests)-1
	case DirnStop:
		return true
	default:
		return true
	}
}

func (e ElevatorState) chooseDirection() Dirn {
	switch e.Direction {
	case DirnUp:
		if e.requestsAbove() {
			return DirnUp
		}
		if e.anyRequestsAtFloor() {
			return DirnStop
		}
		if e.requestsBelow() {
			return DirnDown
		}
		return DirnStop
	case DirnDown, DirnStop:
		if e.requestsBelow() {
			return DirnDown
		}
		if e.anyRequestsAtFloor() {
			return DirnStop
		}
		if e.requestsAbove() {
			return DirnUp
		}
		return DirnStop
	default:
		return DirnStop
	}
}

func clearReqsAtFloor(e ElevatorState, onClearedRequest func(CallType)) ElevatorState {
	e2 := ElevatorState{
		Behaviour: e.Behaviour,
		Floor:     e.Floor,
		Direction: e.Direction,
		Requests:  copyRequests(e.Requests),
	}

	clear := func(c CallType) {
		if e2.Requests[e2.Floor][c] {
			if onClearedRequest != nil {
				onClearedRequest(c)
			}
			e2.Requests[e2.Floor][c] = false
		}
	}

	switch ClearRequests {
	case ClearAll:
		for c := CallTypeHallUp; c <= CallTypeCab; c++ {
			clear(c)
		}
	case ClearInDirn:
		clear(CallTypeCab)
		switch e.Direction {
		case DirnUp:
			if e2.Requests[e2.Floor][CallTypeHallUp] {
				clear(CallTypeHallUp)
			} else if !e2.requestsAbove() {
				clear(CallTypeHallDown)
			}
		case DirnDown:
			if e2.Requests[e2.Floor][CallTypeHallDown] {
				clear(CallTypeHallDown)
			} else if !e2.requestsBelow() {
				clear(CallTypeHallUp)
			}
		case DirnStop:
			clear(CallTypeHallUp)
			clear(CallTypeHallDown)
		}
	}

	return e2
}

func OptimalHallRequests(hallReqs [][2]bool, elevatorStates map[string]LocalElevatorState) (map[string][][]bool, error) {
	numFloors := len(hallReqs)
	if numFloors == 0 {
		return nil, errors.New("no floors in hallRequests")
	}
	if len(elevatorStates) == 0 {
		return nil, errors.New("no elevator states provided")
	}
	for _, s := range elevatorStates {
		if len(s.CabRequests) != numFloors {
			return nil, errors.New("hall and cab requests do not all have the same length")
		}
		if s.Floor < 0 || s.Floor >= numFloors {
			return nil, errors.New("some elevator is at an invalid floor")
		}
		if s.Behaviour == ElevatorMoving {
			next := s.Floor + int(s.Direction)
			if next < 0 || next >= numFloors {
				return nil, errors.New("some elevator is moving away from an end floor")
			}
		}
	}

	reqs := toReq(hallReqs)
	states := initialStates(elevatorStates)

	for i := range states {
		performInitialMove(&states[i], reqs)
	}

	for {
		sort.Slice(states, func(i, j int) bool {
			if states[i].Time == states[j].Time {
				return states[i].ID < states[j].ID
			}
			return states[i].Time < states[j].Time
		})

		done := true
		if anyUnassigned(reqs) {
			done = false
		}
		if unvisitedAreImmediatelyAssignable(reqs, states) {
			assignImmediate(reqs, states)
			done = true
		}
		if done {
			break
		}

		performSingleMove(&states[0], reqs)
	}

	result := make(map[string][][]bool, len(elevatorStates))
	for id := range elevatorStates {
		width := 2
		if IncludeCab {
			width = 3
		}
		result[id] = make([][]bool, numFloors)
		for f := 0; f < numFloors; f++ {
			result[id][f] = make([]bool, width)
			if IncludeCab {
				result[id][f][2] = elevatorStates[id].CabRequests[f]
			}
		}
	}

	for f := 0; f < numFloors; f++ {
		for c := 0; c < 2; c++ {
			if reqs[f][c].Active {
				result[reqs[f][c].AssignedTo][f][c] = true
			}
		}
	}

	return result, nil
}

func toReq(hallReqs [][2]bool) [][]Req {
	r := make([][]Req, len(hallReqs))
	for f := 0; f < len(hallReqs); f++ {
		row := make([]Req, 2)
		for c := 0; c < 2; c++ {
			row[c] = Req{Active: hallReqs[f][c]}
		}
		r[f] = row
	}
	return r
}

func initialStates(states map[string]LocalElevatorState) []State {
	ids := make([]string, 0, len(states))
	for id := range states {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	out := make([]State, 0, len(states))
	for i, id := range ids {
		out = append(out, State{
			ID:    id,
			State: states[id],
			Time:  int64(i),
		})
	}
	return out
}

func performInitialMove(s *State, reqs [][]Req) {
	switch s.State.Behaviour {
	case ElevatorDoorOpen:
		s.Time += int64(config.DoorOpenDuration.Milliseconds()) * 1000 / 2
		fallthrough
	case ElevatorIdle:
		for c := 0; c < 2; c++ {
			if reqs[s.State.Floor][c].Active {
				reqs[s.State.Floor][c].AssignedTo = s.ID
				s.Time += int64(config.DoorOpenDuration.Milliseconds()) * 1000
			}
		}
	case ElevatorMoving:
		s.State.Floor += int(s.State.Direction)
		s.Time += TravelDurationMs * 1000 / 2
	}
}

func performSingleMove(s *State, reqs [][]Req) {
	e := withRequests(s.State, filterReq(reqs, func(r Req) bool { return isUnassigned(r) }))

	onClearRequest := func(c CallType) {
		switch c {
		case CallTypeHallUp, CallTypeHallDown:
			reqs[s.State.Floor][c].AssignedTo = s.ID
		case CallTypeCab:
			s.State.CabRequests[s.State.Floor] = false
		}
	}

	switch s.State.Behaviour {
	case ElevatorMoving:
		if e.shouldStop() {
			s.State.Behaviour = ElevatorDoorOpen
			s.Time += int64(config.DoorOpenDuration.Milliseconds()) * 1000
			_ = clearReqsAtFloor(e, onClearRequest)
		} else {
			s.State.Floor += int(s.State.Direction)
			s.Time += TravelDurationMs * 1000
		}
	case ElevatorIdle, ElevatorDoorOpen:
		s.State.Direction = e.chooseDirection()
		if s.State.Direction == DirnStop {
			if e.anyRequestsAtFloor() {
				_ = clearReqsAtFloor(e, onClearRequest)
				s.Time += int64(config.DoorOpenDuration.Milliseconds()) * 1000
				s.State.Behaviour = ElevatorDoorOpen
			} else {
				s.State.Behaviour = ElevatorIdle
			}
		} else {
			s.State.Behaviour = ElevatorMoving
			s.State.Floor += int(s.State.Direction)
			s.Time += TravelDurationMs * 1000
		}
	}
}

func anyUnassigned(reqs [][]Req) bool {
	for f := 0; f < len(reqs); f++ {
		for c := 0; c < 2; c++ {
			if isUnassigned(reqs[f][c]) {
				return true
			}
		}
	}
	return false
}

func isUnassigned(r Req) bool {
	return r.Active && r.AssignedTo == ""
}

func filterReq(reqs [][]Req, fn func(Req) bool) [][2]bool {
	out := make([][2]bool, len(reqs))
	for f := 0; f < len(reqs); f++ {
		out[f][0] = fn(reqs[f][0])
		out[f][1] = fn(reqs[f][1])
	}
	return out
}

func unvisitedAreImmediatelyAssignable(reqs [][]Req, states []State) bool {
	for _, s := range states {
		if anyBool(s.State.CabRequests) {
			return false
		}
	}
	for f := 0; f < len(reqs); f++ {
		countActive := 0
		for c := 0; c < 2; c++ {
			if reqs[f][c].Active {
				countActive++
			}
		}
		if countActive == 2 {
			return false
		}
		for c := 0; c < 2; c++ {
			if isUnassigned(reqs[f][c]) {
				found := false
				for _, s := range states {
					if s.State.Floor == f && !anyBool(s.State.CabRequests) {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
	}
	return true
}

func assignImmediate(reqs [][]Req, states []State) {
	for f := 0; f < len(reqs); f++ {
		for c := 0; c < 2; c++ {
			for i := range states {
				if isUnassigned(reqs[f][c]) {
					if states[i].State.Floor == f && !anyBool(states[i].State.CabRequests) {
						reqs[f][c].AssignedTo = states[i].ID
						states[i].Time += int64(config.DoorOpenDuration.Milliseconds()) * 1000
					}
				}
			}
		}
	}
}

func anyInRow(row []bool) bool {
	for _, v := range row {
		if v {
			return true
		}
	}
	return false
}

func anyBool(vals []bool) bool {
	for _, v := range vals {
		if v {
			return true
		}
	}
	return false
}

func copyRequests(reqs [][]bool) [][]bool {
	out := make([][]bool, len(reqs))
	for i := range reqs {
		out[i] = make([]bool, len(reqs[i]))
		copy(out[i], reqs[i])
	}
	return out
}
