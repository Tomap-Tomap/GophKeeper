package columns

import (
	"context"
	"fmt"
	"time"

	"github.com/Tomap-Tomap/GophKeeper/tui/banksdatamodel"
	"github.com/Tomap-Tomap/GophKeeper/tui/tablemodel"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// BanksColumns represents the columns of the bank table and the associated model.
type BanksColumns struct {
	ctx        context.Context
	client     Client
	model      banksdatamodel.Model
	id         column
	name       column
	cardNumber column
	cvc        column
	owner      column
	exp        column
	meta       column
	updateAt   column
}

// NewBanksColumns initializes and returns a new BanksColumns instance.
func NewBanksColumns(ctx context.Context, client Client) BanksColumns {
	return BanksColumns{
		ctx:        ctx,
		client:     client,
		model:      banksdatamodel.New(nil, "", "", "", "", "", "", ""),
		id:         column{"ID", 0},
		name:       column{"Name", 1},
		cardNumber: column{"Card number", 2},
		cvc:        column{"CVC", 3},
		owner:      column{"Owner", 4},
		exp:        column{"Exp", 5},
		meta:       column{"Meta", 6},
		updateAt:   column{"Update at", 7},
	}
}

// Len returns the number of columns.
func (c BanksColumns) Len() int {
	return 8
}

// GetColums returns the table columns with the specified width.
func (c BanksColumns) GetColums(w int) []table.Column {
	return []table.Column{
		{Title: c.id.name, Width: w},
		{Title: c.name.name, Width: w},
		{Title: c.cardNumber.name, Width: w},
		{Title: c.cvc.name, Width: w},
		{Title: c.owner.name, Width: w},
		{Title: c.exp.name, Width: w},
		{Title: c.meta.name, Width: w},
		{Title: c.updateAt.name, Width: w},
	}
}

// GetRows fetches the bank data from the client and returns it as table rows.
func (c BanksColumns) GetRows() ([]table.Row, error) {
	data, err := c.client.GetAllBanks(c.ctx)

	if err != nil {
		return nil, fmt.Errorf("cannot get banks data: %w", err)
	}

	rows := make([]table.Row, 0, len(data))

	for _, v := range data {
		rows = append(rows, table.Row{
			v.ID,
			v.Name,
			v.CardNumber,
			v.CVC,
			v.Owner,
			v.Exp,
			v.Meta,
			v.UpdateAt.Format(time.RFC1123),
		})
	}

	return rows, nil
}

// GetInfo returns a string representing the type of data.
func (c BanksColumns) GetInfo() string {
	return "banks"
}

// GetHelp returns a string with help instructions for navigating the table.
func (c BanksColumns) GetHelp() string {
	return "↑: move up • ↓: move down • enter: apply/back"
}

// InitInsert initializes the model for inserting a new bank.
func (c BanksColumns) InitInsert(enterCmd tea.Cmd) tablemodel.Columner {
	c.model = banksdatamodel.New(enterCmd, "", "", "", "", "", "", "")
	return c
}

// InitUpdate initializes the model for updating an existing bank.
func (c BanksColumns) InitUpdate(enterCmd tea.Cmd, row table.Row) tablemodel.Columner {
	c.model = banksdatamodel.New(
		enterCmd,
		row[c.id.idx],
		row[c.name.idx],
		row[c.cardNumber.idx],
		row[c.cvc.idx],
		row[c.owner.idx],
		row[c.exp.idx],
		row[c.meta.idx],
	)

	return c
}

// InitOpen initializes the model for opening an existing bank.
func (c BanksColumns) InitOpen(enterCmd tea.Cmd, row table.Row) tablemodel.Columner {
	c.model = banksdatamodel.New(
		enterCmd,
		row[c.id.idx],
		row[c.name.idx],
		row[c.cardNumber.idx],
		row[c.cvc.idx],
		row[c.owner.idx],
		row[c.exp.idx],
		row[c.meta.idx],
	)

	return c
}

// Update updates the model based on the received message.
func (c BanksColumns) Update(msg tea.Msg) (tablemodel.Columner, tea.Cmd) {
	var cmd tea.Cmd

	c.model, cmd = c.model.Update(msg)

	return c, cmd
}

// View returns the string representation of the model's view.
func (c BanksColumns) View() string {
	return c.model.View()
}

// SetSize updates the model with the new window size.
func (c BanksColumns) SetSize(msg tea.WindowSizeMsg) tablemodel.Columner {
	c.model, _ = c.model.Update(msg)

	return c
}

// Insert inserts a new bank using the client and returns the updated rows.
func (c BanksColumns) Insert() ([]table.Row, error) {
	_, name, number, cvc, owner, exp, meta := c.model.GetResult()

	err := c.client.CreateBank(c.ctx, name, number, cvc, owner, exp, meta)

	if err != nil {
		return nil, err
	}

	return c.GetRows()
}

// UpdateData updates an existing bank using the client and returns the updated rows.
func (c BanksColumns) UpdateData() ([]table.Row, error) {
	id, name, number, cvc, owner, exp, meta := c.model.GetResult()

	err := c.client.UpdateBank(c.ctx, id, name, number, cvc, owner, exp, meta)

	if err != nil {
		return nil, err
	}

	return c.GetRows()
}

// Open opens an existing bank using the client and returns the updated rows.
func (c BanksColumns) Open() ([]table.Row, error) {
	return c.GetRows()
}

// Delete deletes an existing bank using the client and returns the updated rows.
func (c BanksColumns) Delete(deleteRow table.Row) ([]table.Row, error) {
	err := c.client.DeleteBank(c.ctx, deleteRow[c.id.idx])

	if err != nil {
		return nil, err
	}

	return c.GetRows()
}
