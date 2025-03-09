package frontend

import (
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
)

// ConfigureRoutes sets up all HTTP routes for the frontend application.
// This function is called by the infrastructure layer during server initialization.
func ConfigureRoutes(f *Frontend, r *chi.Mux) {
	// Use the Frontend's RegisterRoutes method to configure all routes
	f.RegisterRoutes(r)
}

// withCacheControl adds Cache-Control header to the response.
func withCacheControl() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			next.ServeHTTP(w, r)
		})
	}
}

// withPrefix adds a prefix to the request path.
func withPrefix(prefix string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = path.Join(prefix, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}
