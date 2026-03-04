package main

// Entry point. Sets up channels and starts goroutines.
// Contains no logic — only wires modules together via channels.

import (
	"elevator/internal/config"
	"elevator/internal/consensus"
	"elevator/internal/dispatch"
	"elevator/internal/localControll/hardware"
	"elevator/internal/types"
	"flag"
	"fmt"
	"os"
)

func main() {
	selfID := parseSelfID()

	// -- Channels --

	localControlEvents 		:= make(chan types.FromLocalToDM, 			config.ChannelBufferSize)
	newLocalOrders 			:= make(chan types.CabOrderTable, 			config.ChannelBufferSize)
	localSystemState 		:= make(chan types.LocalSystemState, 		config.ChannelBufferSize)
	convergedSystemState 	:= make(chan types.ConvergedSystemState,	config.ChannelBufferSize)
	hallLightUpdates 		:= make(chan types.HallOrderTable, 			config.ChannelBufferSize)
	incomingMessages 		:= make(chan types.Message, 				config.ChannelBufferSize)
	outgoingMessages 		:= make(chan types.Message, 				config.ChannelBufferSize)
	nodeRegistryEvents 		:= make(chan types.GlobalNodeRegistry, 		config.ChannelBufferSize)

	// -- Goroutines --
	hardware.Init(config.Addr, config.NFloors) //Shuld be moved inside a localContoll run
	//go localControll.Run(newLocalOrders, localControlEvents)

	go dispatch.RunDispatch(
		newLocalOrders,
		localSystemState,
		hallLightUpdates,
		localControlEvents,
		convergedSystemState,
		selfID,
	)

	go consensus.RunConsensus(
		incomingMessages,
		outgoingMessages,
		nodeRegistryEvents,
		localSystemState,
		convergedSystemState,
		selfID,
	)

	// Rest of go rutines

	select {}
}

func parseSelfID() int {
	id := flag.Int("id", 0, "Elevator node ID (0-2)")
	flag.Parse()
	if *id < 0 || *id >= config.NElevators {
		fmt.Fprintf(os.Stderr, "invalid --id %d: must be 0..%d\n", *id, config.NElevators-1)
		os.Exit(1)
	}
	return *id
}
