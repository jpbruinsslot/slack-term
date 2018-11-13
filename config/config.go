package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/erroneousboat/termui"
)

const (
	NotifyAll     = "all"
	NotifyMention = "mention"
)

// Config is the definition of a Config struct
type Config struct {
	SlackToken   string                `json:"slack_token"`
	Workspaces   map[string]Config     `json:"workspaces"`
	Notify       string                `json:"notify"`
	Emoji        bool                  `json:"emoji"`
	SidebarWidth int                   `json:"sidebar_width"`
	MainWidth    int                   `json:"-"`
	KeyMap       map[string]keyMapping `json:"key_map"`
	Theme        Theme                 `json:"theme"`
}

type keyMapping map[string]string

// NewConfig loads the config file and returns a Config struct
func NewConfig(filepath string, workspaceName string) (*Config, error) {
	cfg := getDefaultConfig()

	file, err := os.Open(filepath)
	if err != nil {
		return &cfg, fmt.Errorf("couldn't find the slack-term config file: %v", err)
	}

	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return &cfg, fmt.Errorf("the slack-term config file isn't valid json: %v", err)
	}

	// If no workspace is specified, select the first (ABC-order).
	if workspaceName == "" {
		keys := make([]string, len(cfg.Workspaces))
		i := 0
		for k := range cfg.Workspaces {
			keys[i] = k
			i++
		}
		sort.Strings(keys)
		workspaceName = keys[0]
	}
	workspaceConfig := cfg.Workspaces[workspaceName]

	// Overwrite all options if they're specified in the
	// workspace-specific config:
	cfg.SlackToken = workspaceConfig.SlackToken

	if workspaceConfig.Notify != "" {
		cfg.Notify = workspaceConfig.Notify
	}

	// TODO: There's no way to distinguish between falsy and unset.
	// cfg.Emoji = workspaceConfig.Emoji

	if workspaceConfig.SidebarWidth != 0 {
		cfg.SidebarWidth = workspaceConfig.SidebarWidth
	}
	if workspaceConfig.MainWidth != 0 {
		cfg.MainWidth = workspaceConfig.MainWidth
	}
	if workspaceConfig.KeyMap != nil {
		cfg.KeyMap = workspaceConfig.KeyMap
	}
	if workspaceConfig.Theme != *new(Theme) {
		cfg.Theme = workspaceConfig.Theme
	}

	if cfg.SidebarWidth < 1 || cfg.SidebarWidth > 11 {
		return &cfg, errors.New("please specify the 'sidebar_width' between 1 and 11")
	}

	cfg.MainWidth = 12 - cfg.SidebarWidth

	switch cfg.Notify {
	case NotifyAll, NotifyMention, "":
		break
	default:
		return &cfg, fmt.Errorf("unsupported setting for notify: %s", cfg.Notify)
	}

	termui.ColorMap = map[string]termui.Attribute{
		"fg":        termui.StringToAttribute(cfg.Theme.View.Fg),
		"bg":        termui.StringToAttribute(cfg.Theme.View.Bg),
		"border.fg": termui.StringToAttribute(cfg.Theme.View.BorderFg),
		"border.bg": termui.StringToAttribute(cfg.Theme.View.BorderBg),
		"label.fg":  termui.StringToAttribute(cfg.Theme.View.LabelFg),
		"label.bg":  termui.StringToAttribute(cfg.Theme.View.LabelBg),
	}

	return &cfg, nil
}

func getDefaultConfig() Config {
	return Config{
		SidebarWidth: 1,
		MainWidth:    11,
		Notify:       "",
		Emoji:        false,
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
				"n":          "channel-search-next",
				"N":          "channel-search-prev",
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
		Theme: Theme{
			View: View{
				Fg:       "white",
				Bg:       "default",
				BorderFg: "white",
				BorderBg: "",
				LabelFg:  "green,bold",
				LabelBg:  "",
			},
			Channel: Channel{
				Prefix: "",
				Icon:   "",
				Text:   "",
			},
			Message: Message{
				Time:       "",
				TimeFormat: "15:04",
				Name:       "",
				Text:       "",
			},
		},
	}
}
