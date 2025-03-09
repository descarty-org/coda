package llm

import (
	"errors"
	"fmt"
)

// Common error types for LLM operations
var (
	// Input-related errors
	ErrTokenLimitReached     = errors.New("token limit reached")
	ErrContextLengthExceeded = errors.New("context length exceeded")
	ErrInvalidArguments      = errors.New("invalid arguments")
	ErrNoMessages            = errors.New("no messages returned")

	// Authentication errors
	ErrInvalidAPIKey        = errors.New("invalid API key")
	ErrAuthenticationFailed = errors.New("authentication failed")

	// Service errors
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrTooManyRequests    = errors.New("too many requests")
	ErrTimeout            = errors.New("request timed out")
	ErrRateLimited        = errors.New("rate limited")

	// Content errors
	ErrContentFiltered   = errors.New("content filtered by safety system")
	ErrContentNotAllowed = errors.New("content not allowed")

	// Model errors
	ErrModelNotFound   = errors.New("model not found")
	ErrModelOverloaded = errors.New("model is currently overloaded")

	// Function calling errors
	ErrInvalidFunctionCall = errors.New("invalid function call")
	ErrFunctionNotFound    = errors.New("function not found")

	// Internal errors
	ErrInternalError = errors.New("internal error")
)

// LLMError represents a structured error from an LLM operation.
type LLMError struct {
	// The underlying error
	Err error

	// Provider is the LLM provider that generated the error
	Provider string

	// Model is the model that generated the error
	Model string

	// StatusCode is the HTTP status code (if applicable)
	StatusCode int

	// ErrorCode is the provider-specific error code
	ErrorCode string

	// Error message is detailed error information
	ErrorMessage string

	// RequestID is the provider-specific request ID
	RequestID string

	// Retryable indicates if the error is retryable
	Retryable bool
}

// Error implements the error interface.
func (e *LLMError) Error() string {
	if e.Err == nil {
		return "unknown LLM error"
	}

	base := e.Err.Error()
	if e.Provider != "" {
		base = fmt.Sprintf("[%s] %s", e.Provider, base)
	}

	if e.ErrorCode != "" {
		base = fmt.Sprintf("%s (code: %s)", base, e.ErrorCode)
	}

	return base
}

// Unwrap returns the underlying error.
func (e *LLMError) Unwrap() error {
	return e.Err
}

// NewLLMError creates a new LLMError.
func NewLLMError(err error, provider, model string) *LLMError {
	return &LLMError{
		Err:      err,
		Provider: provider,
		Model:    model,
	}
}

// WithStatusCode adds a status code to the error.
func (e *LLMError) WithStatusCode(code int) *LLMError {
	e.StatusCode = code
	return e
}

// WithErrorCode adds an error code to the error.
func (e *LLMError) WithErrorCode(code string) *LLMError {
	e.ErrorCode = code
	return e
}

// WithErrorMessage adds an error message to the error.
func (e *LLMError) WithErrorMessage(msg string) *LLMError {
	e.ErrorMessage = msg
	return e
}

// WithRequestID adds a request ID to the error.
func (e *LLMError) WithRequestID(id string) *LLMError {
	e.RequestID = id
	return e
}

// WithRetryable sets whether the error is retryable.
func (e *LLMError) WithRetryable(retryable bool) *LLMError {
	e.Retryable = retryable
	return e
}

// IsRetryable returns true if the error is retryable.
func IsRetryable(err error) bool {
	var llmErr *LLMError
	if errors.As(err, &llmErr) {
		return llmErr.Retryable
	}

	// Default retryable errors
	return errors.Is(err, ErrServiceUnavailable) ||
		errors.Is(err, ErrTooManyRequests) ||
		errors.Is(err, ErrTimeout) ||
		errors.Is(err, ErrRateLimited) ||
		errors.Is(err, ErrModelOverloaded)
}
