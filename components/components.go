package components

import "github.com/erroneousboat/gocui"

// TODO: documentation
// Component
type Component struct {
	Name   string
	X      int
	Y      int
	Width  int
	Height int
	View   *gocui.View
}
