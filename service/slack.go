package service

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"

	"github.com/erroneousboat/slack-term/components"
	"github.com/erroneousboat/slack-term/config"
)

const (
	ChannelTypeChannel = "channel"
	ChannelTypeGroup   = "group"
	ChannelTypeIM      = "im"
)

type SlackService struct {
	Config          *config.Config
	Client          *slack.Client
	RTM             *slack.RTM
	SlackChannels   []interface{}
	Channels        []Channel
	UserCache       map[string]string
	CurrentUserID   string
	CurrentUsername string
}

// NewSlackService is the constructor for the SlackService and will initialize
// the RTM and a Client
func NewSlackService(config *config.Config) (*SlackService, error) {
	svc := &SlackService{
		Config:    config,
		Client:    slack.New(config.SlackToken),
		UserCache: make(map[string]string),
	}

	// Get user associated with token, mainly
	// used to identify user when new messages
	// arrives
	authTest, err := svc.Client.AuthTest()
	if err != nil {
		return nil, errors.New("not able to authorize client, check your connection and or slack-token")
	}
	svc.CurrentUserID = authTest.UserID

	// Create RTM
	svc.RTM = svc.Client.NewRTM()
	go svc.RTM.ManageConnection()

	// Creation of user cache this speeds up
	// the uncovering of usernames of messages
	users, _ := svc.Client.GetUsers()
	for _, user := range users {
		// only add non-deleted users
		if !user.Deleted {
			svc.UserCache[user.ID] = user.Name
		}
	}

	// Get name of current user
	currentUser, err := svc.Client.GetUserInfo(svc.CurrentUserID)
	if err != nil {
		svc.CurrentUsername = "slack-term"
	}
	svc.CurrentUsername = currentUser.Name

	return svc, nil
}

// GetChannels will retrieve all available channels, groups, and im channels.
// Because the channels are of different types, we will append them to
// an []interface as well as to a []Channel which will give us easy access
// to the id and name of the Channel.
func (s *SlackService) GetChannels() []string {
	var chans []Channel

	// Channel
	slackChans, err := s.Client.GetChannels(true)
	if err != nil {
		chans = append(chans, Channel{})
	}
	for _, chn := range slackChans {
		if chn.IsMember {
			s.SlackChannels = append(s.SlackChannels, chn)
			chans = append(
				chans, Channel{
					ID:     chn.ID,
					Name:   chn.Name,
					Topic:  chn.Topic.Value,
					Type:   ChannelTypeChannel,
					UserID: "",
				},
			)
		}
	}

	// Groups
	slackGroups, err := s.Client.GetGroups(true)
	if err != nil {
		chans = append(chans, Channel{})
	}
	for _, grp := range slackGroups {
		s.SlackChannels = append(s.SlackChannels, grp)
		chans = append(
			chans, Channel{
				ID:     grp.ID,
				Name:   grp.Name,
				Topic:  grp.Topic.Value,
				Type:   ChannelTypeGroup,
				UserID: "",
			},
		)
	}

	// IM
	slackIM, err := s.Client.GetIMChannels()
	if err != nil {
		chans = append(chans, Channel{})
	}
	for _, im := range slackIM {

		// FIXME: err
		presence, _ := s.GetUserPresence(im.User)

		// Uncover name, when we can't uncover name for
		// IM channel this is then probably a deleted
		// user, because we won't add deleted users
		// to the UserCache, so we skip it
		name, ok := s.UserCache[im.User]

		if ok {
			chans = append(
				chans,
				Channel{
					ID:       im.ID,
					Name:     name,
					Topic:    "",
					Type:     ChannelTypeIM,
					UserID:   im.User,
					Presence: presence,
				},
			)
			s.SlackChannels = append(s.SlackChannels, im)
		}
	}

	s.Channels = chans

	var channels []string
	for _, chn := range s.Channels {
		channels = append(
			channels,
			chn.ToString(
				s.Config.Theme.Channel.Prefix,
				s.Config.Theme.Channel.Icon,
				s.Config.Theme.Channel.Name,
			),
		)
	}
	return channels
}

// ChannelsToString will relay the string representation for a channel
func (s *SlackService) ChannelsToString() []string {
	var channels []string
	for _, chn := range s.Channels {
		channels = append(
			channels,
			chn.ToString(
				s.Config.Theme.Channel.Prefix,
				s.Config.Theme.Channel.Icon,
				s.Config.Theme.Channel.Name,
			),
		)
	}
	return channels
}

// SetPresenceChannelEvent will set the presence of a IM channel
func (s *SlackService) SetPresenceChannelEvent(userID string, presence string) {
	// Get the correct Channel from svc.Channels
	var index int
	for i, channel := range s.Channels {
		if userID == channel.UserID {
			index = i
			break
		}
	}
	s.Channels[index].Presence = presence
}

// GetSlackChannel returns the representation of a slack channel
func (s *SlackService) GetSlackChannel(selectedChannel int) interface{} {
	return s.SlackChannels[selectedChannel]
}

// GetUserPresence will get the presence of a specific user
func (s *SlackService) GetUserPresence(userID string) (string, error) {
	presence, err := s.Client.GetUserPresence(userID)
	if err != nil {
		return "", err
	}

	return presence.Presence, nil
}

// SetChannelReadMark will set the read mark for a channel, group, and im
// channel based on the current time.
func (s *SlackService) SetChannelReadMark(channel interface{}) {
	switch channel := channel.(type) {
	case slack.Channel:
		s.Client.SetChannelReadMark(
			channel.ID, fmt.Sprintf("%f",
				float64(time.Now().Unix())),
		)
	case slack.Group:
		s.Client.SetGroupReadMark(
			channel.ID, fmt.Sprintf("%f",
				float64(time.Now().Unix())),
		)
	case slack.IM:
		s.Client.MarkIMChannel(
			channel.ID, fmt.Sprintf("%f",
				float64(time.Now().Unix())),
		)
	}
}

// MarkAsRead will set the channel as read
func (s *SlackService) MarkAsRead(channelID int) {
	channel := s.Channels[channelID]

	if channel.Notification {
		channel.Notification = false

		switch channel.Type {
		case ChannelTypeChannel:
			s.Client.SetChannelReadMark(
				channel.ID, fmt.Sprintf("%f",
					float64(time.Now().Unix())),
			)
		case ChannelTypeGroup:
			s.Client.SetGroupReadMark(
				channel.ID, fmt.Sprintf("%f",
					float64(time.Now().Unix())),
			)
		case ChannelTypeIM:
			s.Client.MarkIMChannel(
				channel.ID, fmt.Sprintf("%f",
					float64(time.Now().Unix())),
			)
		}
	}
}

// MarkAsUnread will set the channel as unread
func (s *SlackService) MarkAsUnread(channelID string) {
	var index int
	for i, channel := range s.Channels {
		if channel.ID == channelID {
			index = i
			break
		}
	}
	s.Channels[index].Notification = true
}

// SendMessage will send a message to a particular channel
func (s *SlackService) SendMessage(channelID int, message string) {

	// https://godoc.org/github.com/nlopes/slack#PostMessageParameters
	postParams := slack.PostMessageParameters{
		AsUser:   true,
		Username: s.CurrentUsername,
	}

	// https://godoc.org/github.com/nlopes/slack#Client.PostMessage
	s.Client.PostMessage(s.Channels[channelID].ID, message, postParams)
}

// GetMessages will get messages for a channel, group or im channel delimited
// by a count.
func (s *SlackService) GetMessages(channel interface{}, count int) []components.Message {
	// https://api.slack.com/methods/channels.history
	historyParams := slack.HistoryParameters{
		Count:     count,
		Inclusive: false,
		Unreads:   false,
	}

	// https://godoc.org/github.com/nlopes/slack#History
	history := new(slack.History)
	var err error
	switch chnType := channel.(type) {
	case slack.Channel:
		history, err = s.Client.GetChannelHistory(chnType.ID, historyParams)
		if err != nil {
			log.Fatal(err) // FIXME
		}
	case slack.Group:
		history, err = s.Client.GetGroupHistory(chnType.ID, historyParams)
		if err != nil {
			log.Fatal(err) // FIXME
		}
	case slack.IM:
		history, err = s.Client.GetIMHistory(chnType.ID, historyParams)
		if err != nil {
			log.Fatal(err) // FIXME
		}
	}

	// Construct the messages
	var messages []components.Message
	for _, message := range history.Messages {
		msg := s.CreateMessage(message)
		messages = append(messages, msg...)
	}

	// Reverse the order of the messages, we want the newest in
	// the last place
	var messagesReversed []components.Message
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
// This returns an array of string because we will try to uncover attachments
// associated with messages.
func (s *SlackService) CreateMessage(message slack.Message) []components.Message {
	var msgs []components.Message
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

	// When there are attachments append them
	if len(message.Attachments) > 0 {
		msgs = append(msgs, createMessageFromAttachments(message.Attachments)...)
	}

	// Parse time
	floatTime, err := strconv.ParseFloat(message.Timestamp, 64)
	if err != nil {
		floatTime = 0.0
	}
	intTime := int64(floatTime)

	// Format message
	msg := components.Message{
		Time:    time.Unix(intTime, 0),
		Name:    name,
		Content: parseMessage(s, message.Text),
	}

	msgs = append(msgs, msg)

	return msgs
}

func (s *SlackService) CreateMessageFromMessageEvent(message *slack.MessageEvent) []components.Message {

	var msgs []components.Message
	var name string

	// Append (edited) when an edited message is received
	if message.SubType == "message_changed" {
		message = &slack.MessageEvent{Msg: *message.SubMessage}
		message.Text = fmt.Sprintf("%s (edited)", message.Text)
	}

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

	// When there are attachments append them
	if len(message.Attachments) > 0 {
		msgs = append(msgs, createMessageFromAttachments(message.Attachments)...)
	}

	// Parse time
	floatTime, err := strconv.ParseFloat(message.Timestamp, 64)
	if err != nil {
		floatTime = 0.0
	}
	intTime := int64(floatTime)

	// Format message
	msg := components.Message{
		Time:    time.Unix(intTime, 0),
		Name:    name,
		Content: parseMessage(s, message.Text),
	}

	msgs = append(msgs, msg)

	return msgs
}

// parseMessage will parse a message string and find and replace:
//	- emoji's
//	- mentions
func parseMessage(s *SlackService, msg string) string {
	// NOTE: Commented out because rendering of the emoji's
	// creates artifacts from the last view because of
	// double width emoji's
	// msg = parseEmoji(msg)

	msg = parseMentions(s, msg)

	return msg
}

// parseMentions will try to find mention placeholders in the message
// string and replace them with the correct username with and @ symbol
//
// Mentions have the following format:
//	<@U12345|erroneousboat>
// 	<@U12345>
func parseMentions(s *SlackService, msg string) string {
	r := regexp.MustCompile(`\<@(\w+\|*\w+)\>`)

	return r.ReplaceAllStringFunc(
		msg, func(str string) string {
			rs := r.FindStringSubmatch(str)
			if len(rs) < 1 {
				return str
			}

			var userID string
			split := strings.Split(rs[1], "|")
			if len(split) > 0 {
				userID = split[0]
			} else {
				userID = rs[1]
			}

			name, ok := s.UserCache[userID]
			if !ok {
				user, err := s.Client.GetUserInfo(userID)
				if err != nil {
					name = "unknown"
					s.UserCache[userID] = name
				} else {
					name = user.Name
					s.UserCache[userID] = user.Name
				}
			}

			if name == "" {
				name = "unknown"
			}

			return "@" + name
		},
	)
}

// parseEmoji will try to find emoji placeholders in the message
// string and replace them with the correct unicode equivalent
func parseEmoji(msg string) string {
	r := regexp.MustCompile("(:\\w+:)")

	return r.ReplaceAllStringFunc(
		msg, func(str string) string {
			code, ok := config.EmojiCodemap[str]
			if !ok {
				return str
			}
			return code
		},
	)
}

// createMessageFromAttachments will construct a array of string of the Field
// values of Attachments from a Message.
func createMessageFromAttachments(atts []slack.Attachment) []components.Message {
	var msgs []components.Message
	for _, att := range atts {
		for i := len(att.Fields) - 1; i >= 0; i-- {
			msgs = append(msgs, components.Message{
				Content: fmt.Sprintf(
					"%s %s",
					att.Fields[i].Title,
					att.Fields[i].Value,
				),
			},
			)
		}

		if att.Text != "" {
			msgs = append(
				msgs,
				components.Message{Content: fmt.Sprintf("%s", att.Text)},
			)
		}

		if att.Title != "" {
			msgs = append(
				msgs,
				components.Message{Content: fmt.Sprintf("%s", att.Title)},
			)
		}
	}

	return msgs
}
