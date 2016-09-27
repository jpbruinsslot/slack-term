package components

import (
	"github.com/erroneousboat/slack-term/src/service"
	"github.com/gizak/termui"
)

type Channels struct {
	List            *termui.List
	SlackChannels   []SlackChannel
	SelectedChannel int
}

type SlackChannel struct {
	Name string
	ID   string
}

func CreateChannels(svc *service.SlackService, inputHeight int) *Channels {
	channels := &Channels{
		List: termui.NewList(),
	}

	channels.List.BorderLabel = "Channels"
	channels.List.Height = termui.TermHeight() - inputHeight

	// TODO: should be SetSelectedChannel
	channels.SelectedChannel = 4

	channels.GetChannels(svc)

	return channels
}

// Buffer implements interface termui.Bufferer
func (c *Channels) Buffer() termui.Buffer {
	buf := c.List.Buffer()

	for y, item := range c.List.Items {
		var cells []termui.Cell
		if y == c.SelectedChannel {
			cells = termui.DefaultTxBuilder.Build(
				item, termui.ColorBlack, termui.ColorWhite)
		} else {
			cells = termui.DefaultTxBuilder.Build(
				item, c.List.ItemFgColor, c.List.ItemBgColor)
		}

		cells = termui.DTrimTxCls(cells, c.List.InnerWidth())

		x := 0
		for _, cell := range cells {
			width := cell.Width()
			buf.Set(
				c.List.InnerBounds().Min.X+x,
				c.List.InnerBounds().Min.Y+y,
				cell,
			)
			x += width
		}
	}

	return buf
}

func (c *Channels) Add() {

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
		c.SlackChannels = append(
			c.SlackChannels,
			SlackChannel{
				ID:   slackChan.ID,
				Name: slackChan.Name,
			},
		)
	}
}

func (c *Channels) SetSelectedChannel(num int) {
	c.SelectedChannel = num
}

func (c *Channels) MoveCursorUp() {
	if c.SelectedChannel > 0 {
		c.SetSelectedChannel(c.SelectedChannel - 1)
	}
}

func (c *Channels) MoveCursorDown() {
	if c.SelectedChannel < len(c.List.Items)-1 {
		c.SetSelectedChannel(c.SelectedChannel + 1)
	}
}

func (c *Channels) NewMessage() {
}
