package config

import (
	"encoding/json"
	"os"
)

// Config is the definition of a Config struct
type Config struct {
	SlackToken string `json:"slack_token"`
}

// NewConfig loads the config file and returns a Config struct
func NewConfig(filepath string) (*Config, error) {
	var cfg Config

	file, err := os.Open(filepath)
	if err != nil {
		return &cfg, err
	}

	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return &cfg, err
	}

	return &cfg, nil
}
