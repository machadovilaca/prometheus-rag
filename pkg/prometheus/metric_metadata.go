package prometheus

import (
	"errors"
	"strings"
)

// MetricMetadata represents metadata associated with a Prometheus metric
type MetricMetadata struct {
	// Name is the name of the metric
	Name string `json:"name"`

	// Help provides a description of what the metric represents
	Help string `json:"help"`

	// Type indicates the type of metric (counter, gauge, histogram, etc)
	Type string `json:"type"`

	// Labels contains the label names associated with the metric
	Labels []string `json:"labels,omitempty"`
}

// Validate validates the metric metadata
func (m *MetricMetadata) Validate() error {
	if m.Name == "" {
		return errors.New("name is required")
	}

	return nil
}

// ToMap converts the metric metadata to a map
func (m *MetricMetadata) ToMap() map[string]any {
	return map[string]any{
		"name":   m.Name,
		"help":   m.Help,
		"type":   m.Type,
		"labels": strings.Join(m.Labels, ", "),
	}
}
