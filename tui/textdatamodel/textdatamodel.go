// Package textdatamodel provides the data model and logic for managing text data
// within the TUI application. It handles user interactions, updates the state
// based on user input, and renders the text data view.
package textdatamodel

import (
	"fmt"

	"github.com/Tomap-Tomap/GophKeeper/tui/colors"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	maxIdx  = 2
	nameIdx = 0
	metaIdx = 1
	textIdx = 2
)

// Model represents the state of the text data input form.
type Model struct {
	elementStyle        lipgloss.Style
	focusedElementStyle lipgloss.Style
	enterCmd            tea.Cmd

	name textinput.Model
	meta textinput.Model
	text textarea.Model

	id string

	focused int
}

// New creates a new Model with the provided initial values and command to execute on enter.
func New(enterCmd tea.Cmd, id, name, meta, text string) Model {
	elementStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colors.MainColor).Align(lipgloss.Center, lipgloss.Center)

	focusedElementStyle := lipgloss.NewStyle().BorderForeground(colors.FocusColor).Inherit(elementStyle)

	nameTI := textinput.New()
	nameTI.Placeholder = "Name"
	nameTI.Focus()
	nameTI.SetValue(name)

	metaTI := textinput.New()
	metaTI.Placeholder = "Meta"
	metaTI.SetValue(meta)

	textTA := textarea.New()
	textTA.SetValue(text)

	return Model{
		elementStyle:        elementStyle,
		focusedElementStyle: focusedElementStyle,
		enterCmd:            enterCmd,

		name: nameTI,
		meta: metaTI,
		text: textTA,

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
			m.focused = min(m.focused+1, maxIdx)
		case tea.KeyUp:
			m.focused = max(m.focused-1, 0)
		case tea.KeyEnter:
			return m, m.enterCmd
		}

		m.name.Blur()
		m.meta.Blur()
		m.text.Blur()

		switch m.focused {
		case nameIdx:
			m.name.Focus()
		case metaIdx:
			m.meta.Focus()
		case textIdx:
			m.text.Focus()
		}
	}

	cmds := make([]tea.Cmd, maxIdx+1)

	m.name, cmds[nameIdx] = m.name.Update(msg)
	m.meta, cmds[metaIdx] = m.meta.Update(msg)
	m.text, cmds[textIdx] = m.text.Update(msg)

	return m, tea.Batch(cmds...)
}

// View renders the model as a string.
func (m Model) View() string {
	elementsView := make([]string, maxIdx+1)

	elementsView[nameIdx] = fmt.Sprintf("%s %s", m.name.Placeholder, m.name.View())
	elementsView[metaIdx] = fmt.Sprintf("%s %s", m.meta.Placeholder, m.meta.View())
	elementsView[nameIdx] = m.elementStyle.Render(elementsView[nameIdx])
	elementsView[metaIdx] = m.elementStyle.Render(elementsView[metaIdx])
	elementsView[textIdx] = m.elementStyle.Render(m.text.View())

	switch m.focused {
	case nameIdx:
		elementsView[m.focused] = m.focusedElementStyle.Render(m.name.View())
	case metaIdx:
		elementsView[m.focused] = m.focusedElementStyle.Render(m.meta.View())
	case textIdx:
		elementsView[m.focused] = m.focusedElementStyle.Render(m.text.View())
	}

	return lipgloss.JoinVertical(lipgloss.Center, elementsView...)
}

// GetResult returns the values of all input fields in the model.
func (m Model) GetResult() (id, name, meta, text string) {
	id = m.id
	name = m.name.Value()
	meta = m.meta.Value()
	text = m.text.Value()

	return
}

func (m Model) calculateSize(msg tea.WindowSizeMsg) Model {
	newWidth := msg.Width - m.elementStyle.GetHorizontalFrameSize()
	newHeight := msg.Height/(maxIdx+1) - m.elementStyle.GetVerticalFrameSize()

	m.elementStyle = m.elementStyle.
		Width(newWidth).
		Height(newHeight)

	m.focusedElementStyle = m.focusedElementStyle.
		Width(newWidth).
		Height(newHeight)

	m.text.SetWidth(newWidth)

	return m
}
