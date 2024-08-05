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

var testPasswordRow = table.Row{
	testID,
	testName,
	testLogin,
	testPassword,
	testMeta,
	testTime.Format(time.RFC1123),
}

func TestPasswordColumns_Len(t *testing.T) {
	c := NewPasswordColumns(context.Background(), new(MockClient))
	assert.Equal(t, 6, c.Len())
}

func TestPasswordColumns_GetColums(t *testing.T) {
	c := NewPasswordColumns(context.Background(), new(MockClient))
	assert.Equal(t, []table.Column{
		{Title: "ID", Width: 0},
		{Title: "Name", Width: 0},
		{Title: "Login", Width: 0},
		{Title: "Password", Width: 0},
		{Title: "Meta", Width: 0},
		{Title: "Update at", Width: 0},
	}, c.GetColums(0))
}

func TestPasswordColumns_GetRows(t *testing.T) {
	wantPasswords := []storage.Password{
		{
			ID:       testID,
			Name:     testName,
			Login:    testLogin,
			Password: testPassword,
			Meta:     testMeta,
			UpdateAt: testTime,
		},
	}

	mc := new(MockClient)
	mc.On("GetAllPasswords").Return(nil, errors.New("test error")).Once()
	mc.On("GetAllPasswords").Return(wantPasswords, nil).Once()
	defer mc.AssertExpectations(t)

	t.Run("cannot get passwords", func(t *testing.T) {
		c := NewPasswordColumns(context.Background(), mc)
		row, err := c.GetRows()
		assert.ErrorContains(t, err, "cannot get passwords")
		assert.Nil(t, row)
	})

	t.Run("positive test", func(t *testing.T) {
		c := NewPasswordColumns(context.Background(), mc)
		row, err := c.GetRows()
		assert.NoError(t, err)
		assert.Equal(t, []table.Row{testPasswordRow}, row)
	})
}

func TestPasswordColumns_GetInfo(t *testing.T) {
	c := NewPasswordColumns(context.Background(), new(MockClient))
	assert.Equal(t, "passwords", c.GetInfo())
}

func TestPasswordColumns_GetHelp(t *testing.T) {
	c := NewPasswordColumns(context.Background(), new(MockClient))
	assert.Equal(t, "↑: move up • ↓: move down • enter: apply/back", c.GetHelp())
}

func TestPasswordColumns_InitInsert(t *testing.T) {
	c := NewPasswordColumns(context.Background(), new(MockClient))
	columner := c.InitInsert(nil)
	require.IsType(t, PasswordColumns{}, columner)
	c = columner.(PasswordColumns)
	id, name, login, password, meta := c.model.GetResult()
	assert.Equal(t, "", id)
	assert.Equal(t, "", name)
	assert.Equal(t, "", login)
	assert.Equal(t, "", password)
	assert.Equal(t, "", meta)
}

func TestPasswordColumns_InitUpdate(t *testing.T) {
	c := NewPasswordColumns(context.Background(), new(MockClient))
	columner := c.InitUpdate(nil, testPasswordRow)
	require.IsType(t, PasswordColumns{}, columner)
	c = columner.(PasswordColumns)
	id, name, login, password, meta := c.model.GetResult()
	assert.Equal(t, testID, id)
	assert.Equal(t, testName, name)
	assert.Equal(t, testLogin, login)
	assert.Equal(t, testPassword, password)
	assert.Equal(t, testMeta, meta)
}

func TestPasswordColumns_InitOpen(t *testing.T) {
	c := NewPasswordColumns(context.Background(), new(MockClient))
	columner := c.InitOpen(nil, testPasswordRow)
	require.IsType(t, PasswordColumns{}, columner)
	c = columner.(PasswordColumns)
	id, name, login, password, meta := c.model.GetResult()
	assert.Equal(t, testID, id)
	assert.Equal(t, testName, name)
	assert.Equal(t, testLogin, login)
	assert.Equal(t, testPassword, password)
	assert.Equal(t, testMeta, meta)
}

func TestPasswordColumns_Update(t *testing.T) {
	c := NewPasswordColumns(context.Background(), new(MockClient))
	columner, cmd := c.Update(nil)
	require.IsType(t, PasswordColumns{}, columner)
	assert.Nil(t, cmd)
}

func TestPasswordColumns_View(t *testing.T) {
	c := NewPasswordColumns(context.Background(), new(MockClient))
	s := c.View()
	assert.Contains(t, s, "Name")
	assert.Contains(t, s, "Login")
	assert.Contains(t, s, "Password")
	assert.Contains(t, s, "Meta")
}

func TestPasswordColumns_SetSize(t *testing.T) {
	c := NewPasswordColumns(context.Background(), new(MockClient))
	columner := c.SetSize(tea.WindowSizeMsg{})
	require.IsType(t, PasswordColumns{}, columner)
}

func TestPasswordColumns_Insert(t *testing.T) {
	mc := new(MockClient)
	mc.On("CreatePassword", "", "", "", "").Return(errors.New("test error")).Once()
	mc.On("CreatePassword", "", "", "", "").Return(nil).Once()
	mc.On("GetAllPasswords").Return(nil, nil)

	defer mc.AssertExpectations(t)

	c := NewPasswordColumns(context.Background(), mc)
	columner := c.InitInsert(nil)
	require.IsType(t, PasswordColumns{}, columner)

	t.Run("error test", func(t *testing.T) {
		_, err := columner.Insert()
		assert.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		_, err := columner.Insert()
		assert.NoError(t, err)
	})
}

func TestPasswordColumns_UpdateData(t *testing.T) {
	mc := new(MockClient)
	mc.On("UpdatePassword", testID, testName, testLogin, testPassword, testMeta).Return(errors.New("test error")).Once()
	mc.On("UpdatePassword", testID, testName, testLogin, testPassword, testMeta).Return(nil).Once()
	mc.On("GetAllPasswords").Return(nil, nil)

	defer mc.AssertExpectations(t)

	c := NewPasswordColumns(context.Background(), mc)
	columner := c.InitUpdate(nil, testPasswordRow)
	require.IsType(t, PasswordColumns{}, columner)

	t.Run("error test", func(t *testing.T) {
		_, err := columner.UpdateData()
		assert.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		_, err := columner.UpdateData()
		assert.NoError(t, err)
	})
}

func TestPasswordColumns_Open(t *testing.T) {
	mc := new(MockClient)
	mc.On("GetAllPasswords").Return(nil, nil)

	defer mc.AssertExpectations(t)

	c := NewPasswordColumns(context.Background(), mc)
	columner := c.InitOpen(nil, testPasswordRow)
	require.IsType(t, PasswordColumns{}, columner)

	t.Run("positive test", func(t *testing.T) {
		_, err := columner.Open()
		assert.NoError(t, err)
	})
}

func TestPasswordColumns_Delete(t *testing.T) {
	mc := new(MockClient)
	mc.On("DeletePassword", testID).Return(errors.New("test error")).Once()
	mc.On("DeletePassword", testID).Return(nil).Once()
	mc.On("GetAllPasswords").Return(nil, nil)

	defer mc.AssertExpectations(t)

	c := NewPasswordColumns(context.Background(), mc)

	t.Run("error test", func(t *testing.T) {
		_, err := c.Delete(testPasswordRow)
		assert.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		_, err := c.Delete(testPasswordRow)
		assert.NoError(t, err)
	})
}
