package llm

import (
	"coda/internal/config"
	"context"
	"fmt"
	"sync"
	"time"
)

// LLM is an interface that represents a large language model.
type LLM interface {
	// Complete processes the given parameters and returns a completion response.
	Complete(ctx context.Context, params CompleteParams) (*CompleteResponse, error)
}

// ModelInfo provides metadata about a language model.
type ModelInfo struct {
	Model        Model
	Capabilities ModelCapabilities
}

// ModelCapabilities describes what features a model supports.
type ModelCapabilities struct {
	SupportsStreaming bool
	SupportsFunctions bool
	SupportsVision    bool
	SupportsJSON      bool
}

// CompleteParams contains parameters for the Complete method.
type CompleteParams struct {
	Messages    []Message
	MaxTokens   *int
	Temperature *float32
	TopP        *float32
	N           *int
	Stream      bool
	Functions   []FunctionDefinition `json:"functions,omitempty"`
	JSONMode    bool                 `json:"json_mode,omitempty"`
}

// FunctionDefinition defines a function that can be called by the model.
type FunctionDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters"`
}

// CompleteResponse contains the response from a completion request.
type CompleteResponse struct {
	Messages []Message
	Usage    *Usage
	// Metadata about the completion
	Metadata CompletionMetadata
}

// CompletionMetadata contains additional information about a completion.
type CompletionMetadata struct {
	ModelName     string
	FinishReason  string
	CompletionID  string
	LatencyMs     int64
	ProcessedAt   time.Time
	RequestTokens int
}

// Usage contains token usage information.
type Usage struct {
	Unit             string `json:"unit,omitempty"`
	PromptTokens     int    `json:"promptTokens,omitempty"`
	CompletionTokens int    `json:"completionTokens,omitempty"`
	TotalTokens      int    `json:"totalTokens,omitempty"`
}

// APIKeyFunc is a function that returns an API key.
type APIKeyFunc func() string

// Registry of supported models
var (
	modelRegistryMu sync.RWMutex
	supportedModels = map[Model]SupportedModels{}
)

// SupportedModels contains information about a supported model.
type SupportedModels struct {
	Constructor func(Config) (LLM, error)
	Model       Model
}

// Constructor is a function that creates a new LLM instance.
type Constructor func(cfg Config) (LLM, error)

// Registry contains a list of available models.
type Registry struct {
	models []Model
}

// NewRegistry initializes a new model registry with the given configuration.
func NewRegistry(cfg *config.Config) *Registry {
	modelRegistryMu.RLock()
	defer modelRegistryMu.RUnlock()

	models := make([]Model, 0, len(supportedModels))
	for model := range supportedModels {
		switch model.Provider {
		case OpenAI:
			models = append(models, model)
		case Ollama:
			if !cfg.LLM.Ollama.IsConfigured() {
				continue
			}
			models = append(models, model)
		}
	}

	return &Registry{
		models: models,
	}
}

// RegisterLLM registers models with their constructor function.
func RegisterLLM(constructor Constructor, models []Model) {
	modelRegistryMu.Lock()
	defer modelRegistryMu.Unlock()

	for _, model := range models {
		supportedModels[model] = SupportedModels{
			Constructor: constructor,
			Model:       model,
		}
	}
}

// New creates a new LLM instance for the specified model.
func New(cfg Config) (LLM, error) {
	if cfg.APIKeyFunc == nil {
		return nil, fmt.Errorf("API key function is required")
	}

	modelRegistryMu.RLock()
	supportedModel, ok := supportedModels[cfg.Model]
	modelRegistryMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("model %q is not supported", cfg.Model.Name)
	}

	return supportedModel.Constructor(cfg)
}
