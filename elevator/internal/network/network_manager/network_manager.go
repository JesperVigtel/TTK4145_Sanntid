package networkmanager

import (
	"elevator/internal/config"
	"elevator/internal/network/peers"
	"elevator/internal/types"
	"strconv"
	"time"
)
// Ansvar (NetworkManager)
// Oversette peers.PeerUpdate → types.NetworkNodeRegistry (og sende videre til consensus)
// Motta “lokal snapshot som skal broadcastes” (fra consensus eller DM) og:
// pakke i types.Message
// legge på Seq, BootID
// sende periodisk (tick/gossip)
// Motta meldinger fra UDP (msgRx) og:
// droppe self
// droppe stale/out-of-order per avsender
// sende videre på incomingMessages til consensus
// Viktig: Consensus trenger ikke vite om UDP. Den trenger bare:
// incomingMessages <-chan types.Message
// peerEvents <-chan types.NetworkNodeRegistry
// Det er akkurat slik RunConsensusManager allerede er skrevet.

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



