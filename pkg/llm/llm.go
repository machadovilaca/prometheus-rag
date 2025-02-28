package llm

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/rs/zerolog/log"

	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
)

// Client interface for interacting with the LLM
type Client interface {
	// Run runs a query against the LLM
	Run(query string) (string, error)
}

// Config represents the configuration for the LLM
type Config struct {
	BaseURL string
	APIKey  string
	Model   string

	VectorDBClient vectordb.Client
}

type llm struct {
	client *openai.Client
	config Config

	vectorDBClient vectordb.Client
}

// New creates a new LLM client
func New(config Config) (Client, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	if config.VectorDBClient == nil {
		return nil, fmt.Errorf("VectorDBClient is required")
	}

	if config.Model == "" {
		config.Model = ModelGranite318bInstruct
	}

	return &llm{
		client: openai.NewClient(
			option.WithBaseURL(config.BaseURL),
			option.WithAPIKey(config.APIKey),
		),
		config:         config,
		vectorDBClient: config.VectorDBClient,
	}, nil
}

func (l *llm) Run(query string) (string, error) {
	metrics, err := l.vectorDBClient.SearchMetrics(query, 10)
	if err != nil {
		return "", fmt.Errorf("failed to search metrics: %w", err)
	}

	prompt, err := BuildPrompt(metrics)
	if err != nil {
		return "", fmt.Errorf("failed to build prompt: %w", err)
	}

	chatCompletion, err := l.client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(prompt),
			openai.UserMessage(query),
		}),
		Model: openai.F(l.config.Model),
	})
	if err != nil {
		return "", fmt.Errorf("failed to run llm: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		log.Error().Msg("no choices returned")
		return "", fmt.Errorf("no choices returned")
	}

	parsed, err := parseXMLExtract(chatCompletion.Choices[0].Message.Content)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse XML response")
		return "", fmt.Errorf("failed to parse XML response: %w", err)
	}

	return parsed, nil
}

type xmlResponse struct {
	Query struct {
		PromQL string `xml:"promql"`
	} `xml:"query"`
}

func parseXMLExtract(xmlStr string) (string, error) {
	var response xmlResponse
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	err := decoder.Decode(&response)
	if err != nil {
		return "", fmt.Errorf("failed to parse XML response: %w", err)
	}

	return response.Query.PromQL, nil
}
