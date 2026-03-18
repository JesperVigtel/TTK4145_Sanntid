package dispatch

import (
	"elevator/internal/config"
	"elevator/internal/types"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

func getHRAPath() string {
	_, currentFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(currentFile)

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(dir, "hall_request_assigner_mac")
	default: // linux and others
		return filepath.Join(dir, "hall_request_assigner")
	}
}

func computeAssignedOrders(
	convergedState types.ConvergedSystemState,
	localSystemState types.LocalSystemState,
	elevatorID int,
) types.AssignedOrderTable {
	input := buildHRAInput(convergedState, localSystemState, elevatorID)
	if len(input.States) == 0 {
		return fallbackAssignedOrders(localSystemState)
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("computeAssignedOrders: json.Marshal:", err)
		return fallbackAssignedOrders(localSystemState)
	}
	hraPath := getHRAPath()
	raw, err := exec.Command(hraPath, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("computeAssignedOrders: exec:", err, string(raw))
		return fallbackAssignedOrders(localSystemState)
	}

	output := make(map[string][][2]bool)
	if err := json.Unmarshal(raw, &output); err != nil {
		fmt.Println("computeAssignedOrders: json.Unmarshal:", err)
		return fallbackAssignedOrders(localSystemState)
	}

	return buildAssignedOrderTable(output, localSystemState, elevatorID)
}

func buildHRAInput(
	convergedState types.ConvergedSystemState,
	localSystemState types.LocalSystemState,
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
			elevState = localSystemState.ElevatorState
		}
		if !elevState.Assignable {
			continue
		}
		if elevState.Floor < 0 || elevState.Floor >= config.NFloors {
			continue
		}
		orders := convergedState.OrderTables[id]
		if id == elevatorID {
			orders = localSystemState.OrderStates
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
	localSystemState types.LocalSystemState,
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
		result[floor][types.BtnCab] = types.IsActiveOrder(localSystemState.OrderStates[floor][types.BtnCab])
	}

	return result
}

func fallbackAssignedOrders(localSystemState types.LocalSystemState) types.AssignedOrderTable {
	var result types.AssignedOrderTable
	for floor := range config.NFloors {
		result[floor][types.BtnCab] = types.IsActiveOrder(localSystemState.OrderStates[floor][types.BtnCab])
		if !localSystemState.ElevatorState.Assignable {
			continue
		}
		for btn := types.BtnHallUp; btn <= types.BtnHallDown; btn++ {
			state := localSystemState.OrderStates[floor][btn]
			result[floor][btn] = state == types.OrderPending || state == types.OrderAssigned
		}
	}
	return result
}


