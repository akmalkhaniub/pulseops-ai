# SPEC-07: PulseOps AI — Relational Schemas, TimescaleDB & Redis Keys

## 1. Storage Topology
- **PostgreSQL 17 + TimescaleDB:** High-cardinality cluster metrics, incident timelines, and audit logs.
- **Redis 7.4 Streams:** Telemetry ingestion stream and incident notification events.

---

## 2. PostgreSQL DDL Schema

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "timescaledb" CASCADE;

CREATE TYPE incident_severity AS ENUM ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW');
CREATE TYPE incident_status AS ENUM ('TRIGGERED', 'INVESTIGATING', 'AUTO_HEALING', 'RESOLVED');

CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,
    severity incident_severity NOT NULL,
    status incident_status NOT NULL DEFAULT 'TRIGGERED',
    root_cause_analysis TEXT,
    suggested_action TEXT,
    auto_heal_success BOOLEAN,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS gpu_telemetry_events (
    time TIMESTAMPTZ NOT NULL,
    worker_id VARCHAR(64) NOT NULL,
    gpu_model VARCHAR(32) NOT NULL,
    vram_used_mb INT NOT NULL,
    temperature_c NUMERIC(5,2) NOT NULL,
    p99_latency_ms NUMERIC(7,2) NOT NULL
);

SELECT create_hypertable('gpu_telemetry_events', 'time', if_not_exists => TRUE);
```
