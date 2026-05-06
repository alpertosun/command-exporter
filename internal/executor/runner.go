package executor

import (
	"bytes"
	"context"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/alpertosun/command-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	commandResult = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "command_result",
			Help: "Parsed numeric result of command stdout.",
		},
		[]string{"command"},
	)

	commandSuccess = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "command_success",
			Help: "1 if the last execution succeeded and produced a valid float, 0 otherwise.",
		},
		[]string{"command"},
	)

	commandExitCode = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "command_exit_code",
			Help: "Exit code of the last execution.",
		},
		[]string{"command"},
	)

	commandDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "command_duration_seconds",
			Help: "Duration of the last execution in seconds.",
		},
		[]string{"command"},
	)
)

func init() {
	prometheus.MustRegister(commandResult, commandSuccess, commandExitCode, commandDuration)
}

func ScheduleCommand(ctx context.Context, cmd config.Command, stdoutLimit int) {
	interval, err := time.ParseDuration(cmd.Interval)
	if err != nil {
		slog.Warn("Invalid interval, falling back to default",
			slog.String("command", cmd.Name),
			slog.String("interval", cmd.Interval),
		)
		interval = 10 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	execute(ctx, cmd, stdoutLimit)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			execute(ctx, cmd, stdoutLimit)
		}
	}
}

func execute(ctx context.Context, cmd config.Command, stdoutLimit int) {
	if len(cmd.Command) == 0 {
		slog.Error("Empty command", slog.String("name", cmd.Name))
		commandSuccess.WithLabelValues(cmd.Name).Set(0)
		return
	}

	var stdout, stderr bytes.Buffer
	execCmd := exec.CommandContext(ctx, cmd.Command[0], cmd.Command[1:]...)
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	start := time.Now()
	runErr := execCmd.Run()
	duration := time.Since(start).Seconds()

	exitCode := 0
	if runErr != nil {
		exitCode = 1
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	raw := strings.TrimSpace(stdout.String())
	fields := strings.Fields(raw)
	var (
		val      float64
		parseErr error
	)
	if len(fields) == 0 {
		parseErr = strconv.ErrSyntax
	} else {
		val, parseErr = strconv.ParseFloat(fields[0], 64)
	}

	success := runErr == nil && parseErr == nil

	commandExitCode.WithLabelValues(cmd.Name).Set(float64(exitCode))
	commandDuration.WithLabelValues(cmd.Name).Set(duration)

	if success {
		commandResult.WithLabelValues(cmd.Name).Set(val)
		commandSuccess.WithLabelValues(cmd.Name).Set(1)
		slog.Info("Command executed",
			slog.String("command", cmd.Name),
			slog.Float64("value", val),
			slog.Float64("duration_s", duration),
		)
		return
	}

	commandSuccess.WithLabelValues(cmd.Name).Set(0)
	slog.Error("Command failed",
		slog.String("command", cmd.Name),
		slog.Int("exit_code", exitCode),
		slog.String("stdout", trimAndTruncate(stdout.String(), stdoutLimit)),
		slog.String("stderr", trimAndTruncate(stderr.String(), stdoutLimit)),
		slog.Any("exec_error", runErr),
		slog.Any("parse_error", parseErr),
	)
}

func trimAndTruncate(s string, limit int) string {
	s = strings.TrimSpace(s)
	replacer := strings.NewReplacer("\r", "", "\n", " ", "\t", " ")
	s = replacer.Replace(s)
	if limit <= 0 {
		limit = 50
	}
	if len(s) > limit {
		return s[:limit] + "..."
	}
	return s
}
