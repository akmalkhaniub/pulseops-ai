package pipeline

import "testing"

func TestAdmissionController_MultiTierLoadShedding(t *testing.T) {
	ac := NewAdmissionController()

	// Nominal cluster health: all tiers admitted
	ac.UpdateHealth(ClusterHealth{P99QueueLatencyMs: 250, ActiveGPUWorkers: 128, VRAMSaturationPct: 60})
	if !ac.CanAdmit(TierFree) || !ac.CanAdmit(TierPro) || !ac.CanAdmit(TierVIP) {
		t.Fatal("expected all tiers admitted under nominal load")
	}

	// Moderate stress: Free shed, Pro & VIP admitted
	ac.UpdateHealth(ClusterHealth{P99QueueLatencyMs: 1450, ActiveGPUWorkers: 128, VRAMSaturationPct: 88})
	if ac.CanAdmit(TierFree) {
		t.Fatal("expected FREE tier shed under moderate cluster stress")
	}
	if !ac.CanAdmit(TierPro) || !ac.CanAdmit(TierVIP) {
		t.Fatal("expected PRO and VIP admitted under moderate cluster stress")
	}

	// Severe stress: Only VIP admitted
	ac.UpdateHealth(ClusterHealth{P99QueueLatencyMs: 2900, ActiveGPUWorkers: 128, VRAMSaturationPct: 97})
	if ac.CanAdmit(TierFree) || ac.CanAdmit(TierPro) {
		t.Fatal("expected FREE and PRO shed under severe cluster stress")
	}
	if !ac.CanAdmit(TierVIP) {
		t.Fatal("expected VIP admitted under severe cluster stress")
	}
}

func TestFramePool_ZeroAllocationReuse(t *testing.T) {
	pool := NewFramePool()
	frame := pool.Acquire()
	frame.NodeID = "gpu-node-h100-01"
	frame.CachedModels = append(frame.CachedModels, "flux-1-dev")
	frame.VRAMUsedMB = 72000

	pool.Release(frame)

	reused := pool.Acquire()
	if reused.NodeID != "" || len(reused.CachedModels) != 0 || reused.VRAMUsedMB != 0 {
		t.Fatalf("expected reset frame from pool, got %+v", reused)
	}
}

func TestEvaluateIncident_And_WarmCacheAffinity(t *testing.T) {
	workers := []TelemetryFrame{
		{
			NodeID:        "gpu-oom-node",
			CachedModels:  []string{"flux-1-dev"},
			QueueDepth:    1,
			LastErrorCode: "CUDA_ERR_OUT_OF_MEMORY",
		},
		{
			NodeID:       "gpu-cold-node",
			CachedModels: []string{"sdxl-turbo"},
			QueueDepth:   2,
			VRAMUsedMB:   30000,
			VRAMTotalMB:  80000,
		},
		{
			NodeID:       "gpu-warm-node",
			CachedModels: []string{"flux-1-dev"},
			QueueDepth:   5,
			VRAMUsedMB:   45000,
			VRAMTotalMB:  80000,
		},
	}

	if got := EvaluateIncident(&workers[0]); got != ActionCordonAndDrain {
		t.Fatalf("expected ActionCordonAndDrain for OOM node, got %s", got)
	}

	selected, err := SelectAffinityWorker("flux-1-dev", workers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected != "gpu-warm-node" {
		t.Fatalf("expected warm healthy node gpu-warm-node, got %s", selected)
	}
}
