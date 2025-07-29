package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alpertosun/command-exporter/internal/collector"
	"github.com/alpertosun/command-exporter/internal/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("Failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := collector.StartExporter(cfg, cfg.StdoutLimit)

	slog.Info("Exporter started, waiting for shutdown signal...")

	<-ctx.Done()
	slog.Info("Shutdown signal received, stopping exporter...")

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxTimeout); err != nil {
		slog.Error("Error shutting down HTTP server", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("Exporter shutdown cleanly.")
}
