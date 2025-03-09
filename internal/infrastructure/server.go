package infrastructure

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"coda/internal/config"
	"coda/internal/frontend"
	"coda/internal/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	"go.uber.org/fx"
)

type ServerConfig struct {
	ShutdownTimeout time.Duration
}

type Server struct {
	config     ServerConfig
	appConfig  *config.Config
	httpServer *http.Server
	logger     logger.Logger
	frontend   *frontend.Frontend
}

func NewServer(
	logger logger.Logger,
	config *config.Config,
	frontend *frontend.Frontend,
) *Server {
	serverCfg := ServerConfig{
		ShutdownTimeout: 5 * time.Second,
	}
	return &Server{
		config:    serverCfg,
		appConfig: config,
		logger:    logger,
		frontend:  frontend,
	}
}

func (srv *Server) Serve(ctx context.Context) error {
	requestLogger := httplog.NewLogger("http", httplog.Options{
		LogLevel:         slog.LevelDebug,
		JSON:             true,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "msg",
		ResponseHeaders:  true,
	})

	// create a type that satisfies the `api.ServerInterface`, which contains
	// an implementation of every operation from the generated code
	r := chi.NewMux()
	r.Use(middleware.RealIP)
	r.Use(middleware.Compress(5))
	r.Use(httplog.RequestLogger(requestLogger))
	r.Use(withLogger(srv.logger))
	r.Use(withRecoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   srv.appConfig.Server.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}))

	frontend.ConfigureRoutes(srv.frontend, r)

	addr := net.JoinHostPort(srv.appConfig.Server.Host, strconv.Itoa(srv.appConfig.Server.Port))
	srv.httpServer = &http.Server{
		Handler:           r,
		Addr:              addr,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		srv.logger.Info("Server is starting", "addr", addr)
		if err := srv.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			srv.logger.Error("ListenAndServe()", "err", err)
		}
	}()

	srv.gracefulShutdown(ctx)
	return nil
}

func (srv *Server) Shutdown(ctx context.Context) error {
	if srv.httpServer != nil {
		srv.logger.Info("Server is shutting down")
		return srv.httpServer.Shutdown(ctx)
	}
	return nil
}

func (srv *Server) gracefulShutdown(ctx context.Context) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	srv.logger.Info("Server is shutting down")

	ctx, cancel := context.WithTimeout(ctx, srv.config.ShutdownTimeout)
	defer cancel()

	srv.httpServer.SetKeepAlivesEnabled(false)
	if err := srv.httpServer.Shutdown(ctx); err != nil {
		srv.logger.Error("Could not gracefully shutdown the server", "err", err)
	}

	srv.logger.Info("Server stopped")
}

func ServerLifetimeHooks(lc fx.Lifecycle, srv *Server) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			// Lifecycle hooks must not block. Use a goroutine for long-running tasks.
			// See: https://uber-go.github.io/fx/lifecycle.html
			go func() {
				// Create a new context because the given one has a timeout
				// and we want to keep the server running until it's stopped.
				err := srv.Serve(context.Background())
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					srv.logger.Fatalf("Failed to start server: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
