package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Command struct {
	Name     string   `yaml:"name"`
	Command  []string `yaml:"command"`
	Interval string   `yaml:"interval"`
}

type Config struct {
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
	return &cfg, nil
}
