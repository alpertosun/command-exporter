package collector

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/alpertosun/command-exporter/internal/config"
	"github.com/alpertosun/command-exporter/internal/executor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartExporter(ctx context.Context, cfg *config.Config) *http.Server {
	for _, c := range cfg.Commands {
		go executor.ScheduleCommand(ctx, c, cfg.StdoutLimit)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	addr := cfg.ListenAddr
	if addr == "" {
		addr = ":9860"
	}

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	slog.Info("Listening", slog.String("addr", addr))

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", slog.Any("error", err))
		}
	}()

	return server
}
