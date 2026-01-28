# Update 26 Jan 2026 — Specs Check After S3 Integration

Compliance summary:
- No calls lost: ✅ CAB persistence on disk; verified crash/restart flow.
- Button light guarantee: ✅ Idempotent updates + faster 5ms broadcast; convergence confirmed.
- Fault tolerance: ✅ Heartbeat + motor/obstruction detection; graceful offline/online.
- Packet loss resilience: ✅ Periodic, type-tagged messages; idempotent application.
- Scalability: ✅ HRA assignment; configurable intervals and timeouts.
- Cross-platform: ✅ Avoid auto-reboot; persistence-based recovery; envelope logic portable.

# PRELIMINARY DESIGN - KOMPLETT SAMMENFATNING

## 📋 HVA DU HAR FÅTT

Du har nå fått et komplett designsett bestående av:

### 1. **PDD_PRELIMINARY_DESIGN.md** ← **START HER**
   - Formal preliminary design description (< 1 side)
   - Klart og konsist for innlevering
   - Alle kritiske designbeslutninger
   - Oppfyller alle formelle krav

### 2. **UML_DIAGRAMS.md**
   - Elevator State Machine (5 tilstander)
   - Button Light State Machine
   - Class/Component Diagram
   - Sequence Diagrams (normal operasjon, disconnect, crash)

### 3. **IMPLEMENTASJONSGUIDE.md**
   - Pseudokode for alle 3 hovedmoduler
   - Data structures
   - Main/coordinator setup
   - Testing strategy

### 4. **SYSTEMARKITEKTUR.md**
   - Full system overview
   - Channel communication map
   - Timing analysis for all critical paths
   - Fault tolerance matrix

### 5. **ANALYSE_HEISPROSJEKT.md** (fra tidligere)
   - Komparativ analyse av de to løsningene
   - Forklaring av designvalg

---

## 🎯 DESIGNEN I KORTE TREKK

### Arkitektur: **3-Moduls P2P UDP Mesh System**

```
┌─────────────────────────────────────┐
│  ELEVATOR FSM (5 states)            │
│  • Kjører heisen                    │
│  • Håndterer dør & motor            │
│  • Rapporterer status               │
└──────────────┬──────────────────────┘
               │
    ┌──────────┴──────────┐
    │                     │
    ▼                     ▼
┌─────────────┐    ┌─────────────────┐
│NETWORK MOD  │◄──►│ ORDER MANAGER   │
│• Broadcast  │    │ (Central Brain) │
│• Heartbeat  │    │ • CAB Persist   │
│• Registry   │    │ • Hall Assign   │
└─────────────┘    │ • Lights        │
                   │ • Consensus     │
                   └─────────────────┘
```

### Fault Tolerance: **Persistent Storage + Distributed Takeover**

| Feil | Løsning |
|------|---------|
| Strømbrudd | CAB-ordrer lagres til disk, gjenopprettes ved oppstart |
| Nettverksfeil | Heis fungerer lokalt, andre tar over hall-ordrer innen 3s |
| Pakketap | Idempotent broadcasts gjør tap transparent |
| Dørblock | Timer restartes, dør holder seg åpen |

---

## ✅ INNLEVERING: WHAT YOU NEED

### For PDD PDF (innleveres på Blackboard):

```
Filnavn: PDD-##.pdf (hvor ## er gruppens nummer)
Format: PDF (en side)
Innhold:
  ✓ Lab time + desk number + group number
  ✓ Gruppemedlemmer (navn + email)
  ✓ Fault tolerance strategi
  ✓ Network topology & protokoll
  ✓ Programming language rationale (Go)
  ✓ System modulation (3 modules)
  ✓ Tabeller/diagrammer (kan være håndtegnet hvis klart)
  ✓ Svar på 5 kritiske spørsmål:
    - Button light contract?
    - Network unreliability?
    - Crashes & restarts?
    - Network disconnect + hall orders?
    - CAB order crash?
```

### Template for PDD (kopier fra PDD_PRELIMINARY_DESIGN.md):

Du kan kopiere innholdet direkte fra `PDD_PRELIMINARY_DESIGN.md` og gjøre det til PDF.
- Legg til dine egne navn og epostadresser
- Legg til lab time og desk number
- Legg til gruppe nummer

---

## 🔄 IMPLEMENTASJONSFLYT

### Når du skal kode:
1. **Les IMPLEMENTASJONSGUIDE.md** først
2. **Start med Modul 1: Elevator FSM** 
   - Easiest to test
3. **Legg til Modul 2: Network Module** når FSM fungerer
4. **Legg til Modul 3: Order Manager** til slutt
5. **Test ved hjelp av Testing Strategy** i guide

### Kodestruktur å følge:
```go
// main.go
func main() {
    fsm := &ElevatorFSM{...}
    network := &NetworkModule{...}
    orderMgr := &OrderManager{...}
    
    go fsm.Run()
    go network.Run()
    go network.Sender(...)
    go network.Receiver(...)
    go orderMgr.Run()
    
    select {}
}
```

---

## 🎓 DESIGNBESLUTNINGER FORKLART

### 1. Why Go?
```
✓ Goroutines → enkel parallellisme
✓ Channels → thread-safe kommunikasjon
✓ Raskt + deterministisk
✓ Innebygd UDP support
```

### 2. Why Persistent CAB Storage?
```
✓ Løser "no calls lost" kravet
✓ Enkel implementering (file I/O)
✓ Robust mot vilkårlige krasj
✓ Er det eneste som garanterer CAB-sikkerhet
```

### 3. Why Idempotent Broadcasts?
```
✓ Pakketap blir transparent
✓ Duplikater er harmløse
✓ Ikke kompleks "exactly-once" logikk
✓ Fungerer bra med UDP broadcast
```

### 4. Why Distance-Based Conflict Resolution?
```
✓ Innebygd logikk (no external dependencies)
✓ Enkelt å forstå og teste
✓ Gir "reasonable" ytelse
✓ Ikke optimal, men god nok
```

### 5. Why 3-Second Heartbeat Timeout?
```
✓ "Seconds" magnitude per spec
✓ Lang nok til å tolerere packet loss
✓ Kort nok til å oppdage feil raskt
✓ Standard for distributed systems
```

---

## 🧪 QUICK TEST VALIDATION

### Test 1: Kan du tegne FSM?
- [ ] IDLE state
- [ ] MOVING state
- [ ] DOOR_OPEN state
- [ ] EMERGENCY_STOP state
- [ ] Alle transitions korrekt?

### Test 2: Kan du forklare et CAB-krasj?
```
Q: Hva skjer hvis heis A krasjer med aktiv CAB-ordre?
A: 1. CAB lagret på disk
   2. Heis starter på nytt
   3. LoadCabOrders() restituerer ordre
   4. Heis betjener orden normalt
   ✓ Orden ER IKKE tapt
```

### Test 3: Kan du forklare network disconnect?
```
Q: Hva skjer hvis heis A mister nettverk mens den serverer en hall-ordre?
A: 1. Andre heiser venter på heartbeat (50ms intervals)
   2. Etter 3 sekunder: timeout
   3. A markeres offline
   4. Ordren omkonfigureres til B eller C
   5. Ny heisen betjener ordre
   ✓ Ordre TATT OVER innen 3 sekunder
```

### Test 4: Kan du tegne modul-arkitekturen?
```
Tegn eller beskriv:
├─ Elevator FSM
├─ Network Module  
└─ Order Manager

Tegn channelene mellom dem
```

---

## 📊 SAMMENLIGNING: DESIGN vs. REQUIREMENTS

```
┌─────────────────────────────────────────────────────────┐
│ KRAV                    │ LØSNING                       │
├─────────────────────────────────────────────────────────┤
│ Button light contract   │ Cyclisk state machine         │
│                         │ Immediacy for CAB             │
├─────────────────────────────────────────────────────────┤
│ No calls lost           │ Persistent CAB storage        │
│                         │ Automatic takeover            │
├─────────────────────────────────────────────────────────┤
│ Network unreliability   │ Idempotent broadcasts         │
│                         │ Heartbeat detection           │
├─────────────────────────────────────────────────────────┤
│ Spontaneous crashes     │ Disk persistence + restart    │
│                         │ Network helps recovery        │
├─────────────────────────────────────────────────────────┤
│ Door functionality      │ 3-second timer                │
│                         │ Obstruction handling          │
├─────────────────────────────────────────────────────────┤
│ Efficient serving       │ Distance-based assignment     │
│                         │ Not optimal but good enough   │
└─────────────────────────────────────────────────────────┘
```

---

## 🚀 HAR DU ALT DU TRENGER?

### For Innlevering (PDD):
✅ Du har template i `PDD_PRELIMINARY_DESIGN.md`

### For Designforståelse:
✅ Du har UML-diagrammer i `UML_DIAGRAMS.md`

### For Implementasjon:
✅ Du har pseudokode + datastrukturer i `IMPLEMENTASJONSGUIDE.md`

### For Systemforståelse:
✅ Du har arkitektur + timings i `SYSTEMARKITEKTUR.md`

### For Å Velge Løsning:
✅ Du har analyse i `ANALYSE_HEISPROSJEKT.md`

---

## 🎯 NESTE STEG

### UMIDDELBAR (idag):
1. Les `PDD_PRELIMINARY_DESIGN.md` raskt
2. Fyll inn dine navn/emails/group number
3. Konverter til PDF
4. **Lever før fristen!**

### ETTER LEVERING (denne uken):
1. Diskuter design i gruppen
2. Velg hvem som implementerer hvilken modul
3. Sett opp Git repository
4. Start med Modul 1 (Elevator FSM)

### IMPLEMENTASJON (kommende uker):
1. Følg IMPLEMENTASJONSGUIDE.md
2. Test hver modul separat
3. Integrer moduler gradvis
4. Test fault scenarios

---

## ❓ VANLIGE SPØRSMÅL

**Q: Kan jeg bruke en annen programmeringsspråk?**
A: Ja, men Go er anbefalt pga. goroutines/channels. Python/Rust er også OK.

**Q: Er persistent lagring virkelig nødvendig?**
A: Ja - kravet sier "no calls are lost" selv ved strømbrudd. Persistent lagring er eneste garantia.

**Q: Kan jeg bruke HRA (hall_request_assigner)?**
A: Ja, men det er ikke nødvendig. Distance-based assignment er enklere og tilstrekkelig.

**Q: Hva hvis nettverket partisjonerer i flere grupper?**
A: Spec sier det ikke skjer (< 4 heiser), så vi assumerer det ikke.

**Q: Hvor skal jeg lagre CAB-ordrer?**
A: Tekstfil: `cab_orders_[ID].txt` i samme mappe som programmet.

**Q: Hvor lang skal PDD være?**
A: < 1 side tekst (pluss titler, navn, figurer).

---

## 📝 CHECKLIST FØR INNLEVERING

```
PDD Document:
  [ ] Filnavn: PDD-##.pdf (gruppens nummer)
  [ ] Lab time er listet opp
  [ ] Desk number er listet opp
  [ ] Group number er listet opp
  [ ] Alle gruppemedlemmer navn + email
  [ ] Fault tolerance strategi forklart
  [ ] Network topology beskrevet
  [ ] Go-valg begrunnet
  [ ] Module struktur beskrevet
  [ ] Max 1 side tekst
  [ ] Diagrammer/figurer inkludert (kan være håndtegnet)
  [ ] Leveres som PDF (ikke Word, ikke Google Docs)
  [ ] Lastopplag på Blackboard før deadline

UML Diagrams (For senere bruk):
  [ ] FSM diagram tegnet/gjort
  [ ] Button light states forklart
  [ ] Component diagram klart
  [ ] Sequence diagrams for normale tilfeller

Implementation Ready:
  [ ] Modul 1 pseudokode forstått
  [ ] Modul 2 pseudokode forstått
  [ ] Modul 3 pseudokode forstått
  [ ] Data structures definert
  [ ] Config/konstanter satt
```

---

## 💡 PRO TIPS

1. **Tegn diagrammene for hånd først** - det hjelper forståelsen
2. **Forstå FSM først** - alt annet bygger på det
3. **Test en modul om gangen** - ikke alt på en gang
4. **Bruk channels bevisst** - de er nøkkelen til Go-design
5. **Dokumenter mens du koder** - det er bare en halvparters ekstraarbeid

---

## 📚 RESSURSER I MAPPEN

```
/Users/jespervh/Desktop/mappe uten navn/

├─ PDD_PRELIMINARY_DESIGN.md          ← INNLEVER DENNE (som PDF)
├─ UML_DIAGRAMS.md                    ← REFERANSE UNDER KODING
├─ IMPLEMENTASJONSGUIDE.md            ← PSEUDOKODE + KODE-STRUKTUR
├─ SYSTEMARKITEKTUR.md                ← DETALJER + TIMINGS
├─ ANALYSE_HEISPROSJEKT.md            ← BAKGRUNN PÅ DESIGN-VALG
└─ [Analyse av L1 og L2 løsningene i Elevator-main/ og TTK4145-Heislab-master/]
```

---

## ✨ DIN DESIGN ER READY!

Du har nå:
- ✅ En robust, fault-tolerant design
- ✅ Kombinert beste elementer fra begge referanseløsningene
- ✅ Klar implementasjonsguide med pseudokode
- ✅ UML-diagrammer for alle kritiske components
- ✅ Detaljert systemarkitektur med timings
- ✅ Alt du trenger for å lykkes!

**LYKKE TIL MED PROSJEKTET! 🎉**

