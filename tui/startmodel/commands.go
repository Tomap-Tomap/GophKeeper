package startmodel

import (
	"github.com/Tomap-Tomap/GophKeeper/tui/messages"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func textInputCmd(inputs []textinput.Model, authCmd func(values []string) tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		return textInputMsg{
			inputs:  inputs,
			authCmd: authCmd,
		}
	}
}

func registrationCmd(values []string) tea.Cmd {
	var login, password string

	if values != nil && len(values) == 2 {
		login, password = values[0], values[1]
	}

	return func() tea.Msg {
		return messages.Registration{
			Login:    login,
			Password: password,
		}
	}
}

func signinCmd(values []string) tea.Cmd {
	var login, password string

	if values != nil && len(values) == 2 {
		login, password = values[0], values[1]
	}

	return func() tea.Msg {
		return messages.SignIn{
			Login:    login,
			Password: password,
		}
	}
}
