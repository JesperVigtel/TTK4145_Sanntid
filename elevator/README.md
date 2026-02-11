# Elevator System - Distribuert Heiskontroller

Et distribuert heis-system implementert i Go, basert på `golang-standards/project-layout` og offisiell Go-dokumentasjon.

## Prosjektstruktur

```
├── cmd/
│   └── elevator/
│       └── main.go              Programmets inngangspunkt
├── internal/
│   ├── localControll/
│   │   └── localControll.go      Tilstandsmaskinen (FSM)
│   │   └── hardware/
│   │       └── elevio.go         Driver-grensesnitt mot maskinvare
│   ├── network/
│   │   ├── network.go           Nettverksinitialisering
│   │   ├── peers.go             Peer discovery (online/offline)
│   │   └── transmitter.go       Sending/mottak av tilstandsdata
│   ├── decisionMaker/
│   │   └── decisionMaker.go          Kostnadsfunksjon for ordre-fordeling
│   ├── lights/
│   │   └── lights.go            Synkronisering av knappelys
│   └── types/
│       └── types.go             Felles datatyper og konstanter
└── README.md                    Denne filen
```

## Moduloversikt

| Pakke | Ansvar |
|-------|--------|
| `cmd/elevator` | Orkestrator: Initialiserer kanaler, driver og goroutiner |
| `localControll` | Tilstandsmaskin som kontrollerer heisens adferd |
| `network` | Distribuert statusdeling via UDP broadcast |
| `decisionMaker` | Kostnadsfunksjon for fordeling av hall-ordrer |
| `types` | Felles datastrukturer og konstanter |

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


