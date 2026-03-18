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

	assignedOrderChan := make(chan types.AssignedOrderTable, config.ChannelBufferSize)
	hallLightChan := make(chan types.HallLampTable, config.ChannelBufferSize)
	elevatorEventsChan := make(chan types.ElevatorEvents, config.ChannelBufferSize)

	localSystemStateChan := make(chan types.LocalSystemState, config.ChannelBufferSize)
	convergedSystemStateChan := make(chan types.ConvergedSystemState, config.ChannelBufferSize)

	peerMsgChan := make(chan types.Message, config.ChannelBufferSize)
	broadcastChan := make(chan types.Message, config.ChannelBufferSize)
	peerEventsChan := make(chan types.GlobalNodeRegistry, config.ChannelBufferSize)

	go localControl.Run(
		elevAddr,
		assignedOrderChan,
		hallLightChan,
		elevatorEventsChan,
	)

	go dispatch.Run(
		assignedOrderChan,
		localSystemStateChan,
		hallLightChan,
		elevatorEventsChan,
		convergedSystemStateChan,
		selfID,
	)

	go consensus.Run(
		peerMsgChan,
		broadcastChan,
		peerEventsChan,
		localSystemStateChan,
		convergedSystemStateChan,
		selfID,
	)

	go network.Run(
		selfID,
		broadcastChan,
		peerMsgChan,
		peerEventsChan,
	)

	select {}
}
