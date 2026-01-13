# UDP Klient i Go - TTK4145 Exercise 2

## Hvordan kjøre programmet

### 1. Kjør programmet
```bash
go run udp_client.go
```

### 2. Husk å endre arbeidsstasjonsnummer!
Åpne `udp_client.go` og endre linjen:
```go
workstationNumber := 1 // <-- ENDRE DETTE til riktig stasjonsnummer
```

## Hva programmet gjør

### Del 1: Finn server IP
- Lytter på UDP port **30000**
- Mottar broadcast fra serveren med dens IP-adresse
- Skriver ut server IP til terminalen

### Del 2: Kommuniser med serveren
- Sender meldinger til server på port `20000 + arbeidsstasjonsnummer`
- Mottar og skriver ut svar fra serveren
- Serveren svarer med "You said: " foran meldingen din

## Tips

### Hvis du jobber hjemmefra
Se [working-from-home.md](working-from-home.md) for hvordan du setter opp din egen server.

### Hvis du vil teste broadcast
Du kan også sende til broadcast-adresse (`#.#.#.255` eller `255.255.255.255`):

```go
// Eksempel på broadcast sending
broadcastAddr, _ := net.ResolveUDPAddr("udp", "255.255.255.255:20001")
conn, _ := net.DialUDP("udp", nil, broadcastAddr)
conn.Write([]byte("Broadcast melding!"))
```

### Nyttige Go-pakker for nettverkskommunikasjon
- `net` - Standard Go networking
- `net/udp` - UDP spesifikk funksjonalitet  
- `time` - For delays mellom meldinger

## Feilsøking

**Problem:** "bind: address already in use"
- En annen prosess bruker porten allerede
- Endre porten eller stopp den andre prosessen

**Problem:** Mottar ingen svar fra server
- Sjekk at server-IP er korrekt
- Sjekk at arbeidsstasjonsnummer er riktig
- Sjekk at serveren kjører

**Problem:** Mottar egne meldinger tilbake
- Dette skjer hvis du bruker broadcast
- Filtrer ut meldinger som ikke kommer fra server IP
