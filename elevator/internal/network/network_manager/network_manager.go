package networkmanager

import (
	//"fmt"
	"elevator/internal/config"
	"elevator/internal/network/broadcast"
	"elevator/internal/network/peers"
	"elevator/internal/types"
	"strconv"
	"time"
)

// mulig jeg bør legge inn funksjonalitet for error, feks om det ikke kommer inn en ren int og kanskje
// "elev-1", men skal ikke være noe problem om riktig melding kommer.
// func convertPeerUpdate(update peers.PeerUpdate) types.GlobalNodeRegistry {
// 	registry := types.GlobalNodeRegistry{}

// 	for _, peer := range update.Peers {
// 		id, err := strconv.Atoi(peer)
// 		if err != nil || id < 0 || id >= config.NElevators {
// 			//fmt.Printf("[NETWORK] Ugyldig peer-ID ignorert: %q\n", peer)
// 			continue
// 		}
// 		registry.Nodes = append(registry.Nodes, id)
// 	}

// 	// Ny heis
// 	if update.New != "" {
// 		id, err := strconv.Atoi(update.New)
// 		if err == nil && id >= 0 && id < config.NElevators {
// 			//fmt.Printf("[NETWORK] Ny heis oppdaget: ID=%d\n", id)
// 			registry.New = append(registry.New, id)
// 		}
// 	}

// 	// Heiser som har falt ut
// 	for _, lost := range update.Lost {
// 		id, err := strconv.Atoi(lost)
// 		if err != nil || id < 0 || id >= config.NElevators {
// 			continue
// 		}
// 		//fmt.Printf("[NETWORK] Heis mistet: ID=%d\n", id)
// 		registry.Lost = append(registry.Lost, id)
// 	}

// 	return registry
// }

// Skal man initialisere lastLocalState som noe her frem til første localState kommer?
// Kan vel evnetuelt sette lastLocalState.SenderID = selfID slik at den ignorerer frem til den får
// localState
func Run(
	selfID int,
	localState <-chan types.Message, // fra consensus
	incomingMessages chan<- types.Message, // til consensus
	nodeRegistry chan<- types.GlobalNodeRegistry, // til consensus
) {
	//fmt.Println("[NETWORK] Networkmanager called")
	ticker := time.NewTicker(config.BroadcastRate)
	defer ticker.Stop()
	var lastLocalState types.Message
	hasState := false
	msgTx := make(chan types.Message, config.BroadcastBufferSize)
	msgRx := make(chan types.Message, config.BroadcastBufferSize)
	peerUpdateCh := make(chan peers.PeerUpdate, config.BroadcastBufferSize)

	peerTxEnable := make(chan bool)

	go broadcast.Transmitter(config.BroadcastPort, msgTx)
	go broadcast.Receiver(config.BroadcastPort, msgRx)
	go peers.Transmitter(config.PeersPort, strconv.Itoa(selfID), peerTxEnable)
	go peers.Receiver(config.PeersPort, peerUpdateCh)

	peerTxEnable <- true

	for {
		select {
		case state := <-localState:
			//fmt.Printf("[NETWORK] Ny localState mottatt – SenderID=%d\n", selfID)
			state.SenderID = selfID
			lastLocalState = state
			hasState = true

		case <-ticker.C:
			if hasState {
				//fmt.Printf("[NETWORK] Sender state periodisk – SenderID=%d\n", lastLocalState.SenderID)
				msgTx <- lastLocalState // send periodisk
			}

		case msg := <-msgRx:
			if msg.SenderID == selfID {
				//fmt.Printf("[NETWORK] Dropper egen melding fra ID=%d\n", selfID)
				continue // dropp self
			}
			incomingMessages <- msg // videresend til consensus

		case update := <-peerUpdateCh:
			//fmt.Printf("[NETWORK] PeerUpdate – Aktive: %v | Ny: %q | Mistet: %v\n",
			//update.Peers, update.New, update.Lost)
			nodeRegistry <- peers.ConvertPeerUpdate(update)
		}
	}
}
