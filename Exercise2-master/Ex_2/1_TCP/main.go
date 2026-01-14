package main

import (
    "fmt"
    "net"
    "time"
)

const (
    SERVER_IP         = "10.100.23.242"
    LOCAL_IP          = "10.100.23.233"
    LOCAL_LISTEN_PORT = 34934
)

func receiver(conn net.Conn) {
    buffer := make([]byte, 1024)
    for {
        n, _ := conn.Read(buffer)
        fmt.Printf("Mottatt: %s\n", string(buffer[:n]))
    }
}

func sender(conn net.Conn, messages []string) {
    for _, msg := range messages {
        conn.Write([]byte(msg + "\x00"))
        fmt.Printf("Sendt: %s\n", msg)
        time.Sleep(1 * time.Second)
    }
}

func main() {
    // Del 1: Direkte tilkobling
    fmt.Println("--- Del 1: Direkte tilkobling ---")
    conn1, _ := net.Dial("tcp", SERVER_IP+":34933")
    go receiver(conn1)
    sender(conn1, []string{"Hallo fra Go", "Melding 2", "Melding 3"})
    time.Sleep(2 * time.Second)
    conn1.Close()

    // Del 2: Server kobler seg til oss
    fmt.Println("\n--- Del 2: Server kobler seg til oss ---")
    listener, _ := net.Listen("tcp", ":"+fmt.Sprint(LOCAL_LISTEN_PORT))

    conn2, _ := net.Dial("tcp", SERVER_IP+":33546")
    conn2.Write([]byte(fmt.Sprintf("Connect to: %s:%d\x00", LOCAL_IP, LOCAL_LISTEN_PORT)))
    conn2.Close()

    connServer, _ := listener.Accept()
    go receiver(connServer)
    sender(connServer, []string{"Svar fra reverse connection", "Takk for tilkoblingen"})
    time.Sleep(2 * time.Second)

    connServer.Close()
    listener.Close()
}