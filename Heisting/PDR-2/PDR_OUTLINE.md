# TTK4145 Preliminary Design Description - Outline

**Lab Time:** [Your lab time]  
**Workstation/Desk Number:** [Your desk number]  
**Group Number:** [Your group number]  

**Group Members:**
- [Name 1] - [email@ntnu.no]
- [Name 2] - [email@ntnu.no]
- [Name 3] - [email@ntnu.no]

---

## 1. SYSTEM ARCHITECTURE

### Network Topology
**Peer-to-peer UDP mesh broadcast** - No central coordinator, all elevators communicate as equals.

**Rationale:** 
- Eliminates single point of failure (master-slave rejected for this reason)
- Satisfies "no calls lost" requirement even when elevators join/leave
- UDP broadcast enables eventual consistency with minimal complexity

**Protocol Design:**
- **Broadcast interval:** 100ms (balances latency vs bandwidth)
- **Message structure:** JSON-encoded state containing:
  - Elevator state (floor, direction, behavior)
  - Cab orders (local only)
  - Hall order ACK table (distributed consensus)
  - Heartbeat timestamp
- **Port:** Single UDP broadcast port (20009) for all elevators

---

## 2. FAULT TOLERANCE STRATEGY

### 2.1 Data Persistence - "No Calls Lost"

**Cab Orders:** ✅ Disk-backed persistence
```
Action: Button pressed → Immediate write to "cab_orders_<ID>.txt"
Recovery: On startup → Read file → Restore active cab orders
Format: Binary array [floor] → bool
```
**Justification:** Handles power loss and crashes per specification requirement.

**Hall Orders:** ✅ ACK-based consensus (in-memory)
```
ACK Table: [floor][button_type][elevator_id] → bool
Removal Rule: Order removed when ALL online elevators ACK completion
```
**Justification:** 
- Prevents premature removal during network partitions
- Hall orders naturally redistributed when elevator rejoins (cost function recalculation)
- Disk persistence not needed - other elevators maintain state

### 2.2 Failure Detection - Multi-layered

**Layer 1: Heartbeat Timeout** (1 second)
- Missing 10 consecutive broadcasts → elevator marked offline
- Action: Reassign hall orders via cost function

**Layer 2: Order Progress Timeout** (15 seconds)
- Elevator has active order but no floor changes detected
- Action: Mark order as timed-out (-2), redistribute to network

**Layer 3: Self-Detection**
- Motor failure: No floor sensor changes while motor running (>5s)
- Door obstruction: Door blocked >5s → Disconnect from network, serve cab orders only
- Action: Automatic graceful degradation

---

## 3. MODULE ARCHITECTURE

```
┌─────────────────────────────────────────────────┐
│                    MAIN                         │
│  (Initialization, channel creation, recovery)   │
└──────────────┬──────────────────────────────────┘
               │
       ┌───────┴────────┬─────────────┬──────────────┐
       │                │             │              │
   ┌───▼────┐      ┌────▼────┐   ┌───▼─────┐   ┌───▼──────┐
   │  FSM   │      │ Network │   │ Assigner│   │  Lights  │
   └───┬────┘      └────┬────┘   └───┬─────┘   └────┬─────┘
       │                │             │              │
       │   ┌────────────▼─────────────▼──────────────▼──┐
       └───►      Elevator Driver (Hardware I/O)        │
           └────────────────────────────────────────────┘
```

**FSM (Finite State Machine):**
- States: IDLE, MOVING, DOOR_OPEN
- Events: NewOrder, FloorReached, DoorTimeout, Obstruction
- Responsibility: Single elevator control logic

**Network:**
- Broadcast state to peers (100ms periodic)
- Receive and merge peer states
- Detect online/offline transitions
- Maintain global worldview

**Assigner (Order Management):**
- Accept button events (cab + hall)
- Persist cab orders to disk
- Call external HRA (hall_request_assigner) for optimal distribution
- Manage ACK protocol for hall order completion
- Handle timeout-based reassignment

**Lights:**
- Synchronize button lights based on global state
- Hall lights: Show when ANY elevator has accepted order
- Cab lights: Local only

**Elevator Driver:**
- Hardware abstraction layer (using delivered elevio)
- Polls sensors, controls motor/lights
- No business logic

---

## 4. CRITICAL SCENARIOS

### Scenario A: Button Press → Door Open (Normal Operation)

**Hall Call:**
1. User presses UP on floor 2 → Button event to Assigner
2. Assigner: Add to global hall orders → Broadcast to network
3. All elevators receive → Run cost function (HRA)
4. Best elevator: Adds order to local queue → Light turns ON
5. FSM executes order → Arrives at floor 2 going UP
6. FSM: Door opens, sets ACK[2][UP][own_ID] = true
7. ACK broadcasted → When all elevators ACK → Light turns OFF

**Cab Call:**
1. User inside elevator presses floor 3 → Assigner
2. Assigner: Write to disk → Add to local orders → Light ON
3. FSM executes → Arrives floor 3 → Door opens
4. Assigner: Clear order, delete from disk → Light OFF

**Time to light ON:** <200ms (worst case: 2 × broadcast interval)

### Scenario B: Network Disconnection + Active Hall Request

**Detection Phase:**
1. Elevator A loses network → Stops receiving broadcasts
2. Other elevators: No heartbeat from A after 1 second
3. Network module marks A as offline

**Takeover Phase:**
4. Assigner detects topology change → Triggers cost function recomputation
5. HRA redistributes A's hall orders to remaining online elevators
6. New assignee: Updates local orders → Continues execution

**Disconnected Elevator A:**
- Continues serving existing cab orders (lights already ON)
- Refuses new hall calls (or serves them in single-elevator mode)
- Accepts new cab calls (people can exit)

**Reconnection:**
- A rejoins → Sends state broadcast
- Network: Detects new elevator
- Cost function rebalances orders (may take some back if A is optimal)

**Time:** Detection <1s, Reassignment <500ms

### Scenario C: Node Crash with Active Cab Order

**Crash:**
1. Elevator B crashes while servicing cab order (floor 3 light ON)
2. Power restored → Software restarts

**Recovery:**
3. Main initialization → Calls ReadCabOrderBackup()
4. Reads "cab_orders_B.txt" → Floor 3 = true
5. Turns light ON → Adds to FSM queue
6. Resumes execution → Services order

**User Experience:** Elevator pauses briefly, then continues
**No manual intervention required**

**Time:** Recovery < 5s (assuming fast reboot)

### Scenario D: All Above + Packet Loss

**Network Layer handles packet loss through redundancy:**
- State broadcasted every 100ms
- 10% packet loss = 90% of messages get through
- Missed update caught in next broadcast (100ms later)

**ACK Protocol prevents data loss:**
- Order not removed until all elevators acknowledge
- If ACK packet lost → Order stays active
- Elevator re-broadcasts ACK → Eventually all receive

**Timeout provides backstop:**
- If packets consistently lost → Order progress timeout triggers
- Order redistributed to healthy connection path

**No special packet loss handling needed** - system architecture inherently tolerant

---

## 5. DESIGN DECISIONS RATIONALE

### Why Go?
**Selected for concurrency primitives:**
- Goroutines for concurrent module execution
- Channels for safe inter-module communication
- Matches peer-to-peer architecture (no shared state between modules)

### Why External HRA?
**Delivered code (hall_request_assigner) provides:**
- Proven optimal cost function
- Handles edge cases (e.g., elevator stuck between floors)
- Separation of concerns: We focus on distributed coordination, not optimization

### Why ACK Protocol (Not Timeout Only)?
**Comparison:**
- **Timeout only** (Solution 2): Orders cleared when elevator arrives → Vulnerable to network partition (both sides think they serviced it)
- **ACK consensus** (Solution 3): Orders cleared when ALL confirm → Partition-safe, may duplicate service (acceptable per spec)

**Choice:** ACK protocol - Guarantees "button light contract" even during network issues

### Why 100ms Broadcast Interval?
**Analysis from existing solutions:**
- 2ms (Solution 2): Excessive bandwidth, ~11 MB/hour per elevator
- 200ms (Solution 3): Acceptable but slow light response
- 100ms: Optimal balance - Fast enough (<200ms light latency), reasonable bandwidth

---

## 6. TESTING STRATEGY

**Unit Testing:**
- FSM state transitions
- ACK protocol logic
- Persistence read/write

**Integration Testing:**
- Packet loss simulation (using delivered packetloss.sh)
- Network disconnect/reconnect
- Simultaneous button presses
- Multiple elevator coordination

**Failure Testing:**
- Kill elevator process (crash simulation)
- Disconnect network cable
- Obstruct door for extended period
- Power cycle elevator hardware

---

## RISK ASSESSMENT

**Moderate Complexity Risks:**
- ACK protocol implementation - Mitigation: Start with simple timeout, add ACK later
- Cost function integration - Mitigation: Use delivered HRA binary (proven to work)

**Low Complexity:**
- Disk persistence (simple file I/O)
- UDP broadcast (delivered network code available)
- FSM (well-defined state machine)

**Estimated Implementation:** 3-4 weeks for core functionality + 1 week testing/refinement

---

**This preliminary design satisfies all main requirements while maintaining reasonable complexity through:**
1. Proven architectural patterns (peer-to-peer UDP mesh)
2. Leveraging delivered code (elevio, HRA, network modules)
3. Multi-layered fault tolerance (persistence + ACK + timeout)
4. Clear module separation enabling parallel development
