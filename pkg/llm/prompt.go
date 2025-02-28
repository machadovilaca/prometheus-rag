package llm

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	"github.com/rs/zerolog/log"

	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
)

//go:embed promql_prompt.tmpl
var promptTemplate string

// PromptData is the wrapper for metrics metadata to be used in the prompt
type PromptData struct {
	Metrics []*prometheus.MetricMetadata
}

// BuildPrompt builds a prompt for the LLM using the metrics metadata
func BuildPrompt(metrics []*prometheus.MetricMetadata) (string, error) {
	tmpl, err := template.New("promql_prompt").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var promptBuf bytes.Buffer
	err = tmpl.ExecuteTemplate(&promptBuf, "PromqlSystemPrompt", PromptData{
		Metrics: metrics,
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	log.Debug().Msgf("prompt: %s", promptBuf.String())
	return promptBuf.String(), nil
}
