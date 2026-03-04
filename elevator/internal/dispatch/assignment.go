package dispatch

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
	"encoding/json"
	"fmt"
	"os/exec"
)

func prepareAssignment(
	convergedState ConvergedSystemState,
	localState LocalSystemState,
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
	var result LocalOrderTable

	// // If this elevator is not recognised as alive by the network, only serve
	// // cab calls — hall orders require network agreement to assign safely.
	// if !convergedState.AliveList[elevatorID] {
	// 	return cabOrdersOnly(localState)
	// }

	input := buildHallAssignerInput(convergedState, localState, elevatorID)
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("computeAssignedOrders: json.Marshal:", err)
		return result
	}

	raw, err := exec.Command("hall_request_assigner", "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("computeAssignedOrders: exec:", err, string(raw))
		return result
	}

	output := make(map[string][][2]bool)
	if err := json.Unmarshal(raw, &output); err != nil {
		fmt.Println("computeAssignedOrders: json.Unmarshal:", err)
		return result
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
		for btn := BTHallUp; btn <= BTHallDown; btn++ {
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
			for btn := range BTCab {
				result[floor][btn] = assigned[floor][btn]
			}
		}
	}

	for floor := range NFloors {
		result[floor][BTCab] = localState.ElevatorState.CabRequests[floor]
	}

	return result
}

func computeLightUpdate(convergedState ConvergedSystemState, elevatorID int) HallOrderTable {
	return convergedState.HallOrderTable[elevatorID]
}

func cabOrdersOnly(localState LocalSystemState) LocalOrderTable {
	var result LocalOrderTable
	for floor := range NFloors {
		result[floor][BTCab] = localState.ElevatorState.CabRequests[floor]
	}
	return result
}
