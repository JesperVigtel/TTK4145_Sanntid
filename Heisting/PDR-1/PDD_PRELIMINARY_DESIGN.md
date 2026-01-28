# PRELIMINARY DESIGN DESCRIPTION

## TTK4145 Elevator Project — 2025

**Lab Session:** [Your lab time]  
**Workstation/Desk:** [Your workstation number]  
**Group Number:** [Your group number]

---

## Group Members
- Name: [Name], Email: [email@ntnu.no]
- Name: [Name], Email: [email@ntnu.no]
- Name: [Name], Email: [email@ntnu.no]

---

## 1. Design Strategy Overview

This solution combines **distributed P2P UDP mesh networking** with **local persistent state storage** to guarantee no orders are lost. The system is decomposed into three independent modules:
- **Elevator FSM** (state machine & physical operations)
- **Network Module** (peer discovery, broadcast, consensus)
- **Order Manager** (CAB/hall assignment, light state, persistence)

**Key Innovation:** Hybrid assignment strategy using **Hall Request Assigner (HRA)** for global optimality while maintaining **idempotent broadcasts** and **persistent CAB storage** for fault tolerance.

---

## 2. Fault Tolerance Strategy

### Persistence: CAB Order Backup
- CAB orders are saved to disk after every change: `cab_orders_[ID].txt`
- On crash or power loss, CAB orders are restored from disk at startup
- **Guarantee:** No CAB calls are lost, even after unexpected shutdown

### Offline Resilience
- When network is lost, each elevator operates as standalone unit
- CAB orders are served locally; doors remain closed for hall calls
- Hall orders are automatically reassigned to active peers within **3-second heartbeat timeout**

### Consistency: Idempotent Broadcasts
- Each elevator broadcasts complete state periodically (5ms interval, configurable)
- Contains: `ID`, `floor`, `direction`, `state`, `cab_orders`, `online_status`
- **Idempotent property:** Same message received multiple times has no additional effect
- **Result:** Packet loss, duplication, and out-of-order delivery are transparent to application

### Button Light Contract
- CAB lights turn on **immediately** upon button press (local)
- Hall lights turn on when order is **ORDER_ASSIGNED** to an elevator
- Lights turn off only when order is **fully serviced** across all active peers
- Light state is synchronized via idempotent network propagation

---

## 3. Network Topology & Protocol

**Topology:** UDP broadcast mesh on local network (port 1338)

**Message Envelope (Solution 3 improvement):**
```go
type Envelope struct {
    TypeId      string  // Identifies message type
    PayloadJSON []byte  // Serialized state
}
```
This ensures safe routing and type filtering without parsing all messages.

**Broadcast Interval:** 5ms (tunable, faster than L1's 50ms for improved convergence)

**Heartbeat Protocol:**
- Each elevator sends state message with type-tagged envelope every 5ms
- Fields: `elevator_id`, `current_floor`, `direction`, `behaviour`, `cab_orders`, `online_status`
- Timeout: 3000ms — elevators not heard from are marked offline
- Failed elevators' hall orders are redistributed using HRA

**Consensus Mechanism:**
- Each elevator maintains a local mirror of all peers' states
- When all active peers report the same state → consensus achieved
- Order Manager then invokes HRA assigner with global view

---

## 4. Programming Language: Go

**Rationale:**
1. **Goroutines** enable true parallelism for sensors, network, and timers
2. **Channels** provide thread-safe, deadlock-free IPC
3. **Performance & Determinism** critical for real-time constraints
4. **Built-in UDP** support without heavy dependencies

---

## 5. System Architecture

```
┌─────────────────────────────────────┐
│      Main Coordinator               │
│  (goroutine orchestration,          │
│   channel multiplexing)             │
└──────────┬──────────────┬───────────┘
           │              │
     ┌─────▼──────┐  ┌────▼────────┐
     │ Elevator   │  │   Network   │
     │    FSM     │  │   Module    │
     │            │  │             │
     │ • 5 states │  │ • Sender    │
     │ • Door mgmt│  │ • Receiver  │
     │ • Motor    │  │ • Registry  │
     │ • Fault    │  │ • Online/   │
     │   detection│  │   Offline   │
     └─────┬──────┘  └────┬────────┘
           │              │
           └──────┬───────┘
                  │
          ┌───────▼────────┐
          │  Order Manager │
          │                │
          │ • CAB persist  │
          │ • HRA assign   │
          │ • Light state  │
          │ • Consensus    │
          │ • Fault monitor│
          └────────────────┘
```

---

## 6. Module Descriptions

### 6.1 Elevator FSM
**States:** `INIT_STATE`, `IDLE`, `MOVING`, `DOOR_OPEN`

Manages physical elevator operations:
- Accepts orders from Order Manager
- Reports floor sensor data and door status
- Transitions between states based on sensor inputs and orders

**Obstruction Handling:** If door is obstructed, restart 3-second timer. Report obstruction status to Order Manager.

**Motor/Obstruction Fault Detection (Solution 3 adoption):**
- Motor timeout: ~4 seconds (triggers graceful offline state)
- Obstruction timeout: 5 seconds (triggers network disconnect + door close)
- Graceful online/offline toggling to pause broadcasting during faults

### 6.2 Network Module
- Sends own state every 5ms with type-tagged JSON envelope
- Receives and validates messages from all peers
- Maintains `alive_list` and detects online/offline transitions
- **Failure Detection:** `HeartbeatTimeout = 3s`
  - If elevator not heard from → marked offline
  - Its hall orders are redistributed by HRA to active peers

### 6.3 Order Manager
**Persistent CAB Storage:**
- Saves/loads CAB orders from disk: `cab_orders_[ID].txt`
- Restored on startup or after network recovery

**Order Assignment (Hall Orders) — HRA Strategy:**
- Invokes external `hall_request_assigner` binary with global elevator state
- HRA computes optimal assignment considering:
  - Distance to floor
  - Direction of travel
  - Current and queued orders
  - Scalable to n elevators

**Button Light State Machine:**
- `STANDBY` → `BUTTON_PRESSED` (user presses button)
- `BUTTON_PRESSED` → `ORDER_ASSIGNED` (HRA assigns to elevator)
- `ORDER_ASSIGNED` → `COMPLETED` (elevator opens door at correct floor)
- `COMPLETED` → `STANDBY` (after door closes)

Light turns off when transitioning to `STANDBY`.

---

## 7. Critical Design Decisions

| Decision | Rationale |
|----------|-----------|
| **CAB Persistence** | Guarantees "no calls lost" even after power loss or crash. Only way to meet requirement. |
| **HRA Hall Assignment** | Optimal global assignment minimizes average wait time; robust and proven. |
| **Idempotent Broadcasts** | Packet loss, duplication, and reordering are transparent to application logic. |
| **5ms Broadcast Interval** | Balances convergence speed (faster light response) with network load. |
| **Graceful Offline/Online Toggling** | Motor/obstruction faults pause broadcasting without killing the process; recovery is automatic. |
| **Type-Tagged JSON Envelope** | Safe message routing and version evolution without parsing all payloads. |
| **Local Autonomy per Elevator** | Each elevator can function standalone; hall orders auto-reassign on network loss. |

---

## 8. Handling of Project Challenges

| Challenge | Solution |
|-----------|----------|
| **Button light guarantee** | Immediate CAB light + synced hall light via idempotent broadcasts; state machine ensures consistency. |
| **No orders lost** | Persistent CAB storage + network takeover of hall orders = complete coverage. |
| **Network unreliability** | Idempotent broadcasts, heartbeat-based failure detection, automatic peer takeover. |
| **Elevator crash/power loss** | CAB orders restored from disk; hall orders taken over by peers within 3s. |
| **Hall order on disconnect** | Elevator marked offline; HRA redistributes its orders to active peers. |
| **Packet loss** | Idempotent state propagation makes packet loss transparent. Lights and orders eventually converge. |
| **Motor timeout** | Graceful offline state; door stays closed; hall orders auto-reassign; operator resolves. |
| **Door obstruction** | Repeated door timer restarts; after 5s, network disconnect; lights remain stable via persistence. |

---

## 9. Configuration Parameters

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| `BroadcastRate` | 5ms | Fast convergence without excess CPU. Adopted from Solution 3. |
| `HeartbeatTimeout` | 3000ms | Detect and recover from failures within reasonable time window. |
| `MotorStopTimeout` | 4000ms | Detect stuck elevator; trigger graceful offline. |
| `ObstructionTimeout` | 5000ms | Graceful door obstruction handling before network disconnect. |
| `DoorOpenDuration` | 3000ms | Standard door hold time. |
| `NetworkPort` | 1338 | Fixed broadcast port; avoid conflicts. |

---

## 10. Specs Coverage Verification

✅ **No calls lost:** CAB persistence + network consensus  
✅ **Button light guarantee:** Idempotent state machine + immediate CAB lights  
✅ **Fault tolerance:** Heartbeat + motor/obstruction detection + offline toggling  
✅ **Packet loss resilience:** Idempotent broadcasts + type-tagged envelope  
✅ **Scalability:** HRA scales to n elevators; timing parameters configurable  
✅ **Cross-platform:** No auto-reboot dependency; pure Go + standard UDP

