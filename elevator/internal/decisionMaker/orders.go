package decisionMaker

import (
	."elevator/internal/config"
    ."elevator/internal/types"
    "os/exec"
    "fmt"
    "encoding/json"
    "runtime"
)



func mergeNetworkHallOrders(localDecisionBasis *DecisionBasisFromAssigner, consensusGlobalBasis DecisionBasisFromNetwork, elevatorID int) DecisionBasisFromAssigner{
	return *localDecisionBasis 	//Placeholder
}


//Assignment based on deterministic HRA
func assignLocalOrders(decisionBasis DecisionBasisFromNetwork, elevatorID int) [NFloors][NButtons]bool {
	var orderTable [NFloors][NButtons]bool

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		fmt.Println("OS not supported")
		return orderTable
	}

	// Fyll inn input fra decisionBasis!
	input := HRAInput{
		HallRequests: decisionBasis.HallRequests, 		// ANTATT type er riktig
		States:       decisionBasis.ElevatorStates, 		// ANTATT type er riktig
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return orderTable
	}

	ret, err := exec.Command("../hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return orderTable
	}

	output := make(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return orderTable
	}

	// Tolke outputen for denne elevator, for elevID
	key := fmt.Sprintf("%d", elevatorID)
	assignedOrders, exists := output[key]
	if !exists {
		fmt.Printf("No orders assigned for elevator %d\n", elevatorID)
		return orderTable
	}

	// Her må du bygge opp orderTable fra assignedOrders,
	// antar formatet er: [ [floorIdx, [bool for up/down/internal]] ... ]
	for _, order := range assignedOrders {
		floorIdx := order[0]
		btnStates := order[1] // [true/false per knapp]
		for btnIdx, assigned := range btnStates {
			orderTable[floorIdx][btnIdx] = assigned
		}
	}
	return orderTable
}
