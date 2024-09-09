package startmodel

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type textInputMsg struct {
	inputs  []textinput.Model
	authCmd func(values []string) tea.Cmd
}
