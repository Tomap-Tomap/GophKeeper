//go:build unit

package tablemodel

import (
	"errors"
	"testing"

	"github.com/Tomap-Tomap/GophKeeper/tui/messages"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testCmd tea.Cmd

const (
	testInfo = "testInfo"
)

func TestNew(t *testing.T) {
	cm := new(MockColumner)
	cm.On("GetRows").Return(nil, errors.New("test")).Once()
	cm.On("GetRows").Return([]table.Row{}, nil).Once()
	cm.On("GetInfo").Return(testInfo)
	cm.On("GetColums", 0).Return([]table.Column{})
	defer cm.AssertExpectations(t)

	t.Run("cannot get testInfo rows", func(t *testing.T) {
		m, err := New(cm, nil)
		assert.ErrorContains(t, err, "cannot get testInfo rows")
		assert.Empty(t, m)
	})

	t.Run("positive test", func(t *testing.T) {
		m, err := New(cm, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, m)
	})
}

func TestModel_Init(t *testing.T) {
	m := Model{}
	assert.Nil(t, m.Init())
}

func TestModel_Update(t *testing.T) {
	ctrlzMsg := tea.KeyMsg{Type: tea.KeyCtrlZ}
	cm := new(MockColumner)
	defer cm.AssertExpectations(t)

	cm.On("GetRows").Return([]table.Row{}, nil).Once()
	cm.On("GetColums", 0).Return([]table.Column{})
	cm.On("Update", nil).Return(cm, testCmd).Once()
	cm.On("Update", nil).Return(cm, testCmd).Once()
	cm.On("Update", nil).Return(cm, testCmd).Once()
	cm.On("Update", ctrlzMsg).Return(cm, testCmd).Once()

	m, err := New(cm, nil)
	require.NoError(t, err)

	t.Run("status Main", func(t *testing.T) {
		_, cmd := m.Update(nil)
		assert.Nil(t, cmd)
	})

	t.Run("status Inputs", func(t *testing.T) {
		m.status = statusInputs
		_, cmd := m.Update(nil)
		assert.Nil(t, cmd)
	})

	t.Run("status Updates", func(t *testing.T) {
		m.status = statusUpdate
		_, cmd := m.Update(nil)
		assert.Nil(t, cmd)
	})

	t.Run("status Opens", func(t *testing.T) {
		m.status = statusOpen
		_, cmd := m.Update(nil)
		assert.Nil(t, cmd)
	})

	t.Run("ctrlz", func(t *testing.T) {
		m.status = statusOpen
		m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlZ})
		require.IsType(t, Model{}, m)
		assert.Equal(t, statusMain, m.(Model).status)
		assert.Equal(t, tea.ClearScreen(), cmd())
	})
}

func TestModel_View(t *testing.T) {
	cm := new(MockColumner)
	defer cm.AssertExpectations(t)

	cm.On("GetRows").Return([]table.Row{}, nil).Once()
	cm.On("GetColums", 0).Return([]table.Column{})
	cm.On("View").Return("TestView")

	m, err := New(cm, nil)
	require.NoError(t, err)

	t.Run("status Main", func(t *testing.T) {
		res := m.View()
		assert.NotEmpty(t, res)
	})

	t.Run("status not Main", func(t *testing.T) {
		m.status = statusInputs
		res := m.View()
		assert.Contains(t, res, "TestView")
	})
}

func TestModel_SetSize(t *testing.T) {
	wsm := tea.WindowSizeMsg{
		Width:  100,
		Height: 100,
	}

	cm := new(MockColumner)
	defer cm.AssertExpectations(t)

	cm.On("GetRows").Return([]table.Row{}, nil).Once()
	cm.On("GetColums", 0).Return([]table.Column{}).Once()
	m, err := New(cm, nil)
	require.NoError(t, err)

	cm.On("GetColums", wsm.Width-m.tableStyle.Header.GetHorizontalFrameSize()).Return([]table.Column{}).Once()
	cm.On("SetSize", wsm).Return(cm)
	cm.On("Len").Return(1)

	retModel := m.SetSize(wsm)
	assert.Equal(t, wsm.Height-m.tableStyle.Header.GetVerticalFrameSize(), retModel.table.Height())
}

func TestModel_Insert(t *testing.T) {
	m := Model{}
	cmd := m.Insert()
	assert.Equal(t, inputMsg{}, cmd())
}

func TestModel_UpdateData(t *testing.T) {
	m := Model{}
	cmd := m.UpdateData()
	assert.Equal(t, updateMsg{}, cmd())
}

func TestModel_Open(t *testing.T) {
	m := Model{}
	cmd := m.Open()
	assert.Equal(t, openMsg{}, cmd())
}

func TestModel_Delete(t *testing.T) {
	m := Model{}
	cmd := m.Delete()
	assert.Equal(t, deleteMsg{}, cmd())
}

func TestModel_updateMain(t *testing.T) {
	testErr := errors.New("Test error")
	wsm := tea.WindowSizeMsg{
		Width:  100,
		Height: 100,
	}

	cm := new(MockColumner)
	defer cm.AssertExpectations(t)

	cm.On("GetInfo").Return("TestInfo")
	cm.On("GetHelp").Return("TestHelp")

	m := Model{
		columns:    cm,
		columsSize: wsm,
	}

	t.Run("input msg", func(t *testing.T) {
		cm.On("InitInsert", mock.Anything).Return(cm)
		cm.On("SetSize", wsm).Return(cm)

		m, cmd := m.updateMain(inputMsg{})

		assert.Equal(t, statusInputs, m.status)

		msgInfo := messages.Info{
			Info: "Please input TestInfo data",
			Help: "TestHelp • ctrl+z: go back",
		}

		assert.Equal(t, msgInfo, cmd())
	})

	t.Run("update msg", func(t *testing.T) {
		cm.On("InitUpdate", mock.Anything, mock.Anything).Return(cm)
		m, cmd := m.updateMain(updateMsg{})

		assert.Equal(t, statusUpdate, m.status)

		msgInfo := messages.Info{
			Info: "Please update TestInfo data",
			Help: "TestHelp • ctrl+z: go back",
		}

		assert.Equal(t, msgInfo, cmd())
	})

	t.Run("open msg", func(t *testing.T) {
		cm.On("InitOpen", mock.Anything, mock.Anything).Return(cm)
		m, cmd := m.updateMain(openMsg{})

		assert.Equal(t, statusOpen, m.status)

		msgInfo := messages.Info{
			Info: "This is TestInfo data",
			Help: "TestHelp • ctrl+z: go back",
		}

		assert.Equal(t, msgInfo, cmd())
	})

	t.Run("delete msg err", func(t *testing.T) {
		cm.On("Delete", mock.Anything).Return(nil, testErr).Once()
		m, cmd := m.updateMain(deleteMsg{})

		assert.Equal(t, statusMain, m.status)

		errmsg := messages.Error{
			Err: testErr,
		}

		assert.Equal(t, errmsg, cmd())
	})

	t.Run("delete msg", func(t *testing.T) {
		wantRows := []table.Row{}
		cm.On("Delete", mock.Anything).Return(wantRows, nil).Once()
		m, cmd := m.updateMain(deleteMsg{})

		assert.Equal(t, statusMain, m.status)
		assert.Equal(t, m.table.Rows(), wantRows)

		assert.Nil(t, cmd)
	})
}

func TestModel_updateUpdates(t *testing.T) {
	cm := new(MockColumner)
	defer cm.AssertExpectations(t)

	m := Model{
		columns: cm,
		status:  statusInputs,
	}

	t.Run("enter msg err", func(t *testing.T) {
		testErr := errors.New("Test error")
		cm.On("UpdateData").Return(nil, testErr).Once()

		m, cmd := m.updateUpdates(enterMsg{})

		assert.Equal(t, statusInputs, m.status)
		assert.Equal(t, messages.Error{Err: testErr}, cmd())
	})

	t.Run("enter msg", func(t *testing.T) {
		testRow := []table.Row{}
		cm.On("UpdateData").Return(testRow, nil).Once()

		m, cmd := m.updateUpdates(enterMsg{})

		assert.Equal(t, statusMain, m.status)
		assert.Equal(t, testRow, m.table.Rows(), testRow)
		assert.Equal(t, tea.ClearScreen(), cmd())
	})
}

func TestModel_updateOpen(t *testing.T) {
	cm := new(MockColumner)
	defer cm.AssertExpectations(t)

	m := Model{
		columns: cm,
		status:  statusInputs,
	}

	t.Run("enter msg err", func(t *testing.T) {
		testErr := errors.New("Test error")
		cm.On("Open").Return(nil, testErr).Once()

		m, cmd := m.updateOpen(enterMsg{})

		assert.Equal(t, statusInputs, m.status)
		assert.Equal(t, messages.Error{Err: testErr}, cmd())
	})

	t.Run("enter msg", func(t *testing.T) {
		testRow := []table.Row{}
		cm.On("Open").Return(testRow, nil).Once()

		m, cmd := m.updateOpen(enterMsg{})

		assert.Equal(t, statusMain, m.status)
		assert.Equal(t, testRow, m.table.Rows(), testRow)
		assert.Equal(t, tea.ClearScreen(), cmd())
	})
}

func TestModel_updateInputs(t *testing.T) {
	cm := new(MockColumner)
	defer cm.AssertExpectations(t)

	m := Model{
		columns: cm,
		status:  statusInputs,
	}

	t.Run("enter msg err", func(t *testing.T) {
		testErr := errors.New("Test error")
		cm.On("Insert").Return(nil, testErr).Once()

		m, cmd := m.updateInputs(enterMsg{})

		assert.Equal(t, statusInputs, m.status)
		assert.Equal(t, messages.Error{Err: testErr}, cmd())
	})

	t.Run("enter msg", func(t *testing.T) {
		testRow := []table.Row{}
		cm.On("Insert").Return(testRow, nil).Once()

		m, cmd := m.updateInputs(enterMsg{})

		assert.Equal(t, statusMain, m.status)
		assert.Equal(t, testRow, m.table.Rows(), testRow)
		assert.Equal(t, tea.ClearScreen(), cmd())
	})
}
