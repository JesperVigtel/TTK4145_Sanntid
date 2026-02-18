# Complete LocalControl Implementation Checklist

**Project:** Elevator System - Distributed Controller  
**Module:** LocalControl (FSM)  
**Date:** 18. februar 2026

---

## FILE 1: timer.go

### Package & Imports
- [ ] Change package from `timer` to `localControl`
- [ ] Import `time` package

### DoorTimer Struct (3-second door duration)
- [ ] Create `DoorTimer` struct with fields:
  - [ ] `timeout time.Duration`
  - [ ] `timer *time.Timer`
  - [ ] `active bool`
  - [ ] `timerOut chan bool`

- [ ] **NewDoorTimer(duration)** constructor:
  - [ ] Initialize all fields
  - [ ] Create buffered channel (size 1)
  - [ ] Return pointer to DoorTimer

- [ ] **Start()** method:
  - [ ] Check if not already active
  - [ ] Reset timer to timeout
  - [ ] Set active = true
  - [ ] Launch goroutine to wait for timer and send to channel

- [ ] **Stop()** method:
  - [ ] Check if active
  - [ ] Stop timer
  - [ ] Set active = false

- [ ] **Channel()** method:
  - [ ] Return timerOut as read-only channel

### WatchdogTimer Struct (5-second inactivity monitor)
- [ ] Create `WatchdogTimer` struct with fields:
  - [ ] `timeout time.Duration`
  - [ ] `timer *time.Timer`
  - [ ] `active bool`
  - [ ] `watchdogOut chan bool`
  - [ ] `stopChan chan bool`

- [ ] **NewWatchdogTimer(duration)** constructor:
  - [ ] Initialize all fields
  - [ ] Create buffered channels (size 1)
  - [ ] Return pointer to WatchdogTimer

- [ ] **Start()** method:
  - [ ] Check if not already active
  - [ ] Set active = true
  - [ ] Reset timer
  - [ ] Launch goroutine with select loop monitoring timer and stopChan

- [ ] **Kick()** method:
  - [ ] Check if active
  - [ ] Reset timer to timeout

- [ ] **Stop()** method:
  - [ ] Check if active
  - [ ] Set active = false
  - [ ] Stop timer
  - [ ] Send to stopChan

- [ ] **Channel()** method:
  - [ ] Return watchdogOut as read-only channel

---

## FILE 2: types.go (Add new event types)

### Communication Event Types

#### OrderEvent struct (decisionMaker → localControl)
- [ ] **OrderEvent** - Order assignment from decisionMaker:
  - [ ] `Floor int` - Target floor to service
  - [ ] `Button ButtonType` - Type of button (BTHallUp/BTHallDown/BTCab)

#### ElevatorStateEvent struct (localControl → decisionMaker)
- [ ] **ElevatorStateEvent** - Complete elevator state update:
  - [ ] `CurrentFloor int` - Current elevator position (-1 if between floors)
  - [ ] `Direction MotorDirection` - Motor direction (Up/Down/Stop)
  - [ ] `Behaviour ElevatorBehaviour` - Current FSM state (Idle/Moving/DoorOpen)
  - [ ] `DoorOpen bool` - Door status (true/false)
  - [ ] `Obstructed bool` - Obstruction switch status (true/false)
  - [ ] `ButtonPressed *ButtonEvent` - Button press event (nil if none), contains:
    - [ ] `Floor int` - Which floor button
    - [ ] `Button ButtonType` - BTCab/BTHallUp/BTHallDown
  - [ ] `Requests [NFloors][NButtons]bool` - Current active requests queue
  - [ ] `Error bool` - Error/watchdog timeout flag
  - [ ] `Message string` - Human-readable status/error message

---

## FILE 3: localControll.go

### Package & Imports
- [ ] Package declaration: `package localControl`
- [ ] Import:
  - [ ] `elevator/internal/config`
  - [ ] `elevator/internal/localControll/hardware`
  - [ ] `elevator/internal/types`
  - [ ] `time`
  - [ ] `fmt`

### Main State Variables
- [ ] `currentFloor int` - Last known floor (-1 if unknown)
- [ ] `direction types.MotorDirection` - Current direction (Up/Down/Stop)
- [ ] `behaviour types.ElevatorBehaviour` - Current FSM state (Idle/Moving/DoorOpen)
- [ ] `requests [config.NFloors][config.NButtons]bool` - Order queue matrix
- [ ] `obstructed bool` - Current obstruction status

### Channel Variables
- [ ] `floorSensor chan int` - Floor arrival notifications
- [ ] `buttonPress chan types.ButtonEvent` - Button press events
- [ ] `obstruction chan bool` - Obstruction switch status
- [ ] `doorTimer *DoorTimer` - Door open timer (3 seconds)
- [ ] `watchdog *WatchdogTimer` - Inactivity watchdog (5 seconds)

### Init() Function
- [ ] **Initialize hardware:**
  - [ ] Call `hardware.Init(config.Addr, config.NFloors)`

- [ ] **Initialize state:**
  - [ ] Set all requests to false (nested loop)
  - [ ] Set behaviour to Idle
  - [ ] Set direction to Stop
  - [ ] Set currentFloor to -1
  - [ ] Set obstructed to false

- [ ] **Create channels:**
  - [ ] Create floorSensor channel (buffered, size 100)
  - [ ] Create buttonPress channel (buffered, size 100)
  - [ ] Create obstruction channel (buffered, size 100)

- [ ] **Start polling goroutines:**
  - [ ] Launch `go hardware.PollFloorSensor(floorSensor)`
  - [ ] Launch `go hardware.PollButtons(buttonPress)`
  - [ ] Launch `go hardware.PollObstructionSwitch(obstruction)`

- [ ] **Initialize timers:**
  - [ ] Create doorTimer: `doorTimer = NewDoorTimer(config.DoorOpenDuration)`
  - [ ] Create watchdog: `watchdog = NewWatchdogTimer(5 * time.Second)`
  - [ ] Start watchdog: `watchdog.Start()`

- [ ] **Find initial floor:**
  - [ ] Set motor direction down: `hardware.SetMotorDirection(types.Down)`
  - [ ] Wait for floor sensor: `floor := <-floorSensor`
  - [ ] Update currentFloor: `currentFloor = floor`
  - [ ] Stop motor: `hardware.SetMotorDirection(types.Stop)`
  - [ ] Update floor indicator: `hardware.SetFloorIndicator(currentFloor)`
  - [ ] Set behaviour to Idle

### Run() Function
- [ ] **Function signature:**
  - [ ] `func Run(orderChan <-chan types.OrderEvent, stateChan chan<- types.ElevatorStateEvent)`
  - [ ] No return value (infinite loop)

- [ ] **Main select loop** - `for { select { ... } }`:

#### Case 1: Order Received from DecisionMaker
```go
case order := <-orderChan:
```
- [ ] Kick watchdog: `watchdog.Kick()`
- [ ] Store order: `requests[order.Floor][order.Button] = true`
- [ ] Turn on button lamp: `hardware.SetButtonLamp(order.Button, order.Floor, true)`
- [ ] If behaviour == Idle:
  - [ ] Calculate direction: `direction = chooseDirection()`
  - [ ] If direction != Stop:
    - [ ] Start motor: `hardware.SetMotorDirection(direction)`
    - [ ] Set behaviour = Moving
- [ ] Create state event with current state
- [ ] Send to stateChan

#### Case 2: Floor Arrival
```go
case floor := <-floorSensor:
```
- [ ] Kick watchdog: `watchdog.Kick()`
- [ ] Update currentFloor: `currentFloor = floor`
- [ ] Update indicator: `hardware.SetFloorIndicator(floor)`
- [ ] If `shouldStop(floor, direction)`:
  - [ ] Stop motor: `hardware.SetMotorDirection(types.Stop)`
  - [ ] Open door: `hardware.SetDoorOpenLamp(true)`
  - [ ] Clear requests: `clearRequestsAtFloor(floor)`
  - [ ] Start door timer: `doorTimer.Start()`
  - [ ] Set behaviour = DoorOpen
- [ ] Create state event with all current state info
- [ ] Send to stateChan

#### Case 3: Door Timer Expired
```go
case <-doorTimer.Channel():
```
- [ ] Kick watchdog: `watchdog.Kick()`
- [ ] Close door: `hardware.SetDoorOpenLamp(false)`
- [ ] Calculate next direction: `direction = chooseDirection()`
- [ ] If direction != Stop:
  - [ ] Start motor: `hardware.SetMotorDirection(direction)`
  - [ ] Set behaviour = Moving
- [ ] Else:
  - [ ] Set behaviour = Idle
- [ ] Create state event
- [ ] Send to stateChan

#### Case 4: Obstruction Switch
```go
case obs := <-obstruction:
```
- [ ] Kick watchdog: `watchdog.Kick()`
- [ ] Update state: `obstructed = obs`
- [ ] If obs == true and behaviour == DoorOpen:
  - [ ] Stop door timer: `doorTimer.Stop()`
  - [ ] Restart door timer: `doorTimer.Start()` (resets countdown)
- [ ] Create state event with Obstructed = obs
- [ ] Send to stateChan

#### Case 5: Button Pressed (Cab Buttons)
```go
case btn := <-buttonPress:
```
- [ ] Kick watchdog: `watchdog.Kick()`
- [ ] Store request: `requests[btn.Floor][btn.Button] = true`
- [ ] Turn on lamp: `hardware.SetButtonLamp(btn.Button, btn.Floor, true)`
- [ ] If behaviour == Idle:
  - [ ] Calculate direction: `direction = chooseDirection()`
  - [ ] If direction != Stop:
    - [ ] Start motor: `hardware.SetMotorDirection(direction)`
    - [ ] Set behaviour = Moving
- [ ] Create state event with ButtonPressed = &btn (pointer to button event)
- [ ] Send to stateChan

#### Case 6: Watchdog Timeout
```go
case <-watchdog.Channel():
```
- [ ] Create error state event:
  - [ ] Set Error = true
  - [ ] Set Message = "Watchdog timeout: No activity for 5 seconds"
  - [ ] Include all current state (floor, direction, behaviour, etc.)
- [ ] Send to stateChan
- [ ] Restart watchdog: `watchdog.Start()`
- [ ] Optional: Stop motor for safety

### Helper Functions

#### shouldStop(floor int, dir MotorDirection) bool
- [ ] **Check if should stop at this floor:**
  - [ ] If requests[floor][BTCab] == true: return true (always stop for cab)
  - [ ] If dir == Up:
    - [ ] If requests[floor][BTHallUp] == true: return true
    - [ ] If no more requests above: return true (last stop going up)
  - [ ] If dir == Down:
    - [ ] If requests[floor][BTHallDown] == true: return true
    - [ ] If no more requests below: return true (last stop going down)
  - [ ] Return false

#### chooseDirection() MotorDirection
- [ ] **Determine next movement direction:**
  - [ ] Loop through floors above currentFloor:
    - [ ] If any requests exist: return Up
  - [ ] Loop through floors below currentFloor:
    - [ ] If any requests exist: return Down
  - [ ] If no requests anywhere: return Stop

#### clearRequestsAtFloor(floor int)
- [ ] **Clear all requests at given floor:**
  - [ ] Loop through all button types (0 to NButtons-1):
    - [ ] If requests[floor][btnType] == true:
      - [ ] Set requests[floor][btnType] = false
      - [ ] Turn off lamp: `hardware.SetButtonLamp(btnType, floor, false)`

#### sendStateUpdate(stateChan) or inline in each case
- [ ] **Create ElevatorStateEvent with:**
  - [ ] CurrentFloor: currentFloor
  - [ ] Direction: direction
  - [ ] Behaviour: behaviour
  - [ ] DoorOpen: (behaviour == ElevatorDoorOpen)
  - [ ] Obstructed: obstructed
  - [ ] ButtonPressed: nil (or pointer if this is button event)
  - [ ] Requests: copy of requests array
  - [ ] Error: false (or true for watchdog)
  - [ ] Message: appropriate status message
- [ ] Send to stateChan (consider non-blocking with select/default)

---

## FILE 4: elevio.go (Fix existing issue)

### Fix ButtonType Reference in PollButtons
- [ ] Line ~73 in PollButtons function:
  - [ ] Change: `for b := ButtonType(0); b < 3; b++`
  - [ ] To: `for b := types.ButtonType(0); b < 3; b++`

### Verify All Type References
- [ ] Ensure all uses of ButtonType are `types.ButtonType`
- [ ] Ensure all uses of ButtonEvent are `types.ButtonEvent`
- [ ] Ensure all uses of MotorDirection are `types.MotorDirection`

---

## Integration Points

### Main.go Setup (for reference)
- [ ] Create order channel: `orderChan := make(chan types.OrderEvent, 100)`
- [ ] Create state channel: `stateChan := make(chan types.ElevatorStateEvent, 100)`
- [ ] Start localControl: `go localControl.Run(orderChan, stateChan)`
- [ ] DecisionMaker sends to orderChan
- [ ] DecisionMaker receives from stateChan

---

## Testing Checklist

### Basic FSM Flow
- [ ] **Idle → Order → Moving:**
  - [ ] Send OrderEvent to orderChan
  - [ ] Verify motor starts
  - [ ] Verify ElevatorStateEvent sent with Behaviour=Moving

- [ ] **Moving → Floor Arrival → Stop:**
  - [ ] Trigger floor sensor
  - [ ] Verify motor stops at correct floor
  - [ ] Verify door opens
  - [ ] Verify button lights turn off
  - [ ] Verify ElevatorStateEvent sent with DoorOpen=true

- [ ] **Door Open → Timer → Close:**
  - [ ] Wait 3 seconds
  - [ ] Verify door closes
  - [ ] Verify next movement or idle
  - [ ] Verify ElevatorStateEvent sent

### Obstruction Handling
- [ ] **Obstruction During Door Open:**
  - [ ] Block door while open
  - [ ] Verify door stays open
  - [ ] Verify ElevatorStateEvent with Obstructed=true
  - [ ] Clear obstruction
  - [ ] Verify door closes after 3 seconds
  - [ ] Verify ElevatorStateEvent with Obstructed=false

### Button Events
- [ ] **Cab Button Pressed:**
  - [ ] Press cab button
  - [ ] Verify ElevatorStateEvent with ButtonPressed containing Floor and Button=BTCab
  - [ ] Verify request stored
  - [ ] Verify lamp turns on

- [ ] **Hall Button Pressed:**
  - [ ] Press hall button
  - [ ] Verify ElevatorStateEvent with ButtonPressed containing Floor and Button=BTHallUp/Down
  - [ ] Verify request stored
  - [ ] Verify lamp turns on

### Watchdog
- [ ] **Watchdog Timeout:**
  - [ ] Stop all activity
  - [ ] Wait 5 seconds
  - [ ] Verify ElevatorStateEvent with Error=true and Message about watchdog
  - [ ] Verify watchdog restarts

- [ ] **Watchdog Reset:**
  - [ ] Ensure events keep occurring
  - [ ] Verify watchdog never fires
  - [ ] No error messages

### Edge Cases
- [ ] Multiple orders at different floors
- [ ] Orders in opposite direction while moving
- [ ] Rapid button presses
- [ ] Door obstruction cleared/blocked repeatedly
- [ ] Button pressed while door open

---

## Key Implementation Notes

### State Updates to DecisionMaker
- Send ElevatorStateEvent after EVERY state change
- Always include complete state (floor, direction, behaviour, door, obstruction, requests)
- Use ButtonPressed field only when button event triggers the update
- Include error flag and message for watchdog timeouts

### Watchdog Usage
- Call `watchdog.Kick()` at the START of every case in select loop
- This ensures any activity resets the timeout
- Watchdog fires only if truly stuck (no events for 5 seconds)

### Channel Buffering
- Use buffered channels (size 100) to prevent blocking
- Consider non-blocking sends to stateChan if needed

### Thread Safety
- All state variables accessed only in main Run() loop
- No mutex needed since single goroutine manages state

---

## Configuration Constants (from config.go)

```go
NFloors             = 4
NButtons            = 3
DoorOpenDuration    = 3 * time.Second
ChannelBufferSize   = 100
Addr                = "localhost:15657"
```

---

## Communication Flow Diagram

```
DecisionMaker                LocalControl               Hardware
     |                            |                         |
     |--OrderEvent--------------->|                         |
     |   (Floor, ButtonType)      |                         |
     |                            |--SetMotorDirection----->|
     |                            |                         |
     |                            |<--FloorSensor-----------|
     |                            |   (floor number)        |
     |                            |--SetDoorOpen----------->|
     |                            |--SetButtonLamp(false)-->|
     |<--ElevatorStateEvent-------|                         |
     |   (complete state)         |                         |
     |                            |<--ButtonPress-----------|
     |<--ElevatorStateEvent-------|                         |
     |   (ButtonPressed field)    |                         |
```

---

**End of Checklist**
