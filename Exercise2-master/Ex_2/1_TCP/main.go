package main

//https://prod.liveshare.vsengsaas.visualstudio.com/join?DBB46B5401E901B893FF5F4B2FD2B0559AB7

import (
	"fmt"
	"net"
	"time"
)

const (
	SERVER_IP         = "10.100.23.11"
	LOCAL_IP          = "10.100.23.17"
	LOCAL_LISTEN_PORT = 33546
)

func receiver(conn net.Conn) {
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return // Avslutt når tilkobling lukkes
		}
		fmt.Printf("Mottatt: %s\n", string(buffer[:n]))
	}
}

/*Reciver leser av informasjonen fra serveren*/


func sender(conn net.Conn, messages []string) {
	for _, msg := range messages {
		conn.Write([]byte(msg + "\x00"))
		fmt.Printf("Sendt: %s\n", msg)
		time.Sleep(1 * time.Second)
	}
}

/*sender informasjon til serveren*/

func main() {

	conn1, err := net.Dial("tcp", SERVER_IP+":34933")
	if err != nil {
		fmt.Println("Feil ved tilkobling til server:", err)
		return
	}
	go receiver(conn1)
	sender(conn1, []string{"Pussycats", "mjaumjau", "Sinna"})
	time.Sleep(2 * time.Second)
	conn1.Close()

	/*Del 1: Direkte tilkobling til serveren. tar ibruk go-rutine for at meldinger kan mottas samtidig som de sendes*/

	// Reverse connection

	listener, err := net.Listen("tcp", ":"+fmt.Sprint(LOCAL_LISTEN_PORT))
	if err != nil {
		fmt.Println("Feil ved oppstart av server:", err)
		return
	}

	conn2, err := net.Dial("tcp", SERVER_IP+":33546")
	if err != nil {
		fmt.Println("Feil ved instruksjon til server:", err)
		listener.Close()
		return
	}

	conn2.Write([]byte(fmt.Sprintf("Connect to: %s:%d\x00", LOCAL_IP, LOCAL_LISTEN_PORT)))
	conn2.Close()

	connServer, err := listener.Accept()
	if err != nil {
		fmt.Println("Feil ved akseptering av tilkobling:", err)
		listener.Close()
		return
	}
	go receiver(connServer)
	sender(connServer, []string{"Svar fra reverse connection"})
	time.Sleep(2 * time.Second)

	connServer.Close()
	listener.Close()
}
/*Del 2 forteller serveren at den skal koble seg tilbake til oss. Og hører på hva serveren sender tilbake*/
