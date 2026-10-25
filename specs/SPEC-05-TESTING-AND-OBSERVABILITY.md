# SPEC-05: PulseOps AI — Testing, Observability & Benchmark Strategy

## 1. Testing Pyramid & Quality Gates

To satisfy the Lead Engineer requirement:
> *"Own application quality standards across architecture, code structure, testing, responsiveness, and developer experience... Set and enforce quality standards across both AI-generated and manually written code"*

### 1.1 Quality Matrix

| Scope | Tooling | Coverage Target | Purpose |
| :--- | :--- | :--- | :--- |
| **Go Pipeline & Concurrency** | `go test -race -v` | 85%+ | Verifies race-free execution in goroutine pools and worker queues. |
| **Client Unit & Hooks** | **Vitest 3.x** + React Testing Library | 80%+ | Validates Apollo cache mutation, optimistic responses, and reducer logic. |
| **Web End-to-End** | **Playwright 1.50+** | Critical paths | Tests full triage flow from GraphQL subscription to triage confirmation. |
| **Mobile E2E** | **Maestro CLI** | Flow tests | Tests native incident push reception, lock screen action, and offline sync. |
| **Type Integrity** | `tsc --noEmit` | 100% strict | Zero `any` policy across monorepo packages. |

---

## 2. Load Testing & High-Throughput Benchmarks (ImagineArt Scale)

### 2.1 Benchmark Requirements
- Ingestion Gateway must process **10,000 telemetry events/sec** with p99 latency < 15ms.
- Worker queue backpressure must shed excess load without dropping active GPU tasks.

### 2.2 k6 Benchmark Script (`tests/load/k6-ingest.js`)
```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 1000 },
    { duration: '1m', target: 5000 },
    { duration: '30s', target: 10000 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(99)<25'], // 99% of requests under 25ms
    http_req_failed: ['rate<0.001'], // Less than 0.1% failure
  },
};

export default function () {
  const payload = JSON.stringify({
    workerId: `worker-${__VU % 500}`,
    vramUsedMb: 22400,
    gpuTempC: 74.2,
    activeTasks: 4,
  });

  const res = http.post('http://localhost:8080/api/v1/telemetry', payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(res, { 'is status 200': (r) => r.status === 200 });
  sleep(0.01);
}
```

---

## 3. Observability & Telemetry (OpenTelemetry + Prometheus)
- **Metrics Exported:**
  - `pulseops_gpu_vram_saturation_ratio`
  - `pulseops_backpressure_active_mode` (0: Normal, 1: Shedding Free, 2: VIP Only)
  - `pulseops_dlq_tasks_total`
  - `pulseops_autoheal_resolution_duration_seconds`
- **Tracing:** W3C TraceContext propagated from Next.js web client through GraphQL gateway to Go worker pool.
