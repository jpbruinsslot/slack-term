package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	SlackToken string `json:"slack_token"`
	User       string `json:"user"`
}

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
