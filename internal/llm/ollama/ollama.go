package ollama

import (
	"coda/internal/llm"
	"coda/internal/logger"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ollama/ollama/api"
)

// Supported models
var (
	ModelTinySwallow = llm.Model{
		Name:          "yottahmd/tiny-swallow-1.5b-instruct",
		DisplayName:   "Sakana AI Tiny Swallow 1.5B",
		Provider:      llm.Ollama,
		MaxToken:      32768,
		ContextWindow: 32768,
		PDFSupported:  true,
		Version:       "2024-03-05",
		Family:        "SakanaAI",
		Pricing: &llm.ModelPricing{
			InputPerToken:  0.00000,
			OutputPerToken: 0.00000,
			Currency:       "USD",
		},
		Capabilities: llm.ModelCapabilities{
			SupportsStreaming: true,
			SupportsFunctions: false,
			SupportsVision:    false,
			SupportsJSON:      false,
		},
	}
)

// Ensure Client implements the LLM interface
var _ llm.LLM = (*Client)(nil)

// Client is an Ollama client that implements the LLM interface.
type Client struct {
	cfg llm.Config
}

// New creates a new Ollama client.
func New(cfg llm.Config) (llm.LLM, error) {
	return &Client{cfg: cfg}, nil
}

// Complete processes the given parameters and returns a completion response.
func (c *Client) Complete(
	ctx context.Context,
	params llm.CompleteParams,
) (*llm.CompleteResponse, error) {
	startTime := time.Now()

	httpClient := &http.Client{
		Timeout: c.cfg.Timeout,
	}

	u, err := url.Parse(c.cfg.LLMConfig.Ollama.BaseURL)
	if err != nil {
		logger.Error(ctx, "invalid base URL", "err", err)
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	client := api.NewClient(u, httpClient)

	// Convert messages to Ollama format
	var messages []api.Message
	for _, m := range params.Messages {
		switch m.Role {
		case llm.RoleUser:
			messages = append(messages, api.Message{
				Role:    "user",
				Content: m.Content,
			})
		case llm.RoleAssistant:
			messages = append(messages, api.Message{
				Role:    "assistant",
				Content: m.Content,
			})
		case llm.RoleSystem:
			messages = append(messages, api.Message{
				Role:    "system",
				Content: m.Content,
			})
		case llm.RoleFunction:
			fallthrough
		default:
			return nil, fmt.Errorf("unsupported role: %s", m.Role)
		}
	}

	// Build request parameters
	stream := false
	req := &api.ChatRequest{
		Model:    c.cfg.Model.Name,
		Messages: messages,
		Stream:   &stream,
	}

	var msgs []llm.Message
	var resp api.ChatResponse
	respFunc := func(resp api.ChatResponse) error {
		msgs = append(msgs, llm.Message{
			Role:         llmRole(resp.Message.Role),
			Content:      resp.Message.Content,
			FinishReason: resp.DoneReason,
			Completed:    true,
		})
		return nil
	}

	if err := client.Chat(ctx, req, respFunc); err != nil {
		return nil, c.handleError(err)
	}

	// Build the response
	ret := &llm.CompleteResponse{
		Messages: msgs,
		Metadata: llm.CompletionMetadata{
			ModelName:    c.cfg.Model.Name,
			FinishReason: string(resp.DoneReason),
			LatencyMs:    time.Since(startTime).Milliseconds(),
			ProcessedAt:  time.Now().UTC(),
		},
	}

	return ret, nil
}

// handleError converts Ollama errors to our error types.
func (c *Client) handleError(err error) error {
	var statusError api.StatusError
	if errors.As(err, &statusError) {
		llmErr := llm.NewLLMError(err, string(llm.Ollama), c.cfg.Model.Name).
			WithStatusCode(statusError.StatusCode).
			WithErrorCode(statusError.Status).
			WithErrorMessage(statusError.ErrorMessage)

		// Handle HTTP status code ranges
		if statusError.StatusCode >= 500 && statusError.StatusCode < 600 {
			llmErr.Err = llm.ErrServiceUnavailable
			llmErr.Retryable = true
			return llmErr
		}

		if statusError.StatusCode == 429 {
			llmErr.Err = llm.ErrTooManyRequests
			llmErr.Retryable = true
			return llmErr
		}

		return llmErr
	}

	// For non-API errors, wrap in our error type
	return llm.NewLLMError(err, string(llm.Ollama), c.cfg.Model.Name)
}

func init() {
	// Register supported models
	llm.RegisterLLM(New, []llm.Model{
		ModelTinySwallow,
	})
}

func llmRole(ollamaRole string) llm.Role {
	switch ollamaRole {
	case "user":
		return llm.RoleUser
	case "assistant":
		return llm.RoleAssistant
	case "system":
		return llm.RoleSystem
	default:
		return llm.RoleFunction
	}
}
