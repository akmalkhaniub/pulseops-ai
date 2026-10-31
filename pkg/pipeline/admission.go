package pipeline

import (
	"errors"
	"sync"
	"sync/atomic"
)

// Tier represents customer SLA priority tiers.
type Tier string

const (
	TierFree Tier = "FREE"
	TierPro  Tier = "PRO"
	TierVIP  Tier = "VIP"
)

// TelemetryFrame represents a zero-allocation GPU worker heartbeat metric frame.
type TelemetryFrame struct {
	NodeID            string
	Region            string
	CachedModels      []string
	VRAMUsedMB        int64
	VRAMTotalMB       int64
	QueueDepth        int32
	P99QueueLatencyMs int64
	LastErrorCode     string
}

// Reset clears a TelemetryFrame before returning it to sync.Pool.
func (f *TelemetryFrame) Reset() {
	f.NodeID = ""
	f.Region = ""
	f.CachedModels = f.CachedModels[:0]
	f.VRAMUsedMB = 0
	f.VRAMTotalMB = 0
	f.QueueDepth = 0
	f.P99QueueLatencyMs = 0
	f.LastErrorCode = ""
}

// FramePool provides zero-allocation TelemetryFrame reuse on the hot ingestion path.
type FramePool struct {
	pool sync.Pool
}

func NewFramePool() *FramePool {
	return &FramePool{
		pool: sync.Pool{
			New: func() any {
				return &TelemetryFrame{
					CachedModels: make([]string, 0, 8),
				}
			},
		},
	}
}

func (p *FramePool) Acquire() *TelemetryFrame {
	return p.pool.Get().(*TelemetryFrame)
}

func (p *FramePool) Release(f *TelemetryFrame) {
	if f == nil {
		return
	}
	f.Reset()
	p.pool.Put(f)
}

// ClusterHealth captures atomic cluster-wide pressure signals.
type ClusterHealth struct {
	P99QueueLatencyMs int64
	ActiveGPUWorkers  int32
	VRAMSaturationPct int32 // 0 - 100
}

// AdmissionController enforces dynamic multi-tier load shedding under GPU cluster stress.
type AdmissionController struct {
	health atomic.Pointer[ClusterHealth]
}

func NewAdmissionController() *AdmissionController {
	ac := &AdmissionController{}
	ac.health.Store(&ClusterHealth{
		P99QueueLatencyMs: 120,
		ActiveGPUWorkers:  64,
		VRAMSaturationPct: 45,
	})
	return ac
}

func (ac *AdmissionController) UpdateHealth(h ClusterHealth) {
	ac.health.Store(&h)
}

// CanAdmit evaluates whether a request for the given tier should be admitted or shed.
func (ac *AdmissionController) CanAdmit(tier Tier) bool {
	h := ac.health.Load()
	if h == nil {
		return true
	}

	// Severity 1: Severe Cluster Stress (p99 > 2500ms or VRAM > 94%) -> VIP only
	if h.P99QueueLatencyMs > 2500 || h.VRAMSaturationPct > 94 {
		return tier == TierVIP
	}

	// Severity 2: Moderate Load (p99 > 1200ms or VRAM > 85%) -> VIP and PRO only
	if h.P99QueueLatencyMs > 1200 || h.VRAMSaturationPct > 85 {
		return tier == TierVIP || tier == TierPro
	}

	return true
}

// RemediationAction represents an autonomous self-healing action triggered by TriageEngine.
type RemediationAction string

const (
	ActionNone              RemediationAction = "NONE"
	ActionCordonAndDrain    RemediationAction = "CORDON_AND_DRAIN"
	ActionMicroVMColdReboot RemediationAction = "MICROVM_COLD_REBOOT"
	ActionShiftTrafficPool  RemediationAction = "SHIFT_TRAFFIC_POOL"
)

// EvaluateIncident inspects a worker TelemetryFrame and determines autonomous self-healing action.
func EvaluateIncident(f *TelemetryFrame) RemediationAction {
	if f == nil {
		return ActionNone
	}
	switch f.LastErrorCode {
	case "CUDA_ERR_OUT_OF_MEMORY":
		return ActionCordonAndDrain
	case "NVML_DRIVER_TIMEOUT", "XID_79_GPU_FELL_OFF_BUS":
		return ActionMicroVMColdReboot
	}
	if f.VRAMTotalMB > 0 && (f.VRAMUsedMB*100)/f.VRAMTotalMB >= 96 {
		return ActionShiftTrafficPool
	}
	return ActionNone
}

// SelectAffinityWorker picks the lowest-latency healthy worker with warm VRAM model affinity.
func SelectAffinityWorker(modelID string, workers []TelemetryFrame) (string, error) {
	if len(workers) == 0 {
		return "", errors.New("no active GPU workers available")
	}

	var bestWarm *TelemetryFrame
	var bestCold *TelemetryFrame

	for i := range workers {
		w := &workers[i]
		if EvaluateIncident(w) != ActionNone {
			continue
		}
		hasModel := false
		for _, m := range w.CachedModels {
			if m == modelID {
				hasModel = true
				break
			}
		}
		if hasModel {
			if bestWarm == nil || w.QueueDepth < bestWarm.QueueDepth {
				bestWarm = w
			}
		} else {
			if bestCold == nil || w.QueueDepth < bestCold.QueueDepth {
				bestCold = w
			}
		}
	}

	if bestWarm != nil {
		return bestWarm.NodeID, nil
	}
	if bestCold != nil {
		return bestCold.NodeID, nil
	}
	return "", errors.New("all GPU workers are unhealthy or cordoned")
}
