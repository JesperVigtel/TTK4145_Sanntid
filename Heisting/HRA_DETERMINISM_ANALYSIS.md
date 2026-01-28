# HRA Determinisme & Single Elevator Mode - Analyse

## SPØRSMÅL 1: Er HRA Deterministisk?

### Svar: **JA, MEN med viktige forbehold**

#### Hva sier koden:

**Project-resources README sier:**
> "Deterministic behavior requires identical input. The timeToIdle function simulates elevator execution to calculate cost."

**Implementasjonen (Solution 3 cost.go):**
```go
input := HRAInput{
    HallRequests: myNodeData.AllHallOrders,
    States:       elevStates,  // floor, direction, behavior, cabRequests
}

ret, err := exec.Command(hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
```

**Konklusjon:** HRA er en deterministisk algoritme - samme JSON input → samme JSON output.

#### KRITISK BETINGELSE: Alle elev må ha identisk tilstand

**Problem som ikke tas opp i PDR V3:**

Hvis to elevatorer får hall call samtidig fra ulike workstasjoner:

| Scenario | Elev A | Elev B | Resultat |
|----------|--------|--------|----------|
| **Synkron (ideelt)** | Mottar broadcast T=100ms | Mottar broadcast T=100ms | ✅ Samme HRA input → Samme output |
| **Asynkron (realistisk)** | Mottar broadcast T=100ms | Mottar broadcast T=105ms | ❌ Ulike stater → Ulik HRA output |
| **Med packet loss** | Mottar broadcast T=100ms | Packet lost, mottar T=200ms | ❌ **Utbrudd: Begge elev tror de skal ta ordren** |

#### Eksempel på Problem:

```
Tid 0ms:    Floor 2 UP button pressed
Tid 10ms:   Elev A mottok: aliveList=[A,B], cabOrders=[], hallOrders=[UP2]
Tid 15ms:   Elev A kjører HRA: "Assign to A"
Tid 20ms:   Elev B mottok: aliveList=[A], cabOrders=[], hallOrders=[UP2]
            (B tror A er offline pga packet loss)
            Elev B kjører HRA: "Assign to B"

Resultat:   ❌ Begge tror de skal ta ordren!
            ❌ Light turns ON på begge workstasjoner
            ❌ Kanskje både A og B drar til floor 2
```

### HRA Determinisme - RIKTIG STATEMENT for PDR

**Feil (nåværende V3):**
> "Identical global state produces identical assignments on all elevators"

**Riktig statement bør være:**
> "HRA (hall_request_assigner) is deterministic: given identical global state S, all elevators 
> running HRA(S) produce identical assignments. However, this requires all elevators to have 
> synchronized state BEFORE running HRA. The P2P network ensures eventual consistency, but 
> transient disagreement may occur during network jitter or packet loss. Button light 
> disagreement is resolved by: (1) periodic re-broadcast (100ms), (2) ACK protocol (orders 
> not removed until consensus), (3) timeout-based reassignment (15s). Result: Inconsistencies 
> are temporary (<200ms) and self-healing."

---

## SPØRSMÅL 2: Single Elevator Mode - Bare Cab Calls?

### Original Spec (Fra dokumentet du oppga):

```
"Cab call redundancy with a single elevator or a disconnected elevator is not required"

Meaning: In permitted assumption #2, disconnected elevator doesn't need to sync cab calls 
with network. So a single elevator can safely ignore cab order sync issues.
```

```
"If the elevator is disconnected from the network, it should still serve all the 
currently active calls (i.e. whatever lights are showing)"
```

### Problem Med Current PDR V3 Design

**V3 sier:** "Single-elevator mode: serve cab calls only"

**Spørsmål fra deg:** Burde den ikke servere både hall og cab calls hvis det bare er en elevator?

**Svar: JA! Du har rett! Dette er et arkitekturvalg.**

#### Analyse av to alternativer:

**Alternativ A: Kun Cab Calls (Nåværende PDR V3)**
```
✓ Enkelt: Ingen hall call assignment logikk nødvendig
✓ Sikkert: Hvis network kommer tilbake, ingen stale hall calls
✗ Dårlig UX: Brukere kan ikke trykke på hall buttons
✗ Feil tolkninng av spec: Spec sier "should still serve all currently active calls"
```

**Alternativ B: Både Cab og Hall Calls (Smartere)**
```
✓ Bedre UX: Hall buttons fungerer i single-elev mode
✓ Støtter spec: "serve all currently active calls"
✓ Nærmere "single elevator behavior": System fungerer som en standalone heis
✗ Litt mer komplekst: Må håndtere når network kommer tilbake
```

### Den STORE Utfordringen Du Identifiserte

**Spørsmål:** Hvordan skiller vi mellom:
1. **Nettverksfeil (midlertidig)** → Ska single-elev mode servere hall calls
2. **Kvasipermanent disconnect** → Skal single-elev mode kun servere cab calls
3. **Oppstart uten nettverkstilkobling** → Hva skal gjøres?

**Spec sier dette er UNSPECIFIED - du velger selv.**

#### Tre Mulige Design Choices:

**Choice 1: Always Serve Hall Calls (Foreslått av deg)**
```go
// Disconnected elevator = single elevator behavior
// Serve both cab AND hall calls
// When network returns: Merge state, re-sync hall calls via ACK protocol

Behavior:
  Offline: Treat as single elevator - accept all buttons
  Online: Re-enter P2P mode, sync state
  
Pro: Maksimal resilience, bedre UX
Con: Kompleksere state merging når network kommer tilbake
```

**Choice 2: Refuse Hall Calls When Disconnected (Nåværende V3)**
```go
// Single-elevator mode = cab calls only
// When network unavailable: Only serve cab calls
// Hall calls rejected/ignored

Pro: Enklere, mindre kompleks
Con: Dårligere user experience
```

**Choice 3: Hybrid - Grace Period (Smartest)**
```go
// If offline <5 seconds: Serve both (assume network jitter)
// If offline >5 seconds: Serve cab only (assume real disconnect)
// When network returns: Re-sync

Pro: Best of both worlds
Con: Arbitrary timing thresholds
```

---

## RIKTIG TILNÆRMING FOR PDR

### Issue: Network Partition Detection

**Du identifiserer det rette problemet:**
> "It's challenging to distinguish between network failure vs. actual single-elevator mode"

**Spec acknowledges this:**
> "Unspecified behavior: How the elevator behaves when it cannot connect to the network 
> during initialization - You can either enter a 'single-elevator' mode or refuse to start"

### Rekomandert Design For Din PDR

**Endre V3 til å være eksplisitt om antagelsene:**

```
Single-Elevator Mode (Network Disconnected):

Choice: Elevator continues to serve BOTH cab AND hall calls

Rationale:
1. From spec: "should still serve all currently active calls"
2. If only one elevator available, treating it as standalone maximizes resilience
3. Hall calls during disconnection won't be lost (persisted locally via call queue)
4. When network restores: ACK protocol synchronizes state

Implementation:
- Offline elevator: Accept both cab and hall button presses
- Hall calls queue locally until network available
- When network restored: Broadcast queued hall calls for HRA assignment
- Other online elevators: Run HRA, may take some hall calls from newly-online elevator

State Synchronization on Reconnection:
1. Re-online elevator sends: local queue of hall calls + cab orders
2. Other elevators run HRA with new topology
3. ACK protocol ensures no duplicate service
4. Re-online elevator accepts assignment from HRA

Assumption Made:
- We assume that "disconnected" is temporary and usually resolves
- If permanently disconnected, elevator still provides emergency service (cab calls work)
- Spec permits this choice: "Unspecified behavior - you can either..."
```

---

## SUMMARY OF ISSUES IN PDR V3

### Issue 1: HRA Determinism Statement
**Current:** "Identical global state produces identical assignments"  
**Problem:** Doesn't acknowledge that achieving identical state is non-trivial  
**Fix:** Add paragraph explaining eventual consistency + transient disagreement

### Issue 2: Single-Elevator Mode Too Restrictive
**Current:** "serve cab calls only"  
**Problem:** Violates spec requirement "should still serve all currently active calls"  
**Better:** "serve both cab AND hall calls - treat as standalone elevator"  
**Why:** Du har rett - det gir bedre UX og er nærmere single-elevator semantikk

### Issue 3: Not Clear How to Distinguish Network Failure
**Current:** Not addressed  
**Better:** Add explicit statement that this is design choice (unspecified behavior)

---

## RECOMMENDATIONS FOR UPDATED PDR

### Fix 1: HRA Determinism
Replace in V3 Section 1:
```
OLD: "The hall_request_assigner (HRA) is deterministic—identical global state produces 
identical assignments on all elevators. To guarantee synchronized button lights: 
(1) All elevators wait for network consensus..."

NEW: "HRA (hall_request_assigner) is deterministic: given identical global state, it 
produces identical assignments. Network ensures eventual consistency through: (1) 100ms 
periodic broadcasts (redundancy), (2) ACK protocol (orders not finalized until consensus), 
(3) Timeout-based reassignment (15s). Transient disagreement (<100ms) may occur during 
network jitter; automatically resolved by next broadcast cycle."
```

### Fix 2: Single-Elevator Mode
Replace in V3 Section 5:
```
OLD: "If network unavailable at startup → enter single-elevator mode (serve cab calls only)"

NEW: "If network unavailable at startup OR disconnected: enter single-elevator mode. 
Elevator treats itself as standalone: serves BOTH cab and hall calls, maintaining local 
queue of hall calls. When network restores: broadcasts queued calls; other elevators run 
HRA to rebalance. ACK protocol ensures no duplicate service."
```

### Fix 3: Add Design Choice Statement
Add new bullet in Section 5:
```
"Unspecified behavior choice - Network Unavailability: Rather than refuse to start, 
elevator enters single-elevator mode to maximize resilience. This enables continued 
service (cab calls always work, hall calls queue until network available). When network 
restores, queued hall calls are redistributed via HRA."
```

---

## KONKLUSJON

**Du hadde helt rett på begge spørsmål:**

1. ✅ **HRA determinisme:** JA, men betinget på at alle elev har synkronisert tilstand først
   - PDR V3 er for optimistisk uten å adressere hvordan synkronisering oppnås
   
2. ✅ **Single elevator mode:** BÅDE cab og hall calls gir mer mening
   - PDR V3 er for restriktiv
   - Spec sier "serve all currently active calls" - det inkluderer hall calls
   - Din tolkning er nærmere spec-intensjonen

**Disse to endringene vil gjøre PDR betydelig sterkere.**
