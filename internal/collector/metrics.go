package collector

import (
	"log/slog"
	"net/http"

	"github.com/alpertosun/command-exporter/internal/config"
	"github.com/alpertosun/command-exporter/internal/executor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartExporter(cfg *config.Config, stdoutLimit int) *http.Server {
	for _, c := range cfg.Commands {
		go executor.ScheduleCommand(c, stdoutLimit)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	addr := ":9860" // default port
	slog.Info("Listening", slog.String("addr", addr))

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", slog.Any("error", err))
		}
	}()

	return server
}
