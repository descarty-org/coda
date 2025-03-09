package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-playground/validator"
	"gopkg.in/yaml.v2"
)

// Load loads the configuration from the given directory and environment variables.
// It follows this process:
// 1. Find appropriate config files based on the environment
// 2. Load and parse YAML from these files
// 3. Override with environment variables
// 4. Validate the final configuration
//
// Parameters:
//   - env: The environment to load configuration for (local, development, production)
//   - configDir: Directory containing configuration files
//
// Returns:
//   - A fully loaded and validated configuration
//   - An error if loading or validation fails
func Load(env ENV, configDir string) (*Config, error) {
	// Create options with defaults
	opts := Options{
		Env:       env,
		ConfigDir: configDir,
	}
	opts.setDefaults()

	// Find config files for this environment
	files, err := findConfigFiles(opts)
	if err != nil {
		return nil, fmt.Errorf("get config files: %w", err)
	}

	// Initialize config with environment
	cfg := &Config{
		Global: Global{Env: opts.Env},
	}

	// Load and parse each config file
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("config file is not valid yaml: %w", err)
		}
	}

	// Override with environment variables
	if err := LoadEnvVariables(cfg); err != nil {
		return nil, fmt.Errorf("load envs into config: %w", err)
	}

	// Validate the final configuration
	if err := validator.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("config is not valid: %w", err)
	}

	return cfg, nil
}

// findConfigFiles returns the list of config files for the given environment.
// It searches for files matching the environment name in the config directory.
func findConfigFiles(opts Options) ([]string, error) {
	var absCandidates []string
	var result []string

	// Check each candidate file
	for _, c := range opts.candidates() {
		abs, err := filepath.Abs(filepath.Join(opts.ConfigDir, c))
		if err != nil {
			continue
		}
		absCandidates = append(absCandidates, abs)

		// Add file to results if it exists
		if _, err := os.Stat(abs); err == nil {
			result = append(result, abs)
		}
	}

	// Ensure at least one config file was found
	if len(result) == 0 {
		return nil, fmt.Errorf("no config files found in %s. candidates: %v", opts.ConfigDir, absCandidates)
	}

	return result, nil
}

// LoadEnvVariables loads environment variables into the configuration.
// This allows overriding config file values with environment variables.
// Returns an error if any environment variable has an invalid format.
func LoadEnvVariables(cfg *Config) error {
	// Server configuration
	if v, ok := os.LookupEnv("PORT"); ok {
		port, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
		cfg.Server.Port = port
	}
	if v, ok := os.LookupEnv("HOST"); ok {
		cfg.Server.Host = v
	}
	if v, ok := os.LookupEnv("ALLOWED_ORIGINS"); ok {
		cfg.Server.AllowedOrigins = strings.Split(v, ",")
	}

	// LLM configuration
	if v, ok := os.LookupEnv("OPENAI_API_KEY"); ok {
		cfg.LLM.OpenAI.APIKey = v
	}
	if v, ok := os.LookupEnv("OLLAMA_BASE_URL"); ok {
		cfg.LLM.Ollama.BaseURL = v
	}
	if v, ok := os.LookupEnv("LANGFUSE_PUBLIC_KEY"); ok {
		cfg.LLM.Langfuse.PublicKey = v
	}
	if v, ok := os.LookupEnv("LANGFUSE_PRIVATE_KEY"); ok {
		cfg.LLM.Langfuse.PrivateKey = v
	}

	return nil
}

// Options defines parameters for loading configuration.
type Options struct {
	Env       ENV    // Environment to load configuration for
	ConfigDir string // Directory containing configuration files
}

// setDefaults sets default values for configuration options.
func (o *Options) setDefaults() {
	if o.Env == "" {
		o.Env = ENVLocal
	}
	if o.ConfigDir == "" {
		o.ConfigDir = "config"
	}
}

// candidates returns the list of candidate config files for the environment.
// This determines which files will be searched for in the config directory.
func (o *Options) candidates() []string {
	return []string{
		strings.ToLower(string(o.Env)) + ".yaml",
	}
}
