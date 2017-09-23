package config

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/erroneousboat/termui"
)

// Config is the definition of a Config struct
type Config struct {
	SlackToken   string                `json:"slack_token"`
	Theme        string                `json:"theme"`
	SidebarWidth int                   `json:"sidebar_width"`
	MainWidth    int                   `json:"-"`
	KeyMap       map[string]keyMapping `json:"key_map"`
}

type keyMapping map[string]string

// NewConfig loads the config file and returns a Config struct
func NewConfig(filepath string) (*Config, error) {
	cfg := Config{
		Theme:        "dark",
		SidebarWidth: 1,
		MainWidth:    11,
		KeyMap: map[string]keyMapping{
			"command": {
				"i":          "mode-insert",
				"/":          "mode-search",
				"k":          "channel-up",
				"j":          "channel-down",
				"g":          "channel-top",
				"G":          "channel-bottom",
				"<previous>": "chat-up",
				"C-b":        "chat-up",
				"C-u":        "chat-up",
				"<next>":     "chat-down",
				"C-f":        "chat-down",
				"C-d":        "chat-down",
				"q":          "quit",
				"<f1>":       "help",
			},
			"insert": {
				"<left>":      "cursor-left",
				"<right>":     "cursor-right",
				"<enter>":     "send",
				"<escape>":    "mode-command",
				"<backspace>": "backspace",
				"C-8":         "backspace",
				"<delete>":    "delete",
				"<space>":     "space",
			},
			"search": {
				"<left>":      "cursor-left",
				"<right>":     "cursor-right",
				"<escape>":    "clear-input",
				"<enter>":     "clear-input",
				"<backspace>": "backspace",
				"C-8":         "backspace",
				"<delete>":    "delete",
				"<space>":     "space",
			},
		},
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
