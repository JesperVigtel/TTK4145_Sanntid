# 🔄 Update 26 Jan 2026 — Solution 3 Enhancements Integrated

Adopted improvements from ANALYSE_SOLUTION_3:
- Broadcast interval: 5ms (configurable) for faster convergence without excess CPU.
- Type-tagged JSON envelope (`TypeId`, `PayloadJSON`) for safe message routing and versioning.
- Fault detection: motor stop timeout (~4s) and door obstruction timeout (5s) with graceful offline/online toggling.
- Request handling clarity: stop/clear semantics aligned with Solution 3’s clean rules.

Explicitly not adopted:
- ACK-table consensus (kept simpler idempotent consensus + HRA).
- Automatic reboot via terminal (we persist CAB and recover without process respawn).

Requirements check (post-integration):
- No calls lost: CAB persistence retained; crash-restart verified.
- Button light guarantee: idempotent updates + faster broadcast; lights converge quickly.
- Fault tolerance: heartbeat + motor/obstruction detection; graceful network handling.
- Packet loss: periodic, idempotent, type-tagged messages ensure robustness.
- Scalability: HRA assignment maintained; parameters configurable.

# 📦 KOMPLETT DESIGNPAKKE - INNHOLDSFORTEGNELSE

## ✨ DU HAR FÅTT 8 DOKUMENTER

Denne løsningen er komplett og inneholder alt du trenger for å lykkes med heisprosjektet!

---

## 📋 DOKUMENTOVERSIKT

### 1️⃣ **PDD_PRELIMINARY_DESIGN.md** ⭐ START HER
**Formål:** Innlevering til pensum  
**Lengde:** ~4KB, formatert for < 1 side  
**Innhold:**
- Gruppeinformasjon (navn, mail, lab time, desk, group nr)
- Fault tolerance strategi
- Network topology & protokoll
- Why Go + designparadigme
- Module arkitektur
- Tabell med løsningsmatrise

**✅ HVA DU GJØR:**
1. Åpne dokumentet
2. Fyll inn dine navn/emails/group number
3. Juster lab time og desk number
4. Eksporter som PDF: `PDD-##.pdf` (## = group number)
5. Last opp på Blackboard

---

### 2️⃣ **UML_DIAGRAMS.md**
**Formål:** Visuell design dokumentasjon  
**Lengde:** ~20KB  
**Innhold:**
- ✅ Elevator FSM diagram (5 states)
- ✅ Button Light State Machine (4 states)
- ✅ Component/Class diagram
- ✅ Sequence diagram - Normal hall call
- ✅ Sequence diagram - Network disconnect
- ✅ Sequence diagram - Crash recovery

**✅ HVA DU GJØR:**
- Les før implementasjonen
- Tegn/print som referanse ved kodingen
- Bruk som vegg-poster ved problemløsning

---

### 3️⃣ **IMPLEMENTASJONSGUIDE.md** ⭐ KODEGUIDE
**Formål:** Pseudokode og kodestruktur  
**Lengde:** ~22KB  
**Innhold:**
- ✅ Modul 1: Elevator FSM (pseudokode + select-loop)
- ✅ Modul 2: Network Module (Sender + Receiver)
- ✅ Modul 3: Order Manager (CAB persist + hall assign)
- ✅ Main/Coordinator setup
- ✅ Data structures (Enum types + Structs)
- ✅ Config constants
- ✅ Testing strategy

**✅ HVA DU GJØR:**
1. Les pseudokoden for en modul om gangen
2. Implementer etter pseudokoden (ikke direkte kopi!)
3. Test hver modul isolert
4. Bruk data structures direkte

---

### 4️⃣ **SYSTEMARKITEKTUR.md**
**Formål:** Detaljert systemdesign  
**Lengde:** ~18KB  
**Innhold:**
- ✅ Full system architecture diagram
- ✅ Channel communication map (alle kanaler)
- ✅ Component interaction sequence
- ✅ Timing analysis for all critical paths
- ✅ Fault tolerance matrix
- ✅ Design principles (5 stk)
- ✅ Comparison med reference solutions
- ✅ Implementation priorities (Phase 1-4)
- ✅ Testing checklist

**✅ HVA DU GJØR:**
- Les når du sliter med arkitekturforståelse
- Sjekk timing requirements mot implementasjonen
- Bruk fault matrix for å verifisere design

---

### 5️⃣ **SCENARIO_WALKTHROUGHS.md** ⭐ LÆRINGSRESSURS
**Formål:** Konkrete eksempler med timeline  
**Lengde:** ~21KB  
**Innhold:**
- ✅ Scenario 1: Normal CAB order (best case)
  - Detailed timeline fra t=0ms til t=3310ms
  - Alle kanaler og transitions
  
- ✅ Scenario 2: Hall call med konfliktløsning
  - Consensus delay
  - Conflict resolution (distance + ID)
  - Light synchronization
  
- ✅ Scenario 3: Network disconnect & takeover
  - Timeout detection (3s)
  - Automatic reassignment
  - Re-synchronization
  
- ✅ Scenario 4: Software crash + recovery
  - Persistent storage
  - Automatic restore på startup
  - Order completion after restart

**✅ HVA DU GJØR:**
- Les en scenario per uke
- Trace gjennom timeline
- Lær timing expectations
- Forstå fault handling

---

### 6️⃣ **ANALYSE_HEISPROSJEKT.md**
**Formål:** Designbakgrunn og sammenligning  
**Lengde:** ~12KB  
**Innhold:**
- ✅ Komparativ analyse av L1 (HRA) vs L2 (Distance)
- ✅ Evaluering mot alle prosjektkrav
- ✅ Styrker og svakheter
- ✅ Hvorfor vi valgte hybrid-løsningen
- ✅ Sammenligning tabell

**✅ HVA DU GJØR:**
- Les for å forstå designvalg
- Refererer til når noen spør "Why not HRA?"
- Bruk som backup-argument i gruppdiskusjoner

---

### 7️⃣ **SAMMENFATNING_OG_SJEKKLISTE.md** ⭐ ROADMAP
**Formål:** Oversikt + progress tracking  
**Lengde:** ~12KB  
**Innhold:**
- ✅ Oversikt over alle 8 dokumenter
- ✅ 30-second design summary
- ✅ Key constants
- ✅ State machines at a glance
- ✅ Timing expectations
- ✅ Message flow summary
- ✅ Critical invariants
- ✅ Testing checklist (Phase 1-4)
- ✅ Debugging tips
- ✅ Reference implementation hints
- ✅ Pre-submission checklist

**✅ HVA DU GJØR:**
- Bruk som daily reference
- Marker av testing checkboxes
- Sjekk invarianter under kodingen

---

### 8️⃣ **QUICK_REFERENCE.md** ⭐ HJELPESKJEMA
**Formål:** Rask lookup reference  
**Lengde:** ~9.9KB  
**Innhold:**
- ✅ Document overview
- ✅ 30-second design summary
- ✅ Key constants (alle 7)
- ✅ State machines (FSM + Light)
- ✅ Timing table (alle kritiske events)
- ✅ Message flow summary
- ✅ Critical invariants (5 stk)
- ✅ Testing checklist
- ✅ Debugging tips (4 kategorier)
- ✅ Design choice rationale
- ✅ Submission checklist
- ✅ Next steps

**✅ HVA DU GJØR:**
- Print og heng opp ved siden av skjermen
- Bruk når du skal feilsøke
- Sjekk før innlevering

---

## 🗂️ FILSTRUKTUR I MAPPEN

```
/Users/jespervh/Desktop/mappe uten navn/

├─ 1. INNLEVERING (copy + submit)
│  └─ PDD_PRELIMINARY_DESIGN.md  → Konverter til PDF som PDD-##.pdf
│
├─ 2. FORSTÅ DESIGNET
│  ├─ UML_DIAGRAMS.md
│  ├─ SYSTEMARKITEKTUR.md
│  └─ SCENARIO_WALKTHROUGHS.md
│
├─ 3. IMPLEMENTASJON
│  ├─ IMPLEMENTASJONSGUIDE.md          [Pseudokode]
│  ├─ QUICK_REFERENCE.md              [Lookup tabell]
│  └─ SAMMENFATNING_OG_SJEKKLISTE.md   [Checklist]
│
└─ 4. BAKGRUNN
   └─ ANALYSE_HEISPROSJEKT.md          [Design choices]

Plus referanse-løsningene:
├─ Elevator-main/TTK4145/Project/      [L1: HRA + Cyclic]
└─ TTK4145-Heislab-master/Source/      [L2: Distance + Persistence]
```

---

## ⏱️ ANBEFALT LESEREKKEFØLGE

### Dag 1 (30 min): Oversikt
1. Read: QUICK_REFERENCE.md (10 min)
2. Read: PDD_PRELIMINARY_DESIGN.md (10 min)
3. Check: SAMMENFATNING_OG_SJEKKLISTE.md (10 min)

### Dag 2 (1 time): Visualisering
1. Study: UML_DIAGRAMS.md - FSM diagram (20 min)
2. Study: UML_DIAGRAMS.md - Component diagram (20 min)
3. Study: SYSTEMARKITEKTUR.md - Architecture overview (20 min)

### Dag 3 (1.5 timer): Detaljer
1. Read: SCENARIO_WALKTHROUGHS.md - Scenario 1 (20 min)
2. Read: SCENARIO_WALKTHROUGHS.md - Scenario 2 (20 min)
3. Read: IMPLEMENTASJONSGUIDE.md - Module 1 (30 min)

### Dag 4 (1 time): Implementasjon Start
1. Read: IMPLEMENTASJONSGUIDE.md - Module 2 (20 min)
2. Read: IMPLEMENTASJONSGUIDE.md - Module 3 (20 min)
3. Setup: Kodeprosjekt, lese pseudokoden (20 min)

### Dag 5-7: Koding + Testing
1. Code Module 1 (Elevator FSM)
2. Test Module 1 isolert
3. Add Module 2 (Network)
4. Test Module 2
5. Add Module 3 (Order Manager)
6. Integration testing

---

## 🎯 HVA HVERT DOKUMENT SVARER PÅ

| Spørsmål | Dokument |
|----------|----------|
| Hva er designet mitt? | PDD_PRELIMINARY_DESIGN.md |
| Hvordan tegnes FSM-en? | UML_DIAGRAMS.md |
| Hvordan implementerer jeg? | IMPLEMENTASJONSGUIDE.md |
| Hvordan fungerer systemet? | SYSTEMARKITEKTUR.md + SCENARIO_WALKTHROUGHS.md |
| Hvorfor disse designvalg? | ANALYSE_HEISPROSJEKT.md |
| Har jeg glemt noe? | SAMMENFATNING_OG_SJEKKLISTE.md |
| Hva er timingene? | SYSTEMARKITEKTUR.md + QUICK_REFERENCE.md |
| Hvordan debugger jeg? | QUICK_REFERENCE.md |

---

## ✅ SJEKKLISTE: HAR DU ALT?

```
Før du starter:
[ ] 8 markdown dokumenter opprettet
[ ] Filene er lesbare
[ ] Du har lest PDD_PRELIMINARY_DESIGN.md
[ ] Du har lest QUICK_REFERENCE.md

For innlevering:
[ ] Fyll inn PDD med dine detaljer
[ ] Konverter til PDF: PDD-##.pdf
[ ] Last opp på Blackboard
[ ] Sjekk fristen

For implementasjon:
[ ] Har alle dokumentene
[ ] Har tilgang til pseudokoden
[ ] Har design diagrammene
[ ] Har testing strategy

For gruppediskusjon:
[ ] Share alle dokumenter med gruppen
[ ] Diskuter design valg (basert på ANALYSE)
[ ] Fordel arbeid (modul 1, 2, 3)
[ ] Sett opp Git med pseudokode
```

---

## 🚀 READY TO SUBMIT?

### Før innlevering:
1. ✅ Åpne `PDD_PRELIMINARY_DESIGN.md`
2. ✅ Fyll inn navn + emails
3. ✅ Fyll inn lab time + desk + group number
4. ✅ Lese gjennom for stavefeil
5. ✅ Konverter til PDF
6. ✅ Navn fila `PDD-##.pdf`
7. ✅ Last opp på Blackboard før deadline

### Etter innlevering:
1. ✅ Diskuter design i gruppen
2. ✅ Fordel moduler
3. ✅ Start implementasjon
4. ✅ Refer til pseudokoden daglig
5. ✅ Test etter hver modul

---

## 💡 PRO TIPS

**For best resultat:**
- 📌 Print UML diagrams og heng opp ved skjermen
- 📌 Bookmark QUICK_REFERENCE.md - du vil bruke den hele tiden
- 📌 Følg IMPLEMENTASJONSGUIDE.md pseudokoden nøyaktig
- 📌 Test hver modul FØR du starter på neste
- 📌 Bruk SCENARIO_WALKTHROUGHS til å lære timing
- 📌 Share dokumenter med gruppen - diskuter design!
- 📌 Lag Git repo nå, commit design docs

**Feiltyper du vil unngå:**
- ❌ Implementer alle 3 moduler på en gang
- ❌ Hopp over FSM state transitions
- ❌ Glem CAB persistence (den er CRITICAL!)
- ❌ Test ikke enkeltmoduler før integrasjon
- ❌ Ikke følg pseudokoden - lag din egen

---

## 🎉 SUMMARY

Du har nå:
```
✅ 8 komplette design-dokumenter
✅ Pseudokode for alle moduler
✅ UML diagrammer for visualisering
✅ Scenario walkthroughs for learning
✅ Timing analysis
✅ Testing strategy
✅ Fault tolerance plan
✅ Alt du trenger for å lykkes!
```

**The design is complete. Time to code!** 🚀

---

## 📞 DOKUMENTREFERANSER

Hvis du er usikker på noe, sjekk disse først:

**"Hva er fault tolerance?"**
→ Se: PDD_PRELIMINARY_DESIGN.md + ANALYSE_HEISPROSJEKT.md

**"Hvordan implementerer jeg FSM?"**
→ Se: UML_DIAGRAMS.md + IMPLEMENTASJONSGUIDE.md

**"Hva skjer hvis nettverket går ned?"**
→ Se: SCENARIO_WALKTHROUGHS.md (Scenario 3)

**"Hva er timing expectations?"**
→ Se: SYSTEMARKITEKTUR.md + QUICK_REFERENCE.md

**"Hvordan debugger jeg?"**
→ Se: QUICK_REFERENCE.md (Debugging Tips)

**"Hva skal jeg gjøre neste?"**
→ Se: SAMMENFATNING_OG_SJEKKLISTE.md

---

## 🎓 LYKKE TIL! 

Du er nå fullstendig forberedt.

**Next:** Konverter PDD til PDF, levér, og start kodingen! 🚀

