// Package tablemodel provides a model for managing and displaying tabular data
// in a terminal user interface (TUI). It includes functionality for interacting
// with different storage entities, handling user inputs, and updating the table
// view based on various actions.
package tablemodel

import (
	"fmt"

	"github.com/Tomap-Tomap/GophKeeper/tui/colors"
	"github.com/Tomap-Tomap/GophKeeper/tui/commands"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Columner interface defines methods for managing table columns and rows.
type Columner interface {
	Len() int
	GetColums(w int) []table.Column
	GetRows() ([]table.Row, error)
	GetInfo() string
	GetHelp() string
	InitInsert(enterCmd tea.Cmd) Columner
	InitUpdate(enterCmd tea.Cmd, row table.Row) Columner
	InitOpen(enterCmd tea.Cmd, row table.Row) Columner
	Update(msg tea.Msg) (Columner, tea.Cmd)
	View() string
	SetSize(msg tea.WindowSizeMsg) Columner
	Insert() ([]table.Row, error)
	UpdateData() ([]table.Row, error)
	Open() ([]table.Row, error)
	Delete(deleteRow table.Row) ([]table.Row, error)
}

type status int

const (
	statusMain status = iota
	statusInputs
	statusUpdate
	statusOpen
)

// Model struct represents the table model, including its context, client, and state.
type Model struct {
	tableStyle table.Styles
	table      table.Model
	columns    Columner
	columsSize tea.WindowSizeMsg

	status    status
	returnCmd tea.Cmd
}

// New creates a new Model instance with the provided context, client, columner, and return command.
func New(columner Columner, returnCmd tea.Cmd) (Model, error) {
	tableStyle := table.Styles{
		Header: table.DefaultStyles().Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			Bold(false),
		Cell: table.DefaultStyles().Cell,
		Selected: table.DefaultStyles().Selected.
			Background(colors.FocusColor).
			Bold(false),
	}

	rows, err := columner.GetRows()

	if err != nil {
		return Model{}, fmt.Errorf("cannot get %s rows: %w", columner.GetInfo(), err)
	}

	tabel := table.New(
		table.WithColumns(columner.GetColums(0)),
		table.WithStyles(tableStyle),
		table.WithFocused(true),
		table.WithRows(rows),
	)

	return Model{
		tableStyle: tableStyle,
		table:      tabel,
		columns:    columner,
		status:     statusMain,
		returnCmd:  returnCmd,
	}, nil
}

// Init initializes the Model and returns an initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles the update logic for the Model based on the current status.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.status {
	case statusMain:
		m, cmd = m.updateMain(msg)
	case statusInputs:
		m, cmd = m.updateInputs(msg)
	case statusUpdate:
		m, cmd = m.updateUpdates(msg)
	case statusOpen:
		m, cmd = m.updateOpen(msg)
	}

	if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyCtrlZ && m.status != statusMain {
		m.status = statusMain
		cmd = tea.Batch(tea.ClearScreen, m.returnCmd)
	}

	return m, cmd
}

// View renders the current view of the Model.
func (m Model) View() string {
	var view string

	switch m.status {
	case statusMain:
		view = m.table.View()
	default:
		view = m.columns.View()
	}

	return view
}

// SetSize updates the size of the table and columns based on the provided window size message.
func (m Model) SetSize(msg tea.WindowSizeMsg) Model {
	columnWidth := msg.Width/m.columns.Len() - m.tableStyle.Header.GetHorizontalFrameSize()

	m.table.SetColumns(m.columns.GetColums(columnWidth))
	m.table.SetHeight(msg.Height - m.tableStyle.Header.GetVerticalFrameSize())

	m.columsSize = tea.WindowSizeMsg{
		Width:  msg.Width,
		Height: msg.Height,
	}

	m.columns = m.columns.SetSize(m.columsSize)

	return m
}

// Insert returns a command to handle the insertion of new data.
func (m Model) Insert() tea.Cmd {
	return inputCmd()
}

// UpdateData returns a command to handle the updating of existing data.
func (m Model) UpdateData() tea.Cmd {
	return updateCmd()
}

// Open returns a command to handle the opening of data.
func (m Model) Open() tea.Cmd {
	return openCmd()
}

// Delete returns a command to handle the deletion of data.
func (m Model) Delete() tea.Cmd {
	return deleteCmd()
}

func (m Model) updateMain(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case inputMsg:
		m.status = statusInputs

		m.columns = m.columns.InitInsert(enterCmd()).SetSize(m.columsSize)

		cmd = commands.SetInfo(fmt.Sprintf("Please input %s data", m.columns.GetInfo()), m.columns.GetHelp()+" • ctrl+z: go back")
	case updateMsg:
		m.status = statusUpdate
		m.columns = m.columns.InitUpdate(enterCmd(), m.table.SelectedRow()).SetSize(m.columsSize)

		cmd = commands.SetInfo(fmt.Sprintf("Please update %s data", m.columns.GetInfo()), m.columns.GetHelp()+" • ctrl+z: go back")
	case openMsg:
		m.status = statusOpen
		m.columns = m.columns.InitOpen(enterCmd(), m.table.SelectedRow()).SetSize(m.columsSize)

		cmd = commands.SetInfo(fmt.Sprintf("This is %s data", m.columns.GetInfo()), m.columns.GetHelp()+" • ctrl+z: go back")
	case deleteMsg:
		rows, err := m.columns.Delete(m.table.SelectedRow())

		if err != nil {
			return m, commands.Error(err)
		}

		m.table.SetRows(rows)
	default:
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m Model) updateInputs(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case enterMsg:
		rows, err := m.columns.Insert()

		if err != nil {
			return m, commands.Error(err)
		}

		m.table.SetRows(rows)

		m.status = statusMain
		cmd = tea.Batch(tea.ClearScreen, m.returnCmd)
	default:
		m.columns, cmd = m.columns.Update(msg)
	}

	return m, cmd
}

func (m Model) updateUpdates(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.(type) {
	case enterMsg:
		rows, err := m.columns.UpdateData()

		if err != nil {
			return m, commands.Error(err)
		}

		m.table.SetRows(rows)

		m.status = statusMain
		cmd = tea.Batch(tea.ClearScreen, m.returnCmd)
	default:
		m.columns, cmd = m.columns.Update(msg)
	}

	return m, cmd
}

func (m Model) updateOpen(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.(type) {
	case enterMsg:
		rows, err := m.columns.Open()

		if err != nil {
			return m, commands.Error(err)
		}

		m.table.SetRows(rows)

		m.status = statusMain
		cmd = tea.Batch(tea.ClearScreen, m.returnCmd)
	default:
		m.columns, cmd = m.columns.Update(msg)
	}

	return m, cmd
}
