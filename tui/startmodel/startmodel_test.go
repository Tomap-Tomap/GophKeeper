//go:build unit

package startmodel

import (
	"testing"

	"github.com/Tomap-Tomap/GophKeeper/tui/commands"
	"github.com/Tomap-Tomap/GophKeeper/tui/messages"
	"github.com/Tomap-Tomap/GophKeeper/tui/textinputmodel"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	model := New()

	openConfigCmd := model.buttons[0].Action()
	assert.IsType(t, messages.OpenConfigModel{}, openConfigCmd())

	textInputCmd := model.buttons[1].Action()
	assert.IsType(t, textInputMsg{}, textInputCmd())

	textInputCmd = model.buttons[2].Action()
	assert.IsType(t, textInputMsg{}, textInputCmd())
}

func TestInit(t *testing.T) {
	model := New()
	cmd := model.Init()

	assert.NotNil(t, cmd)
	assert.IsType(t, commands.SetInfo("", ""), cmd)
}

func TestUpdate(t *testing.T) {
	model := New()

	t.Run("window size msg", func(t *testing.T) {
		msg := tea.WindowSizeMsg{
			Width:  100,
			Height: 100,
		}

		wantWidth := msg.Width - model.buttonStyle.GetHorizontalFrameSize()
		wantHeight := msg.Height/3 - model.buttonStyle.GetVerticalFrameSize()

		m, _ := model.Update(msg)
		assert.Equal(t, wantHeight, m.(Model).buttonStyle.GetHeight())
		assert.Equal(t, wantHeight, m.(Model).currentButtonStyle.GetHeight())
		assert.Equal(t, wantWidth, m.(Model).buttonStyle.GetWidth())
		assert.Equal(t, wantWidth, m.(Model).buttonStyle.GetWidth())
	})

	t.Run("key down test", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyDown}
		m, _ := model.Update(msg)
		assert.Equal(t, 1, m.(Model).focused)
	})

	t.Run("key up tets", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyUp}
		m, _ := model.Update(msg)
		assert.Equal(t, 0, m.(Model).focused)
	})

	t.Run("enter tets", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		wantCmd := model.buttons[model.focused].Action()
		_, cmd := model.Update(msg)

		assert.IsType(t, wantCmd(), cmd())
	})

	t.Run("textinput msg", func(t *testing.T) {
		m, cmd := model.Update(textInputMsg{})

		assert.Equal(t, statusTextinput, m.(Model).status)
		assert.NotNil(t, m.(Model).textinput)

		cmds := cmd().(tea.BatchMsg)
		assert.Equal(t, 2, len(cmds))
	})

	t.Run("status textinput", func(t *testing.T) {
		model.status = statusTextinput
		model.textinput = textinputmodel.Model{}

		m, _ := model.Update(nil)

		assert.NotNil(t, m.(Model).textinput)
	})
}

func TestView(t *testing.T) {
	model := New()

	view := model.View()
	assert.Contains(t, view, model.buttons[0].View())
	assert.Contains(t, view, model.buttons[1].View())
	assert.Contains(t, view, model.buttons[2].View())
}

func TestRegistrationCmd(t *testing.T) {
	values := []string{"user", "pass"}
	cmd := registrationCmd(values)
	msg := cmd().(messages.Registration)

	assert.Equal(t, "user", msg.Login)
	assert.Equal(t, "pass", msg.Password)
}

func TestSigninCmd(t *testing.T) {
	values := []string{"user", "pass"}
	cmd := signinCmd(values)
	msg := cmd()

	signInMsg, ok := msg.(messages.SignIn)
	assert.True(t, ok)
	assert.Equal(t, "user", signInMsg.Login)
	assert.Equal(t, "pass", signInMsg.Password)
}
