package config

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	NFloors             = 4
	NButtons            = 3
	NElevators          = 3
	PollRateMS          = 20 * time.Millisecond
	BroadcastRate       = 50 * time.Millisecond
	DoorOpenTime        = 3 * time.Second
	MotorTimeout        = 4 * time.Second
	MotorRecoveryTime   = 3 * time.Second
	BroadcastBufferSize = 4 * 1024
	ChannelBufferSize   = 100
	HeartbeatInterval   = 50 * time.Millisecond
	HeartbeatTimeout    = 3000 * time.Millisecond
	BroadcastPort       = 13333
	PeersPort           = 13334
)

func ParseArgs() (int, string) {
	id := flag.Int("id", 0, "Elevator node ID (0-2)")
	port := flag.Int("port", 15657, "TCP port for elevator hardware/simulator")
	addr := flag.String("addr", "", "Full TCP address for elevator hardware/simulator (overrides --port)")
	flag.Parse()

	if *id < 0 || *id >= NElevators {
		fmt.Fprintf(os.Stderr, "invalid --id %d: must be 0..%d\n", *id, NElevators-1)
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
