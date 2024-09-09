package filesdatamodel

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func filePickerCmd(isDirPicker bool) tea.Cmd {
	return func() tea.Msg {
		return filePickerMsg{
			isDirPicker: isDirPicker,
		}
	}
}

func choosePathCmd(path string) tea.Cmd {
	return func() tea.Msg {
		return choosePathMsg{
			path: path,
		}
	}
}

func textInputCmd(inputs []textinput.Model) tea.Cmd {
	return func() tea.Msg {
		return textInputMsg{
			inputs: inputs,
		}
	}
}

func closeTextInputCmd(values []string) tea.Cmd {
	var value string

	if values != nil {
		value = values[0]
	}

	return func() tea.Msg {
		return closeTextInputMsg{
			value: value,
		}
	}
}
