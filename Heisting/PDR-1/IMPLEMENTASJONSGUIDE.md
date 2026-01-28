# Update 26 Jan 2026 — Implementing Solution 3 Enhancements
# Update 26 Jan 2026 — Implementing Solution 3 Enhancements

Configuration (add to your config):
- `network.broadcast_interval_ms: 5` (tunable 2–50)
- `faults.motor_stop_timeout_s: 4`
- `faults.obstruction_timeout_s: 5`
- `network.type_tagged_envelope: true`

Message envelope (Go):
```go
type Envelope struct {
    TypeId     string
    PayloadJSON []byte
}
```

Network transmitter:
```go
ticker := time.NewTicker(time.Duration(cfg.BroadcastIntervalMs) * time.Millisecond)
for range ticker.C {
    env := Envelope{TypeId: reflect.TypeOf(msg).String(), PayloadJSON: mustJSON(msg)}
    conn.Write(mustJSON(env))
}
```

Network receiver:
```go
var env Envelope
if err := json.Unmarshal(pkt, &env); err == nil {
    switch env.TypeId {
    case "types.PeriodicMsg": handlePeriodic(unmarshal(env.PayloadJSON))
    case "types.FaultMsg":    handleFault(unmarshal(env.PayloadJSON))
    }
}
```

Fault monitoring:
```go
if behaviour == EB_Moving && time.Since(motorStart) > cfg.MotorStopTimeout {
    onlineStatus.Set(false) // pause broadcast; persist CAB ensures recovery
}
if doorObstructed && time.Since(obstructionStart) > cfg.ObstructionTimeout {
    onlineStatus.Set(false)
}
if !doorObstructed { onlineStatus.Set(true) }
```

Notes:
- Do not implement automatic reboot; rely on CAB persistence on restart.
- Keep existing FSM; align request stop/clear rules to Solution 3 semantics.

Specs Coverage Snapshot:
- No calls lost: CAB persistence retained and used during restart flows.
- Button light guarantee: Idempotent state propagation + 5ms broadcast cadence.
- Fault tolerance: Heartbeat + motor/obstruction timeouts toggling online/offline.
- Packet loss: Type-tagged envelope with idempotent handlers; periodic resend.
- Scalability: HRA maintained; intervals and timeouts configurable.

# IMPLEMENTASJONSGUIDE - Kombinert Heisprosjekt Løsning

## Oversikt

Denne guiden forklarer hvordan du implementerer den kombinerte løsningen som ble beskrevet i PDD.

---

## 1. MODUL 1: ELEVATOR FSM

### Ansvar
- Håndtere fysisk heisoperasjon
- Implementere 5-tilstand maskin
- Håndtere obstruksjon og timers

### Kode-struktur (pseudokode)

```go
type ElevatorFSM struct {
    state               State          // IDLE, MOVING, DOOR_OPEN, EMERGENCY_STOP
    currentFloor        int
    direction           Direction      // UP, DOWN, STOP
    doorOpen            bool
    obstruction         bool
    doorTimer           time.Timer
    motorTimer          time.Timer
    
    // Channels
    floorSensor         <-chan int     // Fra hardware når etasje nådd
    obstructionSwitch   <-chan bool    // Fra hardware når obstruksjon
    newOrders           <-chan OrderAssignment // Fra Order Manager
    statusOut           chan<- ElevatorStatus  // Til Order Manager
}

func (e *ElevatorFSM) Run() {
    for {
        select {
        case floor := <-e.floorSensor:
            e.currentFloor = floor
            e.handleFloorReached()
            
        case e.obstruction = <-e.obstructionSwitch:
            e.handleObstruction()
            
        case order := <-e.newOrders:
            e.handleNewOrder(order)
            
        case <-e.doorTimer.C:
            e.handleDoorTimeout()
            
        case <-e.motorTimer.C:
            e.handleMotorTimeout()
        }
    }
}

// Tilstandslogikk
func (e *ElevatorFSM) handleFloorReached() {
    switch e.state {
    case IDLE, EMERGENCY_STOP:
        // Ignorer
        
    case MOVING:
        // Sjekk: Skal vi stoppe her?
        if e.shouldStop() {
            e.setMotor(STOP)
            e.setDoor(OPEN)
            e.state = DOOR_OPEN
            e.startDoorTimer(3 * time.Second)
        }
    }
}

func (e *ElevatorFSM) shouldStop() bool {
    // Sjekk: Er det en ordre på denne etasjen i kjøreretningen?
    // Eller CAB-ordre på denne etasjen?
    // Eller skal vi bytte retning?
    return checkOrdersAtFloor() || checkCabAtFloor() || shouldChangeDirection()
}

func (e *ElevatorFSM) handleObstruction() {
    if e.state == DOOR_OPEN && e.obstruction {
        // Obstruksjon aktiv - hold døren åpen
        e.state = EMERGENCY_STOP
        e.setMotor(STOP)
        e.startDoorTimer(3 * time.Second)  // Reset timeren
    } else if e.state == EMERGENCY_STOP && !e.obstruction {
        // Obstruksjon fjernet - fortsett
        e.state = DOOR_OPEN
        e.startDoorTimer(3 * time.Second)
    }
}

func (e *ElevatorFSM) handleDoorTimeout() {
    if e.state == DOOR_OPEN || e.state == EMERGENCY_STOP {
        e.setDoor(CLOSED)
        e.state = IDLE
        
        // Notifiser Order Manager at vi lukket dør
        e.statusOut <- ElevatorStatus{
            floor: e.currentFloor,
            state: e.state,
            direction: STOP,
        }
    }
}
```

### Key Points
- **Tilstandsmaskinen skal være selvsentrert** - fokus på egen tilstand
- **Døropen varighet: 3 sekunder** - brukes timer
- **Obstruksjon er høyeste prioritet** - avbryter alt
- **Rapportér status til Order Manager hvert millisekund**

---

## 2. MODUL 2: NETWORK MODULE

### Ansvar
- Sende og motta broadcasts
- Holde liste over aktive heiser
- Detektere når heiser forsvinner/kommer tilbake

### Kode-struktur

```go
type NetworkModule struct {
    myID                int
    port                int
    
    // State
    elevatorList        [N]Elevator            // Andre heisers tilstand
    aliveList           [N]bool                // Hvem som er online
    hallOrderList       [N][F][2]ButtonState   // Hall ordre fra alle
    lastSeen            map[int]time.Time      // Heartbeat tracking
    
    // Channels
    broadcastOut        chan<- Message         // Til sender goroutine
    broadcastIn         <-chan Message         // Fra receiver goroutine
    registryUpdates     <-chan RegistryUpdate  // Fra receiver: hvem kom/gikk
    worldviewIn         <-chan LocalState      // Fra Order Manager
    worldviewOut        chan<- GlobalState     // Til Order Manager
}

func (n *NetworkModule) Sender(broadcastChan chan<- Message) {
    ticker := time.NewTicker(50 * time.Millisecond)
    defer ticker.Stop()
    
    for range ticker.C {
        msg := Message{
            SenderId:      n.myID,
            ElevatorList:  n.elevatorList,
            HallOrderList: n.hallOrderList,
            OnlineStatus:  true,
            AliveList:     n.aliveList,
        }
        broadcastChan <- msg
    }
}

func (n *NetworkModule) Receiver(listenAddr string, inChan chan<- Message, registryChan chan<- RegistryUpdate) {
    // Lyt på UDP broadcast
    // Dekod JSON og send på channel
    // Håndter heartbeat-timeout
    
    for {
        // Read UDP packet
        // Unmarshal JSON → Message
        
        select {
        case inChan <- msg:
            n.lastSeen[msg.SenderId] = time.Now()
        }
        
        // Sjekk for timeout
        n.checkHeartbeatTimeout(registryChan)
    }
}

func (n *NetworkModule) checkHeartbeatTimeout(registryChan chan<- RegistryUpdate) {
    now := time.Now()
    for id, lastTime := range n.lastSeen {
        if now.Sub(lastTime) > 3*time.Second {
            // TIMEOUT - heisen er offline
            n.aliveList[id] = false
            registryChan <- RegistryUpdate{
                LostNode: id,
            }
            delete(n.lastSeen, id)
        }
    }
}

func (n *NetworkModule) Run() {
    for {
        select {
        case msg := <-n.broadcastIn:
            // Mottak fra annen heisen
            n.aliveList[msg.SenderId] = true
            n.elevatorList[msg.SenderId] = msg.ElevatorList[msg.SenderId]
            n.hallOrderList[msg.SenderId] = msg.HallOrderList[msg.SenderId]
            
            // Sjekk: Hele nettverket synkronisert?
            if n.checkConsensus() {
                // Send oppdatert worldview til Order Manager
                n.worldviewOut <- GlobalState{
                    AliveList:     n.aliveList,
                    ElevatorList:  n.elevatorList,
                    HallOrderList: n.hallOrderList,
                }
            }
            
        case reg := <-n.registryUpdates:
            // Heisen gikk offline/online
            n.handleNodeChange(reg)
            
        case local := <-n.worldviewIn:
            // Order Manager sender sitt lokale tilstand
            n.elevatorList[n.myID] = local.MyElevator
            n.hallOrderList[n.myID] = local.MyHallOrders
        }
    }
}

func (n *NetworkModule) checkConsensus() bool {
    // Sjekk: Sender alle heiser samme versjon av data?
    // Implementer ved hjelp av versjonsnummer eller checksum
    for i := 0; i < NElevators; i++ {
        if !n.aliveList[i] {
            continue
        }
        // Sammenlik med tidligere mottatte versjon
        // Hvis alle er like → true
    }
    return true
}
```

### Key Points
- **Heartbeat timeout: 3 sekunder** - hvis ikke hørt fra heis på 3s → marked offline
- **Broadcast hver 50ms** - må sendes fra egen goroutine (ikke blocking)
- **Receiver må være fra egen goroutine** - og populere channel
- **Konsensus = alle heiser har samme versjon av data**

---

## 3. MODUL 3: ORDER MANAGER

### Ansvar
- Holde styr på CAB-ordrer (persistent)
- Fordele hall-ordrer
- Håndtere konfliktløsning
- Styre button-lys

### Kode-struktur

```go
type OrderManager struct {
    myID                int
    cabOrders           [F]bool                // Lokale CAB-ordrer
    hallLightState      [N][F][2]ButtonState   // Tilstand for lys
    
    // Consensus tracking
    ackMap              [N]bool                // Hvem som ACK'et siste broadcast
    consensus           bool                   // Hele nettverket synkronisert?
    
    // Channels
    elevatorStatus      <-chan ElevatorStatus  // Fra Elevator FSM
    networkWorldview    <-chan GlobalState     // Fra Network Module
    networkState        chan<- LocalState      // Til Network Module
    ordersToElevator    chan<- [F][2]bool      // Til Elevator FSM
    lightsOut           chan<- LightCommand    // Til hardware
}

// ============================================================
// CAB PERSISTENCE
// ============================================================

func (om *OrderManager) LoadCabOrders() error {
    filename := fmt.Sprintf("cab_orders_%d.txt", om.myID)
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return err
    }
    
    // Parse data format: "0,1,0,1" = [false, true, false, true]
    orders := strings.Split(string(data), ",")
    for i, o := range orders {
        om.cabOrders[i] = (o == "1")
    }
    return nil
}

func (om *OrderManager) SaveCabOrders() error {
    filename := fmt.Sprintf("cab_orders_%d.txt", om.myID)
    
    // Convert to string: [true, false, true, false] → "1,0,1,0"
    var sb strings.Builder
    for i, order := range om.cabOrders {
        if i > 0 {
            sb.WriteString(",")
        }
        if order {
            sb.WriteString("1")
        } else {
            sb.WriteString("0")
        }
    }
    
    return ioutil.WriteFile(filename, []byte(sb.String()), 0644)
}

// ============================================================
// ORDER ASSIGNMENT & CONFLICT RESOLUTION
// ============================================================

func (om *OrderManager) AssignHallOrder(floor int, button int, 
    elevatorList [N]Elevator, aliveList [N]bool) int {
    
    // Hvem skal ta denne ordren? Basert på:
    // 1. Hvem er nærmest?
    // 2. Hvis likt: hvem har lavest ID?
    
    bestElevator := -1
    bestDistance := math.MaxInt
    
    for id := 0; id < N; id++ {
        if !aliveList[id] {
            continue  // Heis er offline
        }
        
        dist := abs(elevatorList[id].floor - floor)
        if dist < bestDistance || (dist == bestDistance && id < bestElevator) {
            bestElevator = id
            bestDistance = dist
        }
    }
    
    return bestElevator  // -1 hvis ingen er online (error)
}

// ============================================================
// BUTTON LIGHT STATE MACHINE
// ============================================================

func (om *OrderManager) UpdateButtonLightStates(hallOrderList [N][F][2]ButtonState) {
    for floor := 0; floor < F; floor++ {
        for btn := 0; btn < 2; btn++ {
            // For hver hall ordre, hold styr på hvor den er i livssyklusen
            currentState := om.hallLightState[0][floor][btn]  // Simplified: use node 0
            
            switch hallOrderList[0][floor][btn] {
            case STANDBY:
                om.setLight(floor, btn, false)
                om.hallLightState[0][floor][btn] = STANDBY
                
            case BUTTON_PRESSED:
                om.setLight(floor, btn, true)
                om.hallLightState[0][floor][btn] = BUTTON_PRESSED
                
            case ORDER_ASSIGNED:
                om.setLight(floor, btn, true)
                om.hallLightState[0][floor][btn] = ORDER_ASSIGNED
                
            case ORDER_COMPLETE:
                om.setLight(floor, btn, false)
                om.hallLightState[0][floor][btn] = STANDBY
            }
        }
    }
}

// ============================================================
// MAIN RUN LOOP
// ============================================================

func (om *OrderManager) Run() {
    // Load CAB orders from disk on startup
    om.LoadCabOrders()
    
    for {
        select {
        case status := <-om.elevatorStatus:
            // Elevator rapporterer sin tilstand
            // (floor, state, direction)
            // → Send til Network Module
            om.networkState <- LocalState{
                MyElevator: status,
                MyHallOrders: om.GetMyHallOrders(),
            }
            
        case worldview := <-om.networkWorldview:
            // Network sender oppdatert global tilstand
            
            // 1. Fordel hall-ordrer
            assignedOrders := om.assignAllOrders(
                worldview.ElevatorList,
                worldview.AliveList,
            )
            
            // 2. Send nye ordrer til Elevator FSM
            if assignedOrders != om.lastAssignedOrders {
                om.ordersToElevator <- assignedOrders
                om.lastAssignedOrders = assignedOrders
            }
            
            // 3. Oppdater button-lys
            om.UpdateButtonLightStates(worldview.HallOrderList)
            
        case buttonPress := <-om.buttonPressChannel:
            // Bruker trykket på en knapp
            floor := buttonPress.Floor
            button := buttonPress.Button
            
            if button == CAB {
                // CAB-ordre
                om.cabOrders[floor] = true
                om.SaveCabOrders()  // PERSIST IMMEDIATELY
                om.setLight(floor, button, true)
                
            } else {
                // HALL-ordre
                // Broadcast til nettverket (network module håndterer)
                // Lys tendes når ORDER_ASSIGNED
            }
        }
    }
}

func (om *OrderManager) setLight(floor, button int, on bool) {
    om.lightsOut <- LightCommand{
        Floor:  floor,
        Button: button,
        On:     on,
    }
}
```

### Key Points
- **Persistent lagring:** Lagre CAB-ordrer til disk UMIDDELBART etter endring
- **Konfliktløsning:** Distanse + ID som tiebreaker
- **Lys-tilstand:** Må synkroniseres fra alle heiser (IKKE lokalt)
- **CAB lys:** ON umiddelbart
- **Hall lys:** ON bare når ORDER_ASSIGNED nådd

---

## 4. MAIN / COORDINATOR

```go
func main() {
    nodeID := parseArgs()  // --id <number>
    
    // Initialize hardware
    hwelevio.Init("localhost:15657")
    
    // Create channels
    floorSensor := make(chan int)
    obstructionSwitch := make(chan bool)
    buttonPress := make(chan ButtonEvent)
    
    motorCommand := make(chan Direction)
    doorCommand := make(chan bool)
    lightsCommand := make(chan LightCommand)
    
    networkBroadcast := make(chan Message)
    networkReceive := make(chan Message)
    registryUpdates := make(chan RegistryUpdate)
    
    elevatorStatus := make(chan ElevatorStatus)
    networkWorldview := make(chan GlobalState)
    networkState := make(chan LocalState)
    ordersToElevator := make(chan [F][2]bool)
    
    // Start hardware polling (goroutines)
    go hwelevio.PollFloorSensor(floorSensor)
    go hwelevio.PollObstructionSwitch(obstructionSwitch)
    go hwelevio.PollButtons(buttonPress)
    
    // Create modules
    fsm := &ElevatorFSM{
        state: IDLE,
        currentFloor: 0,
        floorSensor: floorSensor,
        obstructionSwitch: obstructionSwitch,
        newOrders: ordersToElevator,
        statusOut: elevatorStatus,
    }
    
    network := &NetworkModule{
        myID: nodeID,
        port: 1338,
        broadcastOut: networkBroadcast,
        broadcastIn: networkReceive,
        registryUpdates: registryUpdates,
        worldviewIn: networkState,
        worldviewOut: networkWorldview,
    }
    
    orderMgr := &OrderManager{
        myID: nodeID,
        elevatorStatus: elevatorStatus,
        networkWorldview: networkWorldview,
        networkState: networkState,
        ordersToElevator: ordersToElevator,
        lightsOut: lightsCommand,
    }
    
    // Start modules (goroutines)
    go fsm.Run()
    go network.Run()
    go network.Sender(networkBroadcast)
    go network.Receiver(":1338", networkReceive, registryUpdates)
    go orderMgr.Run()
    
    // Start hardware output drivers
    go driveMotor(motorCommand)
    go driveDoor(doorCommand)
    go driveLights(lightsCommand)
    
    // Infinite loop to keep program running
    select {}
}
```

---

## 5. DATA STRUCTURES

```go
// ============================================================
// ENUM TYPES
// ============================================================

type State int
const (
    INIT_STATE State = iota
    IDLE
    MOVING
    DOOR_OPEN
    EMERGENCY_STOP
)

type Direction int
const (
    DOWN Direction = -1
    STOP Direction = 0
    UP Direction = 1
)

type ButtonType int
const (
    HALL_UP ButtonType = 0
    HALL_DOWN ButtonType = 1
    CAB ButtonType = 2
)

type ButtonLightState int
const (
    STANDBY ButtonLightState = iota
    BUTTON_PRESSED
    ORDER_ASSIGNED
    ORDER_COMPLETE
)

// ============================================================
// MESSAGES & STRUCTS
// ============================================================

type Elevator struct {
    Id        int
    Floor     int
    Direction Direction
    State     State
    CabOrders [F]bool
}

type Message struct {
    SenderId      int
    ElevatorList  [N]Elevator
    HallOrderList [N][F][2]ButtonState
    OnlineStatus  bool
    AliveList     [N]bool
}

type ElevatorStatus struct {
    Floor     int
    State     State
    Direction Direction
}

type GlobalState struct {
    AliveList     [N]bool
    ElevatorList  [N]Elevator
    HallOrderList [N][F][2]ButtonState
}

type LocalState struct {
    MyElevator    Elevator
    MyHallOrders  [F][2]ButtonState
}

type ButtonEvent struct {
    Floor  int
    Button ButtonType
}

type LightCommand struct {
    Floor  int
    Button ButtonType
    On     bool
}

type RegistryUpdate struct {
    NewNode  int
    LostNode int
}
```

---

## 6. CONFIG / CONSTANTS

```go
const (
    N_ELEVATORS         = 3
    N_FLOORS            = 4
    N_BUTTONS           = 3
    
    BROADCAST_PORT      = 1338
    BROADCAST_INTERVAL  = 50 * time.Millisecond
    HEARTBEAT_TIMEOUT   = 3 * time.Second
    DOOR_OPEN_DURATION  = 3 * time.Second
    MOTOR_TIMEOUT       = 4 * time.Second
)
```

---

## 7. TESTING STRATEGY

### Test 1: Normal Hall Call
1. Trykk hall-knapp på en etasje
2. Forsikre at riktig heis tar ordren
3. Heis beveger seg og åpner dør
4. Lys slettes

### Test 2: Persistent CAB
1. Trykk CAB-knapp
2. Restart heisen før orden betjenes
3. Forsikre at CAB-ordre gjenopprettes og betjenes

### Test 3: Network Disconnect
1. En heis sender hall-ordre
2. Network cable fjernes for denne heisen
3. Forsikre at andre heiser tar over ordren innen 3s

### Test 4: Packet Loss
1. Bruk `tc` eller `netem` for å simulere packet loss (10-50%)
2. System skal fortsatt fungere (blir bare litt langsommere)

### Test 5: Door Obstruction
1. Trykk obstruksjon-knapp mens dør åpen
2. Dør skal holde seg åpen
3. Når obstruksjon fjernes, dør lukkes og heis fortsetter

