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
) (CabOrderTable, HallOrderTable) {
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
	for floor := 0; floor < NFloors; floor++ {
		for btn := 0; btn < NButtons; btn++ {
			convergedOrder := convergedState.HallOrderTable[elevatorID][floor][btn]
			localOrder := localState.HallRequests[floor][btn]
			// Keep our completed state: we already finished this order but consensus
			// hasn't caught up yet and would otherwise overwrite it with Assigned
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
) CabOrderTable {
	var result CabOrderTable

	// If this elevator is not recognised as alive by the network, only serve
	// cab calls — hall orders require network agreement to assign safely.
	if !convergedState.AliveList[elevatorID] {
		return cabOrdersOnly(localState)
	}

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

	return buildCabOrderTable(output, localState, elevatorID)
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

	for floor := 0; floor < NFloors; floor++ {
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

func buildCabOrderTable(
	output map[string][][2]bool,
	localState LocalSystemState,
	elevatorID int,
) CabOrderTable {
	var result CabOrderTable
	idStr := fmt.Sprintf("elevator_%d", elevatorID)

	if assigned, found := output[idStr]; found {
		for floor := 0; floor < NFloors && floor < len(assigned); floor++ {
			for btn := BTHallUp; btn < BTCab; btn++ {
				result[floor][btn] = assigned[floor][btn]
			}
		}
	}

	for floor := 0; floor < NFloors; floor++ {
		result[floor][BTCab] = localState.ElevatorState.CabRequests[floor]
	}

	return result
}

func computeLightUpdate(convergedState ConvergedSystemState, elevatorID int) HallOrderTable {
	return convergedState.HallOrderTable[elevatorID]
}

func cabOrdersOnly(localState LocalSystemState) CabOrderTable {
	var result CabOrderTable
	for floor := 0; floor < NFloors; floor++ {
		result[floor][BTCab] = localState.ElevatorState.CabRequests[floor]
	}
	return result
}
