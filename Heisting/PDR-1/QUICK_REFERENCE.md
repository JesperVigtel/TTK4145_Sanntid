# Update 26 Jan 2026 — Quick Config & Message Reference (S3)

Config keys (add or verify):
- `network.broadcast_interval_ms = 5`
- `faults.motor_stop_timeout_s = 4`
- `faults.obstruction_timeout_s = 5`
- `network.type_tagged_envelope = true`

Message Envelope (Go):
```go
type Envelope struct {
  TypeId      string
  PayloadJSON []byte
}
```

Receiver routing example:
```go
switch env.TypeId {
  case "types.PeriodicMsg": handlePeriodic(...)
  case "types.FaultMsg":    handleFault(...)
}
```

# QUICK REFERENCE GUIDE

## 📄 DOKUMENTER I DENNE LØSNINGEN

```
1. PDD_PRELIMINARY_DESIGN.md              [SUBMIT THIS AS PDF]
   └─ Formal design document for hand-in
   └─ < 1 page text + diagrams
   └─ Covers all key design decisions

2. UML_DIAGRAMS.md                        [REFERENCE WHILE CODING]
   ├─ Elevator FSM (5 states)
   ├─ Button Light State Machine
   ├─ Component Architecture
   └─ Sequence Diagrams (4 scenarios)

3. IMPLEMENTASJONSGUIDE.md                [PSEUDOKODE FOR MODULER]
   ├─ Modul 1: Elevator FSM (kod struktur)
   ├─ Modul 2: Network Module (kod struktur)
   ├─ Modul 3: Order Manager (kod struktur)
   ├─ Main/Coordinator (hvordan koble sammen)
   ├─ Data Structures (enum og types)
   └─ Testing Strategy

4. SYSTEMARKITEKTUR.md                    [DESIGNDETALJER]
   ├─ Full system overview
   ├─ Channel communication map
   ├─ Timing analysis (all critical paths)
   ├─ Fault tolerance matrix
   └─ Implementation priorities

5. SCENARIO_WALKTHROUGHS.md               [LÆRE AV EKSEMPLER]
   ├─ CAB order walkthrough (normal case)
   ├─ Hall call with conflict resolution
   ├─ Network disconnect + takeover
   └─ Software crash + recovery

6. ANALYSE_HEISPROSJEKT.md                [BAKGRUNN PÅ VALG]
   ├─ Comparison of two reference solutions
   ├─ Why we chose what we chose
   └─ Pros/cons of our design

7. SAMMENFATNING_OG_SJEKKLISTE.md         [OVERSIKT + PROGRESS]
   ├─ What you got
   ├─ What to submit
   ├─ Implementation roadmap
   └─ Pre-submission checklist
```

---

## 🎯 DITT DESIGN I 30 SEKUNDER

**Architecture:**
```
┌─────────────┐
│ ELEVATOR    │ ← Kjører heisen
│ FSM         │
└──────┬──────┘
       │
    ┌──┴──┐
    │     │
    ▼     ▼
┌────────────────┐
│ NETWORK MODULE │ ← Distribuert kommunikasjon (UDP)
└────────────────┘
    │
    ▼
┌─────────────────────┐
│ ORDER MANAGER       │ ← Hjerne (CAB persist + Hall assign)
└─────────────────────┘
```

**Fault Tolerance:**
```
Strømbrudd?        → CAB lagres på disk
Nettverksfeil?     → Heis fungerer lokalt, andre tar over
Pakketap?          → Idempotent broadcasts gjør det transparent
Dørblock?          → Timer restartes automatisk
```

**Kritisk Design Valg:**
```
1. CAB-ordrer persisteres til disk → Sikrer "no calls lost"
2. Distance + ID for konfliktløsning → Enkel distribuert algoritme
3. 3-sekunders heartbeat timeout → Oppdager feil raskt
4. Idempotent broadcasts → Pakketap blir transparent
```

---

## 🔑 KEY CONSTANTS

```go
const (
    N_ELEVATORS        = 3
    N_FLOORS           = 4
    N_BUTTONS          = 3
    
    BROADCAST_PORT     = 1338
    BROADCAST_INTERVAL = 50 * time.Millisecond
    HEARTBEAT_TIMEOUT  = 3 * time.Second
    DOOR_DURATION      = 3 * time.Second
)
```

---

## 🎭 STATE MACHINES AT A GLANCE

### Elevator States (5):
```
IDLE
  ↓↑ (new order)
MOVING
  ↓↑ (floor reached)
DOOR_OPEN ← EMERGENCY_STOP (obstruction)
  ↓↑
(door timer fires)
  ↓↑
(back to IDLE if no more orders)
```

### Button Light States (4):
```
STANDBY → BUTTON_PRESSED → ORDER_ASSIGNED → ORDER_COMPLETE → STANDBY
  (OFF)     (ON)            (ON)            (OFF)          (OFF)
```

---

## 📊 TIMING EXPECTATIONS

| Event | Time | Notes |
|-------|------|-------|
| CAB button → Light on | ~2ms | Immediate |
| CAB button → Elevator moves | ~80ms | Depends on assignment |
| Hall button → Consensus | ~50-70ms | Network broadcast cycle |
| Elevator moves to floor | ~100-200ms | Depends on distance |
| Door open | 3 seconds | Fixed timer |
| Network timeout | 3 seconds | Heartbeat |
| Disconnect → Takeover | ~3.1 seconds | Detect + reassign |

---

## 🔄 MESSAGE FLOW SUMMARY

```
USER PRESSES CAB BUTTON
  ↓
buttonPress channel
  ↓
ORDER MANAGER (immediate CAB light on + save to disk)
  ↓
ELEVATOR FSM (get order + move to floor + open door)
  ↓
NETWORK MODULE (broadcast status every 50ms)
  ↓
ALL ELEVATORS (receive and acknowledge)
  ↓
Door closes after 3s
  ↓
ORDER MANAGER (turn off CAB light + clear order)
  ↓
NETWORK MODULE (broadcast completion)
  ↓
ALL ELEVATORS (see completion + clear from their lists)
```

---

## 🚨 CRITICAL INVARIANTS TO MAINTAIN

1. **CAB orders are always persisted**
   - Save immediately on change
   - Load on startup
   - Clear after serving

2. **All elevators see same state eventually**
   - Broadcast every 50ms
   - Heartbeat timeout after 3s
   - Idempotent updates

3. **No double-taking**
   - Only one elevator per order
   - Conflict resolution uses distance + ID
   - Check before sending to FSM

4. **Door is never open while moving**
   - Motor STOP before door OPEN
   - Motor never starts if door not CLOSED

5. **Button lights follow state machine**
   - CAB: ON immediately, OFF after serving
   - Hall: ON when assigned, OFF when complete

---

## 🧪 TESTING CHECKLIST (In Order)

### Phase 1: Single Elevator
- [ ] FSM state transitions work
- [ ] Floor sensor → FSM → Motor works
- [ ] Door timer (3s) works
- [ ] Obstruction extends door timer
- [ ] CAB button → Light on immediately
- [ ] CAB order saved to disk
- [ ] CAB order persistence (restart)

### Phase 2: Two Elevators + Network
- [ ] Network sends/receives messages
- [ ] Heartbeat detection works
- [ ] Hall button → Both see it
- [ ] Conflict resolution picks closest
- [ ] Tiebreaker (ID) works
- [ ] Both elevators sync state

### Phase 3: Failures
- [ ] Network disconnect → Timeout in 3s
- [ ] Disconnect → Order reassigned
- [ ] Reconnect → Sync without errors
- [ ] Crash → CAB order recovery
- [ ] Door obstruction → Timer extends

### Phase 4: Edge Cases
- [ ] Simultaneous button presses
- [ ] Rapid network changes
- [ ] CAB + Hall mixed
- [ ] Multiple restarts
- [ ] Packet loss (tc/netem)

---

## 🐛 DEBUGGING TIPS

### If lights don't turn on:
```
1. Check: Is buttonPress channel receiving?
2. Check: Is ORDER MANAGER setting lightsOut channel?
3. Check: Is LIGHTS module listening to lightsOut?
4. Check: Is Hardware responding to lamp commands?
```

### If FSM doesn't move:
```
1. Check: Is ordersToElevator channel populated?
2. Check: Is FSM state actually IDLE?
3. Check: Is setMotor() being called?
4. Check: Is motor hardware responding?
```

### If network doesn't work:
```
1. Check: Is SENDER goroutine running?
2. Check: Is RECEIVER goroutine running?
3. Check: Is network port 1338 available?
4. Check: Is UDP broadcast enabled on interface?
```

### If orders aren't saved:
```
1. Check: Is SaveCabOrders() being called?
2. Check: File permissions in directory
3. Check: Is file actually being written?
4. Check: Is LoadCabOrders() called at startup?
```

---

## 📚 REFERENCE IMPLEMENTATION HINTS

### Go Channel Pattern:
```go
// Use non-blocking sends with select
select {
case ch <- data:
    // Sent successfully
case <-time.After(timeout):
    // Timeout
}

// Receive loop
for {
    select {
    case data := <-ch:
        handleData(data)
    case <-ticker.C:
        doPeriodicWork()
    }
}
```

### JSON Marshaling:
```go
// Create message
msg := Message{SenderId: id, Floor: 2}

// Marshal to JSON
data, _ := json.Marshal(msg)

// Send over UDP
conn.WriteTo(data, addr)

// Receive and unmarshal
n, _, _ := conn.ReadFrom(buf[:])
var msg Message
json.Unmarshal(buf[:n], &msg)
```

### File I/O:
```go
// Write
ioutil.WriteFile("file.txt", []byte("data"), 0644)

// Read
data, _ := ioutil.ReadFile("file.txt")
str := string(data)
```

---

## 🎓 WHY EACH DESIGN CHOICE

| Choice | Why |
|--------|-----|
| CAB persistence | Only way to guarantee "no calls lost" |
| Distance-based assign | Simple, robust, no external deps |
| Idempotent broadcasts | Packet loss becomes transparent |
| 3s timeout | "Seconds" magnitude per spec |
| Go + goroutines | Natural parallelism for I/O |
| UDP broadcast | Simpler than TCP, suitable for LAN |
| Separate modules | Easier to test and debug independently |

---

## 📝 SUBMISSION CHECKLIST

Before submitting PDD:

```
[ ] Filnavn: PDD-##.pdf (group number, not desk)
[ ] PDF format (not Word/Google Docs)
[ ] Lab time listed
[ ] Desk number listed  
[ ] Group number listed
[ ] All group members named + email
[ ] Fault tolerance strategy explained
[ ] Network topology described
[ ] Why Go explained
[ ] Module breakdown included
[ ] < 1 page text (diagrams don't count)
[ ] Diagrams/figures clear (hand-drawn OK)
[ ] Addresses all 5 critical questions
[ ] Uploaded to Blackboard before deadline
```

---

## 🚀 YOUR NEXT STEPS

### Today:
1. Read PDD_PRELIMINARY_DESIGN.md
2. Fill in your group details
3. Submit as PDF to Blackboard

### This Week:
1. Read IMPLEMENTASJONSGUIDE.md
2. Understand the 3 modules
3. Review SCENARIO_WALKTHROUGHS.md

### Next Week:
1. Start implementing Modul 1 (FSM)
2. Refer to pseudokode in guide
3. Test single elevator first

### Key Success Factors:
```
✓ Understand the FSM deeply
✓ Get persistent CAB storage working first
✓ Test one module at a time
✓ Use the sequence diagrams to understand flows
✓ Document as you code
```

---

## 💡 FINAL WISDOM

**Don't:**
- ❌ Try to implement all 3 modules at once
- ❌ Skip the FSM state machine
- ❌ Forget about CAB persistence
- ❌ Use blocking I/O for network
- ❌ Submit without testing single elevator first

**Do:**
- ✅ Test FSM independently first
- ✅ Verify CAB saves to disk
- ✅ Use channels properly (non-blocking where possible)
- ✅ Handle timeout scenarios
- ✅ Document your design decisions

---

## 🎉 YOU'RE READY!

You have:
- ✅ Complete design document
- ✅ UML diagrams
- ✅ Pseudokode for implementation
- ✅ Timing analysis
- ✅ Scenario walkthroughs
- ✅ Testing strategy

Everything you need to build a robust, fault-tolerant elevator system!

**Good luck! 🚀**

