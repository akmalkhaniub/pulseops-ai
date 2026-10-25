# SPEC-08: PulseOps AI — Architecture Decision Records (ADRs)

## ADR-001: TimescaleDB Extension on PostgreSQL vs. InfluxDB
* **Status:** Accepted
* **Context:** We need to store high-cardinality GPU metrics alongside transactional incident relational data.
* **Decision:** Use **TimescaleDB hypertables on PostgreSQL 17**.
* **Consequences:** Eliminates running two distinct databases; enables SQL JOINs between telemetry spikes and incident reports.

---

## ADR-002: iOS Critical Alert Entitlement vs. Standard Push Notifications
* **Status:** Accepted
* **Context:** When a 50-node GPU cluster suffers a cascading CUDA OOM during prime-time hours, on-call engineers must be awakened even if their devices are muted.
* **Decision:** Implement **Apple Critical Alert Entitlements** in Swift 6.
* **Consequences:** Bypasses Do Not Disturb and volume switches, strictly adhering to high-stakes Lead SRE expectations.
