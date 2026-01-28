# Update 26 Jan 2026 — Added Fault Scenarios (Solution 3)

Scenario A: Obstruction > 5s
- Given: Door obstructed continuously
- When: Obstruction exceeds timeout
- Then: Set `online=false` (pause network broadcast), keep CAB persisted
- And: When obstruction clears, set `online=true` and rebroadcast; lights converge.

Scenario B: Motor stop timeout
- Given: Behaviour `EB_Moving` without floor arrival
- When: Movement exceeds ~4s threshold
- Then: Set `online=false`, halt motion; CAB persists; hall orders redistributed by peers via HRA.
- And: Operator resolves issue → resume (`online=true`), continue normal operation.

Specs Coverage Snapshot:
- Scenarios demonstrate: no calls lost (CAB persists), light guarantees (hall/cab), fault tolerance (disconnect/obstruction), and packet loss resilience (idempotent effects).

# SCENARIO WALKTHROUGHS - Konkrete Eksempler

## SCENARIO 1: Normal CAB Order (Bedste Tilfelle)

```
SETUP:
├─ Elevator A at floor 0, IDLE
├─ Elevator B at floor 3, IDLE  
├─ Elevator C at floor 2, IDLE
└─ Network: All connected

BRUKER TRYKKER: CAB knapp på floor 2 inne i Elevator A

════════════════════════════════════════════════════════════════

TIMELINE:

t=0ms:
  Hardware → buttonPress channel
  Button: {floor: 2, button: CAB}

t=1ms:
  ORDER MANAGER receives on buttonPress
  ├─ cabOrders[2] = true
  ├─ SaveCabOrders() → file "cab_orders_0.txt"
  └─ lightsOut ← {floor: 2, button: CAB, on: true}
  
  ✓ LIGHT IS ON

t=2ms:
  LIGHTS module receives
  └─ setButtonLamp(CAB, floor 2, ON)
  
  ✓ PHYSICAL LIGHT TURNS ON

t=5ms:
  ORDER MANAGER → Elevator FSM
  ordersToElevator[2][CAB] = true

t=6ms:
  Elevator FSM receives new order
  ├─ currentFloor = 0, targetFloor = 2
  ├─ direction = UP
  ├─ state = IDLE → MOVING
  └─ setMotor(UP)

t=10ms:
  HARDWARE: Motor starts spinning

t=150ms:
  HARDWARE: Floor sensor detects floor 1
  floorSensor ← 1

t=151ms:
  FSM receives floor 1
  ├─ Check: shouldStop? No (order is at floor 2)
  ├─ Continue moving
  └─ statusOut ← "floor 1, moving up"

t=300ms:
  HARDWARE: Floor sensor detects floor 2
  floorSensor ← 2

t=301ms:
  FSM receives floor 2
  ├─ Check: shouldStop? YES (CAB order at floor 2)
  ├─ setMotor(STOP)
  ├─ setDoor(OPEN)
  ├─ state = MOVING → DOOR_OPEN
  ├─ startDoorTimer(3000ms)
  └─ statusOut ← "floor 2, door open"

t=305ms:
  HARDWARE: Door opens (lamp on)

t=310ms:
  NETWORK: Broadcast
  ├─ SenderId: 0
  ├─ floor: 2
  ├─ state: DOOR_OPEN
  └─ All others receive and ACK

t=3301ms:
  TIMER FIRED: doorTimer.C

t=3302ms:
  FSM receives doorTimer
  ├─ Check: obstruction? No
  ├─ setDoor(CLOSED)
  ├─ cabOrders[2] = false (cleared)
  ├─ SaveCabOrders() → UPDATE file
  ├─ state = DOOR_OPEN → IDLE
  └─ statusOut ← "floor 2, idle"

t=3305ms:
  HARDWARE: Door closes

t=3310ms:
  ORDER MANAGER sees cabOrders[2] = false
  ├─ hallLightState[?][2][CAB] = COMPLETED
  └─ lightsOut ← {floor: 2, button: CAB, on: false}
  
  ✓ LIGHT IS OFF

t=3311ms:
  LIGHTS module receives
  └─ setButtonLamp(CAB, floor 2, OFF)
  
  ✓ PHYSICAL LIGHT TURNS OFF

════════════════════════════════════════════════════════════════

RESULT:
✓ CAB order taken within 6ms
✓ Light on within 2ms
✓ Elevator moved to floor 2
✓ Door opened for 3 seconds
✓ Light off after 3310ms total
✓ Order completed successfully

KEY POINTS:
- Fast response (2ms to light)
- Persistent (saved to file at t=1ms)
- Never lost (saved again at t=3302ms)
```

---

## SCENARIO 2: Hall Call with Conflict Resolution

```
SETUP:
├─ Elevator A at floor 0, IDLE
├─ Elevator B at floor 3, IDLE  
├─ Elevator C at floor 2, IDLE
├─ All connected via network
└─ All have synced state

BRUKER TRYKKER: HALL UP button on floor 1 (physical button panel at floor 1)

════════════════════════════════════════════════════════════════

TIMELINE:

t=0ms:
  HARDWARE: Button pressed
  buttonPress ← {floor: 1, button: HALL_UP}

t=1ms:
  ORDER MANAGER receives
  ├─ This is a HALL button (not CAB)
  ├─ hallOrderList[0][1][HALL_UP] = BUTTON_PRESSED
  └─ (Not saved to file - only CAB orders are persisted)

t=5ms:
  ORDER MANAGER → NETWORK MODULE
  networkState ← {
      myHallOrders: [... BUTTON_PRESSED at floor 1 ...]
  }

t=10ms:
  NETWORK SENDER broadcasts
  Message {
      SenderId: 0 (or 1 or 2, whoever is local)
      hallOrderList: [...BUTTON_PRESSED at floor 1...]
      aliveList: [true, true, true]
  }

t=11ms-59ms:
  All elevators receive broadcast
  ├─ Elevator A stores: hallOrderList[0][1][HALL_UP] = BUTTON_PRESSED
  ├─ Elevator B stores: hallOrderList[0][1][HALL_UP] = BUTTON_PRESSED
  ├─ Elevator C stores: hallOrderList[0][1][HALL_UP] = BUTTON_PRESSED
  └─ All see the same state

t=60ms:
  NETWORK SENDER broadcasts again (50ms interval)
  └─ All still see: BUTTON_PRESSED

t=65ms:
  ORDER MANAGER checks consensus
  ├─ All three elevators see BUTTON_PRESSED = true
  ├─ Run conflict resolution!
  │
  │  Distance calculation:
  │  ├─ Elevator A: dist = |0 - 1| = 1 floor  ← CLOSEST
  │  ├─ Elevator B: dist = |3 - 1| = 2 floors
  │  └─ Elevator C: dist = |2 - 1| = 1 floor
  │
  │  TIE BETWEEN A and C! (both 1 floor away)
  │  Tiebreaker: Use elevator ID
  │  ├─ A.ID = 0
  │  ├─ C.ID = 2
  │  └─ 0 < 2 → A wins!
  │
  └─ ASSIGN TO ELEVATOR A

t=70ms:
  ORDER MANAGER → NETWORK
  networkState ← {
      myHallOrders: [hallOrderList[0][1][HALL_UP] = ORDER_ASSIGNED]
  }

t=72ms:
  All elevators see UPDATE:
  ├─ hallOrderList[0][1][HALL_UP] changed to ORDER_ASSIGNED
  └─ All ACK this change

t=75ms:
  Light state updated:
  └─ LIGHT TURNS ON (was already on, stays on)

t=80ms:
  ORDER MANAGER → Elevator A FSM
  ordersToElevator[1][HALL_UP] = true

t=81ms:
  Elevator A FSM receives order
  ├─ currentFloor = 0, targetFloor = 1
  ├─ direction = UP
  ├─ state = IDLE → MOVING
  └─ setMotor(UP)

t=85ms:
  HARDWARE: Motor starts

t=200ms:
  HARDWARE: Floor 1 reached
  floorSensor ← 1

t=201ms:
  FSM receives floor 1
  ├─ shouldStop? YES (HALL order at floor 1 in direction UP)
  ├─ setMotor(STOP)
  ├─ setDoor(OPEN)
  ├─ state = MOVING → DOOR_OPEN
  └─ startDoorTimer(3000ms)

t=205ms:
  HARDWARE: Door opens

t=210ms:
  NETWORK broadcasts:
  ├─ Elevator A: floor 1, state DOOR_OPEN
  └─ All others acknowledge

t=3211ms:
  TIMER FIRED: doorTimer.C

t=3212ms:
  FSM:
  ├─ setDoor(CLOSED)
  ├─ hallOrderList[0][1][HALL_UP] = ORDER_COMPLETE
  ├─ state = DOOR_OPEN → IDLE
  └─ statusOut ← "floor 1, idle"

t=3220ms:
  NETWORK broadcasts:
  ├─ Elevator A: floor 1, state IDLE
  ├─ hallOrderList[0][1][HALL_UP] = ORDER_COMPLETE
  └─ All others acknowledge

t=3225ms:
  All elevators' ORDER MANAGER sees:
  ├─ hallOrderList[0][1][HALL_UP] = ORDER_COMPLETE
  └─ Light should turn OFF

t=3230ms:
  Light state updated globally:
  └─ LIGHT TURNS OFF

════════════════════════════════════════════════════════════════

RESULT:
✓ Button press detected
✓ Consensus reached (65ms)
✓ Correct elevator assigned (A won the conflict)
✓ Elevator moved to floor 1
✓ Door opened for 3 seconds
✓ Light off after order complete
✓ All elevators see the same sequence

KEY INSIGHTS:
- Consensus delay ~70ms (acceptable)
- Conflict resolution used distance + ID
- Light stayed on throughout (~3230ms total)
- All elevators synchronized throughout
```

---

## SCENARIO 3: Network Disconnect - Takeover

```
SETUP:
├─ Elevator A serving hall call on floor 3
│  ├─ floor: 3
│  ├─ state: DOOR_OPEN (holding door for 3s)
│  ├─ order: HALL UP button on floor 3
│  └─ door_timer: 1500ms remaining
├─ Elevator B at floor 0, IDLE
├─ Elevator C at floor 4, moving down
└─ Network: All connected

DISASTER: Network cable to Elevator A is PULLED OUT!

════════════════════════════════════════════════════════════════

TIMELINE:

t=0ms:
  NETWORK DISCONNECT
  └─ Elevator A's UDP socket stops sending/receiving

t=0ms (Elevator A's perspective):
  ├─ Still in DOOR_OPEN state
  ├─ Door timer continues ticking
  ├─ Isolated but still functioning

t=1500ms (Elevator A):
  └─ DOOR TIMER FIRES
      ├─ setDoor(CLOSED)
      ├─ state = DOOR_OPEN → IDLE
      ├─ Try to broadcast status... NO NETWORK!
      └─ Order is served (from A's perspective) but not reported

════════════════════════════════════════════════════════════════

MEANWHILE (Elevators B & C):

t=0ms:
  Last received message from A:
  ├─ floor: 3
  ├─ state: DOOR_OPEN
  ├─ timestamp: t=-50ms
  └─ hallOrderList[0][3][HALL_UP] = ORDER_ASSIGNED

t=50ms-2950ms:
  No message from A
  ├─ B's lastSeen[A] still has old timestamp
  ├─ C's lastSeen[A] still has old timestamp
  └─ Wait... still within timeout window

t=3000ms:
  TIMEOUT OCCURS!
  ├─ now = 3000ms
  ├─ lastSeen[A] = 0ms
  ├─ now - lastSeen[A] = 3000ms > 3000ms? YES!
  └─ A is DEAD

t=3001ms:
  NETWORK RECEIVER detects timeout
  ├─ Remove A from lastSeen
  ├─ Send to registryUpdates: LostNode = 0
  └─ Send alarm

t=3010ms:
  ORDER MANAGER (both B and C) receives:
  ├─ aliveList[0] = false
  ├─ Check orders from A:
  │   └─ hallOrderList[0][3][HALL_UP] = still ORDER_ASSIGNED
  └─ "A was supposed to take this order but A is dead!"

t=3015ms:
  REASSIGN the order!
  ├─ Check distance:
  │   ├─ B at floor 0: dist = |0 - 3| = 3
  │   └─ C at floor 4: dist = |4 - 3| = 1 ← CLOSER
  ├─ C wins! (already close and moving toward floor 3)
  └─ hallOrderList[0][3][HALL_UP] = ORDER_ASSIGNED (to C)

t=3020ms:
  B and C both broadcast:
  ├─ SenderId: 1 (B)
  ├─ hallOrderList[0][3][HALL_UP] = ORDER_ASSIGNED (to C)
  ├─ aliveList[0] = false
  └─ All receive and acknowledge

t=3025ms:
  Elevator C FSM gets order:
  ├─ currentFloor = 4 (moving down from previous order)
  ├─ targetFloor = 3 (new assignment)
  ├─ direction = DOWN (already going down)
  ├─ state = MOVING
  └─ Continue to floor 3

t=3100ms:
  Floor sensor: floor 3
  ├─ FSM: shouldStop? YES
  ├─ setMotor(STOP), setDoor(OPEN)
  ├─ state = MOVING → DOOR_OPEN
  └─ startDoorTimer(3000ms)

t=3105ms:
  HARDWARE: Door opens

t=3200ms:
  Broadcast sent by C and B:
  ├─ Elevator C: floor 3, state DOOR_OPEN
  ├─ hallOrderList[0][3][HALL_UP] = ORDER_ASSIGNED (C serving)
  └─ All receive (B and C receive each other)

t=6100ms:
  DOOR TIMER FIRES (C):
  ├─ setDoor(CLOSED)
  ├─ hallOrderList[0][3][HALL_UP] = ORDER_COMPLETE
  ├─ state = DOOR_OPEN → IDLE
  └─ Broadcast: "Order completed"

════════════════════════════════════════════════════════════════

WHAT ABOUT ELEVATOR A? (Meanwhile...)

t=6100ms (A's internal time):
  ├─ A has been idle for 3000ms
  ├─ Its copy of hallOrderList[0][3][HALL_UP] = ORDER_ASSIGNED
  ├─ But A served it and closed the door
  ├─ A doesn't know it's being reassigned
  └─ A is just sitting at floor 3, waiting for new orders

[Network restored]

t=6500ms:
  NETWORK CABLE PLUGGED BACK IN!
  └─ Elevator A comes back online

t=6501ms:
  A broadcasts:
  ├─ SenderId: 0
  ├─ floor: 3
  ├─ state: IDLE
  ├─ [stale data about orders]
  └─ "Hello! I'm back"

t=6505ms:
  B and C receive from A:
  ├─ A is back!
  ├─ aliveList[0] = true
  ├─ But notice: A doesn't know it was reassigned
  └─ A's view is stale

t=6510ms:
  B and C broadcast THEIR state:
  ├─ hallOrderList[0][3][HALL_UP] = ORDER_COMPLETE
  ├─ aliveList[0] = true (back online)
  └─ Sync A with latest data

t=6515ms:
  A receives sync:
  ├─ Updates hallOrderList[0][3][HALL_UP] = ORDER_COMPLETE
  ├─ Now A knows the order was already done
  ├─ A accepts new orders normally
  └─ SYNCHRONIZED!

════════════════════════════════════════════════════════════════

SUMMARY:

From button press to A going down:    0ms to 210ms
A serving order:                       210ms to 3210ms
Network disconnect:                    3210ms
A fails silently (still working):      3210ms to 6100ms
Timeout detected (B & C):              3000ms after disconnect
Reassignment done:                     3015ms
C gets order:                          3015ms
C reaches floor 3:                     3100ms
C serves order completely:             6100ms
A comes back online:                   6500ms
System fully synced:                   6515ms

KEY METRICS:
✓ Disconnect detected in:             3 seconds (as designed)
✓ Takeover decision made in:          15ms
✓ New elevator serving in:             100ms
✓ Order completed before A knew:       3000ms
✓ Full system recovery:                300ms after reconnect

KEY INSIGHTS:
- A continues working offline (CAB orders still work)
- B & C immediately detect the problem
- Automatic reassignment to nearest elevator
- Order still completed successfully
- System converges to consistent state within 300ms
- NO ORDERS LOST
```

---

## SCENARIO 4: Software Crash During CAB Order

```
SETUP:
├─ Elevator A has CAB order at floor 2
│  ├─ current floor: 1
│  ├─ state: MOVING
│  ├─ cabOrders: [false, false, true, false]
│  └─ Saved to disk: "cab_orders_0.txt" = "0,0,1,0"
├─ Network: Connected
└─ Hardware: Working

DISASTER: Software crash (segfault, panic, OOM...)

════════════════════════════════════════════════════════════════

TIMELINE:

t=0ms:
  CRASH!
  ├─ Main process dies
  ├─ All goroutines stopped
  ├─ Memory cleared
  ├─ BUT: File on disk still exists!
  │   └─ "cab_orders_0.txt" = "0,0,1,0" (saved at t=-100ms)
  └─ No message sent about crash

t=0ms (Other elevators):
  ├─ B and C keep broadcasting
  ├─ Elevator A's messages stop
  └─ They'll detect timeout in 3 seconds

t=100ms (User):
  └─ "Hey, elevator A crashed! Let me restart it..."

t=2000ms (User restarts):
  └─ ./main --id 0

t=2001ms:
  MAIN starts
  ├─ Initialize hardware
  ├─ Create channels
  ├─ Create modules
  └─ Call orderMgr.LoadCabOrders()

t=2002ms:
  LoadCabOrders() runs:
  ├─ Open file: "cab_orders_0.txt"
  ├─ Read content: "0,0,1,0"
  ├─ Parse: cabOrders = [false, false, true, false]
  ├─ Close file
  └─ ✓ CAB ORDER RESTORED!

t=2005ms:
  All goroutines start:
  ├─ fsm.Run()
  ├─ network.Run()
  ├─ network.Sender()
  ├─ network.Receiver()
  ├─ orderMgr.Run()
  └─ Hardware polling

t=2010ms:
  ORDER MANAGER starts FSM with:
  ├─ cabOrders[2] = true (RESTORED!)
  ├─ Send to Elevator FSM: "You have CAB order at floor 2"
  └─ ordersToElevator[2][CAB] = true

t=2015ms:
  Elevator FSM receives:
  ├─ currentFloor = 0 (safety: reset to floor 0)
  ├─ targetFloor = 2
  ├─ direction = UP
  ├─ state = IDLE → MOVING
  └─ setMotor(UP)

t=2020ms:
  NETWORK broadcasts first time:
  ├─ SenderId: 0
  ├─ floor: 0
  ├─ state: MOVING
  ├─ cabOrders: [false, false, true, false]
  └─ "I'm back with CAB order!"

t=2025ms:
  B and C receive:
  ├─ A is alive! (after 3 seconds of silence)
  ├─ See that A has CAB at floor 2
  ├─ Update aliveList[0] = true
  └─ ACK

t=2100ms:
  Floor sensor: floor 1

t=2150ms:
  Floor sensor: floor 2

t=2151ms:
  FSM at floor 2:
  ├─ Check: shouldStop? YES (CAB order at floor 2)
  ├─ setMotor(STOP), setDoor(OPEN)
  ├─ state = MOVING → DOOR_OPEN
  └─ startDoorTimer(3000ms)

t=2155ms:
  Door opens

t=5155ms:
  Door timer fires:
  ├─ setDoor(CLOSED)
  ├─ cabOrders[2] = false (CLEAR!)
  ├─ SaveCabOrders() → UPDATE file: "0,0,0,0"
  ├─ state = DOOR_OPEN → IDLE
  └─ Broadcast: "Order completed"

t=5160ms:
  All elevators see:
  ├─ hallOrderList[0][2][CAB] = false
  ├─ Elevator A is back in service
  └─ All synced

════════════════════════════════════════════════════════════════

RESULT:

From crash to detection:               0ms (immediate crash)
From crash to restart (user action):   2000ms
From restart to CAB restored:          4ms
From restart to moving:                15ms
From restart to floor 2:               150ms
From restart to order complete:        3155ms

KEY FACTS:
✓ CAB order was NOT lost (persistent on disk)
✓ Restored automatically on startup
✓ Elevator served order completely
✓ System recovered automatically within 3 seconds
✓ Other elevators detected absence and compensated
✓ Full sync achieved within 160ms of restart

KEY INSIGHTS:
- Persistent storage is CRITICAL
- File survives the crash
- LoadCabOrders() is called at startup
- Order is restored and served
- This is the ONLY way to guarantee "no calls are lost"
```

════════════════════════════════════════════════════════════════

## SUMMARY TABLE

| Scenario | Detection | Recovery | CAB Lost? |
|----------|-----------|----------|-----------|
| Normal CAB | Immediate | N/A | NO |
| Normal Hall | ~70ms | N/A | NO |
| Network DC | 3s timeout | Auto reassign | NO |
| SW Crash | 3s timeout | Restart+restore | NO |
| Door Block | Immediate | Timer extends | NO |
| Packet Loss | N/A | Idempotent retry | NO |

════════════════════════════════════════════════════════════════

