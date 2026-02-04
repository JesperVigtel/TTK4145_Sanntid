# Elevator System - Distribuert Heiskontroller

Et distribuert heis-system implementert i Go, basert på `golang-standards/project-layout` og offisiell Go-dokumentasjon.

## Prosjektstruktur

```
.
├── cmd/
│   └── elevator/
│       └── main.go              Programmets inngangspunkt
├── internal/
│   ├── fsm/
│   │   └── fsm.go               Tilstandsmaskinen (FSM)
│   ├── network/
│   │   ├── network.go           Nettverksinitialisering
│   │   ├── peers.go             Peer discovery (online/offline)
│   │   └── transmitter.go       Sending/mottak av tilstandsdata
│   ├── assigner/
│   │   └── assigner.go          Kostnadsfunksjon for ordre-fordeling
│   ├── lights/
│   │   └── lights.go            Synkronisering av knappelys
│   ├── hardware/
│   │   └── elevio.go            Driver-grensesnitt mot maskinvare
│   ├── store/
│   │   └── store.go             Persistens av kabinordrer
│   └── types/
│       └── types.go             Felles datatyper og konstanter
└── README.md                    Denne filen
```

## Moduloversikt

| Pakke | Ansvar |
|-------|--------|
| `cmd/elevator` | Orkestrator: Initialiserer kanaler, driver og goroutiner |
| `fsm` | Tilstandsmaskin som kontrollerer heisens adferd |
| `network` | Distribuert statusdeling via UDP broadcast |
| `assigner` | Kostnadsfunksjon for fordeling av hall-ordrer |
| `lights` | Synkronisering og oppdatering av knappelys |
| `hardware` | Abstraksjonslag mot simulator/fysisk maskinvare |
| `store` | Persistens av ordrer ved krasj/strømbrudd |
| `types` | Felles datastrukturer og konstanter |

## Design-prinsipper

- **Separasjon av ansvar:** Hver pakke har ett klart formål
- **Minimal main:** `main.go` fungerer kun som "lim-funksjon"
- **Beskyttet kjerne:** Alt i `/internal` hindrer uønsket ekstern import
- **Testabilitet:** Funksjonal oppdeling gjør enhetstest lett
- **Best practices:** Følger offisiell Go-prosjektstruktur

## Komme i gang

1. Implementer datastrukturer i `internal/types/types.go`
2. Implementer tilstandsmaskinen i `internal/fsm/fsm.go`
3. Implementer nettverksmodulen i `internal/network/`
4. Implementer assigner, lights, hardware og store
5. Bind alt sammen i `cmd/elevator/main.go`

## Notat om `/internal`

Alle koder er plassert i `/internal` for å sikre at prosjektet fungerer som en standalone applikasjon, ikke som et bibliotek. Dette hindrer at ekstern kode uforvarende importerer interne implementasjonsdetaljer.
