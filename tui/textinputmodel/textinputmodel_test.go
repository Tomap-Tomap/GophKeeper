//go:build unit

package textinputmodel

import (
	"testing"

	"github.com/Tomap-Tomap/GophKeeper/tui/messages"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ti := []textinput.Model{
		textinput.New(),
		textinput.New(),
	}
	m := New(ti, nil, "test")

	assert.Equal(t, ti, m.inputs)
	assert.Equal(t, "test", m.infoText)
	assert.Nil(t, m.returnCmd)
}

func TestModel_Init(t *testing.T) {
	m := Model{
		infoText: "test",
	}

	cmd := m.Init()
	assert.Equal(t, messages.Info{
		Info: "test",
		Help: helpText,
	}, cmd())
}

func TestModel_Update(t *testing.T) {
	ti := []textinput.Model{
		textinput.New(),
	}
	model := Model{
		inputs: ti,
		returnCmd: func(values []string) tea.Cmd {
			return func() tea.Msg {
				return nil
			}
		},
	}
	t.Run("windows size test", func(t *testing.T) {
		msg := tea.WindowSizeMsg{
			Width:  100,
			Height: 100,
		}

		wantWidth := msg.Width - model.inputStyle.GetHorizontalFrameSize()
		wantHeight := msg.Height - model.inputStyle.GetVerticalFrameSize()

		m, _ := model.Update(msg)
		assert.Equal(t, wantHeight, m.(Model).inputStyle.GetHeight())
		assert.Equal(t, wantHeight, m.(Model).focusedInputStyle.GetHeight())
		assert.Equal(t, wantWidth, m.(Model).inputStyle.GetWidth())
		assert.Equal(t, wantWidth, m.(Model).inputStyle.GetWidth())
	})

	t.Run("key down test", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyDown}
		m, _ := model.Update(msg)
		assert.Equal(t, 0, m.(Model).focused)
	})

	t.Run("key up tets", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyUp}
		m, _ := model.Update(msg)
		assert.Equal(t, 0, m.(Model).focused)
	})

	t.Run("enter tets", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, cmd := model.Update(msg)

		assert.Nil(t, cmd())
	})
}

func TestModel_View(t *testing.T) {
	ti := textinput.New()
	ti.Placeholder = "Test"
	tis := []textinput.Model{
		ti,
	}
	m := Model{
		inputs: tis,
	}

	res := m.View()

	assert.Contains(t, res, "Test")
}
