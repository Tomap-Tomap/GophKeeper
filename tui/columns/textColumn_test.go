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

var testTextRow = table.Row{
	testID,
	testName,
	testText,
	testMeta,
	testTime.Format(time.RFC1123),
}

func TestTextColumns_Len(t *testing.T) {
	c := NewTextColumns(context.Background(), new(MockClient))
	assert.Equal(t, 5, c.Len())
}

func TestTextColumns_GetColums(t *testing.T) {
	c := NewTextColumns(context.Background(), new(MockClient))
	assert.Equal(t, []table.Column{
		{Title: "ID", Width: 0},
		{Title: "Name", Width: 0},
		{Title: "Text", Width: 0},
		{Title: "Meta", Width: 0},
		{Title: "Update at", Width: 0},
	}, c.GetColums(0))
}

func TestTextColumns_GetRows(t *testing.T) {
	wantTexts := []storage.Text{
		{
			ID:       testID,
			Name:     testName,
			Text:     testText,
			Meta:     testMeta,
			UpdateAt: testTime,
		},
	}

	mc := new(MockClient)
	mc.On("GetAllTexts").Return(nil, errors.New("test error")).Once()
	mc.On("GetAllTexts").Return(wantTexts, nil).Once()
	defer mc.AssertExpectations(t)

	t.Run("cannot get text data", func(t *testing.T) {
		c := NewTextColumns(context.Background(), mc)
		row, err := c.GetRows()
		assert.ErrorContains(t, err, "cannot get text data")
		assert.Nil(t, row)
	})

	t.Run("positive test", func(t *testing.T) {
		c := NewTextColumns(context.Background(), mc)
		row, err := c.GetRows()
		assert.NoError(t, err)
		assert.Equal(t, []table.Row{testTextRow}, row)
	})
}

func TestTextColumns_GetInfo(t *testing.T) {
	c := NewTextColumns(context.Background(), new(MockClient))
	assert.Equal(t, "texts", c.GetInfo())
}

func TestTextColumns_GetHelp(t *testing.T) {
	c := NewTextColumns(context.Background(), new(MockClient))
	assert.Equal(t, "↑: move up • ↓: move down • enter: apply/back", c.GetHelp())
}

func TestTextColumns_InitInsert(t *testing.T) {
	c := NewTextColumns(context.Background(), new(MockClient))
	columner := c.InitInsert(nil)
	require.IsType(t, TextColumns{}, columner)
	c = columner.(TextColumns)
	id, name, meta, text := c.model.GetResult()
	assert.Equal(t, "", id)
	assert.Equal(t, "", name)
	assert.Equal(t, "", meta)
	assert.Equal(t, "", text)
}

func TestTextColumns_InitUpdate(t *testing.T) {
	c := NewTextColumns(context.Background(), new(MockClient))
	columner := c.InitUpdate(nil, testTextRow)
	require.IsType(t, TextColumns{}, columner)
	c = columner.(TextColumns)
	id, name, meta, text := c.model.GetResult()
	assert.Equal(t, testID, id)
	assert.Equal(t, testName, name)
	assert.Equal(t, testMeta, meta)
	assert.Equal(t, testText, text)
}

func TestTextColumns_InitOpen(t *testing.T) {
	c := NewTextColumns(context.Background(), new(MockClient))
	columner := c.InitOpen(nil, testTextRow)
	require.IsType(t, TextColumns{}, columner)
	c = columner.(TextColumns)
	id, name, meta, text := c.model.GetResult()
	assert.Equal(t, testID, id)
	assert.Equal(t, testName, name)
	assert.Equal(t, testMeta, meta)
	assert.Equal(t, testText, text)
}

func TestTextColumns_Update(t *testing.T) {
	c := NewTextColumns(context.Background(), new(MockClient))
	columner, cmd := c.Update(nil)
	require.IsType(t, TextColumns{}, columner)
	assert.Nil(t, cmd)
}

func TestTextColumns_View(t *testing.T) {
	c := NewTextColumns(context.Background(), new(MockClient))
	s := c.View()
	assert.Contains(t, s, "Name")
	assert.Contains(t, s, "Meta")
}

func TestTextColumns_SetSize(t *testing.T) {
	c := NewTextColumns(context.Background(), new(MockClient))
	columner := c.SetSize(tea.WindowSizeMsg{})
	require.IsType(t, TextColumns{}, columner)
}

func TestTextColumns_Insert(t *testing.T) {
	mc := new(MockClient)
	mc.On("CreateText", "", "", "").Return(errors.New("test error")).Once()
	mc.On("CreateText", "", "", "").Return(nil).Once()
	mc.On("GetAllTexts").Return(nil, nil)

	defer mc.AssertExpectations(t)

	c := NewTextColumns(context.Background(), mc)
	columner := c.InitInsert(nil)
	require.IsType(t, TextColumns{}, columner)

	t.Run("error test", func(t *testing.T) {
		_, err := columner.Insert()
		assert.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		_, err := columner.Insert()
		assert.NoError(t, err)
	})
}

func TestTextColumns_UpdateData(t *testing.T) {
	mc := new(MockClient)
	mc.On("UpdateText", testID, testName, testText, testMeta).Return(errors.New("test error")).Once()
	mc.On("UpdateText", testID, testName, testText, testMeta).Return(nil).Once()
	mc.On("GetAllTexts").Return(nil, nil)

	defer mc.AssertExpectations(t)

	c := NewTextColumns(context.Background(), mc)
	columner := c.InitUpdate(nil, testTextRow)
	require.IsType(t, TextColumns{}, columner)

	t.Run("error test", func(t *testing.T) {
		_, err := columner.UpdateData()
		assert.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		_, err := columner.UpdateData()
		assert.NoError(t, err)
	})
}

func TestTextColumns_Open(t *testing.T) {
	mc := new(MockClient)
	mc.On("GetAllTexts").Return(nil, nil)

	defer mc.AssertExpectations(t)

	c := NewTextColumns(context.Background(), mc)
	columner := c.InitOpen(nil, testTextRow)
	require.IsType(t, TextColumns{}, columner)

	t.Run("positive test", func(t *testing.T) {
		_, err := columner.Open()
		assert.NoError(t, err)
	})
}

func TestTextColumns_Delete(t *testing.T) {
	mc := new(MockClient)
	mc.On("DeleteText", testID).Return(errors.New("test error")).Once()
	mc.On("DeleteText", testID).Return(nil).Once()
	mc.On("GetAllTexts").Return(nil, nil)

	defer mc.AssertExpectations(t)

	c := NewTextColumns(context.Background(), mc)

	t.Run("error test", func(t *testing.T) {
		_, err := c.Delete(testTextRow)
		assert.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		_, err := c.Delete(testTextRow)
		assert.NoError(t, err)
	})
}
