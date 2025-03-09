// Package config provides configuration management for the application.
// It handles loading configuration from files and environment variables,
package config

// Config represents the complete application configuration.
// It contains all settings needed for the application to run.
type Config struct {
	Global  Global  `yaml:"global"`  // Global application settings
	Logging Logging `yaml:"logging"` // Logging configuration
	Server  Server  `yaml:"server"`  // HTTP server configuration
	LLM     LLM     `yaml:"llm"`     // Language model configuration
}

// Global contains application-wide settings.
// These settings apply across all components of the application.
type Global struct {
	Env ENV `yaml:"-"` // Environment type (local, development, production)
}

// ENV represents the application environment.
// It determines behavior like logging verbosity, feature flags, etc.
type ENV string

// Environment constants define the possible application environments.
const (
	ENVTest        ENV = "test"        // Test environment for automated testing
	ENVLocal       ENV = "local"       // Local development environment
	ENVDevelopment ENV = "development" // Shared development environment
	ENVProduction  ENV = "production"  // Production environment
)

// Logging configures the application's logging behavior.
type Logging struct {
	Format string `yaml:"format"` // Log format (json, text)
}

// Server configures the HTTP server.
type Server struct {
	Host           string   `yaml:"host" validate:"required"` // Server hostname or IP
	Port           int      `yaml:"port" validate:"required"` // Server port
	AllowedOrigins []string `yaml:"allowedOrigins"`           // CORS allowed origins
}

// LLM configures language model services.
type LLM struct {
	OpenAI   OpenAI   `yaml:"openai" validate:"required"`   // OpenAI API configuration
	Ollama   Ollama   `yaml:"ollama" validate:"required"`   // Ollama API configuration
	Langfuse Langfuse `yaml:"langfuse" validate:"required"` // Langfuse observability configuration
}

// OpenAI configures the OpenAI API client.
type OpenAI struct {
	APIKey string `yaml:"apiKey" validate:"required"` // OpenAI API key
}

// Langfuse configures the Langfuse observability platform.
type Langfuse struct {
	PrivateKey string `yaml:"privateKey"` // Langfuse private key
	PublicKey  string `yaml:"publicKey"`  // Langfuse public key
}

// IsConfigured checks if the Langfuse configuration is complete.
func (l *Langfuse) IsConfigured() bool {
	return l.PrivateKey != "" && l.PublicKey != ""
}

// Ollama configures the Ollama API client.
type Ollama struct {
	BaseURL string `yaml:"baseURL"` // Ollama API base URL
}

func (o *Ollama) IsConfigured() bool {
	return o.BaseURL != ""
}
