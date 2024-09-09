package columns

import (
	"context"
	"fmt"
	"time"

	"github.com/Tomap-Tomap/GophKeeper/tui/passworddatamodel"
	"github.com/Tomap-Tomap/GophKeeper/tui/tablemodel"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// PasswordColumns represents the columns of the password table and the associated model.
type PasswordColumns struct {
	ctx      context.Context
	client   Client
	model    passworddatamodel.Model
	id       column
	name     column
	login    column
	password column
	meta     column
	updateAt column
}

// NewPasswordColumns initializes and returns a new PasswordColumns instance.
func NewPasswordColumns(ctx context.Context, client Client) PasswordColumns {
	return PasswordColumns{
		ctx:      ctx,
		client:   client,
		model:    passworddatamodel.New(nil, "", "", "", "", ""),
		id:       column{"ID", 0},
		name:     column{"Name", 1},
		login:    column{"Login", 2},
		password: column{"Password", 3},
		meta:     column{"Meta", 4},
		updateAt: column{"Update at", 5},
	}
}

// Len returns the number of columns.
func (c PasswordColumns) Len() int {
	return 6
}

// GetColums returns the table columns with the specified width.
func (c PasswordColumns) GetColums(w int) []table.Column {
	return []table.Column{
		{Title: c.id.name, Width: w},
		{Title: c.name.name, Width: w},
		{Title: c.login.name, Width: w},
		{Title: c.password.name, Width: w},
		{Title: c.meta.name, Width: w},
		{Title: c.updateAt.name, Width: w},
	}
}

// GetRows fetches the password data from the client and returns it as table rows.
func (c PasswordColumns) GetRows() ([]table.Row, error) {
	data, err := c.client.GetAllPasswords(c.ctx)

	if err != nil {
		return nil, fmt.Errorf("cannot get passwords: %w", err)
	}

	rows := make([]table.Row, 0, len(data))

	for _, v := range data {
		rows = append(rows, table.Row{
			v.ID,
			v.Name,
			v.Login,
			v.Password,
			v.Meta,
			v.UpdateAt.Format(time.RFC1123),
		})
	}

	return rows, nil
}

// GetInfo returns a string representing the type of data.
func (c PasswordColumns) GetInfo() string {
	return "passwords"
}

// GetHelp returns a string with help instructions for navigating the table.
func (c PasswordColumns) GetHelp() string {
	return "↑: move up • ↓: move down • enter: apply/back"
}

// InitInsert initializes the model for inserting a new password.
func (c PasswordColumns) InitInsert(enterCmd tea.Cmd) tablemodel.Columner {
	c.model = passworddatamodel.New(enterCmd, "", "", "", "", "")
	return c
}

// InitUpdate initializes the model for updating an existing password.
func (c PasswordColumns) InitUpdate(enterCmd tea.Cmd, row table.Row) tablemodel.Columner {
	c.model = passworddatamodel.New(
		enterCmd,
		row[c.id.idx],
		row[c.name.idx],
		row[c.login.idx],
		row[c.password.idx],
		row[c.meta.idx],
	)

	return c
}

// InitOpen initializes the model for opening an existing password.
func (c PasswordColumns) InitOpen(enterCmd tea.Cmd, row table.Row) tablemodel.Columner {
	c.model = passworddatamodel.New(
		enterCmd,
		row[c.id.idx],
		row[c.name.idx],
		row[c.login.idx],
		row[c.password.idx],
		row[c.meta.idx],
	)

	return c
}

// Update updates the model based on the received message.
func (c PasswordColumns) Update(msg tea.Msg) (tablemodel.Columner, tea.Cmd) {
	var cmd tea.Cmd

	c.model, cmd = c.model.Update(msg)

	return c, cmd
}

// View returns the string representation of the model's view.
func (c PasswordColumns) View() string {
	return c.model.View()
}

// SetSize updates the model with the new window size.
func (c PasswordColumns) SetSize(msg tea.WindowSizeMsg) tablemodel.Columner {
	c.model, _ = c.model.Update(msg)

	return c
}

// Insert inserts a new password using the client and returns the updated rows.
func (c PasswordColumns) Insert() ([]table.Row, error) {
	_, name, login, password, meta := c.model.GetResult()

	err := c.client.CreatePassword(c.ctx, name, login, password, meta)

	if err != nil {
		return nil, err
	}

	return c.GetRows()
}

// UpdateData updates an existing password using the client and returns the updated rows.
func (c PasswordColumns) UpdateData() ([]table.Row, error) {
	id, name, login, password, meta := c.model.GetResult()

	err := c.client.UpdatePassword(c.ctx, id, name, login, password, meta)

	if err != nil {
		return nil, err
	}

	return c.GetRows()
}

// Open opens an existing password using the client and returns the updated rows.
func (c PasswordColumns) Open() ([]table.Row, error) {
	return c.GetRows()
}

// Delete deletes an existing password using the client and returns the updated rows.
func (c PasswordColumns) Delete(deleteRow table.Row) ([]table.Row, error) {
	err := c.client.DeletePassword(c.ctx, deleteRow[c.id.idx])

	if err != nil {
		return nil, err
	}

	return c.GetRows()
}
