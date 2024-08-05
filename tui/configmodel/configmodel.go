package configmodel

import (
	"crypto/aes"
	"fmt"
	"strings"

	"github.com/Tomap-Tomap/GophKeeper/crypto"
	"github.com/Tomap-Tomap/GophKeeper/tui/buttons"
	"github.com/Tomap-Tomap/GophKeeper/tui/colors"
	"github.com/Tomap-Tomap/GophKeeper/tui/commands"
	"github.com/Tomap-Tomap/GophKeeper/tui/config"
	"github.com/Tomap-Tomap/GophKeeper/tui/filepickermodel"
	"github.com/Tomap-Tomap/GophKeeper/tui/textinputmodel"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	infoText = "Please configurate configuration"
	helpText = "↑: move up • ↓: move down • enter: select"
)

const buttonCount = 4

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

type Model struct {
	elementStyle        lipgloss.Style
	focusedElementStyle lipgloss.Style

	filepicker tea.Model
	textinput  tea.Model

	buttons [buttonCount]buttonValue
	focused int
	status  status
}

func New(config *config.Config) Model {
	elementStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colors.MainColor).Align(lipgloss.Center, lipgloss.Center)

	focusedElementStyle := lipgloss.NewStyle().BorderForeground(colors.FocusColor).Inherit(elementStyle)

	ti := textinput.New()
	ti.Placeholder = "Address to service"

	buttons := [buttonCount]buttonValue{
		{button: buttons.New("Path to secret key:", filePickerCmd(false, choosePathCmd)), value: config.PathToSecretKey},
		{button: buttons.New("Address to service:", textInputCmd([]textinput.Model{ti})), value: config.AddrToService},
		{button: buttons.New("Create secret key", filePickerCmd(true, createKeyCmd))},
		{button: buttons.New("Save and close", closeConfigCmd())},
	}

	return Model{
		elementStyle:        elementStyle,
		focusedElementStyle: focusedElementStyle,

		buttons: buttons,
		status:  statusMain,
	}
}

func (m Model) Init() tea.Cmd {
	return commands.SetInfo(infoText, helpText)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m = m.calculateSize(msg)
	case choosePathMsg:
		m.status = statusMain
		m.buttons[m.focused].value = msg.path

		return m, tea.Batch(commands.SetInfo(infoText, ""), tea.ClearScreen)
	case closeTextInputMsg:
		m.status = statusMain
		m.buttons[m.focused].value = msg.value
		return m, tea.Batch(commands.SetInfo(infoText, ""), tea.ClearScreen)
	case createKeyMsg:
		m.status = statusMain
		_, path, err := crypto.NewCrypter(aes.BlockSize*2, msg.path)

		if err != nil {
			return m, commands.Error(err)
		}

		m.buttons[0].value = path
		return m, tea.ClearScreen
	case closeConfigMsg:
		return m, tea.Batch(
			commands.CloseConfigModel(
				m.buttons[0].value,
				m.buttons[1].value,
			),
			tea.ClearScreen,
		)
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
			fp, err := filepickermodel.New(msg.choosePathCmd, msg.isDirPicker)

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
			m.textinput = textinputmodel.New(msg.inputs, closeTextInputCmd, "Please input address")
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
