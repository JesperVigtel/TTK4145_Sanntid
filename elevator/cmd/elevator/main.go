package main

//https://prod.liveshare.vsengsaas.visualstudio.com/join?653BD7ECD9B27DA732A843D4219CA6EA2563

import (
	"elevator/internal/config"
	"elevator/internal/consensus"
	"elevator/internal/dispatch"
	"elevator/internal/lights"
	"elevator/internal/localControl"
	"elevator/internal/network/network_manager"
	"elevator/internal/network/peers"
	"elevator/internal/network/utility_network"
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

	// -- Channels --

	// -- Channels for LocalControl --
	localOrders := make(chan types.LocalOrderTable, config.ChannelBufferSize)
	elevatorEvents := make(chan types.ElevatorEvents, config.ChannelBufferSize)
	localLightUpdate := make(chan types.LocalLightUpdate, config.ChannelBufferSize)

	// -- Channels for Decision
	localSystemState := make(chan types.LocalSystemState, config.ChannelBufferSize)
	convergedSystemState := make(chan types.ConvergedSystemState, config.ChannelBufferSize)
	lightUpdate := make(chan types.ButtonLightUpdate, config.ChannelBufferSize)

	// -- Channels for network
	peerMsg := make(chan types.Message, config.ChannelBufferSize)
	broadcast := make(chan types.Message, config.ChannelBufferSize)
	peerEvents := make(chan types.GlobalNodeRegistry, config.ChannelBufferSize)
	msgTx := make(chan types.Message, config.BroadcastBufferSize)
	msgRx := make(chan types.Message, config.BroadcastBufferSize)
	peerUpdateCh := make(chan peers.PeerUpdate, config.BroadcastBufferSize)

	// -- Goroutines --

	go localControl.Run(
		elevAddr,
		localOrders,
		elevatorEvents,
		localLightUpdate,
	)

	go lights.Run(	//Potentially just add to localcontrol, and lihtupdate inn ?
		//localLightUpdate,
		lightUpdate)

	go dispatch.Run(
		localOrders,
		localSystemState,
		lightUpdate,
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

	go utilitynetwork.InitNetwork(	//Init funcitons should live inside their module
		strconv.Itoa(selfID),
		msgTx,
		msgRx,
		peerUpdateCh)

	go networkmanager.Run(		//Naming could be just network or continuity?
		selfID,
		msgTx,
		msgRx,
		peerUpdateCh,
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