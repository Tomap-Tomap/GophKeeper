// Package filesdatamodel provides the data model for handling file-related operations
// within the TUI application. It integrates various components such as buttons,
// text inputs, and file pickers to create a cohesive user interface for file management.
package filesdatamodel

import (
	"fmt"
	"strings"

	"github.com/Tomap-Tomap/GophKeeper/tui/buttons"
	"github.com/Tomap-Tomap/GophKeeper/tui/colors"
	"github.com/Tomap-Tomap/GophKeeper/tui/commands"
	"github.com/Tomap-Tomap/GophKeeper/tui/filepickermodel"
	"github.com/Tomap-Tomap/GophKeeper/tui/textinputmodel"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type status int

const (
	statusMain status = iota
	statusFilepicker
	statusTextinput
)

type buttonValue struct {
	button buttons.Button
	value  string
}

const buttonsCount = 3

// Model represents the main data structure for the file data model.
type Model struct {
	elementStyle        lipgloss.Style
	focusedElementStyle lipgloss.Style
	enterCmd            tea.Cmd

	filepicker tea.Model
	textinput  tea.Model

	buttons [buttonsCount]buttonValue
	id      string

	focused int
	status  status
}

// New creates a new Model instance.
func New(enterCmd tea.Cmd, id, name, meta, fileText string, isDirPicker bool) Model {
	elementStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colors.MainColor).Align(lipgloss.Center, lipgloss.Center)

	focusedElementStyle := lipgloss.NewStyle().BorderForeground(colors.FocusColor).Inherit(elementStyle)

	nameTI := textinput.New()
	nameTI.Placeholder = "Name"
	nameTI.Focus()

	metaTI := textinput.New()
	metaTI.Placeholder = "Meta"

	buttons := [buttonsCount]buttonValue{
		{button: buttons.New("Name:", textInputCmd([]textinput.Model{nameTI})), value: name},
		{button: buttons.New("Meta:", textInputCmd([]textinput.Model{metaTI})), value: meta},
		{button: buttons.New(fileText, filePickerCmd(isDirPicker))},
	}

	return Model{
		elementStyle:        elementStyle,
		focusedElementStyle: focusedElementStyle,
		enterCmd:            enterCmd,

		buttons: buttons,

		id: id,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model accordingly.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m = m.calculateSize(msg)
	case choosePathMsg:
		m.status = statusMain
		m.buttons[m.focused].value = msg.path

		return m, tea.Batch(m.enterCmd, tea.ClearScreen)
	case closeTextInputMsg:
		m.status = statusMain
		m.buttons[m.focused].value = msg.value
		return m, tea.ClearScreen
	}

	switch m.status {
	case statusMain:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyDown:
				m.focused = min(m.focused+1, len(m.buttons)-1)
			case tea.KeyUp:
				m.focused = max(m.focused-1, 0)
			case tea.KeyEnter:
				cmd = m.buttons[m.focused].button.Action()
			}
		case filePickerMsg:
			fp, err := filepickermodel.New(choosePathCmd, msg.isDirPicker)

			if err != nil {
				return m, commands.Error(err)
			}

			m.filepicker = fp
			m.status = statusFilepicker
			cmd = tea.Batch(
				commands.SetWindowSize(),
				m.filepicker.Init(),
			)
		case textInputMsg:
			m.textinput = textinputmodel.New(msg.inputs, closeTextInputCmd, "Please input field")
			m.status = statusTextinput

			cmd = tea.Batch(
				commands.SetWindowSize(),
				m.textinput.Init(),
			)
		}
	case statusFilepicker:
		m.filepicker, cmd = m.filepicker.Update(msg)
	case statusTextinput:
		m.textinput, cmd = m.textinput.Update(msg)
	}

	return m, cmd
}

// View renders the view based on the current status.
func (m Model) View() string {
	switch m.status {
	case statusMain:
		var sb strings.Builder

		for idx, v := range m.buttons {
			view := fmt.Sprintf("%s %s", v.button.View(), v.value)
			if idx == m.focused {
				sb.WriteString(m.focusedElementStyle.Render(view))
			} else {
				sb.WriteString(m.elementStyle.Render(view))
			}
			sb.WriteString("\n")
		}

		return sb.String()
	case statusFilepicker:
		return m.filepicker.View()
	case statusTextinput:
		return m.textinput.View()
	}

	return ""
}

// GetResult returns the results from the model.
func (m Model) GetResult() (id, name, path, meta string) {
	id = m.id
	name = m.buttons[0].value
	path = m.buttons[2].value
	meta = m.buttons[1].value

	return
}

func (m Model) calculateSize(msg tea.WindowSizeMsg) Model {
	newWidth := msg.Width - m.elementStyle.GetHorizontalFrameSize()
	newHeight := msg.Height/len(m.buttons) - m.elementStyle.GetVerticalFrameSize()

	m.elementStyle = m.elementStyle.
		Width(newWidth).
		Height(newHeight)

	m.focusedElementStyle = m.focusedElementStyle.
		Width(newWidth).
		Height(newHeight)

	return m
}
