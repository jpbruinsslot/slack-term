package service

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/nlopes/slack"
)

type SlackService struct {
	Client   *slack.Client
	RTM      *slack.RTM
	Channels []slack.Channel
}

type Channel struct {
	ID   string
	Name string
}

func NewSlackService(token string) *SlackService {
	svc := new(SlackService)

	svc.Client = slack.New(token)
	svc.RTM = svc.Client.NewRTM()

	go svc.RTM.ManageConnection()

	return svc
}

func (s *SlackService) Connect() {

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

	// Here we will construct the messages and format them with a username.
	// Because we need to call the API again for an username because we only
	// will get an user ID from a message, we will storage user ID's and names
	// in a map.
	var messages []string
	users := make(map[string]string)
	for _, message := range history.Messages {
		var name string
		name, ok := users[message.User]
		if !ok {
			user, err := s.Client.GetUserInfo(message.User)
			if err == nil {
				name = user.Name
				users[message.User] = user.Name
			} else {
				name = message.Username
				users[message.User] = name
			}
		}

		// TODO: refactor this to CreateMessage

		// Parse the time we get from slack which is a Unix time float
		floatTime, err := strconv.ParseFloat(message.Timestamp, 64)
		if err != nil {
			floatTime = 0.0
		}
		intTime := int64(floatTime)

		msg := fmt.Sprintf(
			"[%s] <%s> %s",
			time.Unix(intTime, 0).Format("15:04"),
			name,
			message.Text,
		)
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
