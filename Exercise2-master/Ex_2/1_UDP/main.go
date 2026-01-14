package main

import (
    "fmt"
    "net"
    "time"
)

func main() {
    // Finn server IP først (én gang)
    serverIP, err := findServerIP()
    if err != nil {
        fmt.Println("Could not find server IP:", err)
        return
    }
    fmt.Println("Server IP found:", serverIP)

    // Start mottaks-tråd
    go Udp_receive(serverIP)

    // Vent litt før vi begynner å sende
    time.Sleep(1 * time.Second)

    // Send meldinger
    UDP_Send(serverIP)

    // Hold programmet kjørende
    wait := make(chan int)
    <-wait
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

    message := string(buffer[:n])
    fmt.Printf("Received broadcast: %s\n", message)
    return senderAddr.IP.String(), nil
}

func Udp_receive(serverIP string) {
    port := 20007 // Server sender svar til 20007 (eller 20000 + workspaceNumber hvis asymmetrisk)

    addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
    conn, _ := net.ListenUDP("udp", addr)
    defer conn.Close()

    fmt.Printf("Listening for server replies on port %d...\n", port)
    buffer := make([]byte, 1024)

    for {
        n, remoteAddr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println("Receive error:", err)
            continue
        }

        // Filtrer ut meldinger som ikke kommer fra serveren
        if remoteAddr.IP.String() == serverIP {
            fmt.Printf("Received from server: %s\n", string(buffer[:n]))
        }
    }
}

func UDP_Send(serverIP string) {
    workspaceNumber := 7
    port := 20000 + workspaceNumber

    addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", serverIP, port))
    if err != nil {
        fmt.Println("Error resolving address:", err)
        return
    }

    conn, err := net.DialUDP("udp", nil, addr)
    if err != nil {
        fmt.Println("Error creating connection:", err)
        return
    }
    defer conn.Close()

    // Send meldinger med intervaller
    for i := 1; i <= 5; i++ {
        message := fmt.Sprintf("Hello from workspace 7 - message %d", i)
        _, err = conn.Write([]byte(message))
        if err != nil {
            fmt.Println("Error sending:", err)
            return
        }
        fmt.Printf("Sent: %s\n", message)
        time.Sleep(2 * time.Second) // Vær snill mot nettverket
    }
}