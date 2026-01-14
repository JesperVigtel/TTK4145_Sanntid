package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	// time.Sleep(5 * time.Second)
	// name, err := findServerIP()
	// if err != nil {
	// 	fmt.Println("Could not find server IP:", err)
	// 	return
	// }
	// fmt.Println("Server IP found:", name)

	UDP_Send()
}

func findServerIP() (string, error) {
	addr, err := net.ResolveUDPAddr("udp", ":30000")
	if err != nil {
		return "", err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	n, senderAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return "", err
	}

	_ = string(buffer[:n]) // optional: use message
	return senderAddr.IP.String(), nil
}

func UDP_Send() {

	serverIP, err := findServerIP()
	if err != nil {
		fmt.Println("Could not find server IP:", err)
		return
	}
	fmt.Println("Server IP found:", serverIP)
	//serverIP := "255.255.255.255"

	workspaceNumber := 7
	port := 20000 + workspaceNumber

	if err := sendUDP(serverIP, port); err != nil {
		fmt.Println("Error sending UDP message:", err)
		return
	}
	time.Sleep(300 * time.Millisecond)
	fmt.Println("UDP message sent successfully")
}

func sendUDP(serverIP string, port int) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", serverIP, port))
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	message := "Hello from workspace 7"
	_, err = conn.Write([]byte(message))
	if err != nil {
		return err
	}

	conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)

	if err != nil {
		return fmt.Errorf("no reply: %w", err)
	}

	fmt.Printf("Received reply: %s\n", string(buf[:n]))

	return nil
}

func readUDP() (string, error) {
	addr, err := net.ResolveUDPAddr("udp", ":30000")
	if err != nil {
		return "", err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	n, senderAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return "", err
	}

	_ = string(buffer[:n]) // optional: use message
	return senderAddr.IP.String(), nil
}
