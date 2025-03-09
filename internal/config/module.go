package config

import (
	"os"

	"go.uber.org/fx"
)

// Module is the config fx module that provides the application configuration.
// It loads configuration from files and environment variables.
var Module = fx.Module("config",
	fx.Provide(NewConfig),
)

// NewConfig loads the application configuration from the environment.
// It uses the CODA_ENV environment variable to determine which config file to load.
// If CODA_ENV is not set, it defaults to "local".
func NewConfig() (*Config, error) {
	// Determine environment from CODA_ENV or default to local
	env := ENV(getEnvOrDefault("CODA_ENV", string(ENVLocal)))

	// Load configuration from files and environment variables
	return Load(env, "config")
}

// getEnvOrDefault returns the value of the environment variable or the default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := getenv(key); exists {
		return value
	}
	return defaultValue
}

// getenv is a wrapper around os.LookupEnv for testing purposes.
var getenv = func(key string) (string, bool) {
	return os.LookupEnv(key)
}
