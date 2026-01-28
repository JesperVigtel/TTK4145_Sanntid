# PRELIMINARY DESIGN DESCRIPTION

## Header Information
**Lab Time:** [Your lab time]  
**Workstation/Desk:** [Your desk number]  
**Group Number:** [0#]  
**Group Members:** [Name], [Name], [Name]

---

## Design Overview

Peer-to-peer UDP broadcast with persistent CAB storage and idempotent state propagation guarantees no orders lost despite failures, crashes, and packet loss.

## Fault Tolerance Strategy

**CAB Orders:** Persisted to disk; restored on crash (backward error recovery).  
**Hall Orders:** Distributed via HRA; redistributed on failure (forward error recovery).  
**Packet Loss:** Idempotent broadcasts transparently mask loss, duplication, reordering.

### CAP Theorem Trade-off

We sacrifice consistency for partition tolerance + availability (inevitable tradeoff in networked systems):
- Light state lag <10ms (acceptable—human imperceptible, self-heals)
- Hall orders duplicate during partition (one becomes backup)
- **CAB orders NEVER lost** (disk-protected, outside partition logic)

Result: System operational during network faults while CAB orders remain guaranteed.

### Failure Modes & Recovery

| Failure Mode | Detection | Recovery | Negation Risk |
|---|---|---|---|
| Software crash | Process exit | Restore CAB from disk; rebroadcast state | Low |
| Network partition | Heartbeat timeout (3s) | Graceful offline; HRA redistributes; CAB persists | Medium |
| Motor stuck | Motor timeout (4s) | Graceful offline; local CAB service | Low |
| Door obstruction | Sensor + timer (5s) | Force close; graceful offline | Low |
| Packet loss | Transparent | Idempotent retransmit (5ms); commutative ops | None |

All five primary failure modes mapped to detection + recovery. No failure results in lost orders.

### Negation Minimization

Distributed system theory: deletion/reset is inconsistency's primary source. We minimize negation:

**Monotonic (append-only):** CAB orders only added, never removed mid-service. Elevator IDs fixed, only offline status changes. Timestamps strictly increase.

**Controlled Negation:** Hall completion marked (logical deletion, not physical). Motor/obstruction faults transition via timeout. Online/offline status heartbeat-based.

Why it works: Monotonic state + controlled negation + idempotent broadcasts = safe even under packet loss and reordering.

## Network Topology & Protocol

**Topology:** UDP broadcast mesh (port 1338), no central coordinator.  
**Message:** Type-tagged JSON with complete state, broadcast every 5ms.  
**Consensus:** Local mirror updated via idempotent broadcasts. HRA invoked when all peers respond within 3s heartbeat window → eventual consistency.  
**Failure Detection:** Heartbeat timeout (3s) marks elevator offline; HRA redistributes assignments. Motor (4s) and door (5s) timeouts trigger graceful offline.

## System Architecture

**Elevator FSM:** Models physical state (INIT, IDLE, MOVING, DOOR_OPEN). Manages sensors, motor, timings.  
**Network Module:** Broadcasts state every 5ms with type tags. Maintains alive list via heartbeat timeout.  
**Order Manager:** Persists CAB orders. Reads global state, invokes HRA, manages light state machine (STANDBY → PRESSED → ASSIGNED → COMPLETED).

## Programming Language: Go

Lightweight goroutines (one per real-time demand) + channels (thread-safe IPC) eliminate race conditions and deadlock—critical for fault-tolerant real-time. Built-in UDP support, no heavy dependencies.

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| P2P Topology | No SPOF; scales to n elevators without election. |
| Idempotent Broadcasts | Loss/duplication/reordering transparent; simple reliable sync. |
| Graceful Offline | Faults don't cascade; local CAB continues; auto-recovery. |
| 5ms Broadcast | ~10ms convergence, manageable network load. |
| HRA for Hall Orders | Optimal global assignment, decoupled from sync. |
| CAB Persistence | Only guarantee against power loss/crash. Backward recovery. |
| Eventual Consistency | CAP trade-off: partition + availability > consistency. Minimal negation. |

---

**Timings:** Broadcast 5ms | Heartbeat 3s | Motor timeout 4s | Obstruction 5s | Door 3s | UDP 1338

**Status:** Pragmatic, viable, theoretically grounded—ready for implementation.
