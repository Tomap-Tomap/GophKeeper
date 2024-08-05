//go:build unit

// Package buttons provides a simple implementation of a Button struct
// that can be used with the Bubble Tea framework. The Button struct
// contains a view string and a command (cmd) of type tea.Cmd.
// It provides methods to retrieve the view and execute the command.
package buttons

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

const testView = "test button"

type testMsg struct{}

func testCmd() tea.Msg {
	return testMsg{}
}

func Test(t *testing.T) {
	b := New(testView, testCmd)

	cmd := b.Action()
	assert.Equal(t, testMsg{}, cmd())

	view := b.View()
	assert.Equal(t, testView, view)
}
