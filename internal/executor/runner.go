package executor

import (
	"bytes"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/alpertosun/command-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
)

var commandMetric = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "command_result",
		Help: "Result of command execution",
	},
	[]string{"command", "exit_code", "stderr_summary", "stdout_summary"},
)

func init() {
	prometheus.MustRegister(commandMetric)
}

func ScheduleCommand(cmd config.Command, stdoutLimit int) {
	interval, err := time.ParseDuration(cmd.Interval)
	if err != nil {
		interval = 10 * time.Second
	}
	ticker := time.NewTicker(interval)
	for {
		execute(cmd)
		<-ticker.C
	}
}

func execute(cmd config.Command) {
	if len(cmd.Command) == 0 {
		slog.Error("Empty command", slog.String("name", cmd.Name))
		return
	}

	var stdout, stderr bytes.Buffer
	execCmd := exec.Command(cmd.Command[0], cmd.Command[1:]...)
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr
	err := execCmd.Run()

	valStr := strings.TrimSpace(stdout.String())
	valStr = strings.Split(valStr, " ")[0]

	val, parseErr := strconv.ParseFloat(valStr, 64)

	exitCode := "0"
	if err != nil || parseErr != nil {
		exitCode = "1"
		val = 1 // fallback

		errMsg := "Command execution failed"
		if parseErr != nil {
			errMsg += ": stdout is not a valid float"
		}

		slog.Error(errMsg,
			slog.String("command", cmd.Name),
			slog.String("stdout", trimAndTruncate(stdout.String())),
			slog.String("stderr", trimAndTruncate(stderr.String())),
			slog.Any("exec_error", err),
			slog.Any("parse_error", parseErr),
		)
	} else {
		slog.Info("Command executed successfully",
			slog.String("command", cmd.Name),
			slog.String("stdout", trimAndTruncate(valStr)),
		)
	}

	commandMetric.With(prometheus.Labels{
		"command":        cmd.Name,
		"exit_code":      exitCode,
		"stderr_summary": trimAndTruncate(stderr.String()),
		"stdout_summary": trimAndTruncate(stdout.String()),
	}).Set(val)
}

func trimAndTruncate(s string) string {
	s = strings.TrimSpace(s)

	replacer := strings.NewReplacer(
		"\r", "",
		"\n", "",
		"\t", "",
	)
	s = replacer.Replace(s)

	if len(s) > 50 {
		return s[:50] + "..."
	}
	return s
}
