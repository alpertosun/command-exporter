package config

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoad_ValidConfig(t *testing.T) {
	yml := `
commands:
  - name: test_command
    command: ["echo", "123"]
    interval: "5s"
    stdout_limit: 50
`

	tmpFile := "test_config.yaml"
	err := os.WriteFile(tmpFile, []byte(yml), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg.Commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(cfg.Commands))
	}
	if cfg.Commands[0].Name != "test_command" {
		t.Errorf("Unexpected command name: %s", cfg.Commands[0].Name)
	}
}

func TestUnmarshal_InvalidInterval(t *testing.T) {
	raw := `
commands:
  - name: bad_interval
    command: ["ls"]
    interval: "notaduration"
`

	var cfg Config
	err := yaml.Unmarshal([]byte(raw), &cfg)
	if err != nil {
		t.Errorf("Expected no YAML error, got %v", err)
	}

	if cfg.Commands[0].Interval != "notaduration" {
		t.Errorf("Expected interval to stay as string, got %s", cfg.Commands[0].Interval)
	}
}
