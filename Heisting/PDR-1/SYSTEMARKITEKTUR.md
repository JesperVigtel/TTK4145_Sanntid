# Update 26 Jan 2026 — Architecture Deltas from Solution 3

Introduced/updated components:
- `FaultMonitor` (new): monitors movement and door states; triggers online/offline.
- `OnlineStatus` (new): gates network broadcasting; reconnects when obstruction clears.
- `MessageEnvelope` (new): type-tagged JSON for network messages.
- `Network.Transmitter` (updated): 5ms configurable ticker.

Design decisions:
- Use HRA for hall orders; keep idempotent convergence (skip ACK table).
- Persist CAB orders for crash-safe recovery; no auto reboot.

Impact on flows:
- Specs Coverage Snapshot:
- No calls lost: CAB persistence component ensures recovery; network takeover handles hall orders.
- Button lights: Deterministic light state via idempotent propagation and faster broadcast.
- Fault tolerance: Heartbeat + FaultMonitor timeouts; graceful offline/online toggling.
- Packet loss: Periodic, idempotent messages; envelope-based routing.
- Scalability: HRA assignment; configurable timing.
- Faults can temporarily set node offline; lights and CAB remain consistent via persistence.
- Faster broadcast improves button light convergence and peer awareness.

# SYSTEMARKITEKTUR & DESIGNDOKUMENT

## 1. FULL SYSTEM ARCHITECTURE OVERVIEW

```
┌────────────────────────────────────────────────────────────────┐
│                     MAIN COORDINATOR                           │
│                                                                │
│  Channels setup + Module initialization                       │
│  Goroutine management                                          │
└────────────────────────────────────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
    ┌────────┐        ┌────────┐       ┌──────────┐
    │Hardware│        │Elevator│       │ Network  │
    │ Polling│        │  FSM   │       │ Module   │
    └────────┘        └────────┘       └──────────┘
        │                 │                 │
        │ Sensor Data     │ Status         │ Messages
        │                 │                 │
        └─────────────────┼─────────────────┘
                          │
                          ▼
                  ┌──────────────────┐
                  │ Order Manager    │
                  │ (Central Brain)  │
                  │                  │
                  │ • CAB Persist    │
                  │ • Hall Assign    │
                  │ • Light State    │
                  │ • Consensus      │
                  └──────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
    ┌────────┐        ┌────────┐       ┌──────────┐
    │Elevator│        │ Lights │       │File I/O  │
    │ Output │        │ Output │       │(Persist) │
    └────────┘        └────────┘       └──────────┘

════════════════════════════════════════════════════════════════
```

---

## 2. CHANNEL COMMUNICATION MAP

```
HARDWARE SENSORS → FSM
  ├─ floorSensor (int)           : "I reached floor 2"
  ├─ obstructionSwitch (bool)    : "Obstruction detected"
  └─ buttonPress (ButtonEvent)   : "Button pressed at floor 3, hall up"

FSM → ORDER MANAGER
  └─ elevatorStatus              : "I'm at floor 2, idle"

NETWORK ↔ ORDER MANAGER
  ├─ networkWorldview            : "Global elevator states"
  └─ networkState                : "My local state to broadcast"

NETWORK ↔ NETWORK
  ├─ broadcastOut → UDP broadcast
  └─ broadcastIn ← UDP receive

ORDER MANAGER → FSM
  └─ ordersToElevator [F][2]bool : "Go to floors {1,2,3}"

ORDER MANAGER → LIGHTS
  └─ lightsCommand               : "Turn on light floor 2, hall up"

ORDER MANAGER → DISK
  └─ SaveCabOrders()             : "Save CAB order state to file"
  └─ LoadCabOrders()             : "Restore CAB orders from file"

════════════════════════════════════════════════════════════════
```

---

## 3. COMPONENT INTERACTION SEQUENCE: HALL CALL

```
SCENARIO: Hall call button pressed at floor 3

TIME: t=0ms
┌─────────────────────────────────────────────────────────────┐
│ HARDWARE EVENT                                              │
│ ├─ Button pressed: floor 3, HALL_UP                         │
│ └─ Signal sent to buttonPress channel                       │
└─────────────────────────────────────────────────────────────┘

TIME: t=5ms
┌─────────────────────────────────────────────────────────────┐
│ ORDER MANAGER receives button event                         │
│ ├─ Create new hall order: orders[3][HALL_UP] = PENDING     │
│ ├─ Broadcast request to network                             │
│ └─ Tentatively turn on light (local)                        │
└─────────────────────────────────────────────────────────────┘

TIME: t=10ms
┌─────────────────────────────────────────────────────────────┐
│ NETWORK MODULE                                              │
│ ├─ Send broadcast: "Order at floor 3, UP"                   │
│ ├─ All elevators receive and acknowledge                    │
│ └─ Wait for consensus                                       │
└─────────────────────────────────────────────────────────────┘

TIME: t=50ms
┌─────────────────────────────────────────────────────────────┐
│ CONSENSUS REACHED                                           │
│ ├─ All elevators agree on state                             │
│ ├─ ORDER MANAGER runs conflict resolution                   │
│ │   - Elevator A: distance = 2 floors              [WINNER] │
│ │   - Elevator B: distance = 5 floors                       │
│ │   - Elevator C: distance = 4 floors                       │
│ └─ Elevator A assigned to take order                        │
└─────────────────────────────────────────────────────────────┘

TIME: t=50ms (Elevator A)
┌─────────────────────────────────────────────────────────────┐
│ ORDER MANAGER → FSM                                         │
│ ├─ Send: ordersToElevator[3][HALL_UP] = true               │
│ ├─ FSM receives new order                                   │
│ ├─ Calculates direction: currently floor 1, target floor 3  │
│ ├─ direction = UP                                           │
│ └─ State transition: IDLE → MOVING                          │
└─────────────────────────────────────────────────────────────┘

TIME: t=55ms (Elevator A)
┌─────────────────────────────────────────────────────────────┐
│ FSM → Hardware                                              │
│ ├─ setMotor(UP)                                             │
│ ├─ Elevator starts moving up                                │
│ └─ Wait for floor sensor                                    │
└─────────────────────────────────────────────────────────────┘

TIME: t=200ms (Elevator A reaches floor 3)
┌─────────────────────────────────────────────────────────────┐
│ HARDWARE → FSM                                              │
│ ├─ floorSensor channel receives: 3                          │
│ ├─ FSM checks: shouldStop() == true (order at floor 3)      │
│ └─ State transition: MOVING → DOOR_OPEN                     │
└─────────────────────────────────────────────────────────────┘

TIME: t=205ms
┌─────────────────────────────────────────────────────────────┐
│ FSM → Hardware                                              │
│ ├─ setMotor(STOP)                                           │
│ ├─ setDoor(OPEN)                                            │
│ ├─ Start doorTimer(3000ms)                                  │
│ └─ FSM → ORDER MANAGER: "I opened door at floor 3"          │
└─────────────────────────────────────────────────────────────┘

TIME: t=210ms
┌─────────────────────────────────────────────────────────────┐
│ ORDER MANAGER                                               │
│ ├─ Receive: elevatorStatus = {floor: 3, state: DOOR_OPEN}   │
│ ├─ Update local state                                       │
│ ├─ Send to NETWORK: "My door is open at floor 3"            │
│ ├─ Set light[3][HALL_UP] to true (already was)             │
│ └─ Wait for door timer                                      │
└─────────────────────────────────────────────────────────────┘

TIME: t=215ms
┌─────────────────────────────────────────────────────────────┐
│ NETWORK BROADCAST                                           │
│ ├─ All elevators see: Elevator A at floor 3, door open      │
│ └─ All acknowledge                                          │
└─────────────────────────────────────────────────────────────┘

TIME: t=3205ms (3 seconds after door opened)
┌─────────────────────────────────────────────────────────────┐
│ DOOR TIMER FIRED                                            │
│ ├─ FSM receives: doorTimer.C                                │
│ ├─ setDoor(CLOSED)                                          │
│ ├─ State transition: DOOR_OPEN → IDLE                       │
│ └─ Send status: "I'm idle at floor 3, door closed"          │
└─────────────────────────────────────────────────────────────┘

TIME: t=3210ms
┌─────────────────────────────────────────────────────────────┐
│ ORDER MANAGER                                               │
│ ├─ Receive: elevatorStatus = {floor: 3, state: IDLE}        │
│ ├─ Order at floor 3 is COMPLETED                            │
│ ├─ Clear order: orders[3][HALL_UP] = STANDBY               │
│ ├─ LIGHT OFF                                                │
│ └─ Broadcast completion to network                          │
└─────────────────────────────────────────────────────────────┘

RESULT: ✓ Hall call successfully served
        ✓ Light turned on (t=50ms)
        ✓ Elevator arrived and opened door (t=205ms)
        ✓ Light turned off after door closed (t=3210ms)

════════════════════════════════════════════════════════════════
```

---

## 4. CRITICAL PATHS & TIMINGS

### Path 1: Button Press → Light On (CAB)
```
buttonPress channel (t=0ms)
    ↓
ORDER MANAGER receives (t<1ms)
    ↓
Turn on light locally (t<2ms)
    ↓
Save to disk (t<5ms)

LATENCY: ~2ms ✓ (acceptable)
```

### Path 2: Button Press → Light On (HALL)
```
buttonPress channel (t=0ms)
    ↓
ORDER MANAGER broadcasts (t=5ms)
    ↓
NETWORK broadcasts every 50ms (t=50ms first broadcast)
    ↓
All elevators ACK (t=60ms)
    ↓
Consensus reached, conflict resolution (t=65ms)
    ↓
Winning elevator gets order, turns on light (t=70ms)

LATENCY: ~70ms ✓ (acceptable)
```

### Path 3: Floor Reached → Door Opens
```
Floor sensor fires (t=0ms)
    ↓
FSM receives floorSensor channel (t<1ms)
    ↓
shouldStop() check (t<2ms)
    ↓
setMotor(STOP), setDoor(OPEN) (t<5ms)

LATENCY: ~5ms ✓ (excellent)
```

### Path 4: Network Disconnect → Takeover
```
Elevator A goes offline (t=0ms)
    ↓
No heartbeat from A (t=0ms)
    ↓
Timeout timer counts down (3000ms)
    ↓
TIMEOUT FIRED (t=3000ms)
    ↓
NETWORK marks A as offline (t=3001ms)
    ↓
ORDER MANAGER sees aliveList[A] = false (t=3050ms)
    ↓
Reassigns A's orders to B or C (t=3055ms)
    ↓
New assignment broadcast (t=3100ms)
    ↓
Light on for new assignment (t=3105ms)

LATENCY: ~3100ms ✓ (within requirement: "seconds")
```

### Path 5: Crash → CAB Recovery
```
Elevator crashes (t=0ms)
    ├─ CAB orders still on disk
    └─ Memory lost
        ↓
Restart system (t=5000ms user restarts)
    ↓
LoadCabOrders() reads from disk (t=5010ms)
    ↓
CAB orders restored (t=5015ms)
    ↓
Light on for restored CAB (t=5020ms)
    ↓
Continue serving order (t=5025ms)

LATENCY: ~25ms after restart ✓
```

════════════════════════════════════════════════════════════════

---

## 5. FAULT TOLERANCE TABLE

```
╔════════════════════════╦═══════════╦═══════════╦═══════════╗
║ Fault Scenario         ║ Detection ║ Recovery  ║ CAB Lost? ║
╠════════════════════════╬═══════════╬═══════════╬═══════════╣
║ Network packet loss    ║ Immediate │ Idempotent║    NO     ║
║ (10-30%)              ║           ║ retry     ║           ║
╠════════════════════════╬═══════════╬═══════════╬═══════════╣
║ Network disconnect     ║ 3 sec HB  ║ Takeover  ║    NO     ║
║ (one elevator)         ║ timeout   ║ by others ║           ║
╠════════════════════════╬═══════════╬═══════════╬═══════════╣
║ Network partition      ║ 3 sec HB  ║ Single    ║    NO     ║
║ (multiple groups)      ║ timeout   ║ elevator  ║ (persist) ║
║                        ║           ║ mode      ║           ║
╠════════════════════════╬═══════════╬═══════════╬═══════════╣
║ Software crash         ║ Immediate ║ Restart   ║    NO     ║
║ (single elevator)      ║           ║ + persist ║ (from     ║
║                        ║           ║           ║  disk)    ║
╠════════════════════════╬═══════════╬═══════════╬═══════════╣
║ Power loss             ║ Immediate ║ Restart   ║    NO     ║
║ (single elevator)      ║           ║ + persist ║ (from     ║
║                        ║           ║           ║  disk)    ║
╠════════════════════════╬═══════════╬═══════════╬═══════════╣
║ Door obstruction       ║ Hardware  ║ Extend    ║    NO     ║
║ (while open)           ║ sensor    ║ timeout   ║           ║
║                        ║ immediate ║ (3s×N)    ║           ║
╠════════════════════════╬═══════════╬═══════════╬═══════════╣
║ Hall call from         ║ 3 sec HB  ║ Reassign  ║    NO     ║
║ disconnected elev      ║ timeout   ║ to others ║ (auto)    ║
╚════════════════════════╩═══════════╩═══════════╩═══════════╝
```

════════════════════════════════════════════════════════════════

---

## 6. KEY DESIGN PRINCIPLES

### 1. **Idempotency**
- Sending the same message multiple times = same effect
- Packet loss/duplication is transparent
- No "exactly-once" complexity

### 2. **Local Autonomy**
- Each elevator can work standalone
- Network is for optimization, not required for basic function
- Graceful degradation when network fails

### 3. **Persistent Storage**
- CAB orders saved immediately
- Never lost even on crash/power loss
- Restore on startup

### 4. **Eventual Consistency**
- Not all elevators need to agree instantly
- Agreement happens within ~50-70ms
- Eventually everyone sees the same state

### 5. **Fail-Safe Defaults**
- No order assignment if no consensus
- Lights stay on until order complete
- Conservative on taking orders (no double-taking)

════════════════════════════════════════════════════════════════

---

## 7. COMPARISON WITH REFERENCE SOLUTIONS

### vs. Solution 1 (HRA + Cyclic Counter)
```
PROS:
  ✓ Optimal order assignment (HRA)
  ✓ Very robust cyclic state machine
  ✓ Explicit consistency protocol

CONS:
  ✗ CAB orders lost on crash
  ✗ Depends on external HRA binary
  ✗ More complex

OUR SOLUTION: Combines robustness WITH persistence
```

### vs. Solution 2 (Distance-based + File backup)
```
PROS:
  ✓ Persistent CAB storage
  ✓ No external dependencies
  ✓ Simpler implementation

CONS:
  ✗ Less robust to packet loss
  ✗ Hall orders require sync time
  ✗ Less optimal assignment

OUR SOLUTION: Adds robustness to Solution 2's persistence
```

════════════════════════════════════════════════════════════════

---

## 8. IMPLEMENTATION PRIORITIES

### PHASE 1 (Weeks 1-2): Core FSM
- ✓ Elevator state machine (5 states)
- ✓ Door timer (3 seconds)
- ✓ Motor control
- ✓ Floor sensor integration
- ✓ Obstruction handling

### PHASE 2 (Weeks 3-4): Network
- ✓ UDP broadcast sender/receiver
- ✓ Message serialization (JSON)
- ✓ Heartbeat detection
- ✓ Alive list tracking

### PHASE 3 (Weeks 5-6): Order Management
- ✓ CAB order persistence
- ✓ Hall order assignment (distance-based)
- ✓ Conflict resolution
- ✓ Button light state machine

### PHASE 4 (Weeks 7-8): Integration & Testing
- ✓ End-to-end testing
- ✓ Network failure scenarios
- ✓ Packet loss simulation
- ✓ Multi-elevator testing
- ✓ Performance tuning

════════════════════════════════════════════════════════════════

---

## 9. TESTING CHECKLIST

- [ ] Single elevator: all floors, all buttons
- [ ] Single elevator: door obstruction handling
- [ ] Two elevators: hall call assignment (correct winner)
- [ ] Two elevators: simultaneous orders
- [ ] CAB persistence: restart and recovery
- [ ] Network disconnect: takeover within 3s
- [ ] Network reconnect: successful re-synchronization
- [ ] Packet loss: system stability (10%, 30%, 50%)
- [ ] Multiple restarts: no order loss
- [ ] All 5 state transitions work correctly
- [ ] Light timing: immediate CAB, delayed HALL

════════════════════════════════════════════════════════════════
