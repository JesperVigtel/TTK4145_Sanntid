package networkmanager

import (
	"elevator/internal/config"
	"elevator/internal/network/broadcast"
	"elevator/internal/network/peers"
	"elevator/internal/types"
	"strconv"
	"time"
)

func Run(
	selfID int,
	localState <-chan types.Message, // fra consensus
	peerMsg chan<- types.Message, // til consensus
	peerEvents chan<- types.GlobalNodeRegistry, // til consensus
) {
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
			state.SenderID = selfID
			lastLocalState = state
			hasState = true

		case <-ticker.C:
			if hasState {
				msgTx <- lastLocalState
			}

		case msg := <-msgRx:
			if msg.SenderID == selfID {
				continue
			}
			peerMsg <- msg // videresend til consensus

		case update := <-peerUpdateCh:
			peerEvents <- peers.ConvertPeerUpdate(update)
		}
	}
}
