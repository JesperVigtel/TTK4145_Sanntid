// package main

// import (
// 	"fmt"
// 	"net"
// 	"time"
// )

// func main() {
// 	wait := make(chan int)
// 	Udp_receive()
// 	<-wait
// }

// func Udp_receive() {
// 	buffer := make([]byte, 1024)
// 	ServerAddr, _ := net.ResolveUDPAddr("udp", ":30000")
// 	conn, _ := net.ListenUDP("udp", ServerAddr)
// 	var localIP, _ = findServerIP()
// 	fmt.Println("LOCAL IP: " + localIP)
// 	for {
// 		fmt.Println("step1")
// 		n, addr, err := conn.ReadFromUDP(buffer)
// 		fmt.Println("\tHER: ", err)
// 		fmt.Println("step2")
// 		if addr.String() != localIP {
// 			fmt.Println(string(buffer[0:n]))
// 		}
// 	}

// }


// func findServerIP() (string, error) {
// 	addr, err := net.ResolveUDPAddr("udp", ":30000")
// 	if err != nil {
// 		return "", err
// 	}

// 	conn, err := net.ListenUDP("udp", addr)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer conn.Close()

// 	buffer := make([]byte, 1024)
// 	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

// 	n, senderAddr, err := conn.ReadFromUDP(buffer)
// 	if err != nil {
// 		return "", err
// 	}

// 	_ = string(buffer[:n]) // optional: use message
// 	return senderAddr.IP.String(), nil
// }