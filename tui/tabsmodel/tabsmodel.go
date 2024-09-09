// Package tabsmodel provides a model for managing multiple tabs in a terminal user interface (TUI).
// It includes functionality for switching between tabs, handling user inputs, and updating the tab view based on various actions.
package tabsmodel

import (
	"fmt"
	"strings"

	"github.com/Tomap-Tomap/GophKeeper/tui/colors"
	"github.com/Tomap-Tomap/GophKeeper/tui/commands"
	"github.com/Tomap-Tomap/GophKeeper/tui/constants"
	"github.com/Tomap-Tomap/GophKeeper/tui/tablemodel"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const helpText = "↑: move up • ↓: move down • ←: previos tab • →: next tab • insert: add new row • ctrl+u: update row • ctrl+o: open row • delete: delete row"

// Model represents the state of the tabs model, including the current focused tab, the list of columns, and the tab names.
type Model struct {
	inactiveTabStyle lipgloss.Style
	activeTabStyle   lipgloss.Style
	windowStyle      lipgloss.Style

	tabs       []string
	tabContent []tablemodel.Model
	focused    int

	blockTabs bool
}

// New creates a new Model instance with the provided columns and tab names.
// It returns an error if the length of columns and tabsName are not equal or if it fails to initialize the columns.
func New(columns []tablemodel.Columner, tabsName []string) (Model, error) {
	if len(columns) != len(tabsName) {
		return Model{}, fmt.Errorf("len columns and tabsName not equal get: %d %d", len(columns), len(tabsName))
	}

	tabContents := make([]tablemodel.Model, 0, len(columns))

	for _, v := range columns {
		table, err := tablemodel.New(v, unblockCmd)

		if err != nil {
			return Model{}, fmt.Errorf("cannot create %s table: %w", v.GetInfo(), err)
		}

		tabContents = append(tabContents, table)
	}

	inactiveTabBorder := tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder := tabBorderWithBottom("┘", " ", "└")

	inactiveTabStyle := lipgloss.NewStyle().
		Border(inactiveTabBorder, true).
		BorderForeground(colors.MainColor).
		Align(lipgloss.Center, lipgloss.Center)

	activeTabStyle := inactiveTabStyle.Border(activeTabBorder, true)

	windowStyle := lipgloss.NewStyle().
		BorderForeground(colors.MainColor).
		Align(lipgloss.Center).
		Border(lipgloss.NormalBorder()).
		UnsetBorderTop()

	return Model{
		inactiveTabStyle: inactiveTabStyle,
		activeTabStyle:   activeTabStyle,
		windowStyle:      windowStyle,
		tabs:             tabsName,
		tabContent:       tabContents,
	}, nil
}

// Init initializes the Model and returns an initial command.
func (m Model) Init() tea.Cmd {
	return commands.SetInfo("", helpText)
}

// Update handles the update logic for the Model based on the received message.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if wm, ok := msg.(tea.WindowSizeMsg); ok {
		m, msg = m.calculateSize(wm)

		for i := range m.tabContent {
			m.tabContent[i] = m.tabContent[i].SetSize(msg.(tea.WindowSizeMsg))
		}
	}

	if !m.blockTabs {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRight:
				m.focused = min(m.focused+1, len(m.tabs)-1)
			case tea.KeyLeft:
				m.focused = max(m.focused-1, 0)
			case tea.KeyInsert:
				cmd := m.tabContent[m.focused].Insert()
				m.blockTabs = true
				return m, cmd
			case tea.KeyCtrlU:
				cmd := m.tabContent[m.focused].UpdateData()
				m.blockTabs = true
				return m, cmd
			case tea.KeyCtrlO:
				cmd := m.tabContent[m.focused].Open()
				m.blockTabs = true
				return m, cmd
			case tea.KeyDelete:
				return m, m.tabContent[m.focused].Delete()
			}
		}
	}

	var cmds []tea.Cmd

	if _, ok := msg.(unblockMsg); ok {
		m.blockTabs = false
		cmds = append(cmds, commands.SetInfo("", helpText))
	}

	tc, cmd := m.tabContent[m.focused].Update(msg)
	m.tabContent[m.focused] = tc.(tablemodel.Model)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the current view of the Model.
func (m Model) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	for i, t := range m.tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.tabs)-1, i == m.focused
		if isActive {
			style = m.activeTabStyle
		} else {
			style = m.inactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(m.windowStyle.Render(m.tabContent[m.focused].View()))

	return doc.String()
}

func (m Model) calculateSize(msg tea.WindowSizeMsg) (Model, tea.WindowSizeMsg) {
	m.inactiveTabStyle = m.inactiveTabStyle.Width(msg.Width/len(m.tabs) - m.inactiveTabStyle.GetHorizontalFrameSize())
	m.activeTabStyle = m.inactiveTabStyle.Border(m.activeTabStyle.GetBorderStyle(), true)

	m.windowStyle = m.windowStyle.
		Height(msg.Height - m.windowStyle.GetVerticalFrameSize() - m.inactiveTabStyle.GetVerticalFrameSize() - constants.StringHeight)

	var renderedTabs []string

	for _, t := range m.tabs {
		renderedTabs = append(renderedTabs, m.inactiveTabStyle.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	width := lipgloss.Width(row) - m.windowStyle.GetHorizontalFrameSize()

	m.windowStyle = m.windowStyle.Width(width)

	return m, tea.WindowSizeMsg{
		Width:  width,
		Height: m.windowStyle.GetHeight() - m.activeTabStyle.GetVerticalFrameSize(),
	}
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}
