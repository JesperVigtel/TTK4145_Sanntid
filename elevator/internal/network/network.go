package network

import (
	"elevator/internal/config"
)

//func Network()
//funksjonen Network() får tilsent worldview gjennom en kanal som (From_assigner_to_network)
//Network() sender hele tiden
//Opprette 2 go-rutiner som kjører parallelt, reciever og sender slik at det blir
//sendt og motatt hele tiden.








// Høy-nivå nettverkslogikk (initialisering av noder).
// Bruker UDP broadcast for å holde alle heiser oppdatert på hverandres tilstand.
// Skal tåle pakketap og noder som faller ut.

// TODO: Implementer nettverksinitialisering
