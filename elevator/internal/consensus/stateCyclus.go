package consensus 

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

// RunConsensusManager is the main goroutine for the consensus module.
//
// Responsibility:
//   - Receives raw broadcasted states from all elevators (via network)
//   - Tracks which elevators are alive
//   - Runs the cyclic counter on the hall order table to advance order states
//   - Tracks acknowledgments (ackMap) to determine when all alive peers agree
//   - When consensus is reached → emits an agreed DecisionBasisFromNetwork upstream to DecisionMaker
//
// It does NOT:
//   - Know about button press semantics
//   - Know about HRA or order assignment
//   - Do any UDP send/receive (that belongs to network module)
func RunConsensusManager(
	incomingMessages   <-chan Message,                  // Raw messages from network module (other elevators)
	peerEvents         <-chan NetworkNodeRegistry,       // Alive/lost peer events from network module
	localBasisUpdates  <-chan DecisionBasisFromAssigner, // Local state published by DecisionMaker
	agreedWorldview    chan<- DecisionBasisFromNetwork,  // Confirmed consensus sent to DecisionMaker
	nodeID             int,
) {
	var (
		hallOrderTable [NElevators]HallOrderTable        // Each elevator's per-floor order states
		elevatorList   [NElevators]HRAElevState          // Each elevator's physical state
		aliveList      [NElevators]bool                  // Which elevators are currently online
		ackMap         [NElevators]bool                  // Which peers agree with our current hallOrderTable
	)

	// TODO: Initialize hallOrderTable to Initial state for all elevators
	// hallOrderTable = initHallOrderTable()

	for {
		select {

		// --- Event 1: Peer alive/lost update from the network module ---
		case reg := <-peerEvents:
			// TODO: For each lost peer: mark aliveList[id] = false, reset their hallOrderTable row
			// TODO: For each new peer:  mark aliveList[id] = true
			// NOTE: If nodeID itself is in reg.New → we have just come online
			_ = reg

		// --- Event 2: A message broadcasted from another elevator arrives ---
		case msg := <-incomingMessages:
			// Step 1: Check if their view matches ours → set ackMap[msg.SenderID]
			// TODO: ackMap[msg.SenderID] = reflect.DeepEqual(hallOrderTable, msg.HallOrderList) && ...

			// Step 2: Merge their state into our local view
			// TODO: elevatorList[msg.SenderID] = msg.ElevatorList[msg.SenderID]
			// TODO: hallOrderTable[msg.SenderID] = msg.HallOrderList[msg.SenderID]
			// TODO: aliveList[msg.SenderID]    = msg.AliveStatus

			// Step 3: Run the cyclic counter to attempt order state transitions for our own row
			// TODO: hallOrderTable = advanceOrderStates(hallOrderTable, nodeID)
			// This is where Gustav's cyclicCounter logic lives.
			// It reads peer states (hallOrderTable[others]) and advances hallOrderTable[nodeID]

			// Step 4: Check if all alive peers now agree → emit agreed worldview upstream
			// TODO: if allAcknowledged(ackMap, aliveList, nodeID) {
			//     reset ackMap for next round
			//     agreedWorldview <- DecisionBasisFromNetwork{...}
			// }
			_ = msg

		// --- Event 3: Our own local state changed (button press, elevator state update) ---
		case localBasis := <-localBasisUpdates:
			// Update our own row in hallOrderTable and elevatorList from the local decision basis
			// TODO: hallOrderTable[nodeID] = localBasis.HallRequests
			// TODO: elevatorList[nodeID]   = localBasis.LocalState[nodeID]  (convert as needed)
			// TODO: aliveList[nodeID]      = localBasis.AliveStatus
			// NOTE: We do NOT run the cyclic counter here — only when we receive a peer message,
			//       because consensus is about agreeing with OTHERS, not with ourselves.
			_ = localBasis
		}
	}
}

// advanceOrderStates runs the cyclic counter logic for nodeID's row.
// It reads all peers' order states and attempts valid transitions.
// This is a pure function — no side effects, takes the full table and returns an updated table.
//
// State machine per (floor, button):
//
//	Standby → (any peer sees ButtonPressed)  → Pending
//	Pending → (all peers see Pending/Assigned) → Assigned
//	Assigned → (any peer sees OrderComplete)   → Complete
//	Complete → (all peers see Complete/Standby) → Standby
//
// If an illegal combination is detected (states diverged), reset to Initial.
func advanceOrderStates(
	orders [NElevators]HallOrderTable,
	myID   int,
) [NElevators]HallOrderTable {
	// TODO: Port Gustav's cyclicCounter logic here directly.
	// Ref: gustavlokna/Elevator TTK4145/Project/network/cylickCounter.go
	return orders
}

// allAcknowledged returns true when every alive peer (excluding self) has acknowledged
// that their view of the hall order table matches ours.
func allAcknowledged(ackMap [NElevators]bool, aliveList [NElevators]bool, myID int) bool {
	for id := 0; id < NElevators; id++ {
		if id == myID {
			continue
		}
		if aliveList[id] && !ackMap[id] {
			return false
		}
	}
	return true
}

// initHallOrderTable initializes all order states to the Initial/Standby state.
func initHallOrderTable() [NElevators]HallOrderTable {
	var table [NElevators]HallOrderTable
	for elev := range table {
		for floor := range table[elev] {
			for btn := range table[elev][floor] {
				table[elev][floor][btn] = OrderStandby // or Initial if you add that state
			}
		}
	}
	return table
}