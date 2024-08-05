// Package startmodel provides the data model and logic for the start screen
// of the TUI application. It handles user interactions, updates the state
// based on user input, and renders the start screen view.
package startmodel

import (
	"github.com/Tomap-Tomap/GophKeeper/tui/buttons"
	"github.com/Tomap-Tomap/GophKeeper/tui/colors"
	"github.com/Tomap-Tomap/GophKeeper/tui/commands"
	"github.com/Tomap-Tomap/GophKeeper/tui/textinputmodel"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	infoText = "Please choose an option"
	helpText = "↑: move up • ↓: move down • enter: select"
)

type status int

const (
	statusMain status = iota
	statusTextinput
)

// Model represents the state of the start screen.
type Model struct {
	buttonStyle        lipgloss.Style
	currentButtonStyle lipgloss.Style

	textinput tea.Model

	buttons []buttons.Button
	focused int
	status  status
}

// New creates a new Model instance for the start screen.
func New() Model {
	buttonBorder := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colors.MainColor).Align(lipgloss.Center, lipgloss.Center)

	focusedBB := lipgloss.NewStyle().BorderForeground(colors.FocusColor).Inherit(buttonBorder)

	tiLogin := textinput.New()
	tiLogin.Placeholder = "Login"

	tiPassword := textinput.New()
	tiPassword.Placeholder = "Password"
	tiPassword.EchoMode = textinput.EchoPassword

	tis := []textinput.Model{
		tiLogin,
		tiPassword,
	}

	buttons := []buttons.Button{
		buttons.New("Config", commands.OpenConfigModel()),
		buttons.New("Registration", textInputCmd(tis, registrationCmd)),
		buttons.New("Sign in", textInputCmd(tis, signinCmd)),
	}

	sm := Model{
		buttonStyle:        buttonBorder,
		currentButtonStyle: focusedBB,
		buttons:            buttons,
		status:             statusMain,
	}

	return sm
}

// Init initializes the start screen model.
func (m Model) Init() tea.Cmd {
	return commands.SetInfo(infoText, helpText)
}

// Update handles incoming messages and updates the model state accordingly.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if wm, ok := msg.(tea.WindowSizeMsg); ok {
		m = m.calculateSize(wm)
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
				cmd = m.buttons[m.focused].Action()
			}
		case textInputMsg:
			m.textinput = textinputmodel.New(msg.inputs, msg.authCmd, "Please input login and password")
			m.status = statusTextinput

			cmd = tea.Batch(
				commands.SetWindowSize(),
				m.textinput.Init(),
			)
		}
	case statusTextinput:
		m.textinput, cmd = m.textinput.Update(msg)
	}

	return m, cmd
}

// View renders the start screen view based on the current state.
func (m Model) View() string {
	var view string

	switch m.status {
	case statusMain:
		buttonsView := ""

		for idx, v := range m.buttons {
			if idx == m.focused {
				buttonsView = lipgloss.JoinVertical(lipgloss.Center, buttonsView, m.currentButtonStyle.Render((v.View())))
			} else {
				buttonsView = lipgloss.JoinVertical(lipgloss.Center, buttonsView, m.buttonStyle.Render(v.View()))
			}
		}

		view = lipgloss.JoinVertical(lipgloss.Center, buttonsView)
	case statusTextinput:
		view = m.textinput.View()
	}

	return view
}

func (m Model) calculateSize(msg tea.WindowSizeMsg) Model {
	newWidth := msg.Width - m.buttonStyle.GetHorizontalFrameSize()
	newHeight := msg.Height/len(m.buttons) - m.buttonStyle.GetVerticalFrameSize()

	m.buttonStyle = m.buttonStyle.
		Width(newWidth).
		Height(newHeight)

	m.currentButtonStyle = m.currentButtonStyle.
		Width(newWidth).
		Height(newHeight)

	return m
}
