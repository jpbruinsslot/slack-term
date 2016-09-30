package service

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/nlopes/slack"
)

type SlackService struct {
	Client    *slack.Client
	RTM       *slack.RTM
	Channels  []slack.Channel
	UserCache map[string]string
}

type Channel struct {
	ID   string
	Name string
}

func NewSlackService(token string) *SlackService {
	svc := &SlackService{
		Client:    slack.New(token),
		UserCache: make(map[string]string),
	}

	svc.RTM = svc.Client.NewRTM()

	go svc.RTM.ManageConnection()

	users, _ := svc.Client.GetUsers()
	for _, user := range users {
		svc.UserCache[user.ID] = user.Name
	}

	return svc
}

func (s *SlackService) GetChannels() []Channel {
	var chans []Channel

	slackChans, err := s.Client.GetChannels(true)
	if err != nil {
		chans = append(chans, Channel{})
	}

	s.Channels = slackChans

	for _, slackChan := range slackChans {
		chans = append(chans, Channel{slackChan.ID, slackChan.Name})
	}

	return chans
}

func (s *SlackService) SendMessage(channel string, message string) {
	// https://godoc.org/github.com/nlopes/slack#PostMessageParameters
	postParams := slack.PostMessageParameters{
		AsUser: true,
	}

	// https://godoc.org/github.com/nlopes/slack#Client.PostMessage
	s.Client.PostMessage(channel, message, postParams)
}

func (s *SlackService) GetMessages(channel string, count int) []string {
	// https://api.slack.com/methods/channels.history
	historyParams := slack.HistoryParameters{
		Count:     count,
		Inclusive: false,
		Unreads:   false,
	}

	// https://godoc.org/github.com/nlopes/slack#History
	history, err := s.Client.GetChannelHistory(channel, historyParams)
	if err != nil {
		log.Fatal(err)
		return []string{""}
	}

	// Construct the messages
	var messages []string
	for _, message := range history.Messages {
		msg := s.CreateMessage(message)
		messages = append(messages, msg)
	}

	// Reverse the order of the messages, we want the newest in
	// the last place
	var messagesReversed []string
	for i := len(messages) - 1; i >= 0; i-- {
		messagesReversed = append(messagesReversed, messages[i])
	}

	return messagesReversed
}

// CreateMessage will create a string formatted message that can be rendered
// in the Chat pane.
//
// [23:59] <erroneousboat> Hello world!
//
func (s *SlackService) CreateMessage(message slack.Message) string {
	var name string

	// Get username from cache
	name, ok := s.UserCache[message.User]

	// Name not in cache
	if !ok {
		if message.BotID != "" {
			// Name not found, perhaps a bot, use Username
			name, ok = s.UserCache[message.BotID]
			if !ok {
				// Not found in cache, add it
				name = message.Username
				s.UserCache[message.BotID] = message.Username
			}
		} else {
			// Not a bot, not in cache, get user info
			user, err := s.Client.GetUserInfo(message.User)
			if err != nil {
				name = "unknown"
				s.UserCache[message.User] = name
			} else {
				name = user.Name
				s.UserCache[message.User] = user.Name
			}
		}
	}

	if name == "" {
		name = "unknown"
	}

	// Parse time
	floatTime, err := strconv.ParseFloat(message.Timestamp, 64)
	if err != nil {
		floatTime = 0.0
	}
	intTime := int64(floatTime)

	// Format message
	msg := fmt.Sprintf(
		"[%s] <%s> %s",
		time.Unix(intTime, 0).Format("15:04"),
		name,
		message.Text,
	)

	return msg
}

func (s *SlackService) CreateMessageFromMessageEvent(message *slack.MessageEvent) string {

	var name string

	// Get username from cache
	name, ok := s.UserCache[message.User]

	// Name not in cache
	if !ok {
		if message.BotID != "" {
			// Name not found, perhaps a bot, use Username
			name, ok = s.UserCache[message.BotID]
			if !ok {
				// Not found in cache, add it
				name = message.Username
				s.UserCache[message.BotID] = message.Username
			}
		} else {
			// Not a bot, not in cache, get user info
			user, err := s.Client.GetUserInfo(message.User)
			if err != nil {
				name = "unknown"
				s.UserCache[message.User] = name
			} else {
				name = user.Name
				s.UserCache[message.User] = user.Name
			}
		}
	}

	if name == "" {
		name = "unknown"
	}

	// Parse time
	floatTime, err := strconv.ParseFloat(message.Timestamp, 64)
	if err != nil {
		floatTime = 0.0
	}
	intTime := int64(floatTime)

	// Format message
	msg := fmt.Sprintf(
		"[%s] <%s> %s",
		time.Unix(intTime, 0).Format("15:04"),
		name,
		message.Text,
	)

	return msg
}
