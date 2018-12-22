package components

import (
	"strings"
	"time"
)

var (
	COLORS = []string{
		"fg-black",
		"fg-red",
		"fg-green",
		"fg-yellow",
		"fg-blue",
		"fg-magenta",
		"fg-cyan",
		"fg-white",
	}
)

type Message struct {
	Time    time.Time
	Name    string
	Content string

	StyleTime string
	StyleName string
	StyleText string

	FormatTime string
}

func (m Message) colorizeName(styleName string) string {
	if strings.Contains(styleName, "colorize") {
		var sum int
		for _, c := range m.Name {
			sum = sum + int(c)
		}

		i := sum % len(COLORS)

		return strings.Replace(m.StyleName, "colorize", COLORS[i], -1)
	}

	return styleName
}
