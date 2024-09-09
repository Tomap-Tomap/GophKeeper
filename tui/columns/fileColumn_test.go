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

var testFileRow = table.Row{
	testID,
	testName,
	testMeta,
	testTime.Format(time.RFC1123),
}

func TestFileColumns_Len(t *testing.T) {
	c := NewFileColumns(context.Background(), new(MockClient))
	assert.Equal(t, 4, c.Len())
}

func TestFileColumns_GetColums(t *testing.T) {
	c := NewFileColumns(context.Background(), new(MockClient))
	assert.Equal(t, []table.Column{
		{Title: "ID", Width: 0},
		{Title: "Name", Width: 0},
		{Title: "Meta", Width: 0},
		{Title: "Update at", Width: 0},
	}, c.GetColums(0))
}

func TestFileColumns_GetRows(t *testing.T) {
	wantFiles := []storage.File{
		{
			ID:       testID,
			Name:     testName,
			Meta:     testMeta,
			UpdateAt: testTime,
		},
	}

	mc := new(MockClient)
	mc.On("GetAllFiles").Return(nil, errors.New("test error")).Once()
	mc.On("GetAllFiles").Return(wantFiles, nil).Once()
	defer mc.AssertExpectations(t)

	t.Run("cannot get text data", func(t *testing.T) {
		c := NewFileColumns(context.Background(), mc)
		row, err := c.GetRows()
		assert.ErrorContains(t, err, "cannot get text data")
		assert.Nil(t, row)
	})

	t.Run("positive test", func(t *testing.T) {
		c := NewFileColumns(context.Background(), mc)
		row, err := c.GetRows()
		assert.NoError(t, err)
		assert.Equal(t, []table.Row{testFileRow}, row)
	})
}

func TestFileColumns_GetInfo(t *testing.T) {
	c := NewFileColumns(context.Background(), new(MockClient))
	assert.Equal(t, "files", c.GetInfo())
}

func TestFileColumns_GetHelp(t *testing.T) {
	c := NewFileColumns(context.Background(), new(MockClient))
	assert.Equal(t, "↑: move up • ↓: move down • enter: open", c.GetHelp())
}

func TestFileColumns_InitInsert(t *testing.T) {
	c := NewFileColumns(context.Background(), new(MockClient))
	columner := c.InitInsert(nil)
	require.IsType(t, FileColumns{}, columner)
	c = columner.(FileColumns)
	id, name, path, meta := c.model.GetResult()
	assert.Equal(t, "", id)
	assert.Equal(t, "", name)
	assert.Equal(t, "", path)
	assert.Equal(t, "", meta)
}

func TestFileColumns_InitUpdate(t *testing.T) {
	c := NewFileColumns(context.Background(), new(MockClient))
	columner := c.InitUpdate(nil, testFileRow)
	require.IsType(t, FileColumns{}, columner)
	c = columner.(FileColumns)
	id, name, path, meta := c.model.GetResult()
	assert.Equal(t, testID, id)
	assert.Equal(t, testName, name)
	assert.Equal(t, "", path)
	assert.Equal(t, testMeta, meta)
}

func TestFileColumns_InitOpen(t *testing.T) {
	c := NewFileColumns(context.Background(), new(MockClient))
	columner := c.InitOpen(nil, testFileRow)
	require.IsType(t, FileColumns{}, columner)
	c = columner.(FileColumns)
	id, name, path, meta := c.model.GetResult()
	assert.Equal(t, testID, id)
	assert.Equal(t, testName, name)
	assert.Equal(t, "", path)
	assert.Equal(t, testMeta, meta)
}

func TestFileColumns_Update(t *testing.T) {
	c := NewFileColumns(context.Background(), new(MockClient))
	columner, cmd := c.Update(nil)
	require.IsType(t, FileColumns{}, columner)
	assert.Nil(t, cmd)
}

func TestFileColumns_View(t *testing.T) {
	c := NewFileColumns(context.Background(), new(MockClient))
	s := c.View()
	assert.Contains(t, s, "Name")
	assert.Contains(t, s, "Meta")
}

func TestFileColumns_SetSize(t *testing.T) {
	c := NewFileColumns(context.Background(), new(MockClient))
	columner := c.SetSize(tea.WindowSizeMsg{})
	require.IsType(t, FileColumns{}, columner)
}

func TestFileColumns_Insert(t *testing.T) {
	mc := new(MockClient)
	mc.On("CreateFile", "", "", "").Return(errors.New("test error")).Once()
	mc.On("CreateFile", "", "", "").Return(nil).Once()
	mc.On("GetAllFiles").Return(nil, nil)

	defer mc.AssertExpectations(t)

	c := NewFileColumns(context.Background(), mc)
	columner := c.InitInsert(nil)
	require.IsType(t, FileColumns{}, columner)

	t.Run("error test", func(t *testing.T) {
		_, err := columner.Insert()
		assert.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		_, err := columner.Insert()
		assert.NoError(t, err)
	})
}

func TestFileColumns_UpdateData(t *testing.T) {
	mc := new(MockClient)
	mc.On("UpdateFile", testID, testName, "", testMeta).Return(errors.New("test error")).Once()
	mc.On("UpdateFile", testID, testName, "", testMeta).Return(nil).Once()
	mc.On("GetAllFiles").Return(nil, nil)

	defer mc.AssertExpectations(t)

	c := NewFileColumns(context.Background(), mc)
	columner := c.InitUpdate(nil, testFileRow)
	require.IsType(t, FileColumns{}, columner)

	t.Run("error test", func(t *testing.T) {
		_, err := columner.UpdateData()
		assert.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		_, err := columner.UpdateData()
		assert.NoError(t, err)
	})
}

func TestFileColumns_Open(t *testing.T) {
	mc := new(MockClient)
	mc.On("GetFile", testID, "").Return(errors.New("test error")).Once()
	mc.On("GetFile", testID, "").Return(nil).Once()
	mc.On("GetAllFiles").Return(nil, nil)

	defer mc.AssertExpectations(t)

	c := NewFileColumns(context.Background(), mc)
	columner := c.InitOpen(nil, testFileRow)
	require.IsType(t, FileColumns{}, columner)

	t.Run("error test", func(t *testing.T) {
		_, err := columner.Open()
		assert.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		_, err := columner.Open()
		assert.NoError(t, err)
	})
}

func TestFileColumns_Delete(t *testing.T) {
	mc := new(MockClient)
	mc.On("DeleteFile", testID).Return(errors.New("test error")).Once()
	mc.On("DeleteFile", testID).Return(nil).Once()
	mc.On("GetAllFiles").Return(nil, nil)

	defer mc.AssertExpectations(t)

	c := NewFileColumns(context.Background(), mc)

	t.Run("error test", func(t *testing.T) {
		_, err := c.Delete(testFileRow)
		assert.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		_, err := c.Delete(testFileRow)
		assert.NoError(t, err)
	})
}
