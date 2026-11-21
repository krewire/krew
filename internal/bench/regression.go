package bench

import (
	"fmt"
	"math"
)

// AlertSeverity indicates the criticality of a performance delta.
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "INFO"
	SeverityWarning  AlertSeverity = "WARNING"
	SeverityCritical AlertSeverity = "CRITICAL"
)

// RegressionAlert represents an observed degradation relative to baseline.
type RegressionAlert struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Metric      string        `json:"metric"` // "latency", "memory", "allocs"
	BaselineVal int64         `json:"baseline_val"`
	CurrentVal  int64         `json:"current_val"`
	DeltaPct    float64       `json:"delta_pct"`
	Severity    AlertSeverity `json:"severity"`
	Message     string        `json:"message"`
}

// RegressionPolicy defines thresholds for regression warnings and critical alerts.
type RegressionPolicy struct {
	// WarnLatencyPct is the percentage increase in execution time triggering a warning (default 15.0%).
	WarnLatencyPct float64 `json:"warn_latency_pct"`
	// CritLatencyPct is the percentage increase in execution time triggering a critical alert (default 30.0%).
	CritLatencyPct float64 `json:"crit_latency_pct"`
	// WarnMemoryPct is the percentage increase in memory/allocations triggering a warning (default 20.0%).
	WarnMemoryPct float64 `json:"warn_memory_pct"`
	// CritMemoryPct is the percentage increase in memory/allocations triggering a critical alert (default 40.0%).
	CritMemoryPct float64 `json:"crit_memory_pct"`
	// NoiseFloorNs is the minimum execution time below which jitter is ignored (default 20_000 ns / 20 µs).
	NoiseFloorNs int64 `json:"noise_floor_ns"`
}

// DefaultPolicy returns the canonical regression policy thresholds.
func DefaultPolicy() RegressionPolicy {
	return RegressionPolicy{
		WarnLatencyPct: 15.0,
		CritLatencyPct: 30.0,
		WarnMemoryPct:  20.0,
		CritMemoryPct:  40.0,
		NoiseFloorNs:   50_000, // 50 µs noise floor for microbenchmarks
	}
}

// CompareReports evaluates a current benchmark run against a baseline report.
func CompareReports(baseline, current *BenchmarkReport, policy RegressionPolicy) []RegressionAlert {
	var alerts []RegressionAlert
	if baseline == nil || current == nil {
		return alerts
	}

	baseMap := make(map[string]BenchmarkResult)
	for _, b := range baseline.Results {
		baseMap[b.ID] = b
	}

	for _, curr := range current.Results {
		base, exists := baseMap[curr.ID]
		if !exists {
			continue
		}

		// 1. Latency comparison
		if base.NsPerOp > policy.NoiseFloorNs && curr.NsPerOp > policy.NoiseFloorNs {
			deltaNs := curr.NsPerOp - base.NsPerOp
			deltaPct := (float64(deltaNs) / float64(base.NsPerOp)) * 100.0

			if deltaPct >= policy.CritLatencyPct {
				alerts = append(alerts, RegressionAlert{
					ID:          curr.ID,
					Name:        curr.Name,
					Metric:      "latency",
					BaselineVal: base.NsPerOp,
					CurrentVal:  curr.NsPerOp,
					DeltaPct:    math.Round(deltaPct*10) / 10,
					Severity:    SeverityCritical,
					Message:     fmt.Sprintf("Critical latency regression: %s increased by +%.1f%% (base: %d ns, current: %d ns)", curr.Name, deltaPct, base.NsPerOp, curr.NsPerOp),
				})
			} else if deltaPct >= policy.WarnLatencyPct {
				alerts = append(alerts, RegressionAlert{
					ID:          curr.ID,
					Name:        curr.Name,
					Metric:      "latency",
					BaselineVal: base.NsPerOp,
					CurrentVal:  curr.NsPerOp,
					DeltaPct:    math.Round(deltaPct*10) / 10,
					Severity:    SeverityWarning,
					Message:     fmt.Sprintf("Latency warning: %s increased by +%.1f%% (base: %d ns, current: %d ns)", curr.Name, deltaPct, base.NsPerOp, curr.NsPerOp),
				})
			}
		}

		// 2. Memory allocations comparison
		if base.BytesPerOp > 0 && curr.BytesPerOp > 0 {
			deltaBytes := curr.BytesPerOp - base.BytesPerOp
			deltaPct := (float64(deltaBytes) / float64(base.BytesPerOp)) * 100.0

			if deltaPct >= policy.CritMemoryPct {
				alerts = append(alerts, RegressionAlert{
					ID:          curr.ID,
					Name:        curr.Name,
					Metric:      "memory",
					BaselineVal: base.BytesPerOp,
					CurrentVal:  curr.BytesPerOp,
					DeltaPct:    math.Round(deltaPct*10) / 10,
					Severity:    SeverityCritical,
					Message:     fmt.Sprintf("Critical memory regression: %s increased by +%.1f%% bytes/op (base: %d B, current: %d B)", curr.Name, deltaPct, base.BytesPerOp, curr.BytesPerOp),
				})
			} else if deltaPct >= policy.WarnMemoryPct {
				alerts = append(alerts, RegressionAlert{
					ID:          curr.ID,
					Name:        curr.Name,
					Metric:      "memory",
					BaselineVal: base.BytesPerOp,
					CurrentVal:  curr.BytesPerOp,
					DeltaPct:    math.Round(deltaPct*10) / 10,
					Severity:    SeverityWarning,
					Message:     fmt.Sprintf("Memory warning: %s increased by +%.1f%% bytes/op (base: %d B, current: %d B)", curr.Name, deltaPct, base.BytesPerOp, curr.BytesPerOp),
				})
			}
		}
	}

	return alerts
}
