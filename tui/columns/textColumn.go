package columns

import (
	"context"
	"fmt"
	"time"

	"github.com/Tomap-Tomap/GophKeeper/tui/tablemodel"
	"github.com/Tomap-Tomap/GophKeeper/tui/textdatamodel"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// TextColumns represents the columns of the text table and the associated model.
type TextColumns struct {
	ctx    context.Context
	client Client
	model  textdatamodel.Model

	id       column
	name     column
	text     column
	meta     column
	updateAt column
}

// NewTextColumns initializes and returns a new TextColumns instance.
func NewTextColumns(ctx context.Context, client Client) TextColumns {
	return TextColumns{
		ctx:      ctx,
		client:   client,
		model:    textdatamodel.New(nil, "", "", "", ""),
		id:       column{"ID", 0},
		name:     column{"Name", 1},
		text:     column{"Text", 2},
		meta:     column{"Meta", 3},
		updateAt: column{"Update at", 4},
	}
}

// Len returns the number of columns.
func (c TextColumns) Len() int {
	return 5
}

// GetColums returns the table columns with the specified width.
func (c TextColumns) GetColums(w int) []table.Column {
	return []table.Column{
		{Title: c.id.name, Width: w},
		{Title: c.name.name, Width: w},
		{Title: c.text.name, Width: w},
		{Title: c.meta.name, Width: w},
		{Title: c.updateAt.name, Width: w},
	}
}

// GetRows fetches the text data from the client and returns it as table rows.
func (c TextColumns) GetRows() ([]table.Row, error) {
	data, err := c.client.GetAllTexts(c.ctx)

	if err != nil {
		return nil, fmt.Errorf("cannot get text data: %w", err)
	}

	rows := make([]table.Row, 0, len(data))

	for _, v := range data {
		rows = append(rows, table.Row{
			v.ID,
			v.Name,
			v.Text,
			v.Meta,
			v.UpdateAt.Format(time.RFC1123),
		})
	}

	return rows, nil
}

// GetInfo returns a string representing the type of data.
func (c TextColumns) GetInfo() string {
	return "texts"
}

// GetHelp returns a string with help instructions for navigating the table.
func (c TextColumns) GetHelp() string {
	return "↑: move up • ↓: move down • enter: apply/back"
}

// InitInsert initializes the model for inserting a new text.
func (c TextColumns) InitInsert(enterCmd tea.Cmd) tablemodel.Columner {
	c.model = textdatamodel.New(enterCmd, "", "", "", "")
	return c
}

// InitUpdate initializes the model for updating an existing text.
func (c TextColumns) InitUpdate(enterCmd tea.Cmd, row table.Row) tablemodel.Columner {
	c.model = textdatamodel.New(
		enterCmd,
		row[c.id.idx],
		row[c.name.idx],
		row[c.meta.idx],
		row[c.text.idx],
	)

	return c
}

// InitOpen initializes the model for opening an existing text.
func (c TextColumns) InitOpen(enterCmd tea.Cmd, row table.Row) tablemodel.Columner {
	c.model = textdatamodel.New(
		enterCmd,
		row[c.id.idx],
		row[c.name.idx],
		row[c.meta.idx],
		row[c.text.idx],
	)

	return c
}

// Update updates the model based on the received message.
func (c TextColumns) Update(msg tea.Msg) (tablemodel.Columner, tea.Cmd) {
	var cmd tea.Cmd

	c.model, cmd = c.model.Update(msg)

	return c, cmd
}

// View returns the string representation of the model's view.
func (c TextColumns) View() string {
	return c.model.View()
}

// SetSize updates the model with the new window size.
func (c TextColumns) SetSize(msg tea.WindowSizeMsg) tablemodel.Columner {
	c.model, _ = c.model.Update(msg)

	return c
}

// Insert inserts a new text using the client and returns the updated rows.
func (c TextColumns) Insert() ([]table.Row, error) {
	_, name, meta, text := c.model.GetResult()

	err := c.client.CreateText(c.ctx, name, text, meta)

	if err != nil {
		return nil, err
	}

	return c.GetRows()
}

// UpdateData updates an existing text using the client and returns the updated rows.
func (c TextColumns) UpdateData() ([]table.Row, error) {
	id, name, meta, text := c.model.GetResult()

	err := c.client.UpdateText(c.ctx, id, name, text, meta)

	if err != nil {
		return nil, err
	}

	return c.GetRows()
}

// Open opens an existing text using the client and returns the updated rows.
func (c TextColumns) Open() ([]table.Row, error) {
	return c.GetRows()
}

// Delete deletes an existing text using the client and returns the updated rows.
func (c TextColumns) Delete(deleteRow table.Row) ([]table.Row, error) {
	err := c.client.DeleteText(c.ctx, deleteRow[c.id.idx])

	if err != nil {
		return nil, err
	}

	return c.GetRows()
}
