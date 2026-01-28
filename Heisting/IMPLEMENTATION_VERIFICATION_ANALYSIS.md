# Implementation Analysis: HRA Determinism & Single Elevator Mode

## KEY FINDING FROM SPEC & README

From readME_elev.md - **Most Critical Requirement:**

```
"If the elevator is disconnected from the network, it should still serve all the 
currently active calls (i.e. whatever lights are showing)"

"It should also keep taking new cab calls so that people can exit the elevator 
even if it is disconnected from the network"
```

**Unspecified behavior:**
```
"How the hall (call up, call down) buttons work when the elevator is disconnected 
from the network - You can optionally refuse to take these new calls"
```

**Your choice is explicit:** "You can optionally refuse" = It's a design choice.

---

## ANALYSIS OF THREE SOLUTIONS

### Solution 1 (TTK4145 - Peer-to-Peer Cyclic Counter)

**Handling of Offline State:**
```go
// network.go - When this elevator goes offline
if lostNode == nodeID {
    online = false
}

// When offline
if !online {
    newAliveList := [NElevators]bool{}
    newAliveList[nodeID] = aliveList[nodeID]
    stateBroadcast <- FromNetworkToAssigner{
        AliveList:     newAliveList,
        ElevatorList:  elevatorList,
        HallOrderList: hallOrderList,  // ← STILL BROADCASTS HALL ORDERS!
    }
}
```

**Key Finding:** 
- ✅ Continues to broadcast state when offline
- ✅ Keeps hall order list in memory
- ✅ Assigner still processes hall calls
- **Conclusion:** Solution 1 SERVES BOTH cab and hall calls when offline

**HRA Determinism:**
- Uses external `hall_request_assigner` binary
- Passes identical JSON input from all elevators with identical `aliveList`
- **Assumption:** All elevators receive broadcasts synchronously
- **Problem:** No explicit synchronization before HRA call

---

### Solution 2 (Superheis - Full State Broadcasting)

**Handling of Offline:**

From LogManagement.go:
```go
func UpdateOrderList(msg Elev, ...) {
    for i := 0; i < numFloors; i++ {
        for j := 0; j < numButtons-1; j++ {  // j < numButtons-1 = ONLY HALL CALLS (0,1)
            if msg.Orders[i][j].Status == 0 && myElevInfo.Orders[i][j].Status == -1 {
                // New order from network
                myElevInfo.Orders[i][j].Status = 0
                // Turn ON light and queue order
            }
        }
    }
}
```

**Key Finding:**
- Only processes hall calls from **other** elevators
- When **this** elevator goes offline, no changes to logic
- **Still accepts and processes hall calls from network**
- **Conclusion:** Solution 2 SERVES BOTH cab and hall calls when offline

**HRA Determinism:**
- Uses custom cost function (no HRA)
- Simple distance-based: "If same distance, lowest ID wins"
- **Deterministic but simple:** Not optimal, but predictable
- **No synchronization issue** because logic is trivial

---

### Solution 3 (Maiken/Khuong - ACK Protocol with Fault Detection)

**Handling of Offline:**

```go
// fault.go - Motor stop detection
if behaviour == types.EB_Moving && time.Since(start) > motorStopTime {
    bcast.UpdateOnlineStatus(false)  // Disconnect from network
    rebootSystem()
}

// obstruction detection
if time.Since(start) > timeTol && onlineStatus {
    bcast.UpdateOnlineStatus(false)  // Disconnect
}
```

**Processing Model:**
```go
// processing.go - Data processing flow
case btnEvent := <-drv_buttons:
    if btnEvent.Button == elevio.BT_Cab {
        myNodeData.CabOrders[btnEvent.Floor] = true
    } else {
        // Hall button
        myNodeData.AllHallOrders[btnEvent.Floor][btnEvent.Button] = true
    }
```

**Key Finding:**
- ✅ **Accepts ALL button types** (cab + hall) regardless of network status
- ✅ **Queues hall calls locally** in `AllHallOrders`
- ✅ When network restores: Re-broadcasts queued hall calls
- **Conclusion:** Solution 3 SERVES BOTH cab and hall calls when offline
- **Most sophisticated:** Explicit queue for offline hall calls

**HRA Determinism:**
- Uses external `hall_request_assigner` binary
- **Passes list of ALL peers** that are currently online
- **Issue:** If peers disagree on who's online → Different HRA input → Different output
- **Mitigation:** ACK protocol ensures no premature order removal
- **Resolution:** Worst case = duplicate service (acceptable per spec)

---

## CRITICAL INSIGHT: All Three Solutions Serve Hall Calls When Offline

| Solution | When Offline | Hall Calls | Cab Calls | Mechanism |
|----------|-------------|-----------|----------|-----------|
| **1** | Yes | ✅ YES | ✅ YES | Continues broadcasting state |
| **2** | Yes | ✅ YES | ✅ YES | Continues processing network messages |
| **3** | Yes | ✅ YES | ✅ YES | Explicit queue + rebroadcast on reconnect |

**None of them refuse hall calls when offline.**

---

## HRA DETERMINISM - REAL PROBLEM IDENTIFIED

**Scenario That Breaks Determinism:**

```
T=0ms:     Hall call "Floor 2 UP" pressed
T=10ms:    Elev A online, receives message: {onlineElevs: [A, B], hallCalls: [UP2]}
           → Calls HRA({onlineElevs: [A,B], hallCalls: [UP2]})
           → HRA says: "A takes it"
           
T=15ms:    Elev B missed message (packet loss), still thinks prev state
           → Elev B calls HRA({onlineElevs: [A, B, C], hallCalls: [UP2, UP3]})
           → HRA says: "B takes it" (different cost due to different hall call list)

RESULT:    A thinks it's assigned, B thinks it's assigned
           ❌ Both lights ON
           ❌ Spec violation
```

**Root cause:** Not all elevators have identical state BEFORE running HRA.

**Solution 1 mitigation:** Cyclic counter provides eventual consistency
**Solution 2 mitigation:** 2ms broadcast = high redundancy masks problem
**Solution 3 mitigation:** ACK protocol = orders not finalized until consensus

---

## RECOMMENDATIONS FOR PDR V3 UPDATE

### 1. HRA Determinism - Fix the Statement

**Replace:**
```
"HRA (hall_request_assigner) is deterministic: given identical global state, 
it produces identical assignments."
```

**With:**
```
"HRA (hall_request_assigner) is deterministic: given identical input (floor, 
direction, behavior, cab requests for each online elevator), produces identical 
assignments. However, achieving identical state across all elevators is non-trivial 
in a peer-to-peer network:

Problem: Elevators may have different views of online peers due to network jitter 
or packet loss, causing HRA input to differ, resulting in different assignments.

Mitigation: 
1. Periodic broadcasts (100ms) provide redundancy - missed broadcasts caught quickly
2. ACK protocol prevents order finalization until consensus reached
3. Timeout mechanism (15s) allows recovery from temporary disagreements
4. Result: Button light inconsistency is temporary (<200ms) and self-healing

Testing assumption: We assume sufficient network health that transient disagreements 
are rare. Spec permits this (assumption #1: at least one elevator not in failure)."
```

### 2. Single Elevator Mode - Change the Statement

**Replace:**
```
"Single-Elevator Mode at Init: If network unavailable at startup → enter 
single-elevator mode (serve cab calls only)"
```

**With:**
```
"Single-Elevator Mode (Network Disconnection):

When elevator is disconnected from network (whether at init or mid-operation):
- Serve BOTH cab calls AND hall calls
- Hall calls queue locally (stored in `cab_orders_<ID>.txt` if needed)
- When network restores: Rebroadcast queued hall calls for HRA redistribution

Rationale:
1. Spec requirement: "should still serve all the currently active calls"
2. If only one elevator functional (per assumption #1): Treat as standalone
3. Better user experience: Users can call elevator from any floor
4. Unspecified behavior (design choice): Spec permits refusing hall calls, but 
   serving both is safer and more resilient

Implementation:
- Offline: Accept all button presses (buttons don't distinguish network status)
- Lights controlled by local FSM (don't require network)
- Hall calls tracked locally in state
- On reconnect: Process queued calls through HRA with other elevators
- ACK protocol ensures no duplicate service even if queued calls already handled
```

### 3. Add New Section: State Synchronization on Reconnection

```
Reconnection Protocol (Implicit in Design):

When disconnected elevator rejoins network:
1. Sends broadcast with: current floor, direction, ALL queued hall orders, cab orders
2. Other elevators: Receive state update
3. All elevators: Run HRA with new topology (now n+1 elevators)
4. HRA may reassign some hall orders to rejoined elevator (may take some or give some)
5. ACK protocol: Synchronizes which hall orders are actually being served
6. Light updates: Button lights stabilize as ACK consensus reached

Time to stabilization: <200ms (2 × broadcast interval)
```

---

## FINAL VERDICT

✅ **Your instinct was correct:**
- Solution 3's approach (queue and rebroadcast) is the most elegant
- All three solutions serve hall calls when offline - it's the right design
- Solution 1's cyclic counter handles HRA determinism implicitly
- Solution 2's high broadcast frequency masks synchronization issues
- Solution 3's ACK protocol explicitly handles the safety aspects

❌ **PDR V3 current statements are too optimistic:**
- HRA determinism needs more nuance (acknowledge eventual consistency)
- Single-elevator mode should serve BOTH types of calls
- Reconnection protocol should be mentioned

---

## RECOMMENDED V3 CHANGES (Summary)

1. **HRA Determinism section:** Add paragraph about eventual consistency + mitigation
2. **Single-elevator mode:** Change to "serve both cab and hall calls"
3. **Reconnection protocol:** Add explicit mention of queue + rebroadcast
4. **Design choice documentation:** Explain why serving both call types is chosen

These changes will make the PDR significantly stronger and more aligned with actual implementations.
