package consensus

import (
	. "elevator/internal/config"
	. "elevator/internal/types"
	"reflect"
)

func initGlobalHallOrders() [NElevators]HallOrderTable {
	var table [NElevators]HallOrderTable
	for id := range table {
		table[id] = resetHallOrders()
	}
	return table
}

func resetHallOrders() HallOrderTable {
	var table HallOrderTable
	for floor := range table {
		for btn := range table[floor] {
			table[floor][btn] = OrderStandby
		}
	}
	return table
}

func applyNodeRegistryChange(
	nodeRegistry     GlobalNodeRegistry,
	nodeAliveStatus  [NElevators]bool,
	globalHallOrders [NElevators]HallOrderTable,
) ([NElevators]bool, [NElevators]HallOrderTable) {
	for _, id := range nodeRegistry.Lost {
		nodeAliveStatus[id] = false
		globalHallOrders[id]  = resetHallOrders()
	}
	for _, id := range nodeRegistry.New {
		nodeAliveStatus[id] = true
	}
	return nodeAliveStatus, globalHallOrders
}

func nodeViewMatchesOurs(
	msg              Message,
	globalHallOrders [NElevators]HallOrderTable,
	globalElevStates [NElevators]HRAElevState,
) bool {
	return reflect.DeepEqual(globalHallOrders, msg.HallOrderList) &&
		reflect.DeepEqual(globalElevStates, msg.ElevatorList)
}

func allNodesConverged(
	nodeConverged   [NElevators]bool,
	nodeAliveStatus [NElevators]bool,
	nodeID          int,
) bool {
	for id := 0; id < NElevators; id++ {
		if id == nodeID {
			continue
		}
		if nodeAliveStatus[id] && !nodeConverged[id] {
			return false
		}
	}
	return true
}

func resetNodeConverged() [NElevators]bool {
	return [NElevators]bool{}
}