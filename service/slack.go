package service

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nlopes/slack"

	"github.com/erroneousboat/slack-term/components"
	"github.com/erroneousboat/slack-term/config"
)

type SlackService struct {
	Config          *config.Config
	Client          *slack.Client
	RTM             *slack.RTM
	Conversations   []slack.Channel
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

func (s *SlackService) GetChannels() []components.ChannelItem {
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

		slackChans = append(slackChans, channels...)
		nextCur = cursor
	}

	// We're creating tempChan, because we want to be able to
	// sort the types of channels into buckets
	type tempChan struct {
		channelItem  components.ChannelItem
		slackChannel slack.Channel
	}

	// Initialize buckets
	buckets := make(map[int]map[string]*tempChan)
	buckets[0] = make(map[string]*tempChan) // Channels
	buckets[1] = make(map[string]*tempChan) // Group
	buckets[2] = make(map[string]*tempChan) // MpIM
	buckets[3] = make(map[string]*tempChan) // IM

	var wg sync.WaitGroup
	for _, chn := range slackChans {
		chanItem := s.createChannelItem(chn)

		if chn.IsChannel {
			if !chn.IsMember {
				continue
			}

			chanItem.Type = components.ChannelTypeChannel

			buckets[0][chn.ID] = &tempChan{
				channelItem:  chanItem,
				slackChannel: chn,
			}
		}

		if chn.IsGroup {
			if !chn.IsMember {
				continue
			}

			chanItem.Type = components.ChannelTypeGroup

			buckets[1][chn.ID] = &tempChan{
				channelItem:  chanItem,
				slackChannel: chn,
			}
		}

		if chn.IsMpIM {
			chanItem.Type = components.ChannelTypeMpIM

			buckets[2][chn.ID] = &tempChan{
				channelItem:  chanItem,
				slackChannel: chn,
			}
		}

		if chn.IsIM {
			// Check if user is deleted, we do this by checking the user id,
			// and see if we have the user in the UserCache
			name, ok := s.UserCache[chn.User]
			if !ok {
				continue
			}

			chanItem.Name = name
			chanItem.Type = components.ChannelTypeIM

			buckets[3][chn.User] = &tempChan{
				channelItem:  chanItem,
				slackChannel: chn,
			}

			wg.Add(1)
			go func(user string, buckets map[int]map[string]*tempChan) {
				defer wg.Done()

				presence, err := s.GetUserPresence(user)
				if err != nil {
					buckets[3][user].channelItem.Presence = "away"
					return
				}

				buckets[3][user].channelItem.Presence = presence
			}(chn.User, buckets)
		}
	}

	wg.Wait()

	// Sort the buckets
	var keys []int
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	var chans []components.ChannelItem
	for _, k := range keys {

		bucket := buckets[k]

		// Sort channels in every bucket
		tcArr := make([]tempChan, 0)
		for _, v := range bucket {
			tcArr = append(tcArr, *v)
		}

		sort.Slice(tcArr, func(i, j int) bool {
			return tcArr[i].channelItem.Name < tcArr[j].channelItem.Name
		})

		// Add ChannelItem and SlackChannel to the SlackService struct
		for _, tc := range tcArr {
			chans = append(chans, tc.channelItem)
			s.Conversations = append(s.Conversations, tc.slackChannel)
		}
	}

	return chans
}

// GetUserPresence will get the presence of a specific user
func (s *SlackService) GetUserPresence(userID string) (string, error) {
	presence, err := s.Client.GetUserPresence(userID)
	if err != nil {
		return "", err
	}

	return presence.Presence, nil
}

// MarkAsRead will set the channel as read
func (s *SlackService) MarkAsRead(channelID string) {

	// TODO: does this work with other channel types? See old one below,
	// test this
	s.Client.SetChannelReadMark(
		channelID, fmt.Sprintf("%f",
			float64(time.Now().Unix())),
	)

	// switch channel.Type {
	// case ChannelTypeChannel:
	// 	s.Client.SetChannelReadMark(
	// 		channel.ID, fmt.Sprintf("%f",
	// 			float64(time.Now().Unix())),
	// 	)
	// case ChannelTypeGroup:
	// 	s.Client.SetGroupReadMark(
	// 		channel.ID, fmt.Sprintf("%f",
	// 			float64(time.Now().Unix())),
	// 	)
	// case ChannelTypeIM:
	// 	s.Client.MarkIMChannel(
	// 		channel.ID, fmt.Sprintf("%f",
	// 			float64(time.Now().Unix())),
	// 	)
	// }
}

// SendMessage will send a message to a particular channel
func (s *SlackService) SendMessage(channelID string, message string) error {

	// https://godoc.org/github.com/nlopes/slack#PostMessageParameters
	postParams := slack.PostMessageParameters{
		AsUser:    true,
		Username:  s.CurrentUsername,
		LinkNames: 1,
	}

	// https://godoc.org/github.com/nlopes/slack#Client.PostMessage
	_, _, err := s.Client.PostMessage(channelID, message, postParams)
	if err != nil {
		return err
	}

	return nil
}

// GetMessages will get messages for a channel, group or im channel delimited
// by a count.
func (s *SlackService) GetMessages(channelID string, count int) []components.Message {
	// TODO: check other parameters
	// https://godoc.org/github.com/nlopes/slack#GetConversationHistoryParameters
	historyParams := slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     count,
		Inclusive: false,
	}

	history, err := s.Client.GetConversationHistory(&historyParams)
	if err != nil {
		log.Fatal(err) // FIXME
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

func (s *SlackService) createChannelItem(chn slack.Channel) components.ChannelItem {
	return components.ChannelItem{
		ID:          chn.ID,
		Name:        chn.Name,
		Topic:       chn.Topic.Value,
		UserID:      chn.User,
		StylePrefix: s.Config.Theme.Channel.Prefix,
		StyleIcon:   s.Config.Theme.Channel.Icon,
		StyleText:   s.Config.Theme.Channel.Text,
	}
}
