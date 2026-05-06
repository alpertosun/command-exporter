package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Command struct {
	Name     string   `yaml:"name"`
	Command  []string `yaml:"command"`
	Interval string   `yaml:"interval"`
}

type Config struct {
	ListenAddr  string    `yaml:"listen_addr"`
	StdoutLimit int       `yaml:"stdout_limit"`
	Commands    []Command `yaml:"commands"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Commands) == 0 {
		return fmt.Errorf("no commands defined")
	}
	seen := make(map[string]struct{}, len(c.Commands))
	for i, cmd := range c.Commands {
		if cmd.Name == "" {
			return fmt.Errorf("commands[%d]: name is required", i)
		}
		if _, dup := seen[cmd.Name]; dup {
			return fmt.Errorf("commands[%d]: duplicate name %q", i, cmd.Name)
		}
		seen[cmd.Name] = struct{}{}

		if len(cmd.Command) == 0 {
			return fmt.Errorf("commands[%d] (%s): command is required", i, cmd.Name)
		}
		if cmd.Interval == "" {
			return fmt.Errorf("commands[%d] (%s): interval is required", i, cmd.Name)
		}
		d, err := time.ParseDuration(cmd.Interval)
		if err != nil {
			return fmt.Errorf("commands[%d] (%s): invalid interval %q: %w", i, cmd.Name, cmd.Interval, err)
		}
		if d <= 0 {
			return fmt.Errorf("commands[%d] (%s): interval must be positive", i, cmd.Name)
		}
	}
	return nil
}
