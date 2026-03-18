package main

import (
	"elevator/internal/config"
	"elevator/internal/consensus"
	"elevator/internal/dispatch"
	"elevator/internal/localControl"
	"elevator/internal/network"
	"elevator/internal/types"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	selfID, elevAddr := parseArgs()
	// selfID := parseArgs()
	// elevAddr := "localhost:15657"

	// -- Channels for LocalControl --
	localOrders := make(chan types.LocalOrderTable, config.ChannelBufferSize)
	orderLightUpdates := make(chan types.OrderTable, config.ChannelBufferSize)
	elevatorEvents := make(chan types.ElevatorEvents, config.ChannelBufferSize)

	// -- Channels for Decision
	localSystemState := make(chan types.LocalSystemState, config.ChannelBufferSize)
	convergedSystemState := make(chan types.ConvergedSystemState, config.ChannelBufferSize)

	// -- Channels for network
	peerMsg := make(chan types.Message, config.ChannelBufferSize)
	broadcast := make(chan types.Message, config.ChannelBufferSize)
	peerEvents := make(chan types.GlobalNodeRegistry, config.ChannelBufferSize)

	// -- Goroutines --

	go localControl.Run(
		elevAddr,
		localOrders,
		orderLightUpdates,
		elevatorEvents,
	)

	go dispatch.Run(
		localOrders,
		localSystemState,
		orderLightUpdates,
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

func parseArgs() (int, string) {
	id := flag.Int("id", 0, "Elevator node ID (0-2)")
	port := flag.Int("port", 15657, "TCP port for elevator hardware/simulator")
	addr := flag.String("addr", "", "Full TCP address for elevator hardware/simulator (overrides --port)")
	flag.Parse()

	if *id < 0 || *id >= config.NElevators {
		fmt.Fprintf(os.Stderr, "invalid --id %d: must be 0..%d\n", *id, config.NElevators-1)
		os.Exit(1)
	}

	elevAddr := *addr
	if elevAddr == "" {
		if *port <= 0 || *port > 65535 {
			fmt.Fprintf(os.Stderr, "invalid --port %d: must be 1..65535\n", *port)
			os.Exit(1)
		}
		elevAddr = net.JoinHostPort("localhost", strconv.Itoa(*port))
	}

	return *id, elevAddr
}

// func parseArgs() int {
// 	var nodeID int
// 	flag.IntVar(&nodeID, "id", 0, "Node ID")
// 	flag.Parse()
// 	return nodeID
// }
