# SPEC-02: PulseOps AI — Asynchronous Go Pipeline & Inference Orchestration

## 1. Scope & Objective
This specification formalizes the **high-throughput asynchronous pipeline engine** in **Go 1.25+** engineered to satisfy ImagineArt's backend requirements:
1. High-throughput job queuing with zero-allocation memory reuse (`sync.Pool`).
2. Adaptive backpressure and dynamic rate limiting to protect GPU clusters under peak load.
3. GPU worker pool orchestration with warm-cache affinity.
4. Fault tolerance: Exponential backoff retries, circuit breakers, and Dead Letter Queues (DLQ).

---

## 2. Ingestion & Admission Control (Backpressure Engine)

### 2.1 Dynamic Admission Control Algorithm
Under Fortune 500 / multi-million request traffic spikes, unrestricted queuing leads to runaway queue latency and memory exhaustion. The admission controller enforces dynamic backpressure:

```go
package backpressure

import (
	"sync/atomic"
	"time"
)

type ClusterHealth struct {
	P99QueueLatencyMs int64
	ActiveGPUWorkers  int32
	VRAMSaturationPct int32 // 0 - 100
}

type AdmissionController struct {
	health           atomic.Pointer[ClusterHealth]
	freeTierLimit    int64
	proTierLimit     int64
	vipTierLimit     int64
}

// CanAdmit evaluates whether a request should be processed, queued, or rejected
func (ac *AdmissionController) CanAdmit(tier string) bool {
	h := ac.health.Load()
	if h == nil {
		return true
	}

	// Severity 1: Severe Cluster Stress (p99 > 2500ms or VRAM > 94%)
	if h.P99QueueLatencyMs > 2500 || h.VRAMSaturationPct > 94 {
		// Hard load shedding: Only Enterprise VIP traffic admitted
		return tier == "VIP"
	}

	// Severity 2: Moderate Load (p99 > 1200ms or VRAM > 85%)
	if h.P99QueueLatencyMs > 1200 || h.VRAMSaturationPct > 85 {
		// Shed free-tier traffic to preserve SLA for paying customers
		return tier == "VIP" || tier == "PRO"
	}

	return true
}
```

---

## 3. Distributed Queue & Worker Orchestration (Go + Redis Streams)

### 3.1 Redis Streams Partitioning Strategy
- **Topic Key:** `genai:tasks:{tier}:{region}`
  - Example: `genai:tasks:vip:us-east-1`
- **Consumer Group:** `inference_orchestrator_v1`
- **Delivery Guarantees:** At-least-once delivery with explicit ACK (`XACK`).

### 3.2 Abandoned Task Reclaiming (`XAUTOCLAIM`)
If a worker node crashes or suffers a kernel panic mid-inference:
- Workers send heartbeats every 1,500ms.
- The supervisor runs `XAUTOCLAIM` every 3,000ms to reassign tasks in the Pending Entries List (PEL) older than 5,000ms to healthy standby workers.

### 3.3 Warm-Cache Model Affinity Routing
To avoid 15–30s model reloading latencies (e.g. FLUX.1 weights ~24GB):
- The orchestrator maintains a real-time bitmask of models cached in worker VRAM.
- Incoming jobs are routed using consistent hashing with affinity weighting:
  `TargetWorker = SelectWorker(Task.ModelID, ActiveWorkers)`

---

## 4. Failure Handling, Dead-Letter Queue (DLQ) & Auto-Heal

### 4.1 Error Classification & Handling Table

| Error Type | Detection Pattern | Recovery Action | Max Retries |
| :--- | :--- | :--- | :--- |
| **CUDA OOM** | Exit code 137 / `CUDA out of memory` | Re-queue to higher VRAM tier (48GB/80GB A100/H100) or auto-reduce batch/resolution. | 2 |
| **Worker Socket Timeout** | TCP RST / Heartbeat missed > 5s | `XAUTOCLAIM` task; isolate worker pod for K8s pod recreation. | 3 |
| **Model Corrupt / Invalid Prompt** | Inference parser error | Immediately divert to DLQ (`genai:dlq`); publish alert to GraphQL subscription. | 0 (No retry) |
| **Upstream S3 Upload Timeout** | MinIO / S3 HTTP 503 | Exponential backoff (100ms, 400ms, 1600ms) with full jitter. | 5 |

---

## 5. High-Throughput Media Processing (FFmpeg 7.x Pipes)
For post-generation media handling:
1. Video frames from diffusion workers are piped through stdout into a Go worker pool.
2. Go executes `ffmpeg` via streaming memory pipes (`io.Pipe`):
   ```bash
   ffmpeg -f rawvideo -pix_fmt rgb24 -s 1024x1024 -r 30 -i - \
          -c:v libsvtav1 -crf 26 -preset 7 -movflags +faststart \
          -f mp4 pipe:1
   ```
3. The compressed stream is piped directly into S3 Multipart Upload without touching local SSD storage, minimizing I/O bottlenecks.
