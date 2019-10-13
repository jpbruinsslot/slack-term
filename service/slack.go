package service

import (
	"errors"
	"fmt"
	"html"
	"log"
	"net/url"
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
	ThreadCache     map[string]string
	MessageCache    map[string]string
	CurrentUserID   string
	CurrentUsername string
}

// NewSlackService is the constructor for the SlackService and will initialize
// the RTM and a Client
func NewSlackService(config *config.Config) (*SlackService, error) {
	svc := &SlackService{
		Config:       config,
		Client:       slack.New(config.SlackToken),
		UserCache:    make(map[string]string),
		ThreadCache:  make(map[string]string),
		MessageCache: make(map[string]string),
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

	// Get name of current user, and set presence to active
	currentUser, err := svc.Client.GetUserInfo(svc.CurrentUserID)
	if err != nil {
		svc.CurrentUsername = "slack-term"
	}
	svc.CurrentUsername = currentUser.Name
	svc.SetUserAsActive()

	return svc, nil
}

func (s *SlackService) GetChannels() ([]components.ChannelItem, error) {
	slackChans := make([]slack.Channel, 0)

	// Initial request
	initChans, initCur, err := s.Client.GetConversations(
		&slack.GetConversationsParameters{
			ExcludeArchived: "true",
			Limit:           1000,
			Types: []string{
				"public_channel",
				"private_channel",
				"im",
				"mpim",
			},
		},
	)
	if err != nil {
		return nil, err
	}

	slackChans = append(slackChans, initChans...)

	// Paginate over additional channels
	nextCur := initCur
	for nextCur != "" {
		channels, cursor, err := s.Client.GetConversations(
			&slack.GetConversationsParameters{
				Cursor:          nextCur,
				ExcludeArchived: "true",
				Limit:           1000,
				Types: []string{
					"public_channel",
					"private_channel",
					"im",
					"mpim",
				},
			},
		)
		if err != nil {
			return nil, err
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

			if chn.UnreadCount > 0 {
				chanItem.Notification = true
			}

			buckets[0][chn.ID] = &tempChan{
				channelItem:  chanItem,
				slackChannel: chn,
			}
		}

		if chn.IsGroup {
			if !chn.IsMember {
				continue
			}

			// This is done because MpIM channels are also considered groups
			if chn.IsMpIM {
				if !chn.IsOpen {
					continue
				}

				chanItem.Type = components.ChannelTypeMpIM

				if chn.UnreadCount > 0 {
					chanItem.Notification = true
				}

				buckets[2][chn.ID] = &tempChan{
					channelItem:  chanItem,
					slackChannel: chn,
				}
			} else {

				chanItem.Type = components.ChannelTypeGroup

				if chn.UnreadCount > 0 {
					chanItem.Notification = true
				}

				buckets[1][chn.ID] = &tempChan{
					channelItem:  chanItem,
					slackChannel: chn,
				}
			}
		}

		// NOTE: user presence is set in the event handler by the function
		// `actionSetPresenceAll`, that is why we set the presence to away
		if chn.IsIM {
			// Check if user is deleted, we do this by checking the user id,
			// and see if we have the user in the UserCache
			name, ok := s.UserCache[chn.User]
			if !ok {
				continue
			}

			chanItem.Name = name
			chanItem.Type = components.ChannelTypeIM
			chanItem.Presence = "away"

			if chn.UnreadCount > 0 {
				chanItem.Notification = true
			}

			buckets[3][chn.User] = &tempChan{
				channelItem:  chanItem,
				slackChannel: chn,
			}
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

	return chans, nil
}

// GetUserPresence will get the presence of a specific user
func (s *SlackService) GetUserPresence(userID string) (string, error) {
	presence, err := s.Client.GetUserPresence(userID)
	if err != nil {
		return "", err
	}

	return presence.Presence, nil
}

// IsNewThread checks if a thread has been previously observed by checking
// for its presence in the thread cache
func (s *SlackService) IsNewThread(threadID string) bool {

	if _, ok := s.ThreadCache[threadID]; ok || threadID == "" {
		return false
	}

	return true
}

// Set current user presence to active
func (s *SlackService) SetUserAsActive() {
	s.Client.SetUserPresence("auto")
}

// MarkAsRead will set the channel as read
func (s *SlackService) MarkAsRead(channelItem components.ChannelItem) {
	switch channelItem.Type {
	case components.ChannelTypeChannel:
		s.Client.SetChannelReadMark(
			channelItem.ID, fmt.Sprintf("%f",
				float64(time.Now().Unix())),
		)
	case components.ChannelTypeGroup:
		s.Client.SetGroupReadMark(
			channelItem.ID, fmt.Sprintf("%f",
				float64(time.Now().Unix())),
		)
	case components.ChannelTypeMpIM:
		s.Client.MarkIMChannel(
			channelItem.ID, fmt.Sprintf("%f",
				float64(time.Now().Unix())),
		)
	case components.ChannelTypeIM:
		s.Client.MarkIMChannel(
			channelItem.ID, fmt.Sprintf("%f",
				float64(time.Now().Unix())),
		)
	}
}

// SendMessage will send a message to a particular channel
func (s *SlackService) SendMessage(channelID string, message string) error {

	// https://godoc.org/github.com/nlopes/slack#PostMessageParameters
	postParams := slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{
		AsUser:    true,
		Username:  s.CurrentUsername,
		LinkNames: 1,
	})

	text := slack.MsgOptionText(message, true)

	// https://godoc.org/github.com/nlopes/slack#Client.PostMessage
	_, _, err := s.Client.PostMessage(channelID, text, postParams)
	if err != nil {
		return err
	}

	return nil
}

// SendReply will send a message to a particular thread, specifying the
// ThreadTimestamp will make it reply to that specific thread. (see:
// https://api.slack.com/docs/message-threading, 'Posting replies')
func (s *SlackService) SendReply(channelID string, threadID string, message string) error {
	// https://godoc.org/github.com/nlopes/slack#PostMessageParameters
	postParams := slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{
		AsUser:          true,
		Username:        s.CurrentUsername,
		LinkNames:       1,
		ThreadTimestamp: threadID,
	})

	text := slack.MsgOptionText(message, true)

	// https://godoc.org/github.com/nlopes/slack#Client.PostMessage
	_, _, err := s.Client.PostMessage(channelID, text, postParams)
	if err != nil {
		return err
	}

	return nil
}

// SendCommand will send a specific command to slack. First we check
// wether we are dealing with a command, and if it is one of the supported
// ones.
//
// NOTE: slack slash commands that are sent to the slack api are undocumented,
// and as such we need to update the message option that direct it to the
// correct api endpoint.
//
// https://github.com/ErikKalkoken/slackApiDoc/blob/master/chat.command.md
func (s *SlackService) SendCommand(channelID string, message string) (bool, error) {
	// First check if it begins with slash and a command
	r, err := regexp.Compile(`^/\w+`)
	if err != nil {
		return false, err
	}

	match := r.MatchString(message)
	if !match {
		return false, nil
	}

	// Execute the the command when supported
	switch r.FindString(message) {
	case "/thread":
		r := regexp.MustCompile(`(?P<cmd>^/\w+) (?P<id>\w+) (?P<msg>.*)`)
		subMatch := r.FindStringSubmatch(message)

		if len(subMatch) < 4 {
			return false, errors.New("'/thread' command malformed")
		}

		msg := subMatch[3]

		if threadID, ok := s.ThreadCache[subMatch[2]]; ok {
			err := s.SendReply(channelID, threadID, msg)
			if err != nil {
				return false, err
			}
		} else if msgID, ok := s.MessageCache[subMatch[2]]; ok {
			err := s.SendReply(channelID, msgID, msg)
			if err != nil {
				return true, err
			}
		}

		return true, nil
	default:
		r := regexp.MustCompile(`(?P<cmd>^/\w+) (?P<text>.*)`)
		subMatch := r.FindStringSubmatch(message)

		if len(subMatch) < 3 {
			return false, errors.New("slash command malformed")
		}

		cmd := subMatch[1]
		text := subMatch[2]

		msgOption := slack.UnsafeMsgOptionEndpoint(
			fmt.Sprintf("%s%s", slack.APIURL, "chat.command"),
			func(urlValues url.Values) {
				urlValues.Add("command", cmd)
				urlValues.Add("text", text)
			},
		)

		_, _, err := s.Client.PostMessage(channelID, msgOption)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

// GetMessages will get messages for a channel, group or im channel delimited
// by a count. It will return the messages, the thread identifiers
// (as ChannelItem), and and error.
func (s *SlackService) GetMessages(channelID string, count int) ([]components.Message, []components.ChannelItem, error) {

	// https://godoc.org/github.com/nlopes/slack#GetConversationHistoryParameters
	historyParams := slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     count,
		Inclusive: false,
	}

	history, err := s.Client.GetConversationHistory(&historyParams)
	if err != nil {
		return nil, nil, err
	}

	// Construct the messages
	var messages []components.Message
	var threads []components.ChannelItem
	for _, message := range history.Messages {
		msg := s.CreateMessage(message, channelID)
		messages = append(messages, msg)

		// FIXME: create boolean isThread
		if msg.Thread != "" {
			threads = append(threads, components.ChannelItem{
				ID:          msg.ID,
				Name:        msg.Thread,
				Type:        components.ChannelTypeGroup,
				StylePrefix: s.Config.Theme.Channel.Prefix,
				StyleIcon:   s.Config.Theme.Channel.Icon,
				StyleText:   s.Config.Theme.Channel.Text,
			})
		}
	}

	// Reverse the order of the messages, we want the newest in
	// the last place
	var messagesReversed []components.Message
	for i := len(messages) - 1; i >= 0; i-- {
		messagesReversed = append(messagesReversed, messages[i])
	}

	return messagesReversed, threads, nil
}

// CreateMessageByID will construct an array of components.Message with only
// 1 message, using the message ID (Timestamp).
//
// For the choice of history parameters see:
// https://api.slack.com/messaging/retrieving
func (s *SlackService) GetMessageByID(messageID string, channelID string) ([]components.Message, error) {

	var msgs []components.Message

	// https://godoc.org/github.com/nlopes/slack#GetConversationHistoryParameters
	historyParams := slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     1,
		Inclusive: true,
		Latest:    messageID,
	}

	history, err := s.Client.GetConversationHistory(&historyParams)
	if err != nil {
		return msgs, err
	}

	// We break because we're only asking for 1 message
	for _, message := range history.Messages {
		msgs = append(msgs, s.CreateMessage(message, channelID))
		break
	}

	return msgs, nil
}

// CreateMessage will create a string formatted message that can be rendered
// in the Chat pane.
//
// [23:59] <erroneousboat> Hello world!
func (s *SlackService) CreateMessage(message slack.Message, channelID string) components.Message {
	var name string

	// Get username from cache
	name, ok := s.UserCache[message.User]

	// Name not in cache
	if !ok {
		if message.BotID != "" {
			name, ok = s.UserCache[message.BotID]
			if !ok {
				if message.Username != "" {
					name = message.Username
					s.UserCache[message.BotID] = message.Username
				} else {
					bot, err := s.Client.GetBotInfo(message.BotID)
					if err != nil {
						name = "unkown"
						s.UserCache[message.BotID] = name
					} else {
						name = bot.Name
						s.UserCache[message.BotID] = bot.Name
					}
				}
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
	msgTime, err := strconv.ParseFloat(message.Timestamp, 64)

	if err != nil {
		msgTime = 0.0
	}
	intTime := int64(msgTime)

	threadTime, err := strconv.ParseFloat(message.ThreadTimestamp, 64)

	if err != nil {
		threadTime = 0.0
	}

	// Format message
	msg := components.Message{
		ID:          message.Timestamp,
		MsgID:       hashID(int(msgTime)),
		ThreadID:    hashID(int(threadTime)),
		Messages:    make(map[string]components.Message),
		Time:        time.Unix(intTime, 0),
		Name:        name,
		Content:     parseMessage(s, message.Text),
		StyleTime:   s.Config.Theme.Message.Time,
		StyleThread: s.Config.Theme.Message.Thread,
		StyleName:   s.Config.Theme.Message.Name,
		StyleText:   s.Config.Theme.Message.Text,
		FormatTime:  s.Config.Theme.Message.TimeFormat,
	}

	// When there are attachments, add them to Messages
	//
	// NOTE: attachments don't have an id or a timestamp that we can
	// use as a key value for the Messages field, so we use the index
	// of the returned array.
	if len(message.Attachments) > 0 {
		atts := s.CreateMessageFromAttachments(message.Attachments)

		for i, a := range atts {
			msg.Messages[strconv.Itoa(i)] = a
		}
	}

	// When there are files, add them to Messages
	if len(message.Files) > 0 {
		files := s.CreateMessageFromFiles(message.Files)
		for _, file := range files {
			msg.Messages[file.ID] = file
		}
	}

	// When the message timestamp and thread timestamp are the same, we
	// have a parent message. This means it contains a thread with replies.
	//
	// Additionally, we set the thread timestamp in the s.ThreadCache with
	// the base62 representation of the timestamp. We do this because
	// we if we want to reply to a thread, we need to reference this
	// timestamp. Which is too long to type, we shorten it and remember the
	// reference in the cache.
	if message.ThreadTimestamp != "" && message.ThreadTimestamp == message.Timestamp {

		// Set the thread identifier for thread cache
		s.ThreadCache[msg.ThreadID] = message.ThreadTimestamp

		// Set thread prefix for message
		msg.Thread = fmt.Sprintf("%s ", msg.ThreadID)

		// Create the message replies from the thread
		replies := s.CreateMessageFromReplies(message.ThreadTimestamp, channelID)
		for _, reply := range replies {
			msg.Messages[reply.ID] = reply
		}
	} else {
		s.MessageCache[msg.MsgID] = message.Timestamp
	}

	return msg
}

// CreateMessageFromReplies will create components.Message struct from
// the conversation replies from slack.
//
// Useful documentation:
//
// https://api.slack.com/docs/message-threading
// https://api.slack.com/methods/conversations.replies
// https://godoc.org/github.com/nlopes/slack#Client.GetConversationReplies
// https://godoc.org/github.com/nlopes/slack#GetConversationRepliesParameters
func (s *SlackService) CreateMessageFromReplies(messageID string, channelID string) []components.Message {
	msgs := make([]slack.Message, 0)

	initReplies, _, initCur, err := s.Client.GetConversationReplies(
		&slack.GetConversationRepliesParameters{
			ChannelID: channelID,
			Timestamp: messageID,
			Limit:     200,
		},
	)
	if err != nil {
		log.Fatal(err) // FIXME
	}

	msgs = append(msgs, initReplies...)

	nextCur := initCur
	for nextCur != "" {
		conversationReplies, _, cursor, err := s.Client.GetConversationReplies(&slack.GetConversationRepliesParameters{
			ChannelID: channelID,
			Timestamp: messageID,
			Cursor:    nextCur,
			Limit:     200,
		})

		if err != nil {
			log.Fatal(err) // FIXME
		}

		msgs = append(msgs, conversationReplies...)
		nextCur = cursor
	}

	var replies []components.Message
	for _, reply := range msgs {
		// Because the conversations api returns an entire thread (a
		// message plus all the messages in reply), we need to check if
		// one of the replies isn't the parent that we started with.
		//
		// Keep in mind that the api returns the replies with the latest
		// as the first element.
		if reply.ThreadTimestamp != "" && reply.ThreadTimestamp == reply.Timestamp {
			continue
		}

		msg := s.CreateMessage(reply, channelID)

		// Set the thread separator
		msg.Thread = "  "

		replies = append(replies, msg)
	}

	return replies
}

// CreateMessageFromAttachments will construct an array of strings from the
// Field values of Attachments of a Message.
func (s *SlackService) CreateMessageFromAttachments(atts []slack.Attachment) []components.Message {
	var msgs []components.Message
	for _, att := range atts {
		for _, field := range att.Fields {
			msgs = append(msgs, components.Message{
				Content: fmt.Sprintf(
					"%s %s",
					field.Title,
					field.Value,
				),
				StyleTime:   s.Config.Theme.Message.Time,
				StyleThread: s.Config.Theme.Message.Thread,
				StyleName:   s.Config.Theme.Message.Name,
				StyleText:   s.Config.Theme.Message.Text,
				FormatTime:  s.Config.Theme.Message.TimeFormat,
			},
			)
		}

		if att.Pretext != "" {
			msgs = append(
				msgs,
				components.Message{
					Content:     fmt.Sprintf("%s", att.Pretext),
					StyleTime:   s.Config.Theme.Message.Time,
					StyleThread: s.Config.Theme.Message.Thread,
					StyleName:   s.Config.Theme.Message.Name,
					StyleText:   s.Config.Theme.Message.Text,
					FormatTime:  s.Config.Theme.Message.TimeFormat,
				},
			)
		}

		if att.Text != "" {
			msgs = append(
				msgs,
				components.Message{
					Content:     fmt.Sprintf("%s", att.Text),
					StyleTime:   s.Config.Theme.Message.Time,
					StyleThread: s.Config.Theme.Message.Thread,
					StyleName:   s.Config.Theme.Message.Name,
					StyleText:   s.Config.Theme.Message.Text,
					FormatTime:  s.Config.Theme.Message.TimeFormat,
				},
			)
		}

		if att.Title != "" {
			msgs = append(
				msgs,
				components.Message{
					Content:     fmt.Sprintf("%s", att.Title),
					StyleTime:   s.Config.Theme.Message.Time,
					StyleThread: s.Config.Theme.Message.Thread,
					StyleName:   s.Config.Theme.Message.Name,
					StyleText:   s.Config.Theme.Message.Text,
					FormatTime:  s.Config.Theme.Message.TimeFormat,
				},
			)
		}
	}

	return msgs
}

// CreateMessageFromFiles will create components.Message struct from
// conversation attached files
func (s *SlackService) CreateMessageFromFiles(files []slack.File) []components.Message {
	var msgs []components.Message

	for _, file := range files {
		msgs = append(msgs, components.Message{
			Content: fmt.Sprintf(
				"%s %s", file.Title, file.URLPrivate,
			),
			StyleTime:   s.Config.Theme.Message.Time,
			StyleThread: s.Config.Theme.Message.Thread,
			StyleName:   s.Config.Theme.Message.Name,
			StyleText:   s.Config.Theme.Message.Text,
			FormatTime:  s.Config.Theme.Message.TimeFormat,
		})

	}

	return msgs
}

func (s *SlackService) CreateMessageFromMessageEvent(message *slack.MessageEvent, channelID string) (components.Message, error) {
	msg := slack.Message{Msg: message.Msg}

	switch message.SubType {
	case "message_changed":
		// Append (edited) when an edited message is received
		msg = slack.Message{Msg: *message.SubMessage}
		msg.Text = fmt.Sprintf("%s (edited)", msg.Text)
	case "message_replied":
		return components.Message{}, errors.New("ignoring reply events")
	}

	return s.CreateMessage(msg, channelID), nil
}

// parseMessage will parse a message string and find and replace:
//	- emoji's
//	- mentions
//	- html unescape
func parseMessage(s *SlackService, msg string) string {
	if s.Config.Emoji {
		msg = parseEmoji(msg)
	}

	msg = parseMentions(s, msg)

	msg = html.UnescapeString(msg)

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

func hashID(input int) string {
	const base62Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	hash := ""
	for input > 0 {
		hash = string(base62Alphabet[input%62]) + hash
		input = int(input / 62)
	}

	return hash
}
