# Elevator System - Distribuert Heiskontroller

Et distribuert heis-system implementert i Go, basert på `golang-standards/project-layout` og offisiell Go-dokumentasjon.

## Prosjektstruktur

/home/runner/work/TTK4145_Sanntid/TTK4145_Sanntid/
├── elevator/                              # Main elevator control system
│   ├── cmd/elevator/
│   │   └── main.go                       # Entry point
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go                 # Configuration constants
│   │   ├── consensus/
│   │   │   ├── consensus.go              # Distributed consensus engine
│   │   │   ├── order_state.go            # Order state machine (Standby→Pending→Assigned→Complete)
│   │   │   └── peer_tracking.go          # Peer availability and hall order merging
│   │   ├── dispatch/
│   │   │   ├── dispatch.go               # Main dispatcher logic
│   │   │   ├── assignment.go             # Hall request assignment via HRA
│   │   │   ├── local_state.go            # Local system state management
│   │   │   ├── elevator_format.go        # Elevator state format conversion
│   │   │   └── hall_request_assigner/    # External binary for order assignment
│   │   ├── lights/
│   │   │   └── lights.go                 # Light control (cab + hall)
│   │   ├── localControl/
│   │   │   ├── localControl.go           # Local elevator FSM
│   │   │   ├── handlers.go               # Order/movement handlers
│   │   │   ├── orders.go                 # Order management
│   │   │   ├── hardware/
│   │   │   │   └── elevio.go             # Hardware interface
│   │   │   └── timer/
│   │   │       └── timer.go              # Watchdog timers
│   │   ├── network/
│   │   │   ├── peers/
│   │   │   │   └── peers.go              # Peer discovery (heartbeat + UDP)
│   │   │   ├── broadcast/
│   │   │   │   └── broadcast.go          # Type-tagged JSON broadcasting
│   │   │   ├── network_manager/
│   │   │   │   └── network_manager.go    # Network orchestration
│   │   │   ├── conn/                     # OS-specific UDP broadcast
│   │   │   │   ├── bcast_conn_linux.go
│   │   │   │   ├── bcast_conn_windows.go
│   │   │   │   └── bcast_conn_darwin.go
│   │   │   └── utility_network/
│   │   │       └── utility_network.go    # Network initialization
│   │   └── types/
│   │       └── types.go                  # Shared type definitions
│   ├── go.mod
│   └── README.md

## Moduloversikt

┌─────────────────────────────────────────────────────────────┐
│                    LOCAL HARDWARE                            │
│  (Floor sensors, buttons, obstruction, motor)               │
└────────────────────┬────────────────────────────────────────┘
                     │ (ElevatorEvents)
                     ▼
        ┌────────────────────────────┐
        │     LOCAL CONTROL FSM      │ (localControl.go)
        │  Current Floor/Behavior    │
        │  Door Control              │
        └────────────┬───────────────┘
                     │ (LocalSystemState + ButtonPress)
                     ▼
        ┌────────────────────────────┐
        │      DISPATCH LAYER        │ (dispatch.go)
        │  Button→Order conversion   │
        │  State tracking            │
        └────────────┬───────────────┘
                     │ (LocalSystemState)
                     ▼
    ┌────────────────────────────────────────┐
    │     DISTRIBUTED CONSENSUS ENGINE       │
    │  Synchronizes hall orders across all   │ (consensus.go)
    │  elevators using cyclic state machine  │
    └────────────┬─────────────────────────┬─┘
                 │                         │
         (ConvergedSystemState)    (broadcast Message)
                 │                         │
                 ▼                         ▼
    ┌────────────────────────┐   ┌──────────────────┐
    │   ASSIGNMENT ENGINE    │   │  NETWORK LAYER   │
    │  (dispatch/assignment) │   │  (peers, bcast)  │
    │  Calls HRA binary      │   │  UDP multicast   │
    └────────────┬───────────┘   └──────────────────┘
                 │                         ▲
         (LocalOrderTable)           (PeerMessages)
                 │                         │
                 ▼                         │
    ┌────────────────────────┐            │
    │      LIGHTS CONTROL    │            │
    │  Hall + Cab indicators │            │
    └────────────────────────┘            │
                                          │
                   All Elevators ─────────┘
                   Exchange state via UDP

## Design-prinsipper

- **Separasjon av ansvar:** Hver pakke har ett klart formål
- **Minimal main:** `main.go` fungerer kun som "lim-funksjon"
- **Beskyttet kjerne:** Alt i `/internal` hindrer uønsket ekstern import
- **Testabilitet:** Funksjonal oppdeling gjør enhetstest lett
- **Best practices:** Følger offisiell Go-prosjektstruktur

## Navnekonvensjoner i Go

I Go er det flere standardnavnekonvensjoner:

1. **PascalCase**: Brukes for offentlige (exported) typer og funksjoner. Eksempel: `MyFunction`, `MyType`.
2. **camelCase**: Brukes for private (unexported) variabler og funksjoner. Eksempel: `myVariable`, `myFunction`.
3. **snake_case**: Brukes sjeldnere, men kan sees i filnavn og pakker. Eksempel: `my_package`.
4. **Forkortelser**: Forkortelser skal skrives med store bokstaver. Eksempel: `HTTPServer`, `URLParser`.

Disse konvensjonene bidrar til lesbarhet og konsistens i Go-kode.


