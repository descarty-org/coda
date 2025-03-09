package llm

import (
	"coda/internal/config"
	"coda/internal/llm/langfuse"
	"coda/internal/logger"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"golang.org/x/sync/errgroup"
)

// RetryConfig defines the configuration for retry logic.
type RetryConfig struct {
	MaxAttempts int           // Maximum number of retry attempts
	InitialWait time.Duration // Initial wait time before first retry
	MaxWait     time.Duration // Maximum wait time between retries
	Factor      float64       // Exponential backoff factor
}

// DefaultRetryConfig provides sensible default values for retry configuration.
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	InitialWait: 500 * time.Millisecond,
	MaxWait:     5 * time.Second,
	Factor:      1.5,
}

// Completer completes prompts using different models with retry and fallback logic.
type Completer interface {
	// Complete completes the prompt set and returns the result.
	Complete(
		ctx context.Context,
		params CompleteParams,
		model Model,
	) (*CompleteResponse, error)

	// CompleteWithFallback attempts to complete using the primary model,
	// falling back to alternative models if the primary fails.
	CompleteWithFallback(
		ctx context.Context,
		params CompleteParams,
		primaryModel Model,
		fallbackModels ...Model,
	) (*CompleteResponse, error)

	// GetAvailableModels returns a list of available models.
	GetAvailableModels() []Model
}

// Ensure completer implements Completer interface
var _ Completer = (*completer)(nil)

type completer struct {
	cfg         *config.Config
	langfuse    *langfuse.Client
	retryConfig RetryConfig
	registry    *Registry
}

// CompleterOption defines functional options for configuring the completer.
type CompleterOption func(*completer)

// WithCompleterRetryConfig sets a custom retry configuration for the completer.
func WithCompleterRetryConfig(rc RetryConfig) CompleterOption {
	return func(c *completer) {
		c.retryConfig = rc
	}
}

// NewCompleter creates a new Completer with the given options.
func NewCompleter(cfg *config.Config, registry *Registry, opts ...CompleterOption) Completer {
	c := &completer{
		cfg:         cfg,
		retryConfig: DefaultRetryConfig,
		registry:    registry,
	}

	if cfg.LLM.Langfuse.IsConfigured() {
		c.langfuse = langfuse.NewClient(cfg)
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// GetAvailableModels returns a list of available models.
func (c *completer) GetAvailableModels() []Model {
	return c.registry.models
}

// Complete completes the prompt set and returns the result with retry logic.
func (c *completer) Complete(
	ctx context.Context,
	params CompleteParams,
	model Model,
) (*CompleteResponse, error) {
	var (
		res *CompleteResponse
		err error
	)

	// Get API key function for the provider
	apiKeyFunc, err := c.getAPIKeyFunc(model.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	// Initialize LLM client
	llm, err := New(Config{
		Model:      model,
		APIKeyFunc: apiKeyFunc,
		Timeout:    120 * time.Second,
		LLMConfig:  c.cfg.LLM,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM client: %w", err)
	}

	// Implement retry logic with exponential backoff
	var lastErr error
	wait := c.retryConfig.InitialWait

	for attempt := 0; attempt < c.retryConfig.MaxAttempts; attempt++ {
		// Check if context is canceled before making the attempt
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// If this is a retry, log the attempt
		if attempt > 0 {
			logger.Info(ctx, "retrying LLM request",
				"attempt", attempt+1,
				"model", model.Name,
				"previous_error", lastErr)

			// Wait before retrying
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
				// Increase wait time for next attempt, but don't exceed max wait
				wait = time.Duration(float64(wait) * c.retryConfig.Factor)
				if wait > c.retryConfig.MaxWait {
					wait = c.retryConfig.MaxWait
				}
			}
		}

		// Attempt to complete
		res, err = llm.Complete(ctx, params)

		// If successful or if error is not retryable, break the loop
		if err == nil {
			break
		}

		// Store the last error
		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			break
		}
	}

	// If all attempts failed, return the last error
	if err != nil {
		return nil, fmt.Errorf("all completion attempts failed: %w", lastErr)
	}

	// Validate response
	if len(res.Messages) == 0 {
		return nil, ErrNoMessages
	}

	// Send trace events to Langfuse asynchronously
	go func() {
		// Recover from any panics
		defer func() {
			if r := recover(); r != nil {
				var err error
				if e, ok := r.(error); ok {
					err = e
				} else {
					err = fmt.Errorf("%v", r)
				}
				logger.Error(ctx, "panic while sending trace events", "error", err)
			}
		}()

		// Create a new context for the background operation
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		c.sendTraceEvents(bgCtx, model, params, res)
	}()

	return res, nil
}

// CompleteWithFallback attempts to complete using the primary model,
// falling back to alternative models if the primary fails.
func (c *completer) CompleteWithFallback(
	ctx context.Context,
	params CompleteParams,
	primaryModel Model,
	fallbackModels ...Model,
) (*CompleteResponse, error) {
	// Try primary model first
	res, err := c.Complete(ctx, params, primaryModel)
	if err == nil {
		return res, nil
	}

	// Log the primary model failure
	logger.Warn(ctx, "primary model failed, trying fallbacks",
		"primary_model", primaryModel.Name,
		"error", err)

	// Try fallback models in sequence
	for i, fallbackModel := range fallbackModels {
		// Check if context is canceled before trying fallback
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		logger.Info(ctx, "attempting fallback model",
			"fallback_model", fallbackModel.Name,
			"fallback_index", i+1)

		res, err = c.Complete(ctx, params, fallbackModel)
		if err == nil {
			return res, nil
		}

		logger.Warn(ctx, "fallback model failed",
			"fallback_model", fallbackModel.Name,
			"error", err)
	}

	// If all models failed, return the last error
	return nil, fmt.Errorf("all models failed: %w", err)
}

// emptyAPIKeyFunc returns an empty API key.
var emptyAPIKeyFunc = func() string {
	return ""
}

// getAPIKeyFunc returns the appropriate API key function for the given provider.
func (c *completer) getAPIKeyFunc(provider Provider) (APIKeyFunc, error) {
	switch provider {
	case OpenAI:
		return func() string {
			return c.cfg.LLM.OpenAI.APIKey
		}, nil
	case Ollama:
		return emptyAPIKeyFunc, nil
	default:
		return nil, fmt.Errorf("provider %q is not supported", provider)
	}
}

// isRetryableError determines if an error should trigger a retry.
func isRetryableError(err error) bool {
	return errors.Is(err, ErrServiceUnavailable) ||
		errors.Is(err, ErrTooManyRequests) ||
		errors.Is(err, context.DeadlineExceeded)
}

// sendTraceEvents sends telemetry data to Langfuse for observability.
func (c *completer) sendTraceEvents(ctx context.Context, model Model, params CompleteParams, res *CompleteResponse) {
	// Skip if Langfuse is not configured
	if c.langfuse == nil {
		return
	}

	genUUID := func() string {
		u, err := uuid.NewV7()
		if err != nil {
			// Fallback to V4 if V7 fails
			u, _ = uuid.NewV4()
		}
		return u.String()
	}

	// Create a unique trace ID for this interaction
	traceID := genUUID()
	generationID := genUUID()

	// Calculate timestamps
	now := time.Now().UTC()
	startTime := now.Format(time.RFC3339Nano)

	// Use actual timestamps if available, otherwise estimate
	completionStartTime := now.Add(500 * time.Millisecond).Format(time.RFC3339Nano)
	endTime := now.Add(1200 * time.Millisecond).Format(time.RFC3339Nano)

	// Extract user ID from context if available, otherwise generate one
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		userID = genUUID()
	}

	// Extract model parameters from the request
	modelParams := extractModelParameters(params)

	generationBody := langfuse.GenerationBody{
		ID:                  generationID,
		TraceID:             traceID,
		Name:                "Model Response",
		StartTime:           startTime,
		CompletionStartTime: completionStartTime,
		EndTime:             endTime,
		Model:               model.Name,
		ModelParameters:     modelParams,
		Input:               params.Messages,
		Output:              res.Messages,
		Level:               "DEFAULT",
	}

	if res.Usage != nil {
		// Track token usage for cost calculation
		generationBody.UsageDetails = map[string]int{
			"prompt_tokens":     res.Usage.PromptTokens,
			"completion_tokens": res.Usage.CompletionTokens,
			"total_tokens":      res.Usage.TotalTokens,
		}
	}

	// Create a batch of events
	batch := []langfuse.Event{
		// Create a trace for this interaction
		langfuse.CreateTrace(
			genUUID(),
			langfuse.TraceBody{
				ID:          traceID,
				Name:        "Model Interaction",
				UserID:      userID,
				Input:       getLastUserMessage(params.Messages),
				Output:      res.Messages[0].Content,
				Timestamp:   startTime,
				Environment: getEnvironment(c.cfg),
				Tags:        []string{model.Name, string(model.Provider)},
			},
		),

		// Track the model usage
		langfuse.CreateGeneration(
			genUUID(),
			generationBody,
		),
	}

	// Send the batch to Langfuse with a timeout
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		resp, err := c.langfuse.Ingest(batch)
		if err != nil {
			logger.Error(gCtx, "failed to send trace events to Langfuse", "err", err)
			return err
		}

		if len(resp.Errors) > 0 {
			logger.Error(gCtx, "failed to ingest some events", "errors", resp.Errors)
		}
		return nil
	})

	// Wait with timeout
	if err := g.Wait(); err != nil {
		logger.Error(ctx, "error sending telemetry", "err", err)
	}
}

// Helper functions

// getUserIDFromContext extracts user ID from context if available.
func getUserIDFromContext(_ context.Context) string {
	// This is a placeholder - implement actual user ID extraction
	// based on your authentication system
	return ""
}

// getEnvironment returns the current environment name.
func getEnvironment(cfg *config.Config) string {
	if cfg.Global.Env == "production" {
		return "production"
	}
	return "development"
}

// getLastUserMessage extracts the last user message from the messages array.
func getLastUserMessage(messages []Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == RoleUser {
			return messages[i].Content
		}
	}
	return ""
}

// extractModelParameters extracts model parameters from the request.
func extractModelParameters(params CompleteParams) map[string]any {
	modelParams := map[string]any{}

	if params.Temperature != nil {
		modelParams["temperature"] = *params.Temperature
	} else {
		modelParams["temperature"] = 0.7 // default
	}

	if params.MaxTokens != nil {
		modelParams["maxTokens"] = *params.MaxTokens
	} else {
		modelParams["maxTokens"] = 2000 // default
	}

	if params.TopP != nil {
		modelParams["topP"] = *params.TopP
	} else {
		modelParams["topP"] = 1.0 // default
	}

	return modelParams
}
