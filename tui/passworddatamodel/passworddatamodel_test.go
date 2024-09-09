//go:build unit

package passworddatamodel

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

const (
	testID       = "testID"
	testName     = "testName"
	testLogin    = "testLogin"
	testPassword = "testPassword"
	testMeta     = "testMeta"
)

type testEnterMsg struct{}

func testEnterCmd() tea.Msg {
	return testEnterMsg{}
}

func TestNew(t *testing.T) {
	model := New(testEnterCmd, testID, testName, testLogin, testPassword, testMeta)

	assert.Equal(t, testID, model.id)
	assert.Equal(t, testName, model.inputs[0].Value())
	assert.Equal(t, testLogin, model.inputs[1].Value())
	assert.Equal(t, testPassword, model.inputs[2].Value())
	assert.Equal(t, testMeta, model.inputs[3].Value())
}

func TestInit(t *testing.T) {
	model := New(testEnterCmd, testID, testName, testLogin, testPassword, testMeta)

	cmd := model.Init()
	assert.Nil(t, cmd)
}

func TestUpdate(t *testing.T) {
	model := New(testEnterCmd, testID, testName, testLogin, testPassword, testMeta)

	t.Run("key down test", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyDown}
		updatedModel, _ := model.Update(msg)
		assert.Equal(t, 1, updatedModel.focused)
	})

	t.Run("key up test", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyUp}
		updatedModel, _ := model.Update(msg)
		assert.Equal(t, 0, updatedModel.focused)
	})

	t.Run("enter test", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, cmd := model.Update(msg)
		assert.Equal(t, testEnterMsg{}, cmd())
	})

	t.Run("window size test", func(t *testing.T) {
		msg := tea.WindowSizeMsg{
			Width:  100,
			Height: 100,
		}

		wantWidth := msg.Width - model.elementStyle.GetHorizontalFrameSize()
		wantHeight := msg.Height/inputsCount - model.elementStyle.GetVerticalFrameSize()

		m, _ := model.Update(msg)
		assert.Equal(t, wantHeight, m.elementStyle.GetHeight())
		assert.Equal(t, wantHeight, m.focusedElementStyle.GetHeight())
		assert.Equal(t, wantWidth, m.elementStyle.GetWidth())
		assert.Equal(t, wantWidth, m.elementStyle.GetWidth())
	})
}

func TestView(t *testing.T) {
	model := New(testEnterCmd, testID, testName, testLogin, testPassword, testMeta)

	view := model.View()
	assert.Contains(t, view, fmt.Sprintf("Name > %s", testName))
	assert.Contains(t, view, fmt.Sprintf("Login > %s", testLogin))
	assert.Contains(t, view, fmt.Sprintf("Password > %s", testPassword))
	assert.Contains(t, view, fmt.Sprintf("Meta > %s", testMeta))
}

func TestGetResult(t *testing.T) {
	model := New(testEnterCmd, testID, testName, testLogin, testPassword, testMeta)

	resultID, resultName, resultLogin, resultPassword, resultMeta := model.GetResult()

	assert.Equal(t, testID, resultID)
	assert.Equal(t, testName, resultName)
	assert.Equal(t, testLogin, resultLogin)
	assert.Equal(t, testPassword, resultPassword)
	assert.Equal(t, testMeta, resultMeta)
}
