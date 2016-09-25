package components

import (
	"github.com/erroneousboat/slack-term/src/service"
	"github.com/gizak/termui"
)

type Channels struct {
	List *termui.List
}

func CreateChannels(svc *service.SlackService, inputHeight int) *Channels {
	channels := &Channels{
		List: termui.NewList(),
	}

	channels.List.BorderLabel = "Channels"
	channels.List.Overflow = "wrap"
	channels.List.Height = termui.TermHeight() - inputHeight

	channels.GetChannels(svc)

	return channels
}

// Buffer implements interface termui.Bufferer
func (c *Channels) Buffer() termui.Buffer {
	buf := c.List.Buffer()
	return buf
}

// GetHeight implements interface termui.GridBufferer
func (c *Channels) GetHeight() int {
	return c.List.Block.GetHeight()
}

// SetWidth implements interface termui.GridBufferer
func (c *Channels) SetWidth(w int) {
	c.List.SetWidth(w)
}

// SetX implements interface termui.GridBufferer
func (c *Channels) SetX(x int) {
	c.List.SetX(x)
}

// SetY implements interface termui.GridBufferer
func (c *Channels) SetY(y int) {
	c.List.SetY(y)
}

func (c *Channels) GetChannels(svc *service.SlackService) {
	for _, slackChan := range svc.GetChannels() {
		c.List.Items = append(c.List.Items, slackChan.Name)
	}
}
