package executor

import (
	"testing"

	"github.com/alpertosun/command-exporter/internal/config"
)

func TestTrimAndTruncate(t *testing.T) {
	input := "this string is over 50 characters long and should be truncated cleanly with no issues"
	expected := input[:50] + "..."
	got := trimAndTruncate(input)
	if got != expected {
		t.Errorf("Expected '%s', got '%s'", expected, got)
	}
}

func TestExecuteCommand_ValidOutput(t *testing.T) {
	cmd := config.Command{
		Name:     "echo_test",
		Command:  []string{"echo", "42"},
		Interval: "1s",
	}

	// should not panic or error
	execute(cmd)
}
