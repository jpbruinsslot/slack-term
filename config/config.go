package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	fp "path/filepath"

	"github.com/OpenPeeDeeP/xdg"
	"github.com/erroneousboat/termui"
)

const (
	NotifyAll     = "all"
	NotifyMention = "mention"
)

// Config is the definition of a Config struct
type Config struct {
	SlackToken   string                `json:"slack_token"`
	Notify       string                `json:"notify"`
	Emoji        bool                  `json:"emoji"`
	SidebarWidth int                   `json:"sidebar_width"`
	MainWidth    int                   `json:"-"`
	ThreadsWidth int                   `json:"threads_width"`
	KeyMap       map[string]keyMapping `json:"key_map"`
	Theme        Theme                 `json:"theme"`
}

type keyMapping map[string]string

// NewConfig loads the config file and returns a Config struct
func NewConfig(filepath string) (*Config, error) {
	cfg := getDefaultConfig()

	// Open config file, and when none is found or present create
	// a default empty one, at the default filepath location
	file, err := os.Open(filepath)
	if err != nil {
		file, err = CreateConfigFile(filepath)
		if err != nil {
			return &cfg, fmt.Errorf("couldn't open the slack-term config file: (%v)", err)
		}
	}

	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return &cfg, fmt.Errorf("the slack-term config file isn't valid json: (%v)", err)
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

func CreateConfigFile(filepath string) (*os.File, error) {
	filepath = fmt.Sprintf("%s/slack-term/%s", xdg.ConfigHome(), "config")

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		os.MkdirAll(fp.Dir(filepath), os.ModePerm)
	}

	payload := "{\"slack_token\": \"\"}"
	err := ioutil.WriteFile(filepath, []byte(payload), 0755)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func getDefaultConfig() Config {
	return Config{
		SidebarWidth: 1,
		MainWidth:    11,
		ThreadsWidth: 1,
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
				"K":          "thread-up",
				"J":          "thread-down",
				"<previous>": "chat-up",
				"C-b":        "chat-up",
				"C-u":        "chat-up",
				"<next>":     "chat-down",
				"C-f":        "chat-down",
				"C-d":        "chat-down",
				"n":          "channel-search-next",
				"N":          "channel-search-prev",
				"'":          "channel-jump",
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
				Thread:     "fg-bold",
				Name:       "",
				Text:       "",
			},
		},
	}
}
