package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ysoding/pokemon-wiki-spider/collect"
	"github.com/Ysoding/pokemon-wiki-spider/engine"
	"github.com/Ysoding/pokemon-wiki-spider/parse/pokemon"
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

func run(context.Context) error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serverErrorSignal := make(chan error, 1)

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	seeds := pokemon.Tasks
	e := engine.NewEngine(engine.WithLogger(logger),
		engine.WithScheduler(engine.NewSchedule()),
		engine.WithSeeds(seeds),
		engine.WithFetcher(collect.BrowserFetch{
			Timeout: 5 * time.Second,
			Logger:  logger,
		}))

	go func() {
		logger.Sugar().Infow("startup")
		serverErrorSignal <- e.Run()
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
		e.Shutdown()
	}
	return nil
}
