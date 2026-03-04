package dispatch

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
	"encoding/json"
	"fmt"
	"os/exec"
)

func prepareAssignment(
	agreedState AgreedSystemState,
	localState  LocalSystemState,
) (CabOrderTable, HallOrderTable) {
	elevatorID     := localState.ElevatorID
	assignedOrders := computeAssignedOrders(agreedState, localState, elevatorID)
	lightUpdate    := computeLightUpdate(agreedState, elevatorID)
	return assignedOrders, lightUpdate
}






func mergeAgreedHallOrders(
	localState  LocalSystemState,
	agreedState AgreedSystemState,
	elevatorID  int,
) LocalSystemState {
	for floor := 0; floor < NFloors; floor++ {
		for btn := 0; btn < NButtons; btn++ {
			agreedOrder := agreedState.HallOrderTable[elevatorID][floor][btn]
			localOrder  := localState.HallRequests[floor][btn]
			// Keep our completed state: we already finished this order but consensus
			// hasn't caught up yet and would otherwise overwrite it with Assigned
			if localOrder == OrderComplete && agreedOrder == OrderAssigned {
				continue
			}
			localState.HallRequests[floor][btn] = agreedOrder
		}
	}
	return localState
}

func computeAssignedOrders(
	agreedState AgreedSystemState,
	localState  LocalSystemState,
	elevatorID  int,
) CabOrderTable {
	var result CabOrderTable

	// If this elevator is not recognised as alive by the network, only serve
	// cab calls — hall orders require network agreement to assign safely.
	if !agreedState.AliveList[elevatorID] {
		return cabOrdersOnly(localState)
	}

	input    := buildHallAssignerInput(agreedState, localState, elevatorID)
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
	agreedState AgreedSystemState,
	localState  LocalSystemState,
	elevatorID  int,
) HRAInput {
	input := HRAInput{
		HallRequests: [NFloors][2]bool{},
		States:       make(map[string]HRAElevState),
	}

	for id, alive := range agreedState.AliveList {
		if !alive {
			continue
		}
		elevState := agreedState.ElevatorList[id]
		if id == elevatorID {
			elevState.CabRequests = localState.ElevatorState.CabRequests
		}
		input.States[fmt.Sprintf("elevator_%d", id)] = elevState
	}

	for floor := 0; floor < NFloors; floor++ {
		for btn := BTHallUp; btn <= BTHallDown; btn++ {
			allAssigned := true
			for id, alive := range agreedState.AliveList {
				if alive && agreedState.HallOrderTable[id][floor][btn] != OrderAssigned {
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
	output     map[string][][2]bool,
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

func computeLightUpdate(agreedState AgreedSystemState, elevatorID int) HallOrderTable {
	return agreedState.HallOrderTable[elevatorID]
}

func cabOrdersOnly(localState LocalSystemState) CabOrderTable {
	var result CabOrderTable
	for floor := 0; floor < NFloors; floor++ {
		result[floor][BTCab] = localState.ElevatorState.CabRequests[floor]
	}
	return result
}