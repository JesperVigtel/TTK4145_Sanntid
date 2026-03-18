package dispatch

import (
	"elevator/internal/config"
	"elevator/internal/types"
	"encoding/json"
	"fmt"
	"os/exec"
)

func computeAssignedOrders(
	convergedState types.ConvergedSystemState,
	localState types.LocalSystemState,
	elevatorID int,
) types.AssignedOrderTable {
	input := buildHallAssignerInput(convergedState, localState, elevatorID)
	if len(input.States) == 0 {
		// External HRA asserts on empty state sets.
		fmt.Println("computeAssignedOrders: no alive elevator states, using local fallback assignment")
		return fallbackAssignedOrders(localState)
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("computeAssignedOrders: json.Marshal:", err)
		return fallbackAssignedOrders(localState)
	}
	hraPath := hallRequestAssignerPath()
	raw, err := exec.Command(hraPath, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("computeAssignedOrders: exec:", err, string(raw))
		return fallbackAssignedOrders(localState)
	}

	output := make(map[string][][2]bool)
	if err := json.Unmarshal(raw, &output); err != nil {
		fmt.Println("computeAssignedOrders: json.Unmarshal:", err)
		return fallbackAssignedOrders(localState)
	}

	return buildAssignedOrderTable(output, localState, elevatorID)
}

func buildHallAssignerInput(
	convergedState types.ConvergedSystemState,
	localState types.LocalSystemState,
	elevatorID int,
) types.HRAInput {
	input := types.HRAInput{
		HallRequests: [config.NFloors][2]bool{},
		States:       make(map[string]types.HRAAssignerState),
	}

	for id, alive := range convergedState.AliveList {
		if !alive {
			continue
		}
		elevState := convergedState.ElevatorList[id]
		if id == elevatorID {
			elevState = localState.ElevatorState
		}
		if !elevState.Assignable {
			continue
		}
		if elevState.Floor < 0 || elevState.Floor >= config.NFloors {
			continue
		}
		orders := convergedState.OrderTables[id]
		if id == elevatorID {
			orders = localState.OrderStates
		}
		input.States[fmt.Sprintf("elevator_%d", id)] = types.NewHRAAssignerState(elevState, orders)
	}

	for floor := range config.NFloors {
		for btn := types.BtnHallUp; btn <= types.BtnHallDown; btn++ {
			input.HallRequests[floor][btn] = convergedState.OrderTables[elevatorID][floor][btn] == types.OrderAssigned
		}
	}
	return input
}

func buildHallLampTable(orderStates types.OrderTable) types.HallLampTable {
	var hallLamps types.HallLampTable
	for floor := range config.NFloors {
		for btn := types.BtnHallUp; btn <= types.BtnHallDown; btn++ {
			hallLamps[floor][btn] = orderStates[floor][btn] == types.OrderAssigned
		}
	}
	return hallLamps
}

func buildAssignedOrderTable(
	output map[string][][2]bool,
	localState types.LocalSystemState,
	elevatorID int,
) types.AssignedOrderTable {
	var result types.AssignedOrderTable
	idStr := fmt.Sprintf("elevator_%d", elevatorID)

	if assigned, found := output[idStr]; found {
		for floor := 0; floor < config.NFloors && floor < len(assigned); floor++ {
			for btn := types.BtnHallUp; btn <= types.BtnHallDown; btn++ {
				result[floor][btn] = assigned[floor][btn]
			}
		}
	}

	for floor := range config.NFloors {
		result[floor][types.BtnCab] = types.IsActiveOrder(localState.OrderStates[floor][types.BtnCab])
	}

	return result
}

func fallbackAssignedOrders(localState types.LocalSystemState) types.AssignedOrderTable {
	var result types.AssignedOrderTable
	for floor := range config.NFloors {
		result[floor][types.BtnCab] = types.IsActiveOrder(localState.OrderStates[floor][types.BtnCab])
		if !localState.ElevatorState.Assignable {
			continue
		}
		for btn := types.BtnHallUp; btn <= types.BtnHallDown; btn++ {
			state := localState.OrderStates[floor][btn]
			result[floor][btn] = state == types.OrderPending || state == types.OrderAssigned
		}
	}
	return result
}
