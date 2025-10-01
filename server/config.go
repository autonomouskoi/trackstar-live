package server

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	MyURL      string `yaml:"my_url"`
	MyKeyInput string `yaml:"my_key"`
	Listen     string `yaml:"listen"`
	LogDebug   bool   `yaml:"log_debug"`
	DBPath     string `yaml:"db_path"`
	HTMLPath   string `yaml:"html_path"`
}

func (c *ServerConfig) Validate() error {
	return nil
}

func LoadConfig(path string) (*ServerConfig, error) {
	var cfg ServerConfig
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}
	return &cfg, nil
}
