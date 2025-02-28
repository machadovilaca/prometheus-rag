package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
)

type LLM interface {
	Run(query string) (string, error)
}

type Config struct {
	BaseURL string
	APIKey  string
	Model   string

	VectorDB vectordb.VectorDB
}

type llm struct {
	client *openai.Client
	config Config

	vectorDB vectordb.VectorDB
}

func New(config Config) (LLM, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	if config.VectorDB == nil {
		return nil, fmt.Errorf("vectorDB is required")
	}

	if config.Model == "" {
		config.Model = ModelGranite318bInstruct
	}

	return &llm{
		client: openai.NewClient(
			option.WithBaseURL(config.BaseURL),
			option.WithAPIKey(config.APIKey),
		),
		config:   config,
		vectorDB: config.VectorDB,
	}, nil
}

func (l *llm) Run(query string) (string, error) {
	metrics, err := l.vectorDB.SearchMetrics(query, 10)
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

	return chatCompletion.Choices[0].Message.Content, nil
}
