package network

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
	localSystemState <-chan types.Message,
	message chan<- types.Message,
	nodeRegistry chan<- types.GlobalNodeRegistry,
) {
	ticker := time.NewTicker(config.BroadcastRate)
	defer ticker.Stop()

	var lastLocalSystemState types.Message
	hasState := false

	messageTransmitChan := make(chan types.Message, config.BroadcastBufferSize)
	messageReceiveChan := make(chan types.Message, config.BroadcastBufferSize)
	peerUpdateChannel := make(chan peers.PeerUpdate, config.BroadcastBufferSize)
	peerTxEnable := make(chan bool)

	go broadcast.Transmitter(config.BroadcastPort, messageTransmitChan)
	go broadcast.Receiver(config.BroadcastPort, messageReceiveChan)
	go peers.Transmitter(config.PeersPort, strconv.Itoa(selfID), peerTxEnable)
	go peers.Receiver(config.PeersPort, peerUpdateChannel)
	peerTxEnable <- true

	for {
		select {
		case state := <-localSystemState:
			state.SenderID = selfID
			lastLocalSystemState = state
			hasState = true

		case <-ticker.C:
			if hasState {
				messageTransmitChan <- lastLocalSystemState
			}

		case msg := <-messageReceiveChan:
			if msg.SenderID == selfID {
				continue
			}
			message <- msg

		case update := <-peerUpdateChannel:
			nodeRegistry <- peers.ConvertPeerUpdate(update)
		}
	}
}
