package columns

import (
	"context"
	"fmt"
	"time"

	"github.com/Tomap-Tomap/GophKeeper/tui/filesdatamodel"
	"github.com/Tomap-Tomap/GophKeeper/tui/tablemodel"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// FileColumns represents the columns of the file table and the associated model.
type FileColumns struct {
	ctx    context.Context
	client Client
	model  filesdatamodel.Model

	id       column
	name     column
	meta     column
	updateAt column
}

// NewFileColumns initializes and returns a new FileColumns instance.
func NewFileColumns(ctx context.Context, client Client) FileColumns {
	fc := FileColumns{
		ctx:      ctx,
		client:   client,
		model:    filesdatamodel.New(nil, "", "", "", "", false),
		id:       column{"ID", 0},
		name:     column{"Name", 1},
		meta:     column{"Meta", 2},
		updateAt: column{"Update at", 3},
	}

	return fc
}

// Len returns the number of columns.
func (c FileColumns) Len() int {
	return 4
}

// GetColums returns the table columns with the specified width.
func (c FileColumns) GetColums(w int) []table.Column {
	return []table.Column{
		{Title: c.id.name, Width: w},
		{Title: c.name.name, Width: w},
		{Title: c.meta.name, Width: w},
		{Title: c.updateAt.name, Width: w},
	}
}

// GetRows fetches the file data from the client and returns it as table rows.
func (c FileColumns) GetRows() ([]table.Row, error) {
	data, err := c.client.GetAllFiles(c.ctx)

	if err != nil {
		return nil, fmt.Errorf("cannot get text data: %w", err)
	}

	rows := make([]table.Row, 0, len(data))

	for _, v := range data {
		rows = append(rows, table.Row{
			v.ID,
			v.Name,
			v.Meta,
			v.UpdateAt.Format(time.RFC1123),
		})
	}

	return rows, nil
}

// GetInfo returns a string representing the type of data.
func (c FileColumns) GetInfo() string {
	return "files"
}

// GetHelp returns a string with help instructions for navigating the table.
func (c FileColumns) GetHelp() string {
	return "↑: move up • ↓: move down • enter: open"
}

// InitInsert initializes the model for inserting a new file.
func (c FileColumns) InitInsert(enterCmd tea.Cmd) tablemodel.Columner {
	c.model = filesdatamodel.New(enterCmd, "", "", "", "Upload", false)
	return c
}

// InitUpdate initializes the model for updating an existing file.
func (c FileColumns) InitUpdate(enterCmd tea.Cmd, row table.Row) tablemodel.Columner {
	c.model = filesdatamodel.New(
		enterCmd,
		row[c.id.idx],
		row[c.name.idx],
		row[c.meta.idx],
		"Upload",
		false,
	)

	return c
}

// InitOpen initializes the model for opening an existing file.
func (c FileColumns) InitOpen(enterCmd tea.Cmd, row table.Row) tablemodel.Columner {
	c.model = filesdatamodel.New(
		enterCmd,
		row[c.id.idx],
		row[c.name.idx],
		row[c.meta.idx],
		"Download",
		true,
	)

	return c
}

// Update updates the model based on the received message.
func (c FileColumns) Update(msg tea.Msg) (tablemodel.Columner, tea.Cmd) {
	var cmd tea.Cmd

	c.model, cmd = c.model.Update(msg)

	return c, cmd
}

// View returns the string representation of the model's view.
func (c FileColumns) View() string {
	return c.model.View()
}

// SetSize updates the model with the new window size.
func (c FileColumns) SetSize(msg tea.WindowSizeMsg) tablemodel.Columner {
	c.model, _ = c.model.Update(msg)

	return c
}

// Insert inserts a new file using the client and returns the updated rows.
func (c FileColumns) Insert() ([]table.Row, error) {
	_, name, path, meta := c.model.GetResult()

	err := c.client.CreateFile(c.ctx, name, path, meta)

	if err != nil {
		return nil, err
	}

	return c.GetRows()
}

// UpdateData updates an existing file using the client and returns the updated rows.
func (c FileColumns) UpdateData() ([]table.Row, error) {
	id, name, path, meta := c.model.GetResult()

	err := c.client.UpdateFile(c.ctx, id, name, path, meta)

	if err != nil {
		return nil, err
	}

	return c.GetRows()
}

// Open opens an existing file using the client and returns the updated rows.
func (c FileColumns) Open() ([]table.Row, error) {
	id, _, path, _ := c.model.GetResult()

	err := c.client.GetFile(c.ctx, id, path)

	if err != nil {
		return nil, err
	}

	return c.GetRows()
}

// Delete deletes an existing file using the client and returns the updated rows.
func (c FileColumns) Delete(deleteRow table.Row) ([]table.Row, error) {
	err := c.client.DeleteFile(c.ctx, deleteRow[c.id.idx])

	if err != nil {
		return nil, err
	}

	return c.GetRows()
}
