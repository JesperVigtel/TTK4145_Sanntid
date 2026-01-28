## System Design

### Overview

Our system is designed as a fully distributed peer-to-peer architecture where all elevators run identical logic and share state continuously. Communication is implemented using UDP broadcast in a logical mesh topology, allowing each node to exchange messages directly without relying on a central coordinator. Fault tolerance is achieved through periodic state broadcasts that include versioned data structures, enabling nodes to discard outdated information and handle duplicate or out-of-order messages gracefully.

### Fault Tolerance Strategy

To detect node failures, each elevator transmits a heartbeat signal; missing heartbeats over a defined interval are interpreted as a node failure, triggering takeover of its active hall requests by the remaining elevators. Cab orders are persisted to disk and restored upon restart, ensuring no loss during crashes. Hall orders are redistributed when a node goes offline, detected via heartbeat timeout.

To mitigate network packet loss, critical state messages are transmitted repeatedly, ensuring eventual consistency despite unreliable communication. All operations are designed to be idempotent, so duplicate or reordered packets do not cause inconsistency. Since all elevators maintain a replicated view of the system state, the system can continue operating correctly when an elevator crashes or disconnects, with other nodes seamlessly assuming responsibility for unserved hall requests.

### Network Topology & Protocol

Messages are broadcast as complete state snapshots via UDP with type tags at regular intervals. A distributed consensus mechanism allows nodes to reach eventual consistency when connectivity is restored. Failure detection relies on heartbeat timeouts: when a node fails to transmit within the timeout window, the remaining elevators redistribute its orders and continue operation. The distributed state representation eliminates the need for election protocols or master nodes.

### Implementation

The system is implemented in Go, leveraging goroutines and channels to handle concurrent networking, state management, and control logic. This approach provides built-in concurrency primitives that eliminate many race conditions and simplify synchronization, critical requirements for reliable real-time systems. The modular design separates network communication, elevator control, and order management into distinct components that communicate via channels.

### Key Design Decisions

- **Peer-to-peer topology:** No single point of failure; scales without election protocols.
- **Idempotent broadcasts:** Loss, duplication, and reordering are handled transparently.
- **Heartbeat-based failure detection:** Simple, reliable, and easy to tune.
- **Persistent cab storage:** Protects against power loss and unscheduled crashes.
- **Graceful degradation:** Failed nodes exit cleanly; surviving nodes take over automatically.
