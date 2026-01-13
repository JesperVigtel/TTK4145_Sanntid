package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

const (
	BROADCAST_PORT = 30000 // Port for å motta server IP
	SERVER_PORT    = 20000 // Base port (legg til arbeidsstasjonsnummer)
	BUFFER_SIZE    = 1024
)

// Lytter etter server IP på broadcast port 30000
func listenForServerIP() string {
	addr := net.UDPAddr{
		Port: BROADCAST_PORT,
		IP:   net.IPv4zero, // Lytter på alle interfaces
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Feil ved opprettelse av UDP listener: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Lytter etter server IP på port %d...\n", BROADCAST_PORT)

	buffer := make([]byte, BUFFER_SIZE)
	n, remoteAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Printf("Feil ved mottak av data: %v\n", err)
		os.Exit(1)
	}

	serverIP := remoteAddr.IP.String()
	message := string(buffer[:n])

	fmt.Printf("Mottok melding fra server: %s\n", message)
	fmt.Printf("Server IP: %s\n", serverIP)

	return serverIP
}

// Sender meldinger til serveren og mottar svar
func communicateWithServer(serverIP string, workstationNumber int) {
	port := SERVER_PORT + workstationNumber
	serverAddr := fmt.Sprintf("%s:%d", serverIP, port)

	fmt.Printf("\nKobler til server på %s\n", serverAddr)

	// Opprett UDP-adresse for serveren
	raddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		fmt.Printf("Feil ved resolving av adresse: %v\n", err)
		return
	}

	// Opprett socket for sending
	sendConn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		fmt.Printf("Feil ved opprettelse av sender socket: %v\n", err)
		return
	}
	defer sendConn.Close()

	// Opprett socket for mottak (lytter på samme port)
	laddr := net.UDPAddr{
		Port: port,
		IP:   net.IPv4zero,
	}
	recvConn, err := net.ListenUDP("udp", &laddr)
	if err != nil {
		fmt.Printf("Feil ved opprettelse av receiver socket: %v\n", err)
		return
	}
	defer recvConn.Close()

	// Goroutine for å motta meldinger
	go func() {
		buffer := make([]byte, BUFFER_SIZE)
		for {
			n, remoteAddr, err := recvConn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Printf("Feil ved mottak: %v\n", err)
				return
			}

			message := string(buffer[:n])
			// Server prefikser svar med "You said: "
			if remoteAddr.IP.String() == serverIP {
				fmt.Printf("Svar fra server: %s\n", message)
			} else {
				fmt.Printf("Melding fra %s: %s\n", remoteAddr.IP, message)
			}
		}
	}()

	// Send meldinger til serveren
	fmt.Println("\nSkriver 'quit' for å avslutte")
	messageCount := 0

	for {
		messageCount++
		message := fmt.Sprintf("Melding #%d fra Go-klient", messageCount)

		_, err := sendConn.Write([]byte(message))
		if err != nil {
			fmt.Printf("Feil ved sending: %v\n", err)
			return
		}

		fmt.Printf("Sendte: %s\n", message)

		// Vent litt mellom meldinger for å være snill mot nettverket
		time.Sleep(2 * time.Second)

		// Send bare noen få testmeldinger
		if messageCount >= 5 {
			fmt.Println("\nSendt 5 testmeldinger. Venter på svar...")
			time.Sleep(3 * time.Second)
			break
		}
	}
}

func main() {
	fmt.Println("=== UDP Klient for TTK4145 Exercise 2 ===\n")

	// Del 1: Finn server IP
	serverIP := listenForServerIP()
	

	// Vent litt før vi fortsetter
	time.Sleep(1 * time.Second)

	// Del 2: Kommuniser med serveren
	// VIKTIG: Endre workstationNumber til nummeret på arbeidsstasjonen din!
	workstationNumber := 7 // <-- ENDRE DETTE

	fmt.Printf("\n*** Bruker arbeidsstasjon nummer: %d ***\n", workstationNumber)
	fmt.Println("*** Hvis du sitter på en annen arbeidsstasjon, endre workstationNumber i main() ***\n")

	communicateWithServer(serverIP, workstationNumber)

	fmt.Println("\nFerdig!")
}
