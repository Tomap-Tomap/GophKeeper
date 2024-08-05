//go:build unit

package tablemodel

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/mock"
)

type MockColumner struct {
	mock.Mock
}

func (m *MockColumner) Len() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockColumner) GetColums(w int) []table.Column {
	args := m.Called(w)
	return args.Get(0).([]table.Column)
}

func (m *MockColumner) GetRows() ([]table.Row, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]table.Row), args.Error(1)
}

func (m *MockColumner) GetInfo() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockColumner) GetHelp() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockColumner) InitInsert(enterCmd tea.Cmd) Columner {
	args := m.Called(enterCmd)
	return args.Get(0).(Columner)
}

func (m *MockColumner) InitUpdate(enterCmd tea.Cmd, row table.Row) Columner {
	args := m.Called(enterCmd, row)
	return args.Get(0).(Columner)
}

func (m *MockColumner) InitOpen(enterCmd tea.Cmd, row table.Row) Columner {
	args := m.Called(enterCmd, row)
	return args.Get(0).(Columner)
}

func (m *MockColumner) Update(msg tea.Msg) (Columner, tea.Cmd) {
	args := m.Called(msg)
	return args.Get(0).(Columner), args.Get(1).(tea.Cmd)
}

func (m *MockColumner) View() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockColumner) SetSize(msg tea.WindowSizeMsg) Columner {
	args := m.Called(msg)
	return args.Get(0).(Columner)
}

func (m *MockColumner) Insert() ([]table.Row, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]table.Row), args.Error(1)
}

func (m *MockColumner) UpdateData() ([]table.Row, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]table.Row), args.Error(1)
}

func (m *MockColumner) Open() ([]table.Row, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]table.Row), args.Error(1)
}

func (m *MockColumner) Delete(deleteRow table.Row) ([]table.Row, error) {
	args := m.Called(deleteRow)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]table.Row), args.Error(1)
}
