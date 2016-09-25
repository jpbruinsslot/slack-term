package service

import (
	"fmt"
	"log"

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

func (s *SlackService) SendMessage(message string) {}

func (s *SlackService) GetMessages(channel string) []string {
	// https://api.slack.com/methods/channels.history
	historyParams := slack.HistoryParameters{
		// Latest: "now",
		// Oldest:    0,
		Count:     50,
		Inclusive: false,
		Unreads:   false,
	}

	// https://godoc.org/github.com/nlopes/slack#History
	history, err := s.Client.GetChannelHistory(channel, historyParams)
	if err != nil {
		log.Fatal(err)
		return []string{""}
	}

	// TODO: this takes a long time, maybe use some dynamic programming
	var messages []string
	for _, message := range history.Messages {
		var name string
		user, err := s.Client.GetUserInfo(message.User)
		if err == nil {
			name = user.Name
		} else {
			name = "unknown"
		}

		msg := fmt.Sprintf("[%s] %s", name, message.Text)
		messages = append(messages, msg)
	}

	return messages
}
