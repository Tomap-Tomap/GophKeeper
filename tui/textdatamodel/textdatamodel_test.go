//go:build unit

package textdatamodel

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

const (
	testID   = "testID"
	testName = "testName"
	testMeta = "testMeta"
	testText = "testText"
)

func TestNew(t *testing.T) {
	model := New(nil, testID, testName, testMeta, testText)

	assert.Equal(t, testID, model.id)
	assert.Equal(t, testName, model.name.Value())
	assert.Equal(t, testMeta, model.meta.Value())
	assert.Equal(t, testText, model.text.Value())
}

func TestModel_Init(t *testing.T) {
	model := New(nil, testID, testName, testMeta, testText)
	cmd := model.Init()

	assert.Nil(t, cmd)
}

func TestModel_Update(t *testing.T) {
	model := New(nil, testID, testName, testMeta, testText)

	t.Run("window size msg", func(t *testing.T) {
		msg := tea.WindowSizeMsg{
			Width:  100,
			Height: 100,
		}

		wantWidth := msg.Width - model.elementStyle.GetHorizontalFrameSize()
		wantHeight := msg.Height/(maxIdx+1) - model.elementStyle.GetVerticalFrameSize()

		m, _ := model.Update(msg)
		assert.Equal(t, wantHeight, m.elementStyle.GetHeight())
		assert.Equal(t, wantHeight, m.focusedElementStyle.GetHeight())
		assert.Equal(t, wantWidth, m.elementStyle.GetWidth())
		assert.Equal(t, wantWidth, m.elementStyle.GetWidth())
	})

	t.Run("key down test", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyDown}
		m, _ := model.Update(msg)
		assert.Equal(t, 1, m.focused)
	})

	t.Run("key up test", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyUp}
		m, _ := model.Update(msg)
		assert.Equal(t, 0, m.focused)
	})

	t.Run("enter test", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, cmd := model.Update(msg)
		assert.Nil(t, cmd)
	})
}

func TestModel_View(t *testing.T) {
	model := New(nil, testID, testName, testMeta, testText)

	view := model.View()
	assert.Contains(t, view, testName)
	assert.Contains(t, view, testMeta)
	assert.Contains(t, view, testText)
}

func TestModel_GetResult(t *testing.T) {
	model := New(nil, testID, testName, testMeta, testText)

	resultID, resultName, resultMeta, resultText := model.GetResult()

	assert.Equal(t, testID, resultID)
	assert.Equal(t, testName, resultName)
	assert.Equal(t, testMeta, resultMeta)
	assert.Equal(t, testText, resultText)
}

func TestModel_calculateSize(t *testing.T) {
	model := New(nil, testID, testName, testMeta, testText)
	msg := tea.WindowSizeMsg{
		Width:  100,
		Height: 100,
	}

	newWidth := msg.Width - model.elementStyle.GetHorizontalFrameSize()
	newHeight := msg.Height/(maxIdx+1) - model.elementStyle.GetVerticalFrameSize()

	m := model.calculateSize(msg)
	assert.Equal(t, newHeight, m.elementStyle.GetHeight())
	assert.Equal(t, newHeight, m.focusedElementStyle.GetHeight())
	assert.Equal(t, newWidth, m.elementStyle.GetWidth())
	assert.Equal(t, newWidth, m.focusedElementStyle.GetWidth())
}
