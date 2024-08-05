// Package buttons provides a simple implementation of a Button struct
// that can be used with the Bubble Tea framework. The Button struct
// contains a view string and a command (cmd) of type tea.Cmd.
// It provides methods to retrieve the view and execute the command.
package buttons

import tea "github.com/charmbracelet/bubbletea"

// Button represents a UI button with a view and an associated command.
type Button struct {
	view string
	cmd  tea.Cmd
}

// New creates a new Button with the given view and command.
func New(view string, cmd tea.Cmd) Button {
	return Button{
		view: view,
		cmd:  cmd,
	}
}

// Action returns the command associated with the Button.
func (b Button) Action() tea.Cmd {
	return b.cmd
}

// View returns the view string of the Button.
func (b Button) View() string {
	return b.view
}
