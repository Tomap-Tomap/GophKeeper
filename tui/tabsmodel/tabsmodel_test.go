//go:build unit

package tabsmodel

import (
	"errors"
	"testing"

	"github.com/Tomap-Tomap/GophKeeper/tui/messages"
	"github.com/Tomap-Tomap/GophKeeper/tui/tablemodel"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	mc := new(MockColumner)
	defer mc.AssertExpectations(t)

	t.Run("len columns and tabsName not equal get", func(t *testing.T) {
		m, err := New(
			[]tablemodel.Columner{
				new(MockColumner),
			},
			[]string{"Test", "Test"},
		)

		assert.ErrorContains(t, err, "len columns and tabsName not equal get")
		assert.Empty(t, m)
	})

	t.Run("cannot create", func(t *testing.T) {
		testErr := errors.New("test error")
		mc.On("GetRows").Return(nil, testErr).Once()
		mc.On("GetInfo").Return("Test").Once()
		mc.On("GetInfo").Return("Test").Once()

		m, err := New(
			[]tablemodel.Columner{
				mc,
			},
			[]string{"Test"},
		)

		assert.ErrorContains(t, err, "cannot create Test table")
		assert.ErrorIs(t, err, testErr)
		assert.Empty(t, m)
	})

	t.Run("positive test", func(t *testing.T) {
		mc.On("GetRows").Return([]table.Row{}, nil).Once()
		mc.On("GetColums", 0).Return([]table.Column{}).Once()

		m, err := New(
			[]tablemodel.Columner{
				mc,
			},
			[]string{"Test"},
		)

		assert.NoError(t, err)
		assert.NotEmpty(t, m)
	})
}

func TestModel_Init(t *testing.T) {
	wantInfo := messages.Info{
		Help: helpText,
	}
	model := Model{}
	cmd := model.Init()
	assert.Equal(t, wantInfo, cmd())
}

func TestModel_Update(t *testing.T) {
	cm := new(MockColumner)
	cm.On("GetRows").Return([]table.Row{}, nil).Once()
	cm.On("GetColums", 0).Return([]table.Column{}).Once()

	defer cm.AssertExpectations(t)

	model, err := New([]tablemodel.Columner{cm}, []string{"TestTab"})
	require.NoError(t, err)

	t.Run("window size message", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 100, Height: 100}
		cm.On("SetSize", mock.Anything).Return(cm).Once()
		cm.On("Len").Return(1).Once()
		cm.On("GetColums", mock.Anything).Return([]table.Column{}).Once()

		updatedModel, cmd := model.Update(msg)
		assert.NotNil(t, updatedModel)
		assert.Nil(t, cmd)
	})

	t.Run("unblock message", func(t *testing.T) {
		wantInfo := messages.Info{
			Help: helpText,
		}
		msg := unblockMsg{}
		model := model
		model.blockTabs = true
		updatedModel, cmd := model.Update(msg)
		assert.NotNil(t, updatedModel)
		assert.Equal(t, wantInfo, cmd())
		assert.False(t, updatedModel.(Model).blockTabs)
	})

	t.Run("key right message", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRight}
		updatedModel, cmd := model.Update(msg)
		assert.NotNil(t, updatedModel)
		assert.Nil(t, cmd)
		assert.Equal(t, 0, updatedModel.(Model).focused)
	})

	t.Run("key left message", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyLeft}
		updatedModel, cmd := model.Update(msg)
		assert.NotNil(t, updatedModel)
		assert.Nil(t, cmd)
		assert.Equal(t, 0, updatedModel.(Model).focused)
	})
	t.Run("key insert message error", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyInsert}
		updatedModel, cmd := model.Update(msg)
		assert.True(t, updatedModel.(Model).blockTabs)
		assert.NotNil(t, cmd)
	})

	t.Run("key update message error", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyCtrlU}
		updatedModel, cmd := model.Update(msg)
		assert.True(t, updatedModel.(Model).blockTabs)
		assert.NotNil(t, cmd)
	})

	t.Run("key open message error", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyCtrlO}
		updatedModel, cmd := model.Update(msg)
		assert.True(t, updatedModel.(Model).blockTabs)
		assert.NotNil(t, cmd)
	})

	t.Run("key delete message error", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyDelete}
		updatedModel, cmd := model.Update(msg)
		assert.False(t, updatedModel.(Model).blockTabs)
		assert.NotNil(t, cmd)
	})
}

func TestModel_View(t *testing.T) {
	cm := new(MockColumner)
	cm.On("GetRows").Return([]table.Row{}, nil).Once()
	cm.On("GetRows").Return([]table.Row{}, nil).Once()
	cm.On("GetColums", 0).Return([]table.Column{}).Once()
	cm.On("GetColums", 0).Return([]table.Column{}).Once()

	defer cm.AssertExpectations(t)

	model, err := New([]tablemodel.Columner{cm, cm}, []string{"TestTab1", "TestTab2"})
	require.NoError(t, err)

	result := model.View()
	assert.Contains(t, result, "TestTab1")
	assert.Contains(t, result, "TestTab2")
}
