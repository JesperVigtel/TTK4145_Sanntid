package networkmanager

import (
	"elevator/internal/config"
	"elevator/internal/network/peers"
	"elevator/internal/types"
	"strconv"
	"time"
)

// mulig jeg bør legge inn funksjonalitet for error, feks om det ikke kommer inn en ren int og kanskje
// "elev-1", men skal ikke være noe problem om riktig melding kommer.
func convertPeerUpdate(update peers.PeerUpdate) types.GlobalNodeRegistry {
	registry := types.GlobalNodeRegistry{}

	for _, peer := range update.Peers {
		id, err := strconv.Atoi(peer)
		if err != nil || id < 0 || id >= config.NElevators {
			continue
		}
		registry.Nodes = append(registry.Nodes, id)
	}

	// Ny heis
	if update.New != "" {
		id, err := strconv.Atoi(update.New)
		if err == nil && id >= 0 && id < config.NElevators {
			registry.New = append(registry.New, id)
		}
	}

	// Heiser som har falt ut
	for _, lost := range update.Lost {
		id, err := strconv.Atoi(lost)
		if err != nil || id < 0 || id >= config.NElevators {
			continue
		}
		registry.Lost = append(registry.Lost, id)
	}

	return registry
}

// Skal man initialisere lastLocalState som noe her frem til første localState kommer?
// Kan vel evnetuelt sette lastLocalState.SenderID = selfID slik at den ignorerer frem til den får
// localState
func Run(
	selfID 				int,
	msgTx 				chan<- types.Message,
	msgRx 				<-chan types.Message,
	peerUpdateCh 		<-chan peers.PeerUpdate,
	localState 			<-chan types.Message, // fra consensus
	incomingMessages 	chan<- types.Message, // til consensus
	nodeRegistry 		chan<- types.GlobalNodeRegistry, // til consensus
) {
	ticker := time.NewTicker(config.BroadcastRate)
	var lastLocalState types.Message
	hasState := false

	for {
		select {
		case state := <-localState:
			state.SenderID = selfID
			lastLocalState = state
			hasState = true

		case <-ticker.C:
			if hasState {
				msgTx <- lastLocalState // send periodisk
			}

		case msg := <-msgRx:
			if msg.SenderID == selfID {
				continue // dropp self
			}
			incomingMessages <- msg // videresend til consensus

		case update := <-peerUpdateCh:
			nodeRegistry <- convertPeerUpdate(update)
		}
	}
}
