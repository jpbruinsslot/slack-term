package components

import (
	"fmt"
	"sort"
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
	ID       string
	Messages map[string]Message

	Time    time.Time
	Thread  string
	Name    string
	Content string

	StyleTime   string
	StyleThread string
	StyleName   string
	StyleText   string

	FormatTime string
}

func (m Message) GetTime() string {
	return fmt.Sprintf(
		"[[%s]](%s) ",
		m.Time.Format(m.FormatTime),
		m.StyleTime,
	)
}

func (m Message) GetThread() string {
	return fmt.Sprintf("[%s](%s)",
		m.Thread,
		m.StyleThread,
	)
}

func (m Message) GetName() string {
	return fmt.Sprintf("[<%s>](%s) ",
		m.Name,
		m.colorizeName(m.StyleName),
	)
}

func (m Message) GetContent() string {
	return fmt.Sprintf("[.](%s)", m.StyleText)
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

func SortMessages(msgs map[string]Message) []Message {
	keys := make([]string, 0)
	for k := range msgs {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	sortedMessages := make([]Message, 0)
	for _, k := range keys {
		sortedMessages = append(sortedMessages, msgs[k])
	}

	return sortedMessages
}
