package main

import (
	"coda/internal/config"
	"coda/internal/infrastructure"
	"coda/internal/llm"
	"coda/internal/review"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/fx"

	// Supported LLM providers
	_ "coda/internal/llm/ollama"
	_ "coda/internal/llm/openai"
)

var cfg *config.Config

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("loading .env file: %w", err)
	}

	var err error
	cfg, err = config.Load(config.ENV(os.Getenv("ENV")), os.Getenv("CONFIG_DIR"))
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverApp(ctx).Run()
	return nil
}

func serverApp(ctx context.Context) *fx.App {
	var opts []fx.Option
	opts = append(opts, infrastructure.Module)
	opts = append(opts, llm.Module)
	opts = append(opts, review.Module)
	opts = append(opts, fx.Supply(cfg))
	opts = append(opts, fx.Invoke(infrastructure.ServerLifetimeHooks))
	if cfg.Global.Env != config.ENVLocal {
		opts = append(opts, fx.NopLogger)
	}
	opts = append(opts, fx.Supply(ctx))
	return fx.New(opts...)
}
