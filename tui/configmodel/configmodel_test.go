//go:build unit

package configmodel

import (
	"fmt"
	"testing"

	"github.com/Tomap-Tomap/GophKeeper/tui/config"
	"github.com/Tomap-Tomap/GophKeeper/tui/filepickermodel"
	"github.com/Tomap-Tomap/GophKeeper/tui/messages"
	"github.com/Tomap-Tomap/GophKeeper/tui/textinputmodel"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		PathToSecretKey: "test/path",
		AddrToService:   "http://localhost",
	}

	model := New(cfg)

	assert.Equal(t, "test/path", model.buttons[0].value)
	assert.Equal(t, "http://localhost", model.buttons[1].value)
	assert.Equal(t, statusMain, model.status)
	assert.Equal(t, 0, model.focused)

	filePickerCmd := model.buttons[0].button.Action()
	assert.IsType(t, filePickerMsg{}, filePickerCmd())
	choosePathCmd := filePickerCmd().(filePickerMsg).choosePathCmd("")
	assert.IsType(t, choosePathMsg{}, choosePathCmd())

	textInputCmd := model.buttons[1].button.Action()
	assert.IsType(t, textInputMsg{}, textInputCmd())

	filePickerCmd = model.buttons[2].button.Action()
	assert.IsType(t, filePickerMsg{}, filePickerCmd())
	choosePathCmd = filePickerCmd().(filePickerMsg).choosePathCmd("")
	assert.IsType(t, createKeyMsg{}, choosePathCmd())

	closeConfigCmd := model.buttons[3].button.Action()
	assert.IsType(t, closeConfigMsg{}, closeConfigCmd())
}

func TestInit(t *testing.T) {
	model := New(&config.Config{})
	cmd := model.Init()

	assert.Equal(t, messages.Info{
		Info: infoText,
		Help: helpText,
	}, cmd())
}

func TestUpdate(t *testing.T) {
	model := New(&config.Config{})

	t.Run("window size msg", func(t *testing.T) {
		msg := tea.WindowSizeMsg{
			Width:  100,
			Height: 100,
		}

		wantWidth := msg.Width - model.elementStyle.GetHorizontalFrameSize()
		wantHeight := msg.Height/buttonCount - model.elementStyle.GetVerticalFrameSize()

		m, _ := model.Update(msg)
		assert.Equal(t, wantHeight, m.(Model).elementStyle.GetHeight())
		assert.Equal(t, wantHeight, m.(Model).focusedElementStyle.GetHeight())
		assert.Equal(t, wantWidth, m.(Model).elementStyle.GetWidth())
		assert.Equal(t, wantWidth, m.(Model).elementStyle.GetWidth())
	})

	t.Run("choose path msg", func(t *testing.T) {
		path := "testPath"
		m, cmd := model.Update(choosePathMsg{
			path: path,
		})

		assert.Equal(t, statusMain, m.(Model).status)
		assert.Equal(t, path, m.(Model).buttons[m.(Model).focused].value)

		cmds := cmd().(tea.BatchMsg)

		assert.Equal(t, messages.Info{
			Info: infoText,
		}, cmds[0]())
		assert.Equal(t, tea.ClearScreen(), cmds[1]())
	})

	t.Run("close text input msg", func(t *testing.T) {
		value := "testValue"
		m, cmd := model.Update(closeTextInputMsg{
			value: value,
		})

		assert.Equal(t, statusMain, m.(Model).status)
		assert.Equal(t, value, m.(Model).buttons[m.(Model).focused].value)

		cmds := cmd().(tea.BatchMsg)

		assert.Equal(t, messages.Info{
			Info: infoText,
		}, cmds[0]())
		assert.Equal(t, tea.ClearScreen(), cmds[1]())
	})

	t.Run("create key msg", func(t *testing.T) {
		path := "testPath"
		m, cmd := model.Update(createKeyMsg{
			path: path,
		})

		assert.Equal(t, statusMain, m.(Model).status)
		assert.IsType(t, messages.Error{}, cmd())
	})

	t.Run("close config msg", func(t *testing.T) {
		_, cmd := model.Update(closeConfigMsg{})

		cmds := cmd().(tea.BatchMsg)

		assert.IsType(t, messages.CloseConfigModel{}, cmds[0]())
		assert.Equal(t, tea.ClearScreen(), cmds[1]())
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

	t.Run("key up tets", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		wantCmd := model.buttons[model.focused].button.Action()
		_, cmd := model.Update(msg)

		assert.IsType(t, wantCmd(), cmd())
	})

	t.Run("filepicker msg", func(t *testing.T) {
		m, cmd := model.Update(filePickerMsg{})

		assert.Equal(t, statusFilepicker, m.(Model).status)
		assert.NotNil(t, m.(Model).filepicker)

		cmds := cmd().(tea.BatchMsg)
		assert.Equal(t, 2, len(cmds))
	})

	t.Run("textinput msg", func(t *testing.T) {
		m, cmd := model.Update(textInputMsg{})

		assert.Equal(t, statusTextinput, m.(Model).status)
		assert.NotNil(t, m.(Model).textinput)

		cmds := cmd().(tea.BatchMsg)
		assert.Equal(t, 2, len(cmds))
	})

	t.Run("status filepicker", func(t *testing.T) {
		model.status = statusFilepicker
		model.filepicker = filepickermodel.Model{}

		m, _ := model.Update(nil)

		assert.NotNil(t, m.(Model).filepicker)
	})

	t.Run("status textinput", func(t *testing.T) {
		model.status = statusTextinput
		model.textinput = textinputmodel.Model{}

		m, _ := model.Update(nil)

		assert.NotNil(t, m.(Model).textinput)
	})
}

func TestView(t *testing.T) {
	t.Run("status main", func(t *testing.T) {
		model := New(&config.Config{})
		s := model.View()

		assert.Contains(t, s, fmt.Sprintf("%s %s", model.buttons[0].button.View(), model.buttons[0].value))
		assert.Contains(t, s, fmt.Sprintf("%s %s", model.buttons[1].button.View(), model.buttons[1].value))
		assert.Contains(t, s, fmt.Sprintf("%s %s", model.buttons[2].button.View(), model.buttons[2].value))
		assert.Contains(t, s, fmt.Sprintf("%s %s", model.buttons[3].button.View(), model.buttons[3].value))
	})
}
