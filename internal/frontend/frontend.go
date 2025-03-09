package frontend

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Frontend represents the web application that serves the user interface.
// It coordinates the different handlers and components of the web interface.
type Frontend struct {
	index *IndexHandler
}

// NewFrontend creates a new Frontend instance with the provided handlers.
// It follows the dependency injection pattern for better testability.
func newFrontend(index *IndexHandler) *Frontend {
	return &Frontend{
		index: index,
	}
}

// RegisterRoutes configures all routes for the frontend application.
// This method is called by the infrastructure layer to set up HTTP routes.
func (f *Frontend) RegisterRoutes(r chi.Router) {
	// Register static file routes
	r.Route("/static", func(r chi.Router) {
		r.Use(withCacheControl())
		r.Use(withPrefix("/assets"))
		r.Handle("/*", http.FileServer(http.FS(staticFS)))
	})

	// Register application routes
	r.Route("/", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			f.index.RegisterRoutes(r)
		})
	})
}
