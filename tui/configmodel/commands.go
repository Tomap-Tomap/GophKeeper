package configmodel

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func filePickerCmd(isDirPicker bool, choosePathCmd func(path string) tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		return filePickerMsg{
			isDirPicker:   isDirPicker,
			choosePathCmd: choosePathCmd,
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

func createKeyCmd(path string) tea.Cmd {
	return func() tea.Msg {
		return createKeyMsg{
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

func closeConfigCmd() tea.Cmd {
	return func() tea.Msg {
		return closeConfigMsg{}
	}
}
