package frontend

import (
	"coda/internal/config"
	"context"

	"go.uber.org/fx"
)

// Module is the frontend fx module that provides all frontend components.
// It registers the frontend handlers, template manager, and lifecycle hooks.
var Module = fx.Module("frontend",
	fx.Provide(newFrontend),          // Provides the main Frontend instance
	fx.Provide(newTemplateManager),   // Provides the template manager
	fx.Provide(newIndex),             // Provides the index page handler
	fx.Invoke(registerLifetimeHooks), // Registers lifecycle hooks
)

// registerLifetimeHooks sets up the lifecycle hooks for the frontend components.
// It handles template hot-reloading in development and proper cleanup on shutdown.
func registerLifetimeHooks(lc fx.Lifecycle, cfg *config.Config, tm *TemplateManager) {
	lc.Append(fx.Hook{
		// OnStart sets up template file watching for hot-reloading in development
		OnStart: func(_ context.Context) error {
			return tm.watchFiles(cfg)
		},
		// OnStop ensures proper cleanup of resources
		OnStop: func(_ context.Context) error {
			tm.Close()
			return nil
		},
	})
}
