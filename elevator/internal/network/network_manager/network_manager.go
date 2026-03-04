package networkmanager

import (
	"elevator/internal/config"
	"elevator/internal/network/peers"
	"elevator/internal/types"
	"strconv"
	"time"
)


func convertPeerUpdate(update peers.PeerUpdate) types.GlobalNodeRegistry {
    registry := types.GlobalNodeRegistry{}

	for _, peer := range update.Peers{
		id, _ := strconv.Atoi(peer)
		registry.Nodes = append(registry.Nodes, id)
	}
	
    // Ny heis
    if update.New != "" {
        id, _ := strconv.Atoi(update.New)
        registry.New = append(registry.New, id)
    }

    // Heiser som har falt ut
    for _, lost := range update.Lost {
        id, _ := strconv.Atoi(lost)
        registry.Lost = append(registry.Lost, id)
    }

    return registry
}
	

func RunNetworkManager(
    selfID          int,
    msgTx           chan<- types.Message,
    msgRx           <-chan types.Message,
    peerUpdateCh    <-chan peers.PeerUpdate,
    localState      <-chan types.Message,        // fra consensus
    incomingMessages chan<- types.Message,        // til consensus
    nodeRegistry    chan<- types.GlobalNodeRegistry, // til consensus
) {
    ticker := time.NewTicker(config.BroadcastRate)
    var lastLocalState types.Message

    for {
        select {
        case state := <-localState:
            lastLocalState = state

        case <-ticker.C:
            msgTx <- lastLocalState  // send periodisk

        case msg := <-msgRx:
            if msg.SenderID == selfID {
                continue  // dropp self
            }
            incomingMessages <- msg  // videresend til consensus

        case update := <-peerUpdateCh:
            nodeRegistry <- convertPeerUpdate(update)
        }
    }
}



