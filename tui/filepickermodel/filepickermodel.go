// Package filepickermodel provides a model for a file picker interface
// using the Bubble Tea framework. It allows users to select files or
// directories from their file system.
package filepickermodel

import (
	"fmt"
	"os"

	"github.com/Tomap-Tomap/GophKeeper/tui/commands"
	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	infoTex  = "Please choose path"
	helpText = "g: first • G: last • ↑/k/ctrl+p: move up • ↓/j/ctrl+n: move down • K/pgup: page down • J/pgdown: page down • ←/h/backspace/esc: back • →/l/enter: open/select"
)

// ReturnFunc is a function type that takes a file path and returns a tea.Cmd.
type ReturnFunc func(path string) tea.Cmd

// Model represents the file picker model.
type Model struct {
	filepicker filepicker.Model
	returnCmd  ReturnFunc
}

// New creates a new file picker model. It takes a ReturnFunc and a boolean indicating if the picker is for directories.
func New(returnCmd ReturnFunc, isDirPicker bool) (Model, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return Model{}, fmt.Errorf("cannot get user home dir: %w", err)
	}

	fp := filepicker.New()
	fp.AutoHeight = true
	fp.CurrentDirectory = homeDir
	fp.FileAllowed = true
	fp.DirAllowed = false

	if isDirPicker {
		fp.FileAllowed = false
		fp.DirAllowed = true
	}

	return Model{
		filepicker: fp,
		returnCmd:  returnCmd,
	}, nil
}

// Init initializes the file picker model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.filepicker.Init(), commands.SetInfo(infoTex, helpText))
}

// Update updates the model based on messages and handles file selection.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.filepicker, cmd = m.filepicker.Update(msg)

	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		cmd = m.returnCmd(path)
	}

	return m, cmd
}

// View returns the string representation of the file picker view.
func (m Model) View() string {
	return m.filepicker.View()
}
