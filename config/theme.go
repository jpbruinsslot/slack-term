package config

type Theme struct {
	View    View    `json:"view"`
	Channel Channel `json:"channel"`
	Message Message `json:"message"`
}

type View struct {
	Fg         string `json:"fg"`
	Bg         string `json:"bg"`
	BorderFg   string `json:"border_fg"`
	LabelFg    string `json:"border_fg"`
	ParFg      string `json:"par_fg"`
	ParLabelFg string `json:"par_label_fg"`
}

type Message struct {
	Time string `json:"time"`
	Name string `json:"name"`
	Text string `json:"text"`
}

type Channel struct {
	Prefix string `json:"prefix"`
	Icon   string `json:"icon"`
	Text   string `json:"text"`
}
