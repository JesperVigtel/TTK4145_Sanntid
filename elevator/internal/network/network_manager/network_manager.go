package networkmanager


// Ansvar (NetworkManager):

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

