# Update 26 Jan 2026 — Solution 3 Integration (UML Deltas)

Added components and relationships (textual UML augmentation):
- Component `FaultMonitor`: observes `ElevatorBehaviour`, timers for `MotorStopTimeout` and `ObstructionTimeout`.
- Component `OnlineStatus`: toggled by `FaultMonitor` to pause/resume broadcasting when obstructed.
- Interface `MessageEnvelope { TypeId, PayloadJSON }`: used by `Network.Transmitter/Receiver` for type-safe routing.
- Configuration `Network.broadcastIntervalMs = 5` (tunable); feeds `Network.Transmitter` ticker.

Data Flow updates:
- `FaultMonitor` → `OnlineStatus(false)` on obstruction; `OnlineStatus(true)` when cleared.
- `Network.Receiver` filters by `TypeId`; dispatches to appropriate handlers.

Non-adopted elements:
# Specs Coverage Snapshot
- No calls lost: CAB persistence modeled in INIT/RECOVERY flows.
- Button lights: State machine includes assignment→completion transitions and light rules.
- Fault tolerance: FaultMonitor and OnlineStatus reflect graceful handling.
- Packet loss: Receiver routing with Envelope avoids misparse; idempotent application.
- Scalability: HRA assignment visible in sequence diagrams.
- `AckTableConsensus` not included; we retain idempotent convergence with HRA.
- `AutoReboot` not included; recovery via persisted CAB orders.

# UML Diagrams for Preliminary Design

## 1. ELEVATOR STATE MACHINE DIAGRAM

```
                              ┌─────────────────┐
                              │   INIT_STATE    │
                              └────────┬────────┘
                                       │
                        [Load CAB orders from disk]
                                       │
                                       ▼
                          ╔═════════════════════╗
                          ║     IDLE            ║
                          ║                     ║
                          ║ • Motor = STOP      ║
                          ║ • Door = CLOSED     ║
                          ║ • Wait for order    ║
                          ╚════════┬════════════╝
                                   │
                    ┌──────────────┼──────────────┐
                    │              │              │
        [New CAB or Hall    [Obstruction]  [No orders]
         order assigned]              │         │
                    │                 │         └─────────┐
                    ▼                 ▼                     │
              ┌───────────┐      ┌──────────────────┐      │
              │ MOVING    │      │ EMERGENCY_STOP   │      │
              │           │      │ (Door obstruct)  │      │
              │ • Motor=  │      │                  │      │
              │   UP/DOWN │      │ • Wait for       │      │
              │ • Drive   │      │   clearance      │      │
              │   to      │      │ • Repeat timers  │      │
              │   floor   │      └──────────┬───────┘      │
              │           │                 │              │
              └─────┬─────┘                 │              │
                    │                       │              │
         [Reached target floor]             │              │
                    │                   [Obstruction      │
                    │                    cleared]         │
                    ▼                       │              │
              ┌──────────────┐              │              │
              │  DOOR_OPEN   │◄─────────────┘              │
              │              │                             │
              │ • Door open  │                             │
              │ • Wait 3s    │                             │
              │ • Check for  │                             │
              │   direction  │                             │
              │   change     │                             │
              └──────┬───────┘                             │
                     │                                    │
       [3s passed OR direction needed]                    │
                     │                                    │
                     └────────────────────────────────────┘

════════════════════════════════════════════════════════════
TRANSITIONS EXPLAINED:

IDLE → MOVING:
  • New order assigned AND floor != current_floor
  • Set motor direction based on order
  • No obstruction active

MOVING → DOOR_OPEN:
  • Reached target floor
  • Check: Is this a stop point? (orders at this floor AND right direction)
  • Open door
  
DOOR_OPEN → MOVING:
  • 3 seconds elapsed
  • Next order at different floor
  • Set new motor direction

DOOR_OPEN → IDLE:
  • 3 seconds elapsed
  • No more orders
  • Motor stays stopped

ANY → EMERGENCY_STOP:
  • Obstruction detected while door opening/open
  • Motor stops immediately
  • Wait for obstruction to clear
  • Resume from DOOR_OPEN state

════════════════════════════════════════════════════════════
```

---

## 2. BUTTON LIGHT STATE MACHINE

```
      ┌─────────────────────────────────────┐
      │ BUTTON_LIGHT_STATES                 │
      │ (Hall & Cab orders)                 │
      └─────────────────────────────────────┘

      ┌──────────┐
      │ STANDBY  │ (Light OFF)
      │          │
      └────┬─────┘
           │
    [Button pressed]
           │
           ▼
    ┌──────────────────┐
    │ BUTTON_PRESSED   │ (Light ON)
    │                  │
    │ Broadcast to     │
    │ network          │
    └────┬─────────────┘
         │
  [Consensus reached:  
   This elevator wins  
   conflict resolution]
         │
         ▼
    ┌─────────────────┐
    │ ORDER_ASSIGNED  │ (Light ON)
    │                 │
    │ Elevator moving │
    │ to floor        │
    └────┬────────────┘
         │
  [Reached floor AND
   opened door]
         │
         ▼
    ┌──────────────────┐
    │ ORDER_COMPLETE   │ (Light OFF)
    │                  │
    │ Door closing     │
    └────┬─────────────┘
         │
  [Door fully closed
   AND confirmed by
   all network peers]
         │
         ▼
    ┌──────────────┐
    │ STANDBY      │ (Light OFF)
    │ (Back to initial)
    └──────────────┘

════════════════════════════════════════════════════════════
CAB vs HALL DIFFERENCES:

CAB Order:
  • Light ON immediately at BUTTON_PRESSED
  • Only local elevator involved
  • Stored persistently to disk
  
HALL Order:
  • Light ON at ORDER_ASSIGNED (after network consensus)
  • Multiple elevators involved
  • Light synchronized across all panels
  
════════════════════════════════════════════════════════════
```

---

## 3. CLASS/COMPONENT DIAGRAM

```
┌─────────────────────────────────────────────────────────────┐
│                       MAIN APPLICATION                      │
└────────────────────────┬──────────────────────────────────┬─┘
                         │                                  │
          ┌──────────────▼────────┐         ┌──────────────▼──┐
          │                       │         │                 │
          │  ElevatorFSM          │         │ NetworkModule   │
          │  ──────────────       │         │ ──────────────  │
          │                       │         │                 │
          │ - state: State        │         │ - local_id: int │
          │ - current_floor: int  │         │ - elevator_list │
          │ - direction: Dir      │         │   [N]Elevator   │
          │ - door_open: bool     │         │ - alive_list    │
          │ - obstruction: bool   │         │   [N]bool       │
          │                       │         │ - hall_orders   │
          │ + run()               │         │   [N][F][2]     │
          │ + on_floor_reached()  │         │                 │
          │ + on_button_pressed() │         │ + broadcast()   │
          │ + on_door_timer()     │         │ + receive()     │
          │ + check_orders()      │         │ + detect_dead() │
          │                       │         │ + update_alive()│
          └──────────┬────────────┘         └────────┬────────┘
                     │                               │
                     └───────────────┬───────────────┘
                                     │
                     ┌───────────────▼────────────────┐
                     │                                │
                     │   OrderManager                 │
                     │   ────────────                 │
                     │                                │
                     │ - my_id: int                   │
                     │ - cab_orders[F]: bool          │
                     │ - hall_state[N][F][2]:         │
                     │   ButtonLightState             │
                     │ - consensus_acks[N]: bool      │
                     │                                │
                     │ + assign_orders()              │
                     │ + resolve_conflict()           │
                     │ + persist_cab_orders()         │
                     │ + load_cab_orders()            │
                     │ + update_lights()              │
                     │ + handle_disconnect()          │
                     │ + check_consensus()            │
                     │                                │
                     └────────────┬─────────────────┘
                                  │
                    ┌─────────────┴──────────────┐
                    │                            │
            ┌───────▼────────┐      ┌───────────▼──┐
            │                │      │               │
            │ HardwareDriver │      │ FileStorage   │
            │ ────────────   │      │ ───────────   │
            │                │      │               │
            │ + motor_dir()  │      │ + save()      │
            │ + set_door()   │      │ + load()      │
            │ + get_floor()  │      │ + delete()    │
            │ + poll_obs()   │      │               │
            │ + set_lights() │      │               │
            │                │      │               │
            └────────────────┘      └───────────────┘

════════════════════════════════════════════════════════════

ENUMS / TYPES:

enum State {
    INIT_STATE,
    IDLE,
    MOVING,
    DOOR_OPEN,
    EMERGENCY_STOP
}

enum ButtonLightState {
    STANDBY,
    BUTTON_PRESSED,
    ORDER_ASSIGNED,
    ORDER_COMPLETE
}

enum Direction {
    DOWN = -1,
    STOP = 0,
    UP = 1
}

struct Elevator {
    int id
    int floor
    Direction direction
    State state
    bool[F] cab_orders
}

struct Message {
    int sender_id
    Elevator[N] elevator_list
    bool[N] online_status
    bool[N] alive_list
}

════════════════════════════════════════════════════════════
```

---

## 4. UDP MESSAGE ENVELOPE (Solution 3 Integration)

```
╔══════════════════════════════════════════════════════════════════════╗
║                    TYPE-TAGGED JSON ENVELOPE                        ║
║                    (Safe message routing)                           ║
╚══════════════════════════════════════════════════════════════════════╝

┌──────────────────────────────────────────────────────────────────────┐
│ UDP Packet (port 1338, 5ms interval)                                │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  {                                                                  │
│    "TypeId": "dataenums.Message",                                  │
│    "PayloadJSON": [                                                │
│      123, 34, 115, 101, 110, 100, 101, ... (bytes)                │
│    ]                                                               │
│  }                                                                  │
│                                                                      │
│  WHERE PayloadJSON when decoded is:                                │
│  {                                                                  │
│    "SenderId": 0,                                                  │
│    "ElevatorList": [                                               │
│      {                                                             │
│        "CurrentFloor": 2,                                          │
│        "Direction": 1,  // UP=-1, STOP=0, UP=1                    │
│        "Requests": [[false, false], ...],                          │
│        "CurrentBehaviour": 0, // IDLE=0, DOOROPEN=1, MOVING=2    │
│        "OnlineStatus": true                                        │
│      },                                                            │
│      ...                                                           │
│    ],                                                              │
│    "HallOrderList": [[[false, false], ...], ...],                 │
│    "OnlineStatus": true,                                           │
│    "AliveList": [true, true, true]                                │
│  }                                                                  │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘

ADVANTAGES:
✓ Type filtering: Receiver checks TypeId; ignores unknown types
✓ Version evolution: Can add new message types without breaking old
✓ Safe parsing: No ambiguity about message format
✓ Idempotent: Same message received multiple times = same effect

BROADCAST INTERVAL: 5ms (tunable from config/config.go)
TIMEOUT: 3000ms = 600 messages without seeing a peer
FAILURE DETECTION: HeartbeatTimeout = 3s

════════════════════════════════════════════════════════════════════════
```

---

## 5. SEQUENCE DIAGRAM - Normal Hall Call Operation (Idempotent Model)

```
Hall Button pressed at floor 3, UP direction

┌────────┐  ┌────────┐  ┌────────┐
│ Elev A │  │ Elev B │  │ Elev C │
│(Floor 1)  │(Floor 3)  │(Floor 2)
└───┬────┘  └───┬────┘  └───┬────┘
    │           │           │
    │ Hall Button [3, UP] Pressed
    │           │           │
    │ State update: hall_orders[3][UP] = BUTTON_PRESSED
    │ Light ON locally        │
    │                        │
    │ Every 5ms: Broadcast State (Envelope)
    ├──────────────────────►  │
    ├──────────────────────►  │
    │                        │
    │ [B & C receive same envelope]
    │                        │
    │ [Idempotent application: same state stored regardless of order/dupes]
    │                        │
    │ Both apply: hall_orders[3][UP] = BUTTON_PRESSED
    │ Light ON on all panels │
    │                        │
    │ [Every elevator computes HRA independently]
    │ [Using global elevator_list from broadcasts]
    │                        │
    │ HRA Output: "Elevator B wins" (nearest to floor 3)
    │                        │
    │ B: hall_orders[3][UP] = ORDER_ASSIGNED
    │ B Motor = UP            │
    │ B starts moving...      │
    │                        │
    │ [B broadcasts every 5ms: floor updates]
    │                        │
    │ ... 10 seconds later ...│
    │                        │
    │ B: Reached floor 3     │
    │ B: state = DOOR_OPEN   │
    │ B: Broadcast           │
    │ ├──────────────────────┼──────────┐
    │ │                      │          │
    │ ◄─────────────────────┤ ◄─────────┤
    │ A & C receive update   │          │
    │ Update: B at floor 3, door open   │
    │                       │          │
    │ [3 seconds pass]      │          │
    │                       │          │
    │ B: Door closes        │          │
    │ B: hall_orders[3][UP] = COMPLETE │
    │ B: Light OFF locally  │          │
    │ B: Broadcast          │          │
    │ ├──────────────────────┼──────────┐
    │ │                      │          │
    │ ◄─────────────────────┤ ◄─────────┤
    │ A & C receive: COMPLETE           │
    │ Light OFF on all panels           │
    │ hall_orders[3][UP] = STANDBY      │
    │                       │          │

════════════════════════════════════════════════════════════════════════
KEY DIFFERENCES FROM ACK MODEL:
- No explicit replies needed
- Same message can arrive out-of-order, duplicated, or lost
- Idempotent application means result is always correct
- Convergence guaranteed by periodic broadcasts (5ms)
- Packet loss handled transparently (next broadcast carries same state)

════════════════════════════════════════════════════════════════════════
```

---

## 6. SEQUENCE DIAGRAM - Network Disconnect & Takeover (Idempotent Recovery)

```
Network cable disconnected (Elevator A loses connection)

        ┌────────┐         ┌────────┐         ┌────────┐
        │ Elev A │         │ Elev B │         │ Elev C │
        │ (disc) │         │        │         │        │
        └───┬────┘         └───┬────┘         └───┬────┘
            │                  │                  │
   [Hall order at floor 2]      │                  │
   Active: ORDER_ASSIGNED       │                  │
   (Elev A was serving it)       │                  │
            │                  │                  │
        [Broadcast stops]       │                  │
            │                  │                  │
            │              Heartbeat timeout      │
            │              (3 seconds)            │
            │                  │                  │
            │                  ├◄──────────────────┤
            │                  │ A missing!        │
            │                  │                  │
            │            [UPDATE ALIVE_LIST]      │
            │            alive_list[A] = FALSE    │
            │                  │                  │
            │      [REASSIGN VIA HRA]       │
            │      floor 2 UP order:        │
            │      Global elevator state:   │
            │      - B distance: 1, available
            │      - C distance: 3, available
            │      HRA output: B wins       │
            │                  │                  │
            │                  ├─────────────────►│
            │                  │ Periodic broadcast│
            │                  │ (5ms intervals)   │
            │                  │ A=offline, B assigned
            │                  │                  │
            │                  ├◄─────────────────┤
            │                  │ Same state       │
            │                  │ (idempotent)     │
            │                  │                  │
            │      B moves to floor 2       │
            │      Opens door...            │
            │                  │                  │
        [Network restored after 10s]       │
            │                  │                  │
            │ ─────Broadcasts resume────►  │
            │ "I'm alive!"                  │
            │ My CAB orders restored        │
            │ from disk                     │
            │                  │                  │
            │                  ├◄──────────────────┤
            │                  │ Welcome back      │
            │                  │ Your current state│
            │                  │ (idempotent)      │
            │                  │                  │
        [State synchronized via periodic  │
         5ms broadcasts]                  │

════════════════════════════════════════════════════════════

KEY: Idempotent design means network recovery is automatic
- No explicit handshakes needed
- Repeated broadcasts converge all nodes
- Packet loss/reorder/duplication transparent

════════════════════════════════════════════════════════════
```

---

## 6. SEQUENCE DIAGRAM - Crash Recovery with Persistence

```
Elevator A crashes while serving CAB order

        BEFORE CRASH:
        ┌────────────────────────────────┐
        │ cab_orders[A][2] = assigned    │
        │ Moving to floor 2              │
        │ [Saved to disk: cab_orders_0]  │
        └────────────────────────────────┘

        CRASH EVENT:
        │
        Power loss / Software crash
        │
        All state lost in memory
        │
        ┌────────────────────────────────┐
        │ Power returns / Restart         │
        └────────────────────────────────┘
        │
        On startup:
        │
        ├─ ReadCabOrderBackup()
        │  Load from: cab_orders_0.txt
        │
        ├─ Found: cab_orders[2] = active
        │
        ├─ Broadcast: "I'm back with CAB order at floor 2"
        │
        ├─ Network acknowledges
        │
        ├─ Resume serving order
        │  (Continue moving to floor 2 if not there yet)
        │
        └─ Complete order normally

        RESULT: ✓ CAB order NOT lost

════════════════════════════════════════════════════════════
```

---

## 7. UDP BROADCAST NETWORK TOPOLOGY (Solution 3 Integration)

```
┌────────────────────────────────────────────────────────────────────┐
│                    Local Network (192.168.x.x)                    │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│         ┌─────────────────┐                                        │
│         │  Switch / Hub   │  UDP Broadcast: 255.255.255.255:1338  │
│         │                 │  Interval: 5ms (tunable)              │
│         │                 │  Timeout: 3000ms per peer             │
│         └────┬──────┬──────┴──────────────┐                        │
│              │      │                     │                        │
│      ┌───────▼─┐ ┌──▼────────┐    ┌──────▼────┐                  │
│      │ Elev A  │ │  Elev B   │    │ Elev C   │                   │
│      │ ID: 0   │ │  ID: 1    │    │ ID: 2   │                   │
│      │         │ │           │    │          │                   │
│      └────┬────┘ └────┬──────┘    └──────┬───┘                   │
│           │           │                  │                        │
│           │ Broadcasts every 5ms:        │                        │
│           │ Envelope{TypeId, Payload}    │                        │
│           │                              │                        │
│         ┌─┴──────────────┬───────────────┴─┐                     │
│         │                │                 │                     │
│    Elev A              All               Elev C                   │
│    sends          receive &              sends                   │
│    state         apply idempotently      state                   │
│    update          state                 update                  │
│                                                                   │
│  [Online] ←──Heartbeat timeout──→ [Offline]                      │
│   3s                                after 3s                      │
│   active         re-broadcast to   no msg                         │
│   broadcast      network            mark offline                  │
│                                                                   │
└────────────────────────────────────────────────────────────────────┘

═══════════════════════════════════════════════════════════════════════

TYPE-TAGGED ENVELOPE STRUCTURE:

{
  "TypeId": "dataenums.Message",
  "PayloadJSON": [... binary ...]
}

RECEIVER FILTER:
1. Parse outer JSON → extract TypeId
2. If TypeId = "dataenums.Message":
     Unmarshal PayloadJSON as Message struct
3. Else:
     Ignore (for future extensibility)

═══════════════════════════════════════════════════════════════════════

MESSAGE PAYLOAD (decoded from PayloadJSON):

{
  "SenderId": 0,
  "ElevatorList": [
    {
      "CurrentFloor": 2,
      "Direction": 1,  // -1=DOWN, 0=STOP, 1=UP
      "Requests": [[false,true], [true,false], ...],
      "CurrentBehaviour": 1,  // 0=IDLE, 1=DOOROPEN, 2=MOVING
      "OnlineStatus": true    // Adopted: motor/obstruction detection
    },
    ... (up to N=3 elevators)
  ],
  "HallOrderList": [  // [elevator][floor][direction]
    [[false,false], [false,false], [false,true], [false,false]],
    ... (per elevator)
  ],
  "OnlineStatus": true,
  "AliveList": [true, true, true]
}

═══════════════════════════════════════════════════════════════════════

BROADCAST CYCLE (5ms):

Timer fires (5ms)
  ↓
Transmitter: Read current state from Order Manager
  ↓
Build Message struct with:
  • My floor, direction, state
  • My online_status (from FaultMonitor)
  • All CAB orders
  • Synced hall orders
  ↓
Wrap in Envelope{TypeId, PayloadJSON}
  ↓
Send via UDP to 255.255.255.255:1338
  ↓
All peers receive (multicast-like on local broadcast)
  ↓
Receiver: Parse Envelope, check TypeId
  ↓
Unmarshal Message payload
  ↓
Apply state idempotently:
  • Update elevator_list[sender_id] = received.elevator
  • Merge hall_orders (union of all active orders)
  • Update online_status for sender
  ↓
Update heartbeat timestamp for sender
  ↓
Dispatch to Order Manager via channel
  ↓
Order Manager computes HRA, updates light state
  ↓
[Loop: repeat every 5ms]

Heartbeat check (parallel, every 500ms):
  For each peer:
    If now - last_seen > 3000ms:
      Mark offline
      Trigger HRA reassignment

═══════════════════════════════════════════════════════════════════════

IDEMPOTENT PROPERTIES:

1. Duplicate message: apply twice = apply once
   - hall_orders[0][2][UP] = true
   - Apply same message twice → still true

2. Out-of-order delivery: doesn't matter
   - Message A: floor=1, state=IDLE
   - Message B: floor=2, state=MOVING
   - Arrive as B, then A → final state is A (latest always wins)

3. Packet loss: transparent
   - Cycle 1: send state S1
   - Cycle 2: lost
   - Cycle 3: send state S1 again → received, idempotent
   - Result: convergence within 10ms worst-case

═══════════════════════════════════════════════════════════════════════
```
```

