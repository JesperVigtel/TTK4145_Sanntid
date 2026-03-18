package main

import (
	"elevator/internal/config"
	"elevator/internal/consensus"
	"elevator/internal/dispatch"
	"elevator/internal/localControl"
	"elevator/internal/network"
	"elevator/internal/types"
)

func main() {
	selfID, elevAddr := config.ParseArgs()

	assignedOrderUpdates := make(chan types.AssignedOrderTable, config.ChannelBufferSize)
	hallLightUpdates := make(chan types.HallLampTable, config.ChannelBufferSize)
	elevatorEvents := make(chan types.ElevatorEvents, config.ChannelBufferSize)

	localSystemState := make(chan types.LocalSystemState, config.ChannelBufferSize)
	convergedSystemState := make(chan types.ConvergedSystemState, config.ChannelBufferSize)

	peerMsg := make(chan types.Message, config.ChannelBufferSize)
	broadcast := make(chan types.Message, config.ChannelBufferSize)
	peerEvents := make(chan types.GlobalNodeRegistry, config.ChannelBufferSize)

	go localControl.Run(
		elevAddr,
		assignedOrderUpdates,
		hallLightUpdates,
		elevatorEvents,
	)

	go dispatch.Run(
		assignedOrderUpdates,
		localSystemState,
		hallLightUpdates,
		elevatorEvents,
		convergedSystemState,
		selfID,
	)

	go consensus.Run(
		peerMsg,
		broadcast,
		peerEvents,
		localSystemState,
		convergedSystemState,
		selfID,
	)

	go network.Run(
		selfID,
		broadcast,
		peerMsg,
		peerEvents,
	)

	select {}
}
