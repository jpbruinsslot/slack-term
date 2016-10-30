package config

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/gizak/termui"
)

// Config is the definition of a Config struct
type Config struct {
	SlackToken   string `json:"slack_token"`
	Theme        string `json:"theme"`
	SidebarWidth int    `json:"sidebar_width"`
	MainWidth    int    `json:"-"`
	EmacsMode    bool   `json:"emacs_mode"`
}

// NewConfig loads the config file and returns a Config struct
func NewConfig(filepath string) (*Config, error) {
	cfg := Config{
		Theme:        "dark",
		SidebarWidth: 1,
		MainWidth:    11,
	}

	file, err := os.Open(filepath)
	if err != nil {
		return &cfg, err
	}

	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return &cfg, err
	}

	if cfg.SlackToken == "" {
		return &cfg, errors.New("couldn't find 'slack_token' parameter")
	}

	if cfg.SidebarWidth < 1 || cfg.SidebarWidth > 11 {
		return &cfg, errors.New("please specify the 'sidebar_width' between 1 and 11")
	}

	cfg.MainWidth = 12 - cfg.SidebarWidth

	if cfg.Theme == "light" {
		termui.ColorMap = map[string]termui.Attribute{
			"fg":           termui.ColorBlack,
			"bg":           termui.ColorWhite,
			"border.fg":    termui.ColorBlack,
			"label.fg":     termui.ColorBlue,
			"par.fg":       termui.ColorYellow,
			"par.label.bg": termui.ColorWhite,
		}
	}

	return &cfg, nil
}
