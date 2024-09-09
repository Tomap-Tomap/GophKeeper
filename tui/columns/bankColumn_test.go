//go:build unit

package columns

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tomap-Tomap/GophKeeper/storage"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testBankRow = table.Row{
	testID,
	testName,
	testNumber,
	testCvc,
	testOwner,
	testExp,
	testMeta,
	testTime.Format(time.RFC1123),
}

func TestBanksColumns_Len(t *testing.T) {
	c := NewBanksColumns(context.Background(), new(MockClient))
	assert.Equal(t, 8, c.Len())
}

func TestBanksColumns_GetColums(t *testing.T) {
	c := NewBanksColumns(context.Background(), new(MockClient))
	assert.Equal(t, []table.Column{
		{Title: "ID", Width: 0},
		{Title: "Name", Width: 0},
		{Title: "Card number", Width: 0},
		{Title: "CVC", Width: 0},
		{Title: "Owner", Width: 0},
		{Title: "Exp", Width: 0},
		{Title: "Meta", Width: 0},
		{Title: "Update at", Width: 0},
	}, c.GetColums(0))
}

func TestBanksColumns_GetRows(t *testing.T) {
	wantBanks := []storage.Bank{
		{
			ID:         testID,
			Name:       testName,
			CardNumber: testNumber,
			CVC:        testCvc,
			Owner:      testOwner,
			Exp:        testExp,
			Meta:       testMeta,
			UpdateAt:   testTime,
		},
	}

	mc := new(MockClient)
	mc.On("GetAllBanks").Return(nil, errors.New("test error")).Once()
	mc.On("GetAllBanks").Return(wantBanks, nil).Once()
	defer mc.AssertExpectations(t)

	t.Run("cannot get banks data", func(t *testing.T) {
		c := NewBanksColumns(context.Background(), mc)
		row, err := c.GetRows()
		assert.ErrorContains(t, err, "cannot get banks data")
		assert.Nil(t, row)
	})

	t.Run("positive test", func(t *testing.T) {
		c := NewBanksColumns(context.Background(), mc)
		row, err := c.GetRows()
		assert.NoError(t, err)
		assert.Equal(t, []table.Row{testBankRow}, row)
	})
}

func TestBanksColumns_GetInfo(t *testing.T) {
	c := NewBanksColumns(context.Background(), new(MockClient))
	assert.Equal(t, "banks", c.GetInfo())
}

func TestBanksColumns_GetHelp(t *testing.T) {
	c := NewBanksColumns(context.Background(), new(MockClient))
	assert.Equal(t, "↑: move up • ↓: move down • enter: apply/back", c.GetHelp())
}

func TestBanksColumns_InitInsert(t *testing.T) {
	c := NewBanksColumns(context.Background(), new(MockClient))
	columner := c.InitInsert(nil)
	require.IsType(t, BanksColumns{}, columner)
	c = columner.(BanksColumns)
	id, name, number, cvc, owner, exp, meta := c.model.GetResult()
	assert.Equal(t, "", id)
	assert.Equal(t, "", name)
	assert.Equal(t, "", number)
	assert.Equal(t, "", cvc)
	assert.Equal(t, "", owner)
	assert.Equal(t, "", exp)
	assert.Equal(t, "", meta)
}

func TestBanksColumns_InitUpdate(t *testing.T) {
	c := NewBanksColumns(context.Background(), new(MockClient))
	columner := c.InitUpdate(nil, testBankRow)
	require.IsType(t, BanksColumns{}, columner)
	c = columner.(BanksColumns)
	id, name, number, cvc, owner, exp, meta := c.model.GetResult()
	assert.Equal(t, testID, id)
	assert.Equal(t, testName, name)
	assert.Equal(t, testNumber, number)
	assert.Equal(t, testCvc, cvc)
	assert.Equal(t, testOwner, owner)
	assert.Equal(t, testExp, exp)
	assert.Equal(t, testMeta, meta)
}

func TestBanksColumns_InitOpen(t *testing.T) {
	c := NewBanksColumns(context.Background(), new(MockClient))
	columner := c.InitOpen(nil, testBankRow)
	require.IsType(t, BanksColumns{}, columner)
	c = columner.(BanksColumns)
	id, name, number, cvc, owner, exp, meta := c.model.GetResult()
	assert.Equal(t, testID, id)
	assert.Equal(t, testName, name)
	assert.Equal(t, testNumber, number)
	assert.Equal(t, testCvc, cvc)
	assert.Equal(t, testOwner, owner)
	assert.Equal(t, testExp, exp)
	assert.Equal(t, testMeta, meta)
}

func TestBanksColumns_Update(t *testing.T) {
	c := NewBanksColumns(context.Background(), new(MockClient))
	columner, cmd := c.Update(nil)
	require.IsType(t, BanksColumns{}, columner)
	assert.Nil(t, cmd)
}

func TestBanksColumns_View(t *testing.T) {
	c := NewBanksColumns(context.Background(), new(MockClient))
	s := c.View()
	assert.Contains(t, s, "Name")
	assert.Contains(t, s, "Card number")
	assert.Contains(t, s, "CVC")
	assert.Contains(t, s, "OWNER")
	assert.Contains(t, s, "MM/YY")
	assert.Contains(t, s, "Meta")
}

func TestBanksColumns_SetSize(t *testing.T) {
	c := NewBanksColumns(context.Background(), new(MockClient))
	columner := c.SetSize(tea.WindowSizeMsg{})
	require.IsType(t, BanksColumns{}, columner)
}

func TestBanksColumns_Insert(t *testing.T) {
	mc := new(MockClient)
	mc.On("CreateBank", "", "", "", "", "", "").Return(errors.New("test error")).Once()
	mc.On("CreateBank", "", "", "", "", "", "").Return(nil).Once()
	mc.On("GetAllBanks").Return(nil, nil)
	defer mc.AssertExpectations(t)

	c := NewBanksColumns(context.Background(), mc)
	columner := c.InitInsert(nil)
	require.IsType(t, BanksColumns{}, columner)

	t.Run("error test", func(t *testing.T) {
		_, err := columner.Insert()
		assert.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		_, err := columner.Insert()
		assert.NoError(t, err)
	})
}

func TestBanksColumns_UpdateData(t *testing.T) {
	mc := new(MockClient)
	mc.On("UpdateBank", testID, testName, testNumber, testCvc, testOwner, testExp, testMeta).Return(errors.New("test error")).Once()
	mc.On("UpdateBank", testID, testName, testNumber, testCvc, testOwner, testExp, testMeta).Return(nil).Once()
	mc.On("GetAllBanks").Return(nil, nil)

	defer mc.AssertExpectations(t)

	c := NewBanksColumns(context.Background(), mc)
	columner := c.InitUpdate(nil, testBankRow)
	require.IsType(t, BanksColumns{}, columner)

	t.Run("error test", func(t *testing.T) {
		_, err := columner.UpdateData()
		assert.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		_, err := columner.UpdateData()
		assert.NoError(t, err)
	})
}

func TestBanksColumns_Open(t *testing.T) {
	mc := new(MockClient)
	mc.On("GetAllBanks").Return(nil, nil)

	defer mc.AssertExpectations(t)

	c := NewBanksColumns(context.Background(), mc)
	columner := c.InitOpen(nil, testBankRow)
	require.IsType(t, BanksColumns{}, columner)

	t.Run("positive test", func(t *testing.T) {
		_, err := columner.Open()
		assert.NoError(t, err)
	})
}

func TestBanksColumns_Delete(t *testing.T) {
	mc := new(MockClient)
	mc.On("DeleteBank", testID).Return(errors.New("test error")).Once()
	mc.On("DeleteBank", testID).Return(nil).Once()
	mc.On("GetAllBanks").Return(nil, nil)

	defer mc.AssertExpectations(t)

	c := NewBanksColumns(context.Background(), mc)

	t.Run("error test", func(t *testing.T) {
		_, err := c.Delete(testBankRow)
		assert.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		_, err := c.Delete(testBankRow)
		assert.NoError(t, err)
	})
}
