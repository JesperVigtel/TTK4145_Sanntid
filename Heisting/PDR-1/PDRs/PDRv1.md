# PRELIMINARY DESIGN DESCRIPTION

## Header Information
**Lab Time:** [Your lab time]  
**Workstation/Desk:** [Your desk number]  
**Group Number:** [0#]  

**Group Members:**
- [Name], [email@ntnu.no]
- [Name], [email@ntnu.no]
- [Name], [email@ntnu.no]

---

## Design Overview

Our system employs **peer-to-peer UDP broadcast topology** combined with **persistent state storage** and **idempotent state propagation** to guarantee no orders are lost while handling network failures, crashes, and packet loss transparently.

## Fault Tolerance Strategy

We address the core challenge—no orders lost despite arbitrary failures—through a hybrid approach grounded in distributed systems theory:

**CAB Orders (Backward Error Recovery):** Each elevator persists cabin call orders to disk after every change. On crash or power loss, orders are restored at startup, guaranteeing recovery within seconds.

**Hall Orders (Forward Error Recovery):** Hall calls are distributed across active elevators via a Hall Request Assigner (HRA) algorithm. When an elevator fails (detected via heartbeat timeout after 3 seconds), its assignments are automatically redistributed by remaining peers to other active elevators.

**Packet Loss Transparency:** All broadcasts carry complete state with idempotent properties—receiving the same message twice has identical effect as receiving once. Packet loss, duplication, and reordering are thus transparent to the application.

### CAP Theorem & Consistency Trade-off

Our system explicitly applies the CAP theorem to justify architectural choices:

**Given Constraints:**
- **Partition Tolerance:** REQUIRED (inevitable in networked elevators—network failures will occur)
- **Availability:** REQUIRED (elevator must continue moving even during failures)
- **Consistency:** SACRIFICED (we accept temporary inconsistency to maintain partition tolerance + availability)

**Trade-off Justification:**  
We prioritize partition tolerance and availability, accepting eventual consistency. This means:
- Light state may lag up to ~10ms between workspaces (acceptable—human imperceptible, self-heals)
- Hall orders may be assigned twice during network partition (acceptable—one becomes backup)
- **CAB orders are NEVER lost** (protected by local persistent disk storage, outside partition logic)

**Result:** System remains operational during network faults while asymmetrically protecting CAB orders (strong consistency via persistence) and allowing hall orders to converge eventually (weak consistency via network propagation).

### Failure Modes & Recovery Coverage

Systematic enumeration of all failure modes and recovery mechanisms:

| Failure Mode | Detection Mechanism | Recovery Strategy | Negation Risk | Outcome |
|---|---|---|---|---|
| **Software crash** | Process exit / file system | Restore CAB orders from disk; broadcast current state | Low | No orders lost; system reintegrates |
| **Network partition** | Heartbeat timeout (3 sec) | Graceful offline state; HRA redistributes hall assignments; CAB persists | Medium | Hall orders reassigned; CAB protected |
| **Motor stuck** | Motor response timeout (4 sec) | Transition to graceful offline; continue serving local CAB orders | Low | No new hall pickups; CAB continues |
| **Door obstruction** | Sensor + timer (5 sec) | Force door close; transition to graceful offline state | Low | Door handled; prevents cascade failures |
| **Packet loss** | (Transparent via idempotency) | Idempotent broadcast retransmit (5ms cycle); commutative operations | None | Loss masked; convergence <10ms |

All five primary failure modes are explicitly mapped. Each has clear detection and recovery with documented risk levels. No failure mode results in lost orders.

### Negation Minimization Strategy

Distributed system theory identifies "negation" (deletion/reset operations) as the primary source of inconsistency. Our design minimizes negation through careful state design:

**Monotonic Elements (Append-Only, Never Deleted):**
- **CAB orders:** Only added to queue; never removed mid-service (queue grows monotonically)
- **Elevator identifiers:** Fixed at startup; only transition to offline state, never erased from system
- **Broadcast timestamps:** Strictly increasing; never reset or go backward

**Controlled Negation Points (With Safeguards):**
- **Hall order completion:** Marked complete with version metadata (logical deletion, not physical erasure)
- **Motor/obstruction faults:** Transition state with timeout confirmation (not immediate state change)
- **Online/offline status:** Heartbeat-based with retry logic (not instantaneous state flip)

**Why This Works:**  
By keeping most state monotonic (append-only) and controlling negation carefully with version metadata and timeouts, we reduce the surface area for distributed inconsistency. Combined with idempotent broadcasts, the remaining negation operations become safe even under packet loss, duplication, and reordering. This is why our system can tolerate arbitrary network failures without losing consistency on critical orders.

## Network Topology & Protocol

**Topology:** UDP broadcast mesh on local network (port 1338). Each elevator broadcasts and receives; no central coordinator avoids single point of failure.

**Message Format:** Type-tagged JSON envelope containing elevator state (floor, direction, behavior, cabin orders, online status) broadcast every 5 milliseconds. Type tag enables safe routing without parsing all payloads.

**Consensus Mechanism:** Each elevator maintains a local mirror of peer states updated via idempotent broadcasts. When all active peers have responded within the 3-second heartbeat window, Order Manager invokes HRA with the current global view, achieving eventual consistency without explicit voting.

**Failure Detection:** Heartbeat timeout—if an elevator is not heard from within 3 seconds, it is marked offline and its hall assignments are immediately redistributed. Motor faults (4-second timeout) and door obstructions (5-second timeout) trigger graceful offline mode: broadcasting pauses but cabin orders continue being served locally.

## System Architecture

The system divides into three independent modules:

1. **Elevator FSM:** Models physical elevator (states: INIT, IDLE, MOVING, DOOR_OPEN). Manages sensors, motor control, door timing. Reports current state to Order Manager each cycle.

2. **Network Module:** Broadcasts complete elevator state every 5ms with type-tagged envelope. Maintains alive list; detects online/offline transitions via heartbeat timeout. Handles all network communication without assuming reliability.

3. **Order Manager:** Maintains persistent cabin order storage. Reads global state from Network Module, invokes HRA when consensus achieved, manages button light state machine (STANDBY → BUTTON_PRESSED → ORDER_ASSIGNED → COMPLETED → STANDBY).

## Programming Language: Go

Go's **concurrency model** enables one lightweight goroutine per real-time demand (sensors, network, timers) without shared variable synchronization. Channels provide thread-safe IPC, eliminating race conditions and deadlock risks—critical for fault-tolerant real-time systems. Built-in UDP support requires no heavy dependencies.

## Handling Project Challenges

**Button Light Guarantee:** Cabin lights turn on immediately (local operation). Hall lights turn on when HRA assigns an elevator, turn off when order completed. Idempotent broadcasts synchronize light state across all workspaces (~10ms convergence).

**No Orders Lost:** Cabin orders backed to disk; hall orders preserved via network consensus and HRA reassignment on failure. Both mechanisms cover crash, power loss, and network disconnect.

**Network Disconnect with Active Hall Orders:** Heartbeat timeout (3s) detects failure. Remaining elevators invoke HRA to redistribute failed elevator's assignments. No hall order is lost—only reassigned.

**Spontaneous Crashes & Restarts:** Cabin orders restored from persistent storage. Elevator broadcasts state; HRA integrates recovered elevator back into system. Light states synchronize automatically via idempotent broadcasts.

**Motor/Door Faults:** Motor timeout (4s) or obstruction timeout (5s) triggers graceful offline—broadcasting pauses, no new hall calls accepted, cabin orders served locally. Once fault resolves, elevator resumes broadcasting and normal operation.

**Packet Loss:** Transparent via idempotent property. Lost packets recovered in next 5ms broadcast cycle. Duplicates have identical effect as single receipt. Out-of-order delivery masked by state version tracking.

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **P2P Topology** | Avoids single point of failure. Scales naturally to n elevators without coordinator election. |
| **Idempotent Broadcasts** | Packet loss/duplication/reordering become transparent. Simple reliable state sync without ACKs. |
| **Graceful Offline** | Motor/obstruction faults don't cascade. Local CAB service continues. Automatic recovery. |
| **5ms Broadcast** | Balances convergence speed (~10ms) with network load. 10x faster than reference design. |
| **HRA for Hall Orders** | Optimal global assignment. Decoupled from distributed state sync. Proven pattern. |
| **CAB Persistence** | Only way to guarantee no calls lost after power loss. Backward error recovery pattern. |
| **Eventual Consistency** | Accept CAP trade-off: partition tolerance + availability > strict consistency. Minimizes negation. |

---

**Timings:** Broadcast 5ms | Heartbeat 3s | Motor timeout 4s | Obstruction timeout 5s | Door open 3s | UDP port 1338

**Design Status:** Pragmatic, viable, theoretically grounded—ready for implementation with high confidence in meeting all fault tolerance requirements.
