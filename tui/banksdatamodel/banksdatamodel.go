// Package banksdatamodel provides a model for handling bank card information.
package banksdatamodel

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Tomap-Tomap/GophKeeper/tui/colors"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	numInputs = 6
)

// Model represents the state of the bank card input form.
type Model struct {
	elementStyle        lipgloss.Style
	focusedElementStyle lipgloss.Style
	enterCmd            tea.Cmd

	inputs [numInputs]textinput.Model

	id string

	focused int
}

// New creates a new Model with the provided initial values and command to execute on enter.
func New(enterCmd tea.Cmd, id, name, number, cvc, owner, exp, meta string) Model {
	elementStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colors.MainColor).Align(lipgloss.Center, lipgloss.Center)

	focusedElementStyle := lipgloss.NewStyle().BorderForeground(colors.FocusColor).Inherit(elementStyle)

	inputs := [numInputs]textinput.Model{
		createTextInput("Name", name, nil, 0, 0, true),
		createTextInput("Card number", number, ccnValidator, 20, 30, false),
		createTextInput("CVC", cvc, cvvValidator, 3, 5, false),
		createTextInput("OWNER", owner, ownerValidator, 0, 0, false),
		createTextInput("MM/YY", exp, expValidator, 5, 5, false),
		createTextInput("Meta", meta, nil, 0, 0, false),
	}

	return Model{
		elementStyle:        elementStyle,
		focusedElementStyle: focusedElementStyle,
		enterCmd:            enterCmd,
		inputs:              inputs,
		id:                  id,
	}
}

func createTextInput(placeholder, value string, validator func(string) error, charLimit, width int, focus bool) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(value)
	ti.Validate = validator
	ti.CharLimit = charLimit
	ti.Width = width
	if focus {
		ti.Focus()
	}
	return ti
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

		m = m.blurAllInputs()
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
	var sb strings.Builder
	for i := range m.inputs {
		elementView := fmt.Sprintf("%s %s", m.inputs[i].Placeholder, m.inputs[i].View())
		if i == m.focused {
			sb.WriteString(m.focusedElementStyle.Render(elementView))
		} else {
			sb.WriteString(m.elementStyle.Render(elementView))
		}
		sb.WriteString("\n")
	}
	return lipgloss.JoinVertical(lipgloss.Center, sb.String())
}

// GetResult returns the values of all input fields in the model.
func (m Model) GetResult() (id, name, number, cvc, owner, exp, meta string) {
	id = m.id
	name = m.inputs[0].Value()
	number = m.inputs[1].Value()
	cvc = m.inputs[2].Value()
	owner = m.inputs[3].Value()
	exp = m.inputs[4].Value()
	meta = m.inputs[5].Value()

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

func (m Model) blurAllInputs() Model {
	for i := range m.inputs {
		m.inputs[i].Blur()
	}

	return m
}

func ccnValidator(s string) error {
	if len(s) > 16+3 {
		return fmt.Errorf("CCN is too long")
	}

	if len(s) == 0 || len(s)%5 != 0 && (s[len(s)-1] < '0' || s[len(s)-1] > '9') {
		return fmt.Errorf("CCN is invalid")
	}

	if len(s)%5 == 0 && s[len(s)-1] != ' ' {
		return fmt.Errorf("CCN must separate groups with spaces")
	}

	c := strings.ReplaceAll(s, " ", "")
	_, err := strconv.ParseInt(c, 10, 64)

	return err
}

func cvvValidator(s string) error {
	_, err := strconv.ParseInt(s, 10, 64)
	return err
}

func expValidator(s string) error {
	e := strings.ReplaceAll(s, "/", "")
	_, err := strconv.ParseInt(e, 10, 64)
	if err != nil {
		return fmt.Errorf("EXP is invalid")
	}

	if len(s) >= 3 && (strings.Index(s, "/") != 2 || strings.LastIndex(s, "/") != 2) {
		return fmt.Errorf("EXP doesn't contain /")
	}

	return nil
}

func ownerValidator(s string) error {
	if s != strings.ToUpper(s) {
		return fmt.Errorf("owner is invalid")
	}

	return nil
}
