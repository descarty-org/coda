package infrastructure

import (
	"coda/internal/frontend"
	"coda/internal/logger"

	"go.uber.org/fx"
)

var Module = fx.Module("infrastructure",
	fx.Provide(NewServer),
	frontend.Module,
	logger.Module,
)
