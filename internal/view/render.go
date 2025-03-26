package view

import (
	"github.com/copygen/copygen/internal/arguments"
)

// Renderer interface with a unified Render method.
type Renderer interface {
	Render(input string)
}

func NewRenderer(vt arguments.ViewType, view *View) Renderer {
	switch vt {
	case arguments.ViewHuman:
		return &HumanRenderer{view}
	default:
		panic("unknown view type")
	}
}

// HumanRenderer for writing human-readable output.
type HumanRenderer struct {
	view *View
}

// Validate that HumanRenderer implements the Renderer interface.
var _ Renderer = (*HumanRenderer)(nil)

// NewHumanRenderer creates a HumanRenderer with a "human" view bound to an output stream.
func NewHumanRenderer(view *View) *HumanRenderer {
	return &HumanRenderer{
		view: view,
	}
}

func (v *HumanRenderer) Render(input string) {
	_, err := v.view.Stream.Writer.Write([]byte(input))
	if err != nil {
		panic(err)
	}
}
