package network
import (
	"elevator/internal/config"
)


//Sender og Reciever blir 2 gorutiner i network.go og vil jobbe i bakgrunnen
//disse vil håndtere all nettverkskommunikasjon
//func sender()
// Jobben til sender er å lytte til kanalen i 
func Sender()


func Reciever()




//func reciever()
//Lytter hele tiden paa kanal fra heisene og når det kommer endringer fra heisene så
//sendes det gjennom sender


//func heartbeat()
//lytter hele tiden og "nullstiller" en slags heartbeat-timer hver gang den får inn en heartbeat fra
//kanalen, dersom det går for lang tid så vil det 

