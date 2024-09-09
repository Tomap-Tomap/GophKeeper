// Package textinputmodel provides a model for handling multiple text inputs
// with focus management and custom styling.
package textinputmodel

import (
	"fmt"

	"github.com/Tomap-Tomap/GophKeeper/tui/colors"
	"github.com/Tomap-Tomap/GophKeeper/tui/commands"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const helpText = "↑: move up • ↓: move down • enter: apply"

// ReturnFunc is a type for the function that handles the return command.
type ReturnFunc func(values []string) tea.Cmd

// Model represents the state and behavior of the text input model.
type Model struct {
	infoText          string
	inputStyle        lipgloss.Style
	focusedInputStyle lipgloss.Style

	inputs    []textinput.Model
	returnCmd ReturnFunc
	focused   int
}

// New initializes a new Model with the given inputs, return command and text for info.
func New(inputs []textinput.Model, returnCmd ReturnFunc, infoText string) Model {
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colors.MainColor).Align(lipgloss.Center, lipgloss.Center)

	focusedInputStyle := lipgloss.NewStyle().BorderForeground(colors.FocusColor).Inherit(inputStyle)

	return Model{
		infoText:          infoText,
		inputStyle:        inputStyle,
		focusedInputStyle: focusedInputStyle,

		inputs:    inputs,
		returnCmd: returnCmd,
	}
}

// Init initializes the model and returns an initial command.
func (m Model) Init() tea.Cmd {
	return commands.SetInfo(m.infoText, helpText)
}

// Update handles incoming messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if wm, ok := msg.(tea.WindowSizeMsg); ok {
		m = m.calculateSize(wm)
	}

	cmds := make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			returnData := make([]string, 0, len(m.inputs))

			for _, v := range m.inputs {
				returnData = append(returnData, v.Value())
			}

			return m, m.returnCmd(returnData)
		case tea.KeyDown:
			m.focused = min(m.focused+1, len(m.inputs)-1)
		case tea.KeyUp:
			m.focused = max(m.focused-1, 0)
		}

		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return m, tea.Batch(cmds...)
}

// View renders the model's view.
func (m Model) View() string {
	var inputsView string

	for idx, v := range m.inputs {
		view := fmt.Sprintf("%s: %s", v.Placeholder, v.View())
		if idx == m.focused {
			inputsView = lipgloss.JoinVertical(lipgloss.Center, inputsView, m.focusedInputStyle.Render(view))
		} else {
			inputsView = lipgloss.JoinVertical(lipgloss.Center, inputsView, m.inputStyle.Render(view))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Center, inputsView)
}

func (m Model) calculateSize(msg tea.WindowSizeMsg) Model {
	newWidth := msg.Width - m.inputStyle.GetHorizontalFrameSize()
	newHeight := msg.Height/len(m.inputs) - m.inputStyle.GetVerticalFrameSize()

	m.inputStyle = m.inputStyle.
		Width(newWidth).
		Height(newHeight)

	m.focusedInputStyle = m.focusedInputStyle.
		Width(newWidth).
		Height(newHeight)

	return m
}
