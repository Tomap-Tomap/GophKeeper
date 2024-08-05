//go:build unit

package filepickermodel

import (
	"testing"

	"github.com/Tomap-Tomap/GophKeeper/tui/messages"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModel_Init(t *testing.T) {
	t.Run("test is dir picker", func(t *testing.T) {
		model, err := New(nil, true)
		require.NoError(t, err)
		cmd := model.Init()

		cmds := cmd().(tea.BatchMsg)
		assert.Equal(t, messages.Info{
			Info: infoTex,
			Help: helpText,
		}, cmds[1]())
	})

	t.Run("test isn't dir picker", func(t *testing.T) {
		model, err := New(nil, false)
		require.NoError(t, err)
		cmd := model.Init()

		cmds := cmd().(tea.BatchMsg)
		assert.Equal(t, messages.Info{
			Info: infoTex,
			Help: helpText,
		}, cmds[1]())
	})
}

func TestModel_Update(t *testing.T) {
	model, err := New(nil, false)
	require.NoError(t, err)

	m, cmd := model.Update(nil)
	assert.NotNil(t, m)
	assert.Nil(t, cmd, nil)
}
