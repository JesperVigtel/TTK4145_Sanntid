package decisionMaker

import (
	."elevator/internal/config"
    ."elevator/internal/types"
    "os/exec"
    "fmt"
    "encoding/json"
    "runtime"
)



func mergeNetworkHallOrders(
	localDecisionBasis 		*DecisionBasisFromAssigner, 
	consensusGlobalBasis 	DecisionBasisFromNetwork, 
	elevatorID int) 		DecisionBasisFromAssigner {


	return *localDecisionBasis 	//Placeholder
}




//Assignment based on deterministic HRA
func assignLocalOrders(decisionBasis DecisionBasisFromNetwork, elevatorID int) CabOrderTable {
	var localOrderTable CabOrderTable

	// Fyll inn input fra decisionBasis!
	input := HRAInput{
		HallRequests: [NFloors][2]bool{}, 		// ANTATT type er riktig
		States:       make(map[string]HRAElevState), 		// ANTATT type er riktig
	}

	for elevatorNum, alive := range decisionBasis.AliveList{
		if alive {
			elevatorNumStr := fmt.Sprintf("elevator_%d", elevatorNum)
			input.States[elevatorNumStr] = decisionBasis.ElevatorList[elevatorNum]
		}
	}

	for floor := 0; floor < NFloors; floor++ {
		for button := BTHallUp; button <= BTHallDown; button ++ {
			orderAssigned := true
			for elevatorNum, alive := range decisionBasis.AliveList {
				if alive && decisionBasis.HallOrderTable[elevatorNum][floor][button] != OrderAssigned{
					orderAssigned = false
					break
				}
			}
			input.HallRequests[floor][button] = orderAssigned
		}
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return localOrderTable
	}

	ret, err := exec.Command("hall_request_assigner", "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return localOrderTable
	}

	output := make(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return localOrderTable
	}

	// Tolke outputen for denne elevator, for elevID
	key := fmt.Sprintf("%d", elevatorID)
	assignedOrders, exists := output[key]
	if !exists {
		fmt.Printf("No orders assigned for elevator %d\n", elevatorID)
		return localOrderTable
	}

	localOrderTable := BuildOrderTable(input, output, elevatorID)

	return localOrderTable
}




func BuildOrderTable(
	hraOutput map[string][][2]bool,
	elevatorID string, 
	cabRequests []bool
	) 

	CabOrderTable {
    var orderTable CabOrderTable

	cabRequests := input.States[elevatorIDStr].CabRequests
	elevatorIDStr := fmt.Sprintf("elevator_%d", elevatorID)

    if assignedOrders, found := hraOutput[elevatorID]; found {
        for floor := 0; floor < NFloors && floor < len(assignedOrders); floor++ {
            for btn := BTHallUp; btn < BTCab; btn++ {
                orderTable[floor][btn] = assignedOrders[floor][btn]
            }
            orderTable[floor][BTCab] = cabRequests[floor]
        }
    }
    return orderTable
}

// //Consider adding to assignOrder
// 	hraExecutable := ""
// 	switch runtime.GOOS {
// 	case "linux":
// 		hraExecutable = "hall_request_assigner"
// 	case "windows":
// 		hraExecutable = "hall_request_assigner.exe"
// 	default:
// 		fmt.Println("OS not supported")
// 		return orderTable
// 	}