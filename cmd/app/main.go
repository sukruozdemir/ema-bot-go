package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sukruozdemir/ema-bot-go/internal/app"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger (NewNop() disables all logging)
	logger := zap.NewNop()
	defer func() {
		if err := logger.Sync(); err != nil {
			// Handle sync error if needed
		}
	}()

	// Create context that cancels on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("Shutdown signal received")
		cancel()
	}()

	// Run the application
	if err := app.Run(ctx, logger); err != nil {
		logger.Fatal("Application failed", zap.Error(err))
	}
}
