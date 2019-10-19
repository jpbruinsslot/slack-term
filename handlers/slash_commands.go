package handlers

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"github.com/erroneousboat/slack-term/components"
	"github.com/erroneousboat/slack-term/context"
	"github.com/erroneousboat/termui"
	"github.com/nlopes/slack"
)

// editCommandHandler accepts user input of the form `/edit msgID` and `/edit msgID Updated message`
//
// When an updated message is not provided, the msg identified by msgID is fetched from the slack service
// and placed in the user input component to be edited by the user.
//
// When an updated message is provided, the update to the message identified by msgID is sent to the Slack
// service to be persisited.
func editCommandHandler(ctx *context.AppContext, channelID, cmd string) (ok bool, err error) {

	r := regexp.MustCompile(`(?P<cmd>^/\w+) (?P<id>\w+) (?P<msg>.*)`)
	subMatch := r.FindStringSubmatch(cmd)

	// check if both the message ID to be edited and the updated message are present
	if len(subMatch) < 3 {

		r := regexp.MustCompile(`(?P<cmd>^/\w+) (?P<id>\w+)$`)
		subMatch := r.FindStringSubmatch(cmd)

		// if only the message ID was provided, fetch the message contents and place it in the input view
		if len(subMatch) == 3 {

			internalMsgID := subMatch[2]
			var ID string
			if threadID, found := ctx.Service.ThreadCache[internalMsgID]; found {
				ID = threadID
			} else if msgID, found := ctx.Service.MessageCache[internalMsgID]; found {
				ID = msgID
			}
			var msgs []components.Message

			if msgs, err = ctx.Service.GetMessageByID(ID, channelID); ID == "" || err != nil {
				ok = false
				err = fmt.Errorf("Sorry. We were not able to find the message with ID: '%s'", ID)
				return
			}

			editInput := fmt.Sprintf("/edit %s %s", internalMsgID, msgs[0].Content)
			ctx.View.Input.SetText(editInput)

			termui.Render(ctx.View.Input)

			ok = true

			return
		}

		err = errors.New("slash command malformed")
		return
	}

	ID := subMatch[2]
	msg := subMatch[3]

	if threadID, found := ctx.Service.ThreadCache[ID]; found {
		err = ctx.Service.UpdateChat(channelID, threadID, msg)
		if err != nil {
			return
		}
	} else if msgID, found := ctx.Service.MessageCache[ID]; found {
		err = ctx.Service.UpdateChat(channelID, msgID, msg)
		if err != nil {
			return
		}
	}

	ok = true

	return

}

// threadCommandHandler accepts user input of the form `/thread msgID message`.
// The message is threaded under the message identified by msgID. If msgID is not yet a threaded
// message, the new thread is created.
func threadCommandHandler(ctx *context.AppContext, channelID, cmd string) (ok bool, err error) {

	r := regexp.MustCompile(`(?P<cmd>^/\w+) (?P<id>\w+) (?P<msg>.*)`)
	subMatch := r.FindStringSubmatch(cmd)

	if len(subMatch) < 4 {
		err = errors.New("'/thread' command malformed")
		return
	}

	msg := subMatch[3]

	if threadID, found := ctx.Service.ThreadCache[subMatch[2]]; found {
		err = ctx.Service.SendReply(channelID, threadID, msg)
		if err != nil {
			return
		}
	} else if msgID, found := ctx.Service.MessageCache[subMatch[2]]; found {
		err = ctx.Service.SendReply(channelID, msgID, msg)
		if err != nil {
			return
		}
	}

	ok = true

	return
}

// defaultCommandHandler is a catch-all for slash commands that are unrecognized by slack-term. These commands
// are passed through to the Slack service.
func defaultCommandHandler(ctx *context.AppContext, channelID, cmd string) (ok bool, err error) {

	r := regexp.MustCompile(`(?P<cmd>^/\w+) (?P<text>.*)`)
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
