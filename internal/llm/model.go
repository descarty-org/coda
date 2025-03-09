package llm

import (
	"coda/internal/config"
	"time"
)

// Config contains configuration for an LLM client.
type Config struct {
	APIKeyFunc APIKeyFunc
	Model      Model
	Timeout    time.Duration
	LLMConfig  config.LLM
}

// Model represents a language model.
type Model struct {
	Provider      Provider
	Name          string
	DisplayName   string
	MaxToken      int
	ContextWindow int
	PDFSupported  bool
	Version       string
	Family        string
	Pricing       *ModelPricing
	Capabilities  ModelCapabilities
}

// ModelPricing contains pricing information for a model.
type ModelPricing struct {
	InputPerToken  float64
	OutputPerToken float64
	Currency       string
}

// Provider represents an LLM provider.
type Provider string

// Supported providers
const (
	OpenAI Provider = "openai"
	Ollama Provider = "ollama"
)

// String returns the string representation of the provider.
func (p Provider) String() string {
	return string(p)
}
