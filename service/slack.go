package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
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
	Channels        []components.ChannelItem
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
		return nil, errors.New("not able to authorize client, check your connection and if your slack-token is set correctly")
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

func (s *SlackService) GetChannelsV2() []string {
	slackChans := make([]slack.Channel, 0)

	// Initial request
	initChans, initCur, err := s.Client.GetConversations(
		&slack.GetConversationsParameters{
			ExcludeArchived: "true",
			Limit:           10,
			Types: []string{
				"public_channel",
				"private_channel",
				"im",
				"mpim",
			},
		},
	)
	if err != nil {
		log.Fatal(err) // FIXME
	}

	slackChans = append(slackChans, initChans...)

	// Paginate over additional channels
	nextCur := initCur
	for nextCur != "" {
		channels, cursor, err := s.Client.GetConversations(
			&slack.GetConversationsParameters{
				Cursor:          nextCur,
				ExcludeArchived: "true",
				Limit:           10,
				Types: []string{
					"public_channel",
					"private_channel",
					"im",
					"mpim",
				},
			},
		)
		if err != nil {
			log.Fatal(err) // FIXME
		}

		log.Printf("len(channels): %d", len(channels))
		log.Printf("nextCur: %s", nextCur)
		log.Printf("cursor: %s", cursor)
		log.Println("---")

		slackChans = append(slackChans, channels...)
		nextCur = cursor
	}
	// os.Exit(0)

	var chans []components.ChannelItem
	for _, chn := range slackChans {

		// Defaults
		if chn.IsChannel {
			if !chn.IsMember {
				continue
				os.Exit(0)
			}
		}

		if chn.IsGroup {
			if !chn.IsMember {
				continue
			}
		}

		if chn.IsMpIM {
		}

		if chn.IsIM {
			// TODO: check if user is deleted. IsUsedDeleted is not present
			// in the `conversation` struct.
		}

		var chanName string
		name, ok := s.UserCache[chn.User]
		if ok {
			chanName = name
		} else {
			chanName = chn.Name
		}

		chans = append(
			chans, components.ChannelItem{
				ID:          chn.ID,
				Name:        chanName,
				Topic:       chn.Topic.Value,
				Type:        components.ChannelTypeChannel,
				UserID:      chn.User,
				Presence:    "",
				StylePrefix: s.Config.Theme.Channel.Prefix,
				StyleIcon:   s.Config.Theme.Channel.Icon,
				StyleText:   s.Config.Theme.Channel.Text,
			},
		)

		s.SlackChannels = append(s.SlackChannels, chn)
	}

	s.Channels = chans

	var channels []string
	for _, chn := range s.Channels {
		channels = append(channels, chn.ToString())
	}

	return channels
}

// GetChannels will retrieve all available channels, groups, and im channels.
// Because the channels are of different types, we will append them to
// an []interface as well as to a []Channel which will give us easy access
// to the id and name of the Channel.
func (s *SlackService) GetChannels() []string {
	var chans []components.ChannelItem

	var wg sync.WaitGroup

	// Channels
	wg.Add(1)
	var slackChans []slack.Channel
	go func() {
		var err error
		slackChans, err = s.Client.GetChannels(true)
		if err != nil {
			chans = append(chans, components.ChannelItem{})
		}
		wg.Done()
	}()

	// Groups
	wg.Add(1)
	var slackGroups []slack.Group
	go func() {
		var err error
		slackGroups, err = s.Client.GetGroups(true)
		if err != nil {
			chans = append(chans, components.ChannelItem{})
		}
		wg.Done()
	}()

	// IM
	wg.Add(1)
	var slackIM []slack.IM
	go func() {
		var err error
		slackIM, err = s.Client.GetIMChannels()
		if err != nil {
			chans = append(chans, components.ChannelItem{})
		}
		wg.Done()
	}()

	wg.Wait()

	// Channels
	for _, chn := range slackChans {
		if chn.IsMember {
			s.SlackChannels = append(s.SlackChannels, chn)
			chans = append(
				chans, components.ChannelItem{
					ID:          chn.ID,
					Name:        chn.Name,
					Topic:       chn.Topic.Value,
					Type:        components.ChannelTypeChannel,
					UserID:      "",
					StylePrefix: s.Config.Theme.Channel.Prefix,
					StyleIcon:   s.Config.Theme.Channel.Icon,
					StyleText:   s.Config.Theme.Channel.Text,
				},
			)
		}
	}

	// Groups
	for _, grp := range slackGroups {
		s.SlackChannels = append(s.SlackChannels, grp)
		chans = append(
			chans, components.ChannelItem{
				ID:          grp.ID,
				Name:        grp.Name,
				Topic:       grp.Topic.Value,
				Type:        components.ChannelTypeGroup,
				UserID:      "",
				StylePrefix: s.Config.Theme.Channel.Prefix,
				StyleIcon:   s.Config.Theme.Channel.Icon,
				StyleText:   s.Config.Theme.Channel.Text,
			},
		)
	}

	// IM
	for _, im := range slackIM {

		// Uncover name, when we can't uncover name for
		// IM channel this is then probably a deleted
		// user, because we won't add deleted users
		// to the UserCache, so we skip it
		name, ok := s.UserCache[im.User]

		if ok {
			chans = append(
				chans,
				components.ChannelItem{
					ID:          im.ID,
					Name:        name,
					Topic:       "",
					Type:        components.ChannelTypeIM,
					UserID:      im.User,
					Presence:    "",
					StylePrefix: s.Config.Theme.Channel.Prefix,
					StyleIcon:   s.Config.Theme.Channel.Icon,
					StyleText:   s.Config.Theme.Channel.Text,
				},
			)
			s.SlackChannels = append(s.SlackChannels, im)
		}
	}

	s.Channels = chans

	// We set presence of IM channels here because we need to separately
	// issue an API call for every channel, this will speed up that process
	s.SetPresenceChannels()

	var channels []string
	for _, chn := range s.Channels {
		channels = append(channels, chn.ToString())
	}

	return channels
}

// ChannelsToString will relay the string representation for a channel
func (s *SlackService) ChannelsToString() []string {
	var channels []string
	for _, chn := range s.Channels {
		channels = append(channels, chn.ToString())
	}
	return channels
}

// SetPresence will set presence for all IM channels
func (s *SlackService) SetPresenceChannels() {
	var wg sync.WaitGroup
	for i, channel := range s.SlackChannels {

		switch channel := channel.(type) {
		case slack.IM:
			wg.Add(1)
			go func(i int) {
				presence, _ := s.GetUserPresence(channel.User)
				s.Channels[i].Presence = presence
				wg.Done()
			}(i)
		}

	}

	wg.Wait()
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
		s.Channels[channelID].Notification = false

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

// FindChannel will loop over s.Channels to find the index where the
// channelID equals the ID
func (s *SlackService) FindChannel(channelID string) int {
	var index int
	for i, channel := range s.Channels {
		if channel.ID == channelID {
			index = i
			break
		}
	}
	return index
}

// MarkAsUnread will set the channel as unread
func (s *SlackService) MarkAsUnread(channelID string) {
	index := s.FindChannel(channelID)
	s.Channels[index].Notification = true
}

// GetChannelName will return the name for a specific channelID
func (s *SlackService) GetChannelName(channelID string) string {
	index := s.FindChannel(channelID)
	return s.Channels[index].Name
}

// SendMessage will send a message to a particular channel
func (s *SlackService) SendMessage(channelID int, message string) error {

	// https://godoc.org/github.com/nlopes/slack#PostMessageParameters
	postParams := slack.PostMessageParameters{
		AsUser:    true,
		Username:  s.CurrentUsername,
		LinkNames: 1,
	}

	// https://godoc.org/github.com/nlopes/slack#Client.PostMessage
	_, _, err := s.Client.PostMessage(s.Channels[channelID].ID, message, postParams)
	if err != nil {
		return err
	}

	return nil
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
		msgs = append(msgs, s.CreateMessageFromAttachments(message.Attachments)...)
	}

	// Parse time
	floatTime, err := strconv.ParseFloat(message.Timestamp, 64)
	if err != nil {
		floatTime = 0.0
	}
	intTime := int64(floatTime)

	// Format message
	msg := components.Message{
		Time:       time.Unix(intTime, 0),
		Name:       name,
		Content:    parseMessage(s, message.Text),
		StyleTime:  s.Config.Theme.Message.Time,
		StyleName:  s.Config.Theme.Message.Name,
		StyleText:  s.Config.Theme.Message.Text,
		FormatTime: s.Config.Theme.Message.TimeFormat,
	}

	msgs = append(msgs, msg)

	return msgs
}

func (s *SlackService) CreateMessageFromMessageEvent(message *slack.MessageEvent) ([]components.Message, error) {

	var msgs []components.Message
	var name string

	switch message.SubType {
	case "message_changed":
		// Append (edited) when an edited message is received
		message = &slack.MessageEvent{Msg: *message.SubMessage}
		message.Text = fmt.Sprintf("%s (edited)", message.Text)
	case "message_replied":
		// Ignore reply events
		return nil, errors.New("ignoring reply events")
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
		msgs = append(msgs, s.CreateMessageFromAttachments(message.Attachments)...)
	}

	// Parse time
	floatTime, err := strconv.ParseFloat(message.Timestamp, 64)
	if err != nil {
		floatTime = 0.0
	}
	intTime := int64(floatTime)

	// Format message
	msg := components.Message{
		Time:       time.Unix(intTime, 0),
		Name:       name,
		Content:    parseMessage(s, message.Text),
		StyleTime:  s.Config.Theme.Message.Time,
		StyleName:  s.Config.Theme.Message.Name,
		StyleText:  s.Config.Theme.Message.Text,
		FormatTime: s.Config.Theme.Message.TimeFormat,
	}

	msgs = append(msgs, msg)

	return msgs, nil
}

// CheckNotifyMention check if the message event is either contains a
// mention or is posted on an IM channel
func (s *SlackService) CheckNotifyMention(ev *slack.MessageEvent) bool {
	channel := s.Channels[s.FindChannel(ev.Channel)]
	switch channel.Type {
	case ChannelTypeIM:
		return true
	}

	// Mentions have the following format:
	//	<@U12345|erroneousboat>
	// 	<@U12345>
	r := regexp.MustCompile(`\<@(\w+\|*\w+)\>`)
	matches := r.FindAllString(ev.Text, -1)
	for _, match := range matches {
		if strings.Contains(match, s.CurrentUserID) {
			return true
		}
	}

	return false
}

func (s *SlackService) CreateNotifyMessage(channelID string) string {
	channel := s.Channels[s.FindChannel(channelID)]

	switch channel.Type {
	case ChannelTypeChannel:
		return fmt.Sprintf("Message received on channel: %s", channel.Name)
	case ChannelTypeGroup:
		return fmt.Sprintf("Message received in group: %s", channel.Name)
	case ChannelTypeIM:
		return fmt.Sprintf("Message received from: %s", channel.Name)
	}

	return ""
}

// parseMessage will parse a message string and find and replace:
//	- emoji's
//	- mentions
func parseMessage(s *SlackService, msg string) string {
	if s.Config.Emoji {
		msg = parseEmoji(msg)
	}

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

// CreateMessageFromAttachments will construct a array of string of the Field
// values of Attachments from a Message.
func (s *SlackService) CreateMessageFromAttachments(atts []slack.Attachment) []components.Message {
	var msgs []components.Message
	for _, att := range atts {
		for i := len(att.Fields) - 1; i >= 0; i-- {
			msgs = append(msgs, components.Message{
				Content: fmt.Sprintf(
					"%s %s",
					att.Fields[i].Title,
					att.Fields[i].Value,
				),
				StyleTime:  s.Config.Theme.Message.Time,
				StyleName:  s.Config.Theme.Message.Name,
				StyleText:  s.Config.Theme.Message.Text,
				FormatTime: s.Config.Theme.Message.TimeFormat,
			},
			)
		}

		if att.Text != "" {
			msgs = append(
				msgs,
				components.Message{
					Content:    fmt.Sprintf("%s", att.Text),
					StyleTime:  s.Config.Theme.Message.Time,
					StyleName:  s.Config.Theme.Message.Name,
					StyleText:  s.Config.Theme.Message.Text,
					FormatTime: s.Config.Theme.Message.TimeFormat,
				},
			)
		}

		if att.Title != "" {
			msgs = append(
				msgs,
				components.Message{
					Content:    fmt.Sprintf("%s", att.Title),
					StyleTime:  s.Config.Theme.Message.Time,
					StyleName:  s.Config.Theme.Message.Name,
					StyleText:  s.Config.Theme.Message.Text,
					FormatTime: s.Config.Theme.Message.TimeFormat,
				},
			)
		}
	}

	return msgs
}
