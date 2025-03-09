package openai

import (
	"coda/internal/llm"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// Supported models
// https://platform.openai.com/docs/models
var (
	ModelGPT4o = llm.Model{
		Name:          "gpt-4o",
		DisplayName:   "OpenAI gpt-4o",
		Provider:      llm.OpenAI,
		MaxToken:      16384,
		ContextWindow: 128_000,
		PDFSupported:  true,
		Version:       "2023-05-15",
		Family:        "GPT-4",
		Pricing: &llm.ModelPricing{
			InputPerToken:  0.00001,
			OutputPerToken: 0.00003,
			Currency:       "USD",
		},
		Capabilities: llm.ModelCapabilities{
			SupportsStreaming: true,
			SupportsFunctions: true,
			SupportsVision:    true,
			SupportsJSON:      true,
		},
	}
)

// Error codes from OpenAI API
const (
	ErrCodeContextLengthExceeded = "context_length_exceeded"
	ErrCodeLength                = "length"
	ErrCodeInvalidAPIKey         = "invalid_api_key"
	ErrCodeRateLimitExceeded     = "rate_limit_exceeded"
	ErrCodeInsufficientQuota     = "insufficient_quota"
	ErrCodeInvalidRequestError   = "invalid_request_error"
)

// Ensure Client implements the LLM interface
var _ llm.LLM = (*Client)(nil)

// Client is an OpenAI client that implements the LLM interface.
type Client struct {
	cfg    llm.Config
	client *openai.Client
}

// New creates a new OpenAI client.
func New(cfg llm.Config) (llm.LLM, error) {
	if cfg.APIKeyFunc == nil {
		return nil, fmt.Errorf("API key function is required")
	}

	// Create the OpenAI client with options
	var client *openai.Client

	// Set custom timeout if provided
	if cfg.Timeout > 0 {
		httpClient := &http.Client{
			Timeout: cfg.Timeout,
		}
		client = openai.NewClient(option.WithHTTPClient(httpClient), option.WithAPIKey(cfg.APIKeyFunc()))
	} else {
		client = openai.NewClient(option.WithAPIKey(cfg.APIKeyFunc()))
	}

	return &Client{
		cfg:    cfg,
		client: client,
	}, nil
}

// GetModelInfo returns information about the model.
func (c *Client) GetModelInfo() llm.ModelInfo {
	return llm.ModelInfo{
		Model: c.cfg.Model,
	}
}

// Complete processes the given parameters and returns a completion response.
func (c *Client) Complete(
	ctx context.Context,
	params llm.CompleteParams,
) (*llm.CompleteResponse, error) {
	startTime := time.Now()

	// Convert messages to OpenAI format
	var messages []openai.ChatCompletionMessageParamUnion
	for _, m := range params.Messages {
		switch m.Role {
		case llm.RoleUser:
			messages = append(messages, openai.UserMessage(m.Content))
		case llm.RoleAssistant:
			messages = append(messages, openai.AssistantMessage(m.Content))
		case llm.RoleSystem:
			messages = append(messages, openai.SystemMessage(m.Content))
		case llm.RoleFunction:
			// Current version doesn't support function messages directly
			// Fallback to a user message
			messages = append(messages, openai.UserMessage(fmt.Sprintf("Function %s returned: %s", m.Name, m.Content)))
		default:
			return nil, fmt.Errorf("unsupported role: %s", m.Role)
		}
	}

	// Build request parameters
	completionParams := openai.ChatCompletionNewParams{
		Messages: openai.F(messages),
		Model:    openai.F(c.cfg.Model.Name),
		Seed:     openai.Int(1), // For reproducibility
	}

	// Add optional parameters if provided
	if params.MaxTokens != nil {
		completionParams.MaxTokens = openai.Int(int64(*params.MaxTokens))
	}

	if params.Temperature != nil {
		temp := float64(*params.Temperature)
		completionParams.Temperature = openai.Float(temp)
	}

	if params.TopP != nil {
		topP := float64(*params.TopP)
		completionParams.TopP = openai.Float(topP)
	}

	if params.N != nil {
		completionParams.N = openai.Int(int64(*params.N))
	}

	// Note: Function calling and JSON mode are not directly supported in this version
	// of the library in the same way. We would need to adapt this based on the actual
	// library version and capabilities.

	// Make the API call
	completion, err := c.client.Chat.Completions.New(ctx, completionParams)
	if err != nil {
		return nil, c.handleError(err)
	}

	// Check for empty response
	if len(completion.Choices) == 0 {
		return nil, llm.ErrNoMessages
	}

	// Convert response to our format
	var msgs []llm.Message
	for _, choice := range completion.Choices {
		msg := llm.Message{
			Role:         llm.Role(choice.Message.Role),
			Content:      choice.Message.Content,
			FinishReason: string(choice.FinishReason),
			Completed:    true,
		}

		// Note: Function calls handling would need to be adapted based on
		// the actual library version

		msgs = append(msgs, msg)
	}

	// Build the response
	ret := &llm.CompleteResponse{
		Messages: msgs,
		Usage: &llm.Usage{
			Unit:             "tokens",
			PromptTokens:     int(completion.Usage.PromptTokens),
			CompletionTokens: int(completion.Usage.CompletionTokens),
			TotalTokens:      int(completion.Usage.TotalTokens),
		},
		Metadata: llm.CompletionMetadata{
			ModelName:     c.cfg.Model.Name,
			FinishReason:  string(completion.Choices[0].FinishReason),
			CompletionID:  completion.ID,
			LatencyMs:     time.Since(startTime).Milliseconds(),
			ProcessedAt:   time.Now().UTC(),
			RequestTokens: int(completion.Usage.PromptTokens),
		},
	}

	return ret, nil
}

// handleError converts OpenAI errors to our error types.
func (c *Client) handleError(err error) error {
	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		llmErr := llm.NewLLMError(err, string(llm.OpenAI), c.cfg.Model.Name).
			WithStatusCode(apiErr.StatusCode).
			WithErrorCode(apiErr.Code)

		// Map specific error codes
		switch apiErr.Code {
		case ErrCodeContextLengthExceeded:
			llmErr.Err = llm.ErrContextLengthExceeded
			return llmErr
		case ErrCodeLength:
			llmErr.Err = llm.ErrTokenLimitReached
			return llmErr
		case ErrCodeInvalidAPIKey:
			llmErr.Err = llm.ErrInvalidAPIKey
			return llmErr
		case ErrCodeRateLimitExceeded:
			llmErr.Err = llm.ErrRateLimited
			llmErr.Retryable = true
			return llmErr
		case ErrCodeInsufficientQuota:
			llmErr.Err = errors.New("insufficient quota")
			return llmErr
		}

		// Handle HTTP status code ranges
		if apiErr.StatusCode >= 500 && apiErr.StatusCode < 600 {
			llmErr.Err = llm.ErrServiceUnavailable
			llmErr.Retryable = true
			return llmErr
		}

		if apiErr.StatusCode == 429 {
			llmErr.Err = llm.ErrTooManyRequests
			llmErr.Retryable = true
			return llmErr
		}

		return llmErr
	}

	// For non-API errors, wrap in our error type
	return llm.NewLLMError(err, string(llm.OpenAI), c.cfg.Model.Name)
}

func init() {
	llm.RegisterLLM(New, []llm.Model{
		ModelGPT4o,
	})
}
