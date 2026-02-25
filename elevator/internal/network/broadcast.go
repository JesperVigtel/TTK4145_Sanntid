package network
import (
	. "elevator/internal/config"
	. "elevator/internal/types"
	"network/conn"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
)


//Sender og Reciever blir 2 gorutiner i network.go og vil jobbe i bakgrunnen
//disse vil håndtere all nettverkskommunikasjon
//func sender()
// Jobben til sender er å lytte til kanalen i 
func Transmitter(port int, messageOut <-chan Message){ //velger en enkel meldingstype som blir sendt hele tiden fremfor en "chans ... interface{}"" dette føles ryddigere
	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
	for message := range messageOut{
		jsonbytes, _ := json.Marshal(message) // konverterer Message-structen til JSON-bytes
		if len(jsonbytes) > BroadcastBufferSize{
			panic(fmt.Sprintf("Tried to send a message longer than the buffer size."))
		}
		conn.WriteTo(jsonbytes, addr)
	}
}


func Reciever()




//func reciever()
//Lytter hele tiden paa kanal fra heisene og når det kommer endringer fra heisene så
//sendes det gjennom sender


//func heartbeat()
//lytter hele tiden og "nullstiller" en slags heartbeat-timer hver gang den får inn en heartbeat fra
//kanalen, dersom det går for lang tid så vil det 

