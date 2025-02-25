package embeddings

import (
	"context"
	"fmt"
	"strings"

	"github.com/nlpodyssey/cybertron/pkg/models/bert"
	"github.com/nlpodyssey/cybertron/pkg/tasks"
	"github.com/nlpodyssey/cybertron/pkg/tasks/textencoding"

	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
)

const (
	modelsDir = "../../_models"
	modelName = textencoding.DefaultModelMulti
)

// Encoder is an interface for encoding queries and metric metadata
type Encoder interface {
	// GetDimension returns the dimension of the encoded vector
	GetDimension() (int, error)

	// EncodeQuery encodes a query into a vector
	EncodeQuery(query string) ([]float32, error)

	// EncodeMetricMetadata encodes a metric metadata into a vector
	EncodeMetricMetadata(metadata prometheus.MetricMetadata) ([]float32, error)
}

// Config is the configuration for the encoder
type Config struct {
	ModelsDir string
	ModelName string
}

type encoder struct {
	config *tasks.Config
	model  textencoding.Interface
}

// NewEncoder creates a new encoder
func NewEncoder(config Config) (Encoder, error) {
	if config.ModelsDir == "" {
		config.ModelsDir = modelsDir
	}

	if config.ModelName == "" {
		config.ModelName = modelName
	}

	tasksConfig := &tasks.Config{
		ModelsDir: config.ModelsDir,
		ModelName: config.ModelName,
	}

	m, err := tasks.LoadModelForTextEncoding(tasksConfig)
	if err != nil {
		return nil, err
	}

	return &encoder{config: tasksConfig, model: m}, nil
}

func (e *encoder) GetDimension() (int, error) {
	result, err := e.model.Encode(context.Background(), "", int(bert.MeanPooling))
	if err != nil {
		return 0, err
	}

	return len(result.Vector.Data().F32()), nil
}

func (e *encoder) EncodeQuery(query string) ([]float32, error) {
	result, err := e.model.Encode(context.Background(), lowercase(query), int(bert.MeanPooling))
	if err != nil {
		return nil, err
	}

	return result.Vector.Data().F32(), nil
}

func (e *encoder) EncodeMetricMetadata(metadata prometheus.MetricMetadata) ([]float32, error) {
	if err := metadata.Validate(); err != nil {
		return nil, fmt.Errorf("invalid metric metadata: %w", err)
	}

	text := fmt.Sprintf("%s %s", metadata.Name, metadata.Help)

	result, err := e.model.Encode(context.Background(), lowercase(text), int(bert.MeanPooling))
	if err != nil {
		return nil, err
	}

	return result.Vector.Data().F32(), nil
}

func lowercase(s string) string {
	return strings.ToLower(s)
}
