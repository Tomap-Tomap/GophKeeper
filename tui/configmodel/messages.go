package configmodel

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type filePickerMsg struct {
	isDirPicker   bool
	choosePathCmd func(path string) tea.Cmd
}

type choosePathMsg struct {
	path string
}

type textInputMsg struct {
	inputs []textinput.Model
}

type closeTextInputMsg struct {
	value string
}

type createKeyMsg struct {
	path string
}

type closeConfigMsg struct {
}
