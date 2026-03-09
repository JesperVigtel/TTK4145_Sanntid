package utilitynetwork

import (
	"fmt"
	// "strconv"
	// "elevator/internal/config"
	. "elevator/internal/config"
	"elevator/internal/network/broadcast"
	"elevator/internal/network/peers"
	"elevator/internal/types"
)

func InitNetwork(id string, msgTx chan types.Message, msgRx chan types.Message, peerUpdate chan peers.PeerUpdate) {
	fmt.Println("[DEBUG] InitNetwork called sucsessfy")
	go broadcast.Transmitter(BroadcastPort, msgTx)
	go broadcast.Receiver(BroadcastPort, msgRx)
	go peers.Transmitter(PeersPort, id, make(chan bool))
	go peers.Receiver(PeersPort, peerUpdate)
}

// func PeerUpdates(peerUpdate <-chan peers.PeerUpdate, aliveList *[config.NElevators]bool){
//     for update := range peerUpdate {
//         // Ny heis koblet til
//         if update.New != "" {
//             id, _ := strconv.Atoi(update.New)
//             aliveList[id] = true
//         }
//         // Heis falt ut
//         for _, lost := range update.Lost {
//             id, _ := strconv.Atoi(lost)
//             aliveList[id] = false
//         }
//     }
// }
