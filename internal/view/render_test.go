package view

import (
	"bytes"
	"testing"

	"github.com/copygen/copygen/internal/arguments"
	"github.com/stretchr/testify/assert"
)

// TestNewRenderer_human tests the NewRenderer function, which should return a HumanRenderer
// and bind provided io.Writer to the view's stream writer.
func TestNewRenderer_human(t *testing.T) {
	b := bytes.Buffer{}
	hv := NewRenderer(arguments.ViewHuman, NewView(&b))

	// Check that the view is a HumanRenderer
	humanRenderer, ok := hv.(*HumanRenderer)
	assert.True(t, ok, "Expected hv to be of type *HumanRenderer")

	assert.IsType(t, &HumanRenderer{}, humanRenderer)

	// Check that the view's stream writer is the same as the buffer
	assert.Equal(t, &b, humanRenderer.view.Stream.Writer)
}
