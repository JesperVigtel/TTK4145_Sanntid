# TTK4145 Preliminary Design Description - 1 Page

**Lab Time:** [Your lab time] | **Desk:** [Desk #] | **Group:** [Group #]  
**Names:** [Name 1], [Name 2], [Name 3] — [emails]

---

## SYSTEM ARCHITECTURE

**Peer-to-peer UDP mesh broadcast** with 100ms broadcast interval. No central coordinator; all elevators communicate as equals, enabling "no calls lost" even during network failures and crashes.

---

## FAULT TOLERANCE STRATEGY

**Cab Orders:** Disk persistence (`cab_orders_<ID>.txt`) survives crashes/power loss. System automatically resumes without manual restart.

**Hall Orders:** ACK-based consensus—orders removed only when ALL online elevators acknowledge completion, preventing duplicate service during network partitions.

**Failure Detection (3-layer):**
1. **Heartbeat timeout (1s):** Missing broadcasts → mark offline → redistribute orders via HRA
2. **Order progress timeout (15s):** No floor progress → redistribute order  
3. **Self-detection:** Motor failure/door obstruction (>5s) → graceful degradation (disconnect from network, continue serving cab orders)

---

## MODULE ARCHITECTURE

**FSM:** Local elevator control (IDLE → MOVING → DOOR_OPEN). Clears only buttons matching direction (UP/DOWN). Direction changes: clear old direction button, 3-second door, then reverse.

**Network:** Periodic broadcast (100ms), state merge, peer detection.

**Assigner:** Button events, disk writes, HRA cost function calls, ACK protocol, timeout handling.

**Lights:** Hall lights (ANY elevator accepted) vs. Cab lights (local only).

**Driver:** Hardware abstraction (using delivered elevio).

---

## CRITICAL SCENARIOS

**A) Normal Operation:** Button press → broadcast → HRA assigns to best elevator → FSM executes → arrives at floor → clears only matching direction → door opens 3s → light off when all ACK. **Latency:** <200ms light-on.

**B) Network Disconnect + Hall Request:** Detect missing heartbeat (1s) → redistribute order to remaining elevators via HRA. Disconnected elevator continues cab orders, refuses hall calls. Auto-reconnect rebalances. **Time:** <1.5s total.

**C) Crash with Active Cab Order:** Restart → read cab order file → restore lights → continue. **No manual intervention.** **Time:** <5s recovery.

**D) Packet Loss (10%):** State re-broadcasted every 100ms (redundancy). ACK protocol ensures no loss (orders stay active if ACK packet lost). Timeout redistributes if needed. **Inherently tolerant.**

---

## DESIGN RATIONALE

**Go:** Goroutines + channels suit peer-to-peer coordination (no shared state between modules).

**External HRA:** Proven optimal cost function; separates coordination logic from optimization.

**ACK Protocol:** Unlike timeout-only approaches, handles network partitions correctly (partition-safe consensus).

**100ms Broadcast:** Balances <200ms light latency with reasonable bandwidth (vs. 2ms or 200ms alternatives).

---

## IMPLEMENTATION APPROACH

- **Persistent storage:** Synchronous disk writes on cab order change
- **Distributed consensus:** ACK table broadcast ensures all peers agree on completion
- **Self-healing:** Multi-layer detection + automatic recovery
- **Directed clearing:** FSM respects direction announcements (only clear UP/DOWN matching direction)
- **Graceful degradation:** Obstruction/motor failure → disconnect from network, continue cab orders

Estimated 4-5 weeks: core (3-4 weeks) + testing (1 week).
