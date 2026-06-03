package templatex

import (
	"sync"
	"testing"
)

type metricCall struct {
	name   string
	value  float64
	labels map[string]string
}

type recordingMetrics struct {
	mu         sync.Mutex
	counters   []metricCall
	histograms []metricCall
	gauges     []metricCall
}

func (m *recordingMetrics) IncCounter(name string, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters = append(m.counters, metricCall{name: name, labels: cloneLabels(labels)})
}

func (m *recordingMetrics) ObserveHistogram(name string, value float64, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.histograms = append(m.histograms, metricCall{name: name, value: value, labels: cloneLabels(labels)})
}

func (m *recordingMetrics) SetGauge(name string, value float64, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges = append(m.gauges, metricCall{name: name, value: value, labels: cloneLabels(labels)})
}

func (m *recordingMetrics) hasCounter(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.counters {
		if call.name == name {
			return true
		}
	}
	return false
}

func (m *recordingMetrics) counterWithLabel(name string, key string, value string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.counters {
		if call.name == name && call.labels[key] == value {
			return true
		}
	}
	return false
}

func (m *recordingMetrics) hasGauge(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.gauges {
		if call.name == name {
			return true
		}
	}
	return false
}

func (m *recordingMetrics) gaugeWithLabels(name string, value float64, labels map[string]string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.gauges {
		if call.name == name && call.value == value && sameLabels(call.labels, labels) {
			return true
		}
	}
	return false
}

func (m *recordingMetrics) hasHistogram(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.histograms {
		if call.name == name {
			return true
		}
	}
	return false
}

func (m *recordingMetrics) histogramWithLabels(name string, labels map[string]string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.histograms {
		if call.name == name && sameLabels(call.labels, labels) {
			return true
		}
	}
	return false
}

func sameLabels(actual map[string]string, expected map[string]string) bool {
	if len(actual) != len(expected) {
		return false
	}
	for key, value := range expected {
		if actual[key] != value {
			return false
		}
	}
	return true
}

func cloneLabels(labels map[string]string) map[string]string {
	if labels == nil {
		return nil
	}
	cloned := make(map[string]string, len(labels))
	for key, value := range labels {
		cloned[key] = value
	}
	return cloned
}

func TestNoopMetricsAcceptsCalls(t *testing.T) {
	metrics := NoopMetrics{}

	metrics.IncCounter(MetricClientCreatedTotal, map[string]string{"name": "templatex"})
	metrics.ObserveHistogram(MetricClientHealthLatencyMS, 1, nil)
	metrics.SetGauge(MetricClientHealthStatus, 1, map[string]string{"status": string(HealthHealthy)})
}
