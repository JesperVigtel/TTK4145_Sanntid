package dispatch

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
	"encoding/json"
	"fmt"
	"os/exec"
)

func prepareAssignment(
	localState LocalSystemState,
	convergedState ConvergedSystemState,
) (LocalOrderTable, HallOrderTable) {
	elevatorID := localState.ElevatorID
	assignedOrders := computeAssignedOrders(convergedState, localState, elevatorID)
	lightUpdate := computeLightUpdate(convergedState, elevatorID)
	return assignedOrders, lightUpdate
}

func mergeConvergedHallOrders(
	localState LocalSystemState,
	convergedState ConvergedSystemState,
	elevatorID int,
) LocalSystemState {
	for floor := range NFloors {
		for btn := range NButtons {
			convergedOrder := convergedState.HallOrderTable[elevatorID][floor][btn]
			localOrder := localState.HallRequests[floor][btn]
			if localOrder == OrderComplete && convergedOrder == OrderAssigned {
				continue
			}
			localState.HallRequests[floor][btn] = convergedOrder
		}
	}
	return localState
}

func computeAssignedOrders(
	convergedState ConvergedSystemState,
	localState LocalSystemState,
	elevatorID int,
) LocalOrderTable {
	input := buildHallAssignerInput(convergedState, localState, elevatorID)
	if len(input.States) == 0 {
		// External HRA asserts on empty state sets.
		fmt.Println("computeAssignedOrders: no alive elevator states, using local fallback assignment")
		return localFallbackOrders(localState)
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("computeAssignedOrders: json.Marshal:", err)
		return localFallbackOrders(localState)
	}

	raw, err := exec.Command("hall_request_assigner", "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("computeAssignedOrders: exec:", err, string(raw))
		return localFallbackOrders(localState)
	}

	output := make(map[string][][2]bool)
	if err := json.Unmarshal(raw, &output); err != nil {
		fmt.Println("computeAssignedOrders: json.Unmarshal:", err)
		return localFallbackOrders(localState)
	}

	return buildLocalOrderTable(output, localState, elevatorID)
}

func buildHallAssignerInput(
	convergedState ConvergedSystemState,
	localState LocalSystemState,
	elevatorID int,
) HRAInput {
	input := HRAInput{
		HallRequests: [NFloors][2]bool{},
		States:       make(map[string]HRAElevState),
	}

	for id, alive := range convergedState.AliveList {
		if !alive {
			continue
		}
		elevState := convergedState.ElevatorList[id]
		if id == elevatorID {
			elevState.CabRequests = localState.ElevatorState.CabRequests
		}
		input.States[fmt.Sprintf("elevator_%d", id)] = elevState
	}

	for floor := range NFloors {
		for btn := BtnHallUp; btn <= BtnHallDown; btn++ {
			allAssigned := true
			for id, alive := range convergedState.AliveList {
				if alive && convergedState.HallOrderTable[id][floor][btn] != OrderAssigned {
					allAssigned = false
					break
				}
			}
			input.HallRequests[floor][btn] = allAssigned
		}

	}
	return input
}

func buildLocalOrderTable(
	output map[string][][2]bool,
	localState LocalSystemState,
	elevatorID int,
) LocalOrderTable {
	var result LocalOrderTable
	idStr := fmt.Sprintf("elevator_%d", elevatorID)

	if assigned, found := output[idStr]; found {
		for floor := 0; floor < NFloors && floor < len(assigned); floor++ {
			for btn := range BtnCab {
				result[floor][btn] = assigned[floor][btn]
			}
		}
	}

	for floor := range NFloors {
		result[floor][BtnCab] = localState.ElevatorState.CabRequests[floor]
	}

	return result
}

func computeLightUpdate(convergedState ConvergedSystemState, elevatorID int) HallOrderTable {
	return convergedState.HallOrderTable[elevatorID]
}

func localFallbackOrders(localState LocalSystemState) LocalOrderTable {
	var result LocalOrderTable
	for floor := range NFloors {
		result[floor][BtnCab] = localState.ElevatorState.CabRequests[floor]
		for btn := BtnHallUp; btn <= BtnHallDown; btn++ {
			state := localState.HallRequests[floor][btn]
			result[floor][btn] = state == OrderPending || state == OrderAssigned
		}
	}
	return result
}
