package config

type Theme struct {
	Message Message `json:"message"`
	Channel Channel `json:"channel"`
}

type Message struct {
	Time    string `json:"time"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type Channel struct {
	Prefix string `json:"prefix"`
	Icon   string `json:"icon"`
	Name   string `json:"name"`
}
