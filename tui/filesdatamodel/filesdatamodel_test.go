//go:build unit

package filesdatamodel

import (
	"fmt"
	"testing"

	"github.com/Tomap-Tomap/GophKeeper/tui/filepickermodel"
	"github.com/Tomap-Tomap/GophKeeper/tui/textinputmodel"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

const (
	testID       = "testID"
	testName     = "testName"
	testMeta     = "testMeta"
	testFileText = "testFileText"
)

func TestNew(t *testing.T) {
	model := New(nil, testID, testName, testMeta, testFileText, false)

	textInputCmd := model.buttons[0].button.Action()
	assert.IsType(t, textInputMsg{}, textInputCmd())

	textInputCmd = model.buttons[1].button.Action()
	assert.IsType(t, textInputMsg{}, textInputCmd())

	filePickerCmd := model.buttons[2].button.Action()
	assert.IsType(t, filePickerMsg{}, filePickerCmd())

	assert.Equal(t, testName, model.buttons[0].value)
	assert.Equal(t, testMeta, model.buttons[1].value)

	id, name, path, meta := model.GetResult()
	assert.Equal(t, testID, id)
	assert.Equal(t, testName, name)
	assert.Equal(t, "", path)
	assert.Equal(t, testMeta, meta)
}

func TestModel_Init(t *testing.T) {
	model := New(nil, testID, testName, testMeta, testFileText, false)
	cmd := model.Init()

	assert.Nil(t, cmd)
}

func TestModel_Update(t *testing.T) {
	model := New(nil, testID, testName, testMeta, testFileText, false)

	t.Run("window size msg", func(t *testing.T) {
		msg := tea.WindowSizeMsg{
			Width:  100,
			Height: 100,
		}

		wantWidth := msg.Width - model.elementStyle.GetHorizontalFrameSize()
		wantHeight := msg.Height/buttonsCount - model.elementStyle.GetVerticalFrameSize()

		m, _ := model.Update(msg)
		assert.Equal(t, wantHeight, m.elementStyle.GetHeight())
		assert.Equal(t, wantHeight, m.focusedElementStyle.GetHeight())
		assert.Equal(t, wantWidth, m.elementStyle.GetWidth())
		assert.Equal(t, wantWidth, m.elementStyle.GetWidth())
	})

	t.Run("choose path msg", func(t *testing.T) {
		path := "testPath"
		m, cmd := model.Update(choosePathMsg{
			path: path,
		})

		assert.Equal(t, statusMain, m.status)
		assert.Equal(t, path, m.buttons[m.focused].value)

		assert.Equal(t, tea.ClearScreen(), cmd())
	})

	t.Run("close text input msg", func(t *testing.T) {
		value := "testValue"
		m, cmd := model.Update(closeTextInputMsg{
			value: value,
		})

		assert.Equal(t, statusMain, m.status)
		assert.Equal(t, value, m.buttons[m.focused].value)
		assert.Equal(t, tea.ClearScreen(), cmd())
	})

	t.Run("key down test", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyDown}
		m, _ := model.Update(msg)
		assert.Equal(t, 1, m.focused)
	})

	t.Run("key up tets", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyUp}
		m, _ := model.Update(msg)
		assert.Equal(t, 0, m.focused)
	})

	t.Run("enter tets", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		wantCmd := model.buttons[model.focused].button.Action()
		_, cmd := model.Update(msg)

		assert.IsType(t, wantCmd(), cmd())
	})

	t.Run("filepicker msg", func(t *testing.T) {
		m, cmd := model.Update(filePickerMsg{})

		assert.Equal(t, statusFilepicker, m.status)
		assert.NotNil(t, m.filepicker)

		cmds := cmd().(tea.BatchMsg)
		assert.Equal(t, 2, len(cmds))
	})

	t.Run("textinput msg", func(t *testing.T) {
		m, cmd := model.Update(textInputMsg{})

		assert.Equal(t, statusTextinput, m.status)
		assert.NotNil(t, m.textinput)

		cmds := cmd().(tea.BatchMsg)
		assert.Equal(t, 2, len(cmds))
	})

	t.Run("status filepicker", func(t *testing.T) {
		model.status = statusFilepicker
		model.filepicker = filepickermodel.Model{}

		m, _ := model.Update(nil)

		assert.NotNil(t, m.filepicker)
	})

	t.Run("status textinput", func(t *testing.T) {
		model.status = statusTextinput
		model.textinput = textinputmodel.Model{}

		m, _ := model.Update(nil)

		assert.NotNil(t, m.textinput)
	})
}

func TestModel_View(t *testing.T) {
	t.Run("status main", func(t *testing.T) {
		model := New(nil, testID, testName, testMeta, testFileText, false)
		s := model.View()

		assert.Contains(t, s, fmt.Sprintf("%s %s", model.buttons[0].button.View(), model.buttons[0].value))
		assert.Contains(t, s, fmt.Sprintf("%s %s", model.buttons[1].button.View(), model.buttons[1].value))
		assert.Contains(t, s, fmt.Sprintf("%s %s", model.buttons[2].button.View(), model.buttons[2].value))
	})
}
