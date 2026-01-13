
// if sending directly to a single remote machine:
    addr = new Address(remoteIP, remotePort)
    sock = new Socket(udp)
    
    // either: set up the socket to use a single remote address
        sock.connect(addr)
        sock.send(message)
    // or: set up the remote address when sending
        sock.sendTo(message, addr)
        
// if sending on broadcast:
// you have to set up the BROADCAST socket option before calling connect / sendTo
    broadcastIP = #.#.#.255 // First three bytes are from the local IP, or just use 255.255.255.255
    addr = new InternetAddress(broadcastIP, port)
    sendSock = new Socket(udp) // UDP, aka SOCK_DGRAM
    sendSock.setOption(broadcast, true)
    sendSock.sendTo(message, addr)