# TTK4145 Preliminary Design Description - V3 (Final)

**Lab Time:** [Your lab time] | **Desk:** [Desk #] | **Group:** [Group #]  
**Names:** [Name 1], [Name 2], [Name 3] — [emails]

---

## SYSTEM ARCHITECTURE

**Peer-to-peer UDP mesh broadcast** with 100ms broadcast interval. No central coordinator; all elevators communicate as equals, enabling "no calls lost" even during network failures and crashes.

**Protocol:** All elevators broadcast their state (floor, direction, cab orders, ACK table) every 100ms to single UDP port. Network convergence: <200ms for light synchronization under normal conditions (2 × broadcast interval).

---

## FAULT TOLERANCE STRATEGY

**Cab Orders:** Disk persistence (`cab_orders_<ID>.txt`). Write on every button press. Recovery on startup reads file and restores lights—system automatically resumes without manual restart.

**Hall Orders:** ACK-based consensus. Order removed only when ALL online elevators acknowledge completion. Prevents duplicate service during network partitions. Orders naturally redistribute when elevator rejoins (cost function recomputation).

**HRA Determinism & Button Light Consistency:** The hall_request_assigner (HRA) is deterministic: identical input produces identical assignments. However, achieving identical state across all elevators is non-trivial in peer-to-peer networks. Problem: Network jitter or packet loss may cause elevators to have different views of online peers, causing different HRA inputs and outputs. Mitigation: (1) Periodic broadcasts (100ms) provide redundancy—missed broadcasts caught quickly, (2) ACK protocol prevents order finalization until consensus reached, (3) Timeout mechanism (15s) allows recovery from transient disagreements. Result: Button light inconsistencies are temporary (<200ms) and self-healing.

**Failure Detection (3-layer):**
1. **Heartbeat (1s):** Missing broadcasts → offline → reassign hall orders via HRA
2. **Progress (15s):** No floor change while order active → redistribute order
3. **Self-Detection:** Motor stall (>5s) or obstruction (>5s) → graceful degradation (disconnect, serve cab-only)

---

## MODULE ARCHITECTURE

**FSM:** Local elevator logic (IDLE ↔ MOVING ↔ DOOR_OPEN states). Clears only buttons matching travel direction. 

**Real-time Obstruction Handling:**
- ObstructionTriggered (immediate): If DOOR_OPEN → restart 3-second timer (prevent close). If MOVING → ignored (door closed).
- Door remains open indefinitely while obstruction persists; can never close mid-obstruction.
- ObstructionCleared: Door can close on next timer expiration.
- After 5+ seconds continuous obstruction: Fault module initiates network disconnect.

**Direction Changes:** When reversing direction (e.g., UP→DOWN at floor 2): (1) Clear UP button, (2) Open door 3s at floor 2, (3) Close door, move to floor 1, (4) Clear DOWN button at floor 1. Ensures users see direction announcement.

**Network:** Periodic broadcast (100ms), state merge, peer detection/online-offline transitions.

**Assigner:** Accept button events, persist cab orders, run HRA cost function on topology changes, manage ACK protocol, timeout-based reassignment.

**Lights:** Hall lights (on when ANY elevator accepts) vs. Cab lights (local only).

**Driver:** Hardware abstraction (using delivered elevio).

---

## CRITICAL SCENARIOS

**A) Normal Hall Call:**  
Button UP (floor 2) → broadcast → HRA assigns to best elevator → FSM executes → arrives going UP → clears only UP button → door 3s → when all elevators ACK → light OFF. **Latency:** <200ms light-on.

**B) Network Disconnect + Reassignment:**  
Heartbeat timeout (1s) → mark offline → HRA redistributes hall orders to remaining elevators. Disconnected elevator continues active cab orders only (refuses new hall calls). When network restored: Auto-reconnect, rebalance orders.  **Time:** <1.5s detection + reassignment.

**C) Crash with Cab Order:**  
Restart → read cab order file → restore lights → resume service. **No manual intervention.** **Time:** <5s.

**D) ACK Deadlock Prevention:**  
If elevator crashes with active hall order: progress timeout (15s) triggers → order redistributed → other elevators' ACKs override missing ACK → light turns OFF when healthy elevator completes. **No deadlock, no infinite light.**

---

## DESIGN DECISIONS

**Go:** Goroutines + channels suit peer-to-peer (no shared state between modules).

**External HRA:** Proven cost function; separates coordination from optimization logic.

**ACK Protocol:** Handles network partitions correctly (partition-safe). Deterministic with synchronized state.

**100ms Broadcast:** Optimal balance—<200ms light latency vs. reasonable bandwidth.

**Single-Elevator Mode (Disconnection):** When disconnected—at startup or during operation—elevator serves BOTH cab and hall calls as a standalone. Hall calls queue locally. On reconnect: queued calls rebroadcast, HRA rebalances with new topology, ACK protocol synchronizes. Rationale: Spec requires "serve all currently active calls"; treating as standalone maximizes resilience; unspecified behavior choice for robustness.

---

## IMPLEMENTATION APPROACH

- **Persistence:** Synchronous disk writes on cab order change (crash recovery)
- **Consensus:** ACK table broadcast ensures agreement on completion
- **Self-Healing:** Multi-layer detection (heartbeat + progress + self-detect)
- **Direction Logic:** FSM respects announcements (directional button clearing, direction reversals)
- **Obstruction:** Real-time signal handling (restart timer, prevent close)
- **Reconnection:** Queued hall orders rebroadcast; HRA rebalances; ACK protocol synchronizes; lights stabilize <200ms
- **Graceful Degradation:** Obstruction/motor failure → disconnect network, continue cab orders

**Implementation:** ~4 weeks core + 1 week testing/hardening.
