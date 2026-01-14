package localip


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

//Hei