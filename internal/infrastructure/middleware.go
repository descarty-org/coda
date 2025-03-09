package infrastructure

import (
	"coda/internal/logger"
	"net/http"
	"runtime/debug"
)

// withLogger is a middleware that logs the request details.
func withLogger(lg logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lg.Info("Request started", "method", r.Method, "path", r.URL.Path, "host", r.Host, "url", r.URL.String(),
				"remote", r.RemoteAddr, "user_agent", r.UserAgent(),
				"referer", r.Referer(), "proto", r.Proto, "content_length", r.ContentLength, "request_uri", r.RequestURI,
			)

			ctx := logger.WithLogger(r.Context(), lg)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// This function is adapted from the `recoverer` middleware from the `chi` package.
func withRecoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(rvr)
				}

				st := string(debug.Stack())
				logger.Error(r.Context(), "Panic occurred", "err", rvr, "st", st)

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
