package peers

import (
	"elevator/internal/config"
	"elevator/internal/network/conn"
	"elevator/internal/types"
	"strconv"
	"fmt"
	"net"
	"sort"
	"time"
)

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

// const interval = 15 * time.Millisecond
// const timeout = 500 * time.Millisecond
const interval = config.HeartbeatInterval
const timeout = config.HeartbeatTimeout

//Hvor lenge vil vi vente før vi markerer noe som lost?
//Hvor ofte vil vi sende ut?

func Transmitter(port int, id string, transmitEnable <-chan bool) {

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		}
		if enable {
			conn.WriteTo([]byte(id), addr)
		}
	}
}

func Receiver(port int, peerUpdateCh chan<- PeerUpdate) {

	var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)
	// Emit one snapshot even when no peers are present so consensus can
	// distinguish "no peers detected" from "peer discovery has not run yet".
	sentInitialSnapshot := false

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		id := string(buf[:n])

		// Adding new connection
		p.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				p.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}

		// Removing dead connection
		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Since(v) > timeout { //if time.Now().Sub(v) > timeout
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Sending update
		if updated || !sentInitialSnapshot {
			p.Peers = make([]string, 0, len(lastSeen))

			for k := range lastSeen { //for k, _ := range lastSeen
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)
			peerUpdateCh <- p
			sentInitialSnapshot = true
		}
	}
}


func ConvertPeerUpdate(update PeerUpdate) types.GlobalNodeRegistry {
	registry := types.GlobalNodeRegistry{}

	for _, peer := range update.Peers {
		id, err := strconv.Atoi(peer)
		if err != nil || id < 0 || id >= config.NElevators {
			//fmt.Printf("[NETWORK] Ugyldig peer-ID ignorert: %q\n", peer)
			continue
		}
		registry.Nodes = append(registry.Nodes, id)
	}

	// Ny heis
	if update.New != "" {
		id, err := strconv.Atoi(update.New)
		if err == nil && id >= 0 && id < config.NElevators {
			//fmt.Printf("[NETWORK] Ny heis oppdaget: ID=%d\n", id)
			registry.New = append(registry.New, id)
		}
	}

	// Heiser som har falt ut
	for _, lost := range update.Lost {
		id, err := strconv.Atoi(lost)
		if err != nil || id < 0 || id >= config.NElevators {
			continue
		}
		//fmt.Printf("[NETWORK] Heis mistet: ID=%d\n", id)
		registry.Lost = append(registry.Lost, id)
	}

	return registry
}