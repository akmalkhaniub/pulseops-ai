# ⚡ PulseOps AI: Autonomous GPU Inference Fleet & Incident Commander

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Next.js](https://img.shields.io/badge/Next.js-16.2.6-black?style=flat&logo=next.js)](https://nextjs.org)
[![React](https://img.shields.io/badge/React-19.2-61DAFB?style=flat&logo=react)](https://react.dev)
[![Expo](https://img.shields.io/badge/Expo-SDK_56-000020?style=flat&logo=expo)](https://expo.dev)
[![GraphQL](https://img.shields.io/badge/GraphQL-Subscriptions-E10098?style=flat&logo=graphql)](https://graphql.org)
[![Redis](https://img.shields.io/badge/Redis-7.4_Streams-DC382D?style=flat&logo=redis)](https://redis.io)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17_TimescaleDB-336791?style=flat&logo=postgresql)](https://timescale.com)

> **Autonomous GPU Inference Fleet & Incident Commander for high-scale GenAI platforms ($35M+ ARR scale).**  
> Solves the catastrophic cost of GPU cluster failures (CUDA OOM, driver timeouts, queue saturation, and inference deadlocks) across thousands of diffusion and LLM worker nodes. Built with a high-throughput **Go 1.25+** asynchronous telemetry pipeline, zero-allocation memory reuse (`sync.Pool`), dynamic admission control, and cross-platform incident command across **Next.js 16.2.6** and **Expo SDK 56** via **GraphQL Subscriptions**.

---

## 🏛️ System Architecture

```mermaid
flowchart TD
    subgraph GPUCluster ["GPU Worker Fleet (Thousands of Nodes)"]
        Node1["Worker Node 1 (FLUX.1 Diffusion)"]
        Node2["Worker Node 2 (Video SVD Render)"]
        NodeN["Worker Node N (LLM Speculative Decoding)"]
        TelemetryAgent["eBPF / Prometheus Telemetry Sidecar"]
    end

    subgraph IngestPipeline ["PulseOps Go 1.25 Pipeline Core"]
        Gateway["Telemetry Ingestion Gateway (HTTP/3 + gRPC)"]
        RingBuffer["Zero-Allocation Ring Buffer (sync.Pool)"]
        TriageEngine["Autonomous Incident Triage Engine"]
        AdmissionCtrl["Adaptive Backpressure Controller"]
    end

    subgraph StorageQueue ["Storage & Stream Topology"]
        Streams["Redis 7.4 Streams (Telemetry & Events)"]
        Timescale["PostgreSQL 17 + TimescaleDB (High-Cardinality Metrics)"]
        DLQ["Dead Letter Queue & Auto-Remediation Workers"]
    end

    subgraph Remediation ["Autonomous Self-Healing Actions"]
        Drain["Worker Cordon & Drain"]
        Restart["Driver Reset / MicroVM Cold Reboot"]
        BatchShift["Traffic Shift to Healthy Pool"]
    end

    subgraph CommandClients ["Command & Control Tiers"]
        WebCommander["Next.js 16.2.6 Web Command Center (Three.js 3D Cluster Topology)"]
        MobileCommander["Expo SDK 56 Mobile On-Call Commander (Swift Critical Alerts)"]
    end

    Node1 & Node2 & NodeN --> TelemetryAgent
    TelemetryAgent -->|High-Frequency Metrics| Gateway
    Gateway --> RingBuffer
    RingBuffer --> TriageEngine
    TriageEngine --> AdmissionCtrl
    AdmissionCtrl --> Streams & Timescale
    TriageEngine -->|CUDA OOM / Deadlock| DLQ
    DLQ --> Drain & Restart & BatchShift
    Streams -->|GraphQL Subscriptions (graphql-ws)| WebCommander & MobileCommander
```

---

## 📐 Complete Specification Index

Every subsystem of PulseOps AI is formalised in detail in [`specs/`](file:///g:/ReplitProjects/pulseops-ai/specs):

| Specification Document | Scope & Technical Focus |
| :--- | :--- |
| [**SPEC-01: Architecture & Topology**](file:///g:/ReplitProjects/pulseops-ai/specs/SPEC-01-ARCHITECTURE.md) | Turborepo 2.x blueprint, global technology matrix, and multi-tier directory structure. |
| [**SPEC-02: Pipeline & Orchestration**](file:///g:/ReplitProjects/pulseops-ai/specs/SPEC-02-PIPELINE-AND-ORCHESTRATION.md) | Go 1.25 telemetry pipeline, dynamic admission control, and GPU worker pool orchestration. |
| [**SPEC-03: GraphQL Schema & Subscriptions**](file:///g:/ReplitProjects/pulseops-ai/specs/SPEC-03-GRAPHQL-SCHEMA.graphql) | Complete GraphQL SDL for cluster health, incident alerts, auto-remediation actions, and live streams. |
| [**SPEC-04: Client & Native Module Spec**](file:///g:/ReplitProjects/pulseops-ai/specs/SPEC-04-CLIENT-AND-NATIVE-SPEC.md) | Next.js 16.2 Web Commander and Expo SDK 56 Mobile on-call app with Swift/Kotlin critical alerts. |
| [**SPEC-05: Testing & Observability**](file:///g:/ReplitProjects/pulseops-ai/specs/SPEC-05-TESTING-AND-OBSERVABILITY.md) | Vitest, Go benchmarks, chaos simulation (injected OOMs), and OpenTelemetry spans. |
| [**SPEC-06: Implementation Plan**](file:///g:/ReplitProjects/pulseops-ai/specs/SPEC-06-IMPLEMENTATION-PLAN.md) | 5-phase engineering roadmap from telemetry daemon to multi-platform incident command. |
| [**SPEC-07: Data Models & Migrations**](file:///g:/ReplitProjects/pulseops-ai/specs/SPEC-07-DATA-MODELS-AND-MIGRATIONS.md) | TimescaleDB hypertable schema, incident lifecycle state machines, and audit log tables. |
| [**SPEC-08: Architectural Decision Records**](file:///g:/ReplitProjects/pulseops-ai/specs/SPEC-08-ADR-DECISIONS.md) | ADR-001 through ADR-004 justifying TimescaleDB vs ClickHouse, Go vs Rust for pipeline, and GraphQL WS. |

---

## 🛠️ Technology Matrix

| Tier | Component | Specification | Rationale |
| :--- | :--- | :--- | :--- |
| **Monorepo** | Workspace Manager | Turborepo 2.x + pnpm 9.x | Symlinked packages, zero-config remote caching. |
| **Web Client** | Web Commander | Next.js 16.2.6 (React 19.2) | Server Actions v2, partial prerendering, Tailwind CSS v4. |
| **Mobile Client** | On-Call App | Expo SDK 56 (React Native 0.85) | Bridgeless New Architecture, 60fps Reanimated 4 gestures. |
| **Native Module** | Critical Alerts | Swift 6 (iOS 18) & Kotlin 2.1 | iOS Critical Alerts (bypasses Do Not Disturb) & Android Foreground Services. |
| **API & Data** | GraphQL Gateway | Go 1.25+ (`gqlgen`) + `graphql-ws` | High-concurrency compiled binary, low-overhead WebSocket fan-out. |
| **Pipeline** | Core Ingestion Daemon | Go 1.25+ goroutine pools | Sub-millisecond latency, zero-allocation memory reuse (`sync.Pool`). |
| **Queue & Stream** | Message Broker | Redis 7.4 Streams (Valkey) | Consumer groups, `XAUTOCLAIM`, sub-millisecond dispatch. |
| **Database** | Persistent Time-Series | PostgreSQL 17 + TimescaleDB | High-cardinality time-series metrics & audit logging. |

---

## 🚀 Monorepo Directory Topology

```text
pulseops-ai/
├── apps/
│   ├── web/                       # Next.js 16.2.6 Web Incident Command Center
│   └── mobile/                    # Expo SDK 56 Mobile On-Call Application
├── packages/
│   ├── api-graphql/               # GraphQL schema & Go resolver contracts
│   ├── graphql-codegen/           # Autogenerated typed React hooks
│   └── ui/                        # Cross-platform command primitives
├── services/
│   └── pipeline-go/               # Go 1.25 Distributed Telemetry & Incident Ingester
└── specs/                         # Complete Spec-Driven Development (SDD) documentation
```

---

## 📜 License & Governance

Licensed under the [MIT License](LICENSE). Architected for enterprise-scale AI infrastructure reliability and autonomous incident mitigation.
