package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ysoding/pokemon-wiki-spider/engine"
	"go.uber.org/zap"
)

const (
	PokemonListURL = "https://wiki.52poke.com/wiki/%E5%AE%9D%E5%8F%AF%E6%A2%A6%E5%88%97%E8%A1%A8%EF%BC%88%E6%8C%89%E5%85%A8%E5%9B%BD%E5%9B%BE%E9%89%B4%E7%BC%96%E5%8F%B7%EF%BC%89"
)

func main() {
	// logger
	zap.ReplaceGlobals(zap.Must(zap.NewProduction()))

	ctx := context.Background()

	if err := run(ctx); err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serverErrorSignal := make(chan error, 1)

	go func() {
		zap.L().Sugar().Infow("startup")
		serverErrorSignal <- engine.Run(ctx)
	}()

	// shutdown
	select {
	case err := <-serverErrorSignal:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}

	case sig := <-shutdown:
		zap.L().Sugar().Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer zap.L().Sugar().Infow("shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		<-ctx.Done()
	}
	return nil
}
