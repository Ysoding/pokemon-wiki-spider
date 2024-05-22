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

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	go func() {
		logger.Sugar().Infow("startup")
		serverErrorSignal <- engine.Run(ctx, logger)
	}()

	// shutdown
	select {
	case err := <-serverErrorSignal:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}

	case sig := <-shutdown:
		logger.Sugar().Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer logger.Sugar().Infow("shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		<-ctx.Done()
	}
	return nil
}
