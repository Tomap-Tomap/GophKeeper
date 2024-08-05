// Package commands provides a set of functions that return commands
// for the Bubble Tea framework. These commands are used to manage
// window size, handle errors, set information messages, and open or
// close configuration models in the TUI application.
package commands

import (
	"os"

	"github.com/Tomap-Tomap/GophKeeper/tui/messages"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
)

// SetWindowSize returns a command that retrieves the current window size
// and sends a WindowSizeMsg with the width and height.
func SetWindowSize() tea.Cmd {
	width, height, err := term.GetSize(os.Stdin.Fd())

	if err != nil {
		return Error(err)
	}

	return func() tea.Msg {
		return tea.WindowSizeMsg{
			Width:  width,
			Height: height,
		}
	}
}

// Error returns a command that sends an error message.
func Error(err error) tea.Cmd {
	return func() tea.Msg {
		return messages.Error{
			Err: err,
		}
	}
}

// SetInfo returns a command that sends an information message with
// the provided info and help strings.
func SetInfo(info, help string) tea.Cmd {
	return func() tea.Msg {
		return messages.Info{
			Info: info,
			Help: help,
		}
	}
}

// OpenConfigModel returns a command that sends a message to open
// the configuration model.
func OpenConfigModel() tea.Cmd {
	return func() tea.Msg {
		return messages.OpenConfigModel{}
	}
}

// CloseConfigModel returns a command that sends a message to close
// the configuration model with the provided path to the key and
// address to the service.
func CloseConfigModel(pathToKey, addrToService string) tea.Cmd {
	return func() tea.Msg {
		return messages.CloseConfigModel{
			PathToKey:     pathToKey,
			AddrToService: addrToService,
		}
	}
}
