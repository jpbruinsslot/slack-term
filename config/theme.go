package config

type Theme struct {
	View    View    `json:"view"`
	Channel Channel `json:"channel"`
	Message Message `json:"message"`
}

type View struct {
	Fg       string `json:"fg"`        // Foreground text
	Bg       string `json:"bg"`        // Background text
	BorderFg string `json:"border_fg"` // Border foreground
	BorderBg string `json:"border_bg"` // Border background
	LabelFg  string `json:"label_fg"`  // Label text foreground
	LabelBg  string `json:"label_bg"`  // Label text background
}

type Message struct {
	Time       string `json:"time"`
	Name       string `json:"name"`
	Thread     string `json:"thread"`
	Text       string `json:"text"`
	TimeFormat string `json:"time_format"`
}

type Channel struct {
	Prefix string `json:"prefix"`
	Icon   string `json:"icon"`
	Text   string `json:"text"`
}
