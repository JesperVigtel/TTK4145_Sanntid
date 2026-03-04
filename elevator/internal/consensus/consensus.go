package consensus

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
)

func RunConsensusManager(
	incomingMessages  	<-chan Message,
	nodeRegistryEvents 	<-chan GlobalNodeRegistry,
	localSystemState  	<-chan LocalSystemState,
	agreedSystemState 	chan<- AgreedSystemState,
	elevatorID        	int,
) {
	var (
		globalHallOrders  [NElevators]HallOrderTable
		globalElevStates  [NElevators]HRAElevState
		nodeAliveStatus   [NElevators]bool
		nodeConverged     [NElevators]bool
	)

	globalHallOrders = initGlobalHallOrders()

	for {
		select {

		case nodeRegistry := <-nodeRegistryEvents:
			nodeAliveStatus, globalHallOrders = applyNodeRegistryChange(nodeRegistry, nodeAliveStatus, globalHallOrders)

		case msg := <-incomingMessages:
			nodeConverged[msg.SenderID]        = nodeViewMatchesOurs(msg, globalHallOrders, globalElevStates)
			globalElevStates[msg.SenderID]     = msg.ElevatorList[msg.SenderID]
			globalHallOrders[msg.SenderID]     = msg.HallOrderList
			nodeAliveStatus[msg.SenderID]      = msg.AliveStatus
			globalHallOrders                   = stepAllOrderStates(globalHallOrders, elevatorID)

			if allNodesConverged(nodeConverged, nodeAliveStatus, elevatorID) {
				nodeConverged = resetNodeConverged()
				agreedSystemState <- AgreedSystemState{
					AliveList:      nodeAliveStatus,
					ElevatorList:   globalElevStates,
					HallOrderTable: globalHallOrders,
				}
			}

		case state := <-localSystemState:
			globalHallOrders[elevatorID]  = state.HallRequests
			globalElevStates[elevatorID]  = state.ElevatorState
			nodeAliveStatus[elevatorID]   = state.AliveStatus
		}
	}
}