// Package passworddatamodel provides a model for handling password-related information
// within the TUI application. It integrates various components such as text inputs
// to create a cohesive user interface for managing passwords.
package passworddatamodel

import (
	"fmt"

	"github.com/Tomap-Tomap/GophKeeper/tui/colors"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const inputsCount = 4

// Model represents the state of the password input form.
type Model struct {
	elementStyle        lipgloss.Style
	focusedElementStyle lipgloss.Style
	enterCmd            tea.Cmd

	inputs [inputsCount]textinput.Model

	id string

	focused int
}

// New creates a new Model with the provided initial values and command to execute on enter.
func New(enterCmd tea.Cmd, id, name, login, password, meta string) Model {
	elementStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colors.MainColor).Align(lipgloss.Center, lipgloss.Center)

	focusedElementStyle := lipgloss.NewStyle().BorderForeground(colors.FocusColor).Inherit(elementStyle)

	nameTI := textinput.New()
	nameTI.Placeholder = "Name"
	nameTI.Focus()
	nameTI.SetValue(name)

	loginTI := textinput.New()
	loginTI.Placeholder = "Login"
	loginTI.SetValue(login)

	passwordTI := textinput.New()
	passwordTI.Placeholder = "Password"
	passwordTI.SetValue(password)

	metaTI := textinput.New()
	metaTI.Placeholder = "Meta"
	metaTI.SetValue(meta)

	return Model{
		elementStyle:        elementStyle,
		focusedElementStyle: focusedElementStyle,
		enterCmd:            enterCmd,

		inputs: [inputsCount]textinput.Model{
			nameTI,
			loginTI,
			passwordTI,
			metaTI,
		},

		id: id,
	}
}

// Init initializes the model and returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the model state accordingly.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if wm, ok := msg.(tea.WindowSizeMsg); ok {
		m = m.calculateSize(wm)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyDown:
			m.focused = min(m.focused+1, len(m.inputs)-1)
		case tea.KeyUp:
			m.focused = max(m.focused-1, 0)
		case tea.KeyEnter:
			return m, m.enterCmd
		}

		for i := range m.inputs {
			m.inputs[i].Blur()
		}

		m.inputs[m.focused].Focus()
	}

	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return m, tea.Batch(cmds...)
}

// View renders the model as a string.
func (m Model) View() string {
	elementsView := make([]string, len(m.inputs))

	for i := range m.inputs {
		elementsView[i] = fmt.Sprintf("%s %s", m.inputs[i].Placeholder, m.inputs[i].View())

		if i == m.focused {
			elementsView[i] = m.focusedElementStyle.Render(elementsView[i])
		} else {
			elementsView[i] = m.elementStyle.Render(elementsView[i])
		}
	}

	return lipgloss.JoinVertical(lipgloss.Center, elementsView...)
}

// GetResult returns the values of all input fields in the model.
func (m Model) GetResult() (id, name, login, password, meta string) {
	id = m.id
	name = m.inputs[0].Value()
	login = m.inputs[1].Value()
	password = m.inputs[2].Value()
	meta = m.inputs[3].Value()

	return
}

func (m Model) calculateSize(msg tea.WindowSizeMsg) Model {
	newWidth := msg.Width - m.elementStyle.GetHorizontalFrameSize()
	newHeight := msg.Height/len(m.inputs) - m.elementStyle.GetVerticalFrameSize()

	m.elementStyle = m.elementStyle.
		Width(newWidth).
		Height(newHeight)

	m.focusedElementStyle = m.focusedElementStyle.
		Width(newWidth).
		Height(newHeight)

	return m
}
