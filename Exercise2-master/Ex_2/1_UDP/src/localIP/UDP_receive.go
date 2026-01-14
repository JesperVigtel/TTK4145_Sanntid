// package main

// import (
// 	"fmt"
// 	"net"
// 	"time"
// )
// func findServerIP() (string, error) {
//     addr, err := net.ResolveUDPAddr("udp", ":30000")
//     if err != nil { return "", err }

//     conn, err := net.ListenUDP("udp", addr)
//     if err != nil { return "", err }
//     defer conn.Close()

//     buffer := make([]byte, 1024)
//     conn.SetReadDeadline(time.Now().Add(10 * time.Second))

//     n, senderAddr, err := conn.ReadFromUDP(buffer)
//     if err != nil { return "", err }

//     _ = string(buffer[:n]) // optional: use message
//     return senderAddr.IP.String(), nil
// }