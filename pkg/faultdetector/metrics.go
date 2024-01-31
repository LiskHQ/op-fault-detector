package faultdetector

import "github.com/prometheus/client_golang/prometheus"

// FaultDetectorMetrics encapsulates faul detector related metrics.
type FaultDetectorMetrics struct {
	highestOutputIndex   prometheus.Gauge
	stateMismatch        prometheus.Gauge
	apiConnectionFailure prometheus.Gauge
}

// NewFaultDetectorMetrics returns [FaultDetectorMetrics] with initialized metrics and registering to prometheus registry.
func NewFaultDetectorMetrics(reg prometheus.Registerer) *FaultDetectorMetrics {
	m := &FaultDetectorMetrics{
		highestOutputIndex: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "fault_detector_highest_output_index",
				Help: "The highest current output index",
			}),
		stateMismatch: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "fault_detector_is_state_mismatch",
			Help: "0 when state is matched, 1 when mismatch",
		}),
		apiConnectionFailure: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "fault_detector_api_connection_failure",
			Help: "Number of times API call failed",
		}),
	}
	reg.MustRegister(m.highestOutputIndex)
	reg.MustRegister(m.stateMismatch)

	return m
}
