# Update 26 Jan 2026 — Adopted Elements from Solution 3

Adopted:
- 5ms broadcast interval (configurable)
- Type-tagged JSON message envelope
- Fault detection for motor stop and door obstruction with graceful offline/online toggling
- Clear request stop/clear semantics

Not adopted:
- ACK table consensus (kept idempotent convergence)
- Automatic reboot (persist and recover CAB orders instead)

Effect on hybrid solution:
- Faster convergence and safer routing without extra complexity
- Stronger fault handling while preserving simplicity and portability

# Komparativ Analyse: To P2P UDP Mesh Implementasjoner av Heisprosjektet

## Oversikt over Løsningene

**Løsning 1: Elevator-main/TTK4145/Project** (heretter L1)
- Bruker sentralisert ordredistribusjon via `hall_request_assigner` (HRA)
- Implementerer cyclisk konsistensmekanisme
- Mer avansert nettverksarkitektur

**Løsning 2: TTK4145-Heislab-master/Source** (heretter L2)
- Bruker distribuert konfliktløsning basert på distanse
- Enklere nettverksarkitektur
- Mer simpel tilnærming til bestillingsfordeling

---

## 1. HÅNDTERING AV KRAV: "Button Lights are a Service Guarantee"

### Løsning 1 (L1)
**Implementering:**
- Bruker `ButtonState` enum med 5 tilstander: `Initial`, `Standby`, `ButtonPressed`, `OrderAssigned`, `OrderComplete`
- Cyclisk konsistensmekanisme i `cyclicCounter.go` sikrer alle heiser blir enige om tilstand
- State transitions må være gyldig på alle heiser før neste tilstand nås
- HRA (hall_request_assigner) distribuerer bestillinger basert på kostnadsoptimalisering

**Fordeler:**
✅ Deterministisk tilstandsmaskine sikrer alle heiser når samme lystilstand  
✅ Cyclisk mekanisme hindrer inkonsistensproblem ved pakketap  
✅ Garantert fordeling av bestillinger via HRA  
✅ Lyset skal antenne så lenge noen heiser er aktive

**Ulemper:**
❌ Mer komplisert logikk og mehrere tilstandsoverganger  
❌ Krav på HRA-eksekverbar som ikke er inkludert  
❌ Pakketap kan forsinke tilstandsendringer

### Løsning 2 (L2)
**Implementering:**
- Orden har status: `-1` (inaktiv), `0` (pending), `elevID` (aktiv/tatt), eller `-2` (timeout)
- Når knapp trykkes: umiddelbar lysenalering hvis `Status == 0`
- For hall calls: bare CAB-ordrer lyses opp umiddelbart, hall-ordrer venter på bekreftelse fra flere heiser

**Fordeler:**
✅ Enklere implementering  
✅ CAB-ordrer får lys umiddelbart  
✅ Færre nettverksavhengigheter  

**Ulemper:**
❌ Hall-ordrer krever synkronisering før lys tent (kan ta tid)  
❌ Hvis nettverket er nede, hall-ordrer lyst ikke  
❌ Mindre robust mot pakketap

**Vinner: Løsning 1** - Garanterer mer robust lysgaranti gjennom cyclisk mekanisme

---

## 2. HÅNDTERING AV KRAV: "No Calls are Lost"

### Løsning 1 (L1)
**Håndtering av nettverksfeil:**
- `HeartbeatTimeout: 3000ms` - heiser markeres som offline hvis ingen meldinger
- `cyclicCounter` sikrer at bestillinger ikke blir "glemt" selv ved midlertidig partition
- Når heiser blir online igjen, mottar den hele tilstanden fra andre via broadcast
- **CAB-ordrer:** Lagres IKKE eksplisitt på disk - krever nettverkskommunikasjon for gjenoppretting

**Håndtering av strømbrudd/krasj:**
- Ingen persistens for CAB-ordrer
- Når heisen starter på nytt, setter den `elevator.ActiveStatus = false`
- Andre heiser vil ta over hall-ordrer
- **Kritisk svakhet:** CAB-ordrer blir tapt ved strømbrudd

### Løsning 2 (L2)
**Håndtering av nettverksfeil:**
- `checkForTimeout()` implementerer timeout-mekanisme
- Heiser som forsvinner fra nettverket fjernes fra `otherElevInfo` etter timeout
- Hvis heisen blir offline:
  - CAB-ordrer fortsetter å fungere (lagret lokalt)
  - Hall-ordrer **kan ikke tas** (krever nettverkskommunikasjon for fordeling)

**Håndtering av strømbrudd/krasj:**
- **CAB-ordrer:** Persisteres til disk via `CabOrderBackup[ID].txt`
- Ved oppstart: `ReadCabOrderBackup()` restituerer CAB-ordrer fra fil
- Hall-ordrer må gjenopprettes fra andre heiser via nettverket

**Implementering:**
```go
// Løsning 2: Backup av CAB-ordrer
func UpdateCabOrderBackup() {
    filename := "CabOrderBackup" + strconv.Itoa(id) + ".txt"
    file.WriteString(cabOrderString)
}

// Lesing ved oppstart
func ReadCabOrderBackup() {
    content, _ := ioutil.ReadFile(filename)
    // Gjenoppretter alle CAB-ordrer fra fil
}
```

### Sammenligning av Feiloversikt:

| Feiltype | Løsning 1 | Løsning 2 |
|----------|-----------|----------|
| Midlertidig nettverkstap | ✅ CAB/Hall bevares | ✅ CAB/Hall bevares |
| Permanent nettverkstap | ✅ Heiser fungerer separat | ✅ Heiser fungerer separat |
| Strømbrudd (CAB-ordrer) | ❌ **TAPT** | ✅ Gjenopprettes fra disk |
| Strømbrudd (Hall-ordrer) | ✅ Gjenopprettes via nett | ✅ Gjenopprettes via nett |
| Software krasj | ❌ CAB-ordrer tapt | ✅ CAB-ordrer gjenopprettes |

**Vinner: Løsning 2** - Implementerer persistent lagring av CAB-ordrer som løser kravene bedre

---

## 3. LYS- OG KNAPPFUNKSJONALITET

### Løsning 1 (L1)
**Hall-knapper:** 
- Alle heiser sender sine knapptrykk via broadcast
- Ordre blir distribuert til en heiser via HRA
- Lyset holdes på via cyclisk mekanisme til orden er utført

**CAB-knapper:**
- Kun den lokale heisen reagerer
- Lyset aktiveres når `OrderAssigned`-tilstand nås

**Prøvetid:** Lys slettes når orden er `OrderComplete`

### Løsning 2 (L2)
**Hall-knapper:**
- Knapptrykk lagres med `Status = 0` (pending)
- Venter på bekreftelse fra andre heiser
- Første heisen som sier "jeg tar denne" blir prioritert

**CAB-knapper:**
- Aktiveres umiddelbart på lokalt panel (`Status = elevatorID`)
- Kan betjenes selv om nettverket er nede

**Prøvetid:** Lys slettes når orden er `Finished = true`

### Ulemper ved begge:

**L1:**
- Avhenger av HRA-eksekverbar
- Hvis HRA krasjer = ingen bestillingsdistribusjon

**L2:**
- Hall-ordrer krever synkronisering (langsom oppstart)
- Enklere å ha inkonsistensproblem

**Vinner: Løsning 1** - Mer robust lysbelysningsmekanisme (men avhenger av HRA)

---

## 4. HÅNDTERING AV NETTVERKSDISKONEKSJON

### Løsning 1 (L1)
**Atferd når offline:**
```
if !online {
    newAliveList := [NElevators]bool{}
    newAliveList[nodeID] = aliveList[nodeID]
    stateBroadcast <- FromNetworkToAssigner{...}
}
```
- Sender kun egen status
- Andre heiser får beskjed om at den er offline
- **Hall-ordrer:** Stoppet - ingen nye kan tas når offline
- **CAB-ordrer:** Fortsetter (men er ikke persistert)

### Løsning 2 (L2)
**Atferd når offline:**
- FSM-staten blir `IDLE` eller `EXECUTE` basert på siste tilstand
- Heisen fortsetter å behandle sine lokale CAB-ordrer
- Hall-ordrer stoppet
- **Restart ikke nødvendig** - enters "single elevator mode"

### Vinner: Løsning 2** - Mer robust offline-atferd med persistering

---

## 5. EFFEKTIVITET: BESTILLINGSDISTRIBUSJON

### Løsning 1 (L1)
**Algoritme:** Bruker `hall_request_assigner` (HRA)
- Optimalt i teori - løser ved å minimere total oppbygningskost
- Konsistent distribusjon av alle heiser
- **Krav:** Eksekverbar `hall_request_assigner` må være tilgjengelig

**Distribusjon:**
```go
ret, err := exec.Command("hall_request_assigner", "-i", string(jsonBytes)).CombinedOutput()
```

### Løsning 2 (L2)
**Algoritme:** Basert på distanse + heiser-ID
```go
func solveConflict(order, elev, conflictElevs) bool {
    myDist := math.Abs(float64(floor - order.Floor))
    for _, other := range conflictElevs {
        theirDist := math.Abs(float64(other.Floor - order.Floor))
        if myDist > theirDist { return false }
        if myDist == theirDist && myId > other.Id { return false }
    }
    return true
}
```
- Velg heisen nærmest (+ bruk ID som tiebreaker)
- Ikke optimalt, men enkelt og distribuert

### Sammenligning:

| Aspekt | L1 (HRA) | L2 (Distanse) |
|--------|---------|---------------|
| Optimalitet | ✅ Teoretisk optimal | ❌ Suboptimal |
| Kompleksitet | ❌ Avhender eksekverbar | ✅ Innebygd logikk |
| Robusthet | ❌ Kan krasje | ✅ Alltid tilgjengelig |
| Oppstarttid | ❌ Initialisering nødvendig | ✅ Umiddelbar |

**Vinner: Løsning 1** (hvis HRA virker), **ellers Løsning 2**

---

## 6. KOMMUNIKASJONSARKITEKTUR

### Løsning 1 (L1)
**Broadcast-mekanisme:**
- Sender hele `Message` struct med JSON
- Inneholder: `ElevatorList`, `HallOrderList`, `OnlineStatus`, `AliveList`
- Bruker UDP broadcast på port 1338
- **Heartbeat:** 50ms broadcast-rate, 3s timeout

**Pakkestruktur:**
```go
type Message struct {
    SenderId      int
    ElevatorList  [NElevators]HRAElevState
    HallOrderList [NElevators][NFloors][NButtons]ButtonState
    OnlineStatus  bool
    AliveList     [NElevators]bool
}
```

### Løsning 2 (L2)
**Broadcast-mekanisme:**
- Sender hele `Elev` struct med JSON
- Inneholder: `Id`, `Floor`, `CurrentOrder`, `State`, `Orders`
- Bruker UDP broadcast på port 20009
- **Heartbeat:** 2ms poll interval, 4 heartbeat ticks threshold

**Pakkestruktur:**
```go
type Elev struct {
    Id           int
    Floor        int
    CurrentOrder Order
    State        int
    Orders       [numFloors][numButtons]Order
}
```

### Sammenligning:

| Aspekt | L1 | L2 |
|--------|----|----|
| Pakkefrekvens | 50ms | 2ms kontroll, async broadcast |
| Pakkesize | Større (ElevatorList hele arrayen) | Mindre (kun egen tilstand) |
| Kompleksitet | Høy (hele nettverkstilstand) | Lav (egen tilstand) |
| Robusthet mot tap | Bedre (redundans) | Dårligere (mindre redundans) |

**Vinner: Løsning 1** - Mer robust kommunikasjonsdesign

---

## 7. HÅNDTERING AV DØRFUNKSJONALITET

### Løsning 1 (L1)
- 3 sekunders dørhold implementert via timer (`DoorOpenDuration = 3 * time.Second`)
- Obstruksjon blokkerer dørlukkingen
- Retningsendring (3s ekstra ventetid) implementert

### Løsning 2 (L2)
- Lignende 3-sekundersmekanisme via `elevcontroller.ElevStopAtFloor()`
- Obstruksjon håndteres via `pollObstructionSwitch()`

**Begge løsningene:** ✅ Implementerer kravet korrekt

---

## SAMMENFATNING OG ANBEFALINGER

### Løsning 1 Styrker:
1. **Robust nettverkskommunikasjon** - cyclisk mekanisme sikrer konsistens
2. **Optimal bestillingsdistribusjon** via HRA
3. **Redundant tilstandslagring** - hele nettverkstilstanden sendes
4. **Robust mot pakketap**

### Løsning 1 Svakheter:
1. ❌ **CAB-ordrer tapt ved strømbrudd** - kritisk problem
2. Avhender ekstern `hall_request_assigner` eksekverbar
3. Høyere kompleksitet
4. Større pakkestørrelse

---

### Løsning 2 Styrker:
1. ✅ **Persistent lagring av CAB-ordrer** - løser strømbruddsproblemet
2. **Enklere implementering** - færre tilstander, intuitiv logikk
3. **Ingen eksterne avhengigheter** - alt er innebygd
4. **Raskere oppstart**
5. **Fungerer offline** - CAB-ordrer fortsetter

### Løsning 2 Svakheter:
1. ❌ **Suboptimal bestillingsdistribusjon** (men akseptabel)
2. **Mindre robust mot pakketap** - mindre redundans
3. **Hall-ordrer krever synkronisering** ved oppstart
4. Mer kompleks konfliktløsning

---

## KRITISK EVALUERING MOT PROSJEKTKRAV

| Krav | L1 | L2 | Kommentar |
|------|----|----|-----------|
| Button lights service guarantee | ✅ | ⚠️ | L1 er bedre, men L2 akseptabel |
| No calls lost | ❌ | ✅ | L2 har persistent lagring, L1 mister CAB ved strøm |
| No network loss | ✅ | ✅ | Begge fungerer offline |
| Lights sync under packet loss | ✅ | ⚠️ | L1 bedre design |
| Reasonable failure recovery | ✅ | ✅ | Begge ~3s recovery |
| Door functionality | ✅ | ✅ | Begge implementerer |
| Efficient call serving | ✅ | ⚠️ | L1 optimal (med HRA), L2 heuristisk |

---

## ANBEFALING FOR DIN LØSNING

**Beste hybrid-løsning:**

Kombiner Løsning 2's persistent lagring av CAB-ordrer med Løsning 1's cycliske konsistensmekanisme:

1. **Implementer persistent CAB-ordre lagring** (fra L2)
   - Lagrer til disk ved hver endring
   - Gjenoppretter ved oppstart
   - Løser strømbruddsproblemet

2. **Bruk cyclisk mekanisme eller simplere tilstandsmaskine** (inspirert L1)
   - Sikrer konsentert tilstander over nettverket
   - Eller implementer simpler idempotent broadcast

3. **Bruk L2's distanse-basert bestillingsdistribusjon**
   - Enklere, ingen eksterne avhengigheter
   - Tilstrekkelig god for prosjektkravet

4. **Legg til heartbeat/alive-detection**
   - Begge har dette, viktig for å vite hvem som er online

Denne kombinasjonen ville tilfredsstille **alle** prosjektkravene mens man unngår svakheter i begge originale løsningene.

