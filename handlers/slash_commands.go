package handlers

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/erroneousboat/slack-term/context"
	"github.com/erroneousboat/termui"
	"github.com/nlopes/slack"
)

func deleteCommandHandler(ctx *context.AppContext, channelID, cmdParams string) (ok bool, err error) {
	r := regexp.MustCompile(`(?P<id>\w+)$`)
	cmdParams = strings.TrimSpace(cmdParams)

	messageID := r.FindString(cmdParams)
	if messageID == "" {
		return false, errors.New("Please provide a message ID. E.g. /delete aFksE8")
	}

	var ts string
	if ts, err = ctx.Service.CacheFetch(messageID); err != nil {
		return
	}

	if err = ctx.Service.DeleteMessage(channelID, messageID); err != nil {
		return
	}

	ctx.View.Chat.DeleteMessage(ts)
	termui.Render(ctx.View.Chat)

	ok = true

	return
}

// editCommandHandler accepts user input of the form `msgID` and `msgID Updated message`
//
// When an updated message is not provided, the msg identified by msgID is fetched from the slack service
// and placed in the user input component to be edited by the user.
//
// When an updated message is provided, the update to the message identified by msgID is sent to the Slack
// service to be persisited.
func editCommandHandler(ctx *context.AppContext, channelID, cmdParams string) (ok bool, err error) {

	r := regexp.MustCompile(`(?P<id>\w+)\s*(?P<msg>.*)$`)

	cmdParams = strings.TrimSpace(cmdParams)
	subMatch := r.FindStringSubmatch(cmdParams)

	if len(subMatch) == 0 {
		err = errors.New("/edit command malformed")
		return
	}

	msgID := subMatch[1]
	msg := subMatch[2]

	// if no message was provided, then we need to fetch the message identified by internalMsgID first
	// and render its contents in the input field
	if msg == "" {

		var ts string
		if ts, err = ctx.Service.CacheFetch(msgID); err != nil {
			return
		}

		// check if the user has a thread selected
		var parentID string
		if thr, ok := ctx.View.Threads.GetSelectedChannel(); ok && thr.ID != channelID {
			parentID = thr.ID
		} else {
			parentID = ts
		}

		if m, found := ctx.View.Chat.GetMessage(parentID, ts); found {
			editInput := fmt.Sprintf("/edit %s %s", msgID, m.Content)
			ctx.View.Input.SetText(editInput)
			termui.Render(ctx.View.Input)
		} else {
			err = fmt.Errorf("Sorry. Unable to find message with ID: '%s'", ts)
			return
		}

		ok = true

		return
	}

	var ts string
	if ts, err = ctx.Service.CacheFetch(msgID); err != nil {
		return
	}

	if err = ctx.Service.UpdateChat(channelID, ts, msg); err != nil {
		return
	}

	ok = true

	return
}

// threadCommandHandler accepts user input of the form `/thread msgID message`.
// The message is threaded under the message identified by msgID. If msgID is not yet a threaded
// message, the new thread is created.
func threadCommandHandler(ctx *context.AppContext, channelID, cmd string) (ok bool, err error) {

	r := regexp.MustCompile(`(?P<id>\w+)\s+(?P<msg>.*)$`)
	cmd = strings.TrimSpace(cmd)
	subMatch := r.FindStringSubmatch(cmd)

	if len(subMatch) == 0 {
		err = errors.New("/thread command malformed")
		return
	}

	msgID := subMatch[1]
	msg := subMatch[2]

	var ts string
	if ts, err = ctx.Service.CacheFetch(msgID); err != nil {
		return
	}

	if err = ctx.Service.SendReply(channelID, ts, msg); err != nil {
		return
	}

	ok = true

	return
}

// defaultCommandHandler is a catch-all for slash commands that are unrecognized by slack-term. These commands
// are passed through to the Slack service.
func defaultCommandHandler(ctx *context.AppContext, channelID, cmd string) (ok bool, err error) {

	r := regexp.MustCompile(`(?P<cmd>^/\w+) (?P<text>.*)`)
	cmd = strings.TrimSpace(cmd)
	subMatch := r.FindStringSubmatch(cmd)

	if len(subMatch) < 3 {
		err = errors.New("slash command malformed")
		return
	}

	c := subMatch[1]
	text := subMatch[2]

	msgOption := slack.UnsafeMsgOptionEndpoint(
		fmt.Sprintf("%s%s", slack.APIURL, "chat.command"),
		func(urlValues url.Values) {
			urlValues.Add("command", c)
			urlValues.Add("text", text)
		},
	)

	_, _, err = ctx.Service.Client.PostMessage(channelID, msgOption)
	if err != nil {
		ok = false
		return
	}

	ok = true

	return
}
