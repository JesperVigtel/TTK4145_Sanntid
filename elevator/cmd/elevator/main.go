package main

import (
	"elevator/internal/config"
	"elevator/internal/consensus"
	"elevator/internal/dispatch"
	"elevator/internal/localControll"
	"elevator/internal/types"
	"flag"
	"fmt"
	"os"
)

func main() {
	selfID := parseSelfID()

	// -- Channels --
	localControlEvents 		:= make(chan types.FromLocalToDM, 			config.ChannelBufferSize)
	newLocalOrders 			:= make(chan types.LocalOrderTable, 		config.ChannelBufferSize)
	localSystemState 		:= make(chan types.LocalSystemState, 		config.ChannelBufferSize)
	convergedSystemState 	:= make(chan types.ConvergedSystemState,	config.ChannelBufferSize)
	hallLightUpdates 		:= make(chan types.HallOrderTable, 			config.ChannelBufferSize)
	incomingMessages 		:= make(chan types.Message, 				config.ChannelBufferSize)
	outgoingMessages 		:= make(chan types.Message, 				config.ChannelBufferSize)
	nodeRegistryEvents 		:= make(chan types.GlobalNodeRegistry, 		config.ChannelBufferSize)

	// -- Goroutines --
	//hardware.Init(config.Addr, config.NFloors) //Shuld be moved inside a localContoll run
	//go localControl.Run(newLocalOrders, localControlEvents)

	go dispatch.Run(
		newLocalOrders,
		localSystemState,
		hallLightUpdates,
		localControlEvents,
		convergedSystemState,
		selfID,
	)

	go consensus.Run(
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
