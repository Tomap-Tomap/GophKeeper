//go:build unit

package tabsmodel

import (
	"github.com/Tomap-Tomap/GophKeeper/tui/tablemodel"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/mock"
)

type MockColumner struct {
	tablemodel.Columner
	mock.Mock
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

func (m *MockColumner) GetColums(w int) []table.Column {
	args := m.Called(w)
	return args.Get(0).([]table.Column)
}

func (m *MockColumner) SetSize(msg tea.WindowSizeMsg) tablemodel.Columner {
	args := m.Called(msg)
	return args.Get(0).(tablemodel.Columner)
}

func (m *MockColumner) Update(msg tea.Msg) (tablemodel.Columner, tea.Cmd) {
	args := m.Called(msg)
	return args.Get(0).(tablemodel.Columner), args.Get(1).(tea.Cmd)
}

func (m *MockColumner) Len() int {
	args := m.Called()
	return args.Int(0)
}
