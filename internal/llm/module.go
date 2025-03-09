package llm

import (
	"coda/internal/config"

	"go.uber.org/fx"
)

// Module exports the LLM module for dependency injection.
var Module = fx.Module("llm",
	fx.Provide(
		// Provide the completer with default configuration
		func(cfg *config.Config) Completer {
			r := NewRegistry(cfg)
			return NewCompleter(cfg, r, WithCompleterRetryConfig(DefaultRetryConfig))
		},
	),
)
