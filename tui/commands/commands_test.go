//go:build unit

package commands_test

import (
	"errors"
	"testing"

	"github.com/Tomap-Tomap/GophKeeper/tui/commands"
	"github.com/Tomap-Tomap/GophKeeper/tui/messages"
	"github.com/stretchr/testify/assert"
)

func TestSetWindowSize(t *testing.T) {
	cmd := commands.SetWindowSize()
	msg := cmd()

	_, ok := msg.(messages.Error)
	assert.True(t, ok)
}

func TestError(t *testing.T) {
	err := errors.New("test error")
	cmd := commands.Error(err)
	msg := cmd()

	errorMsg, ok := msg.(messages.Error)
	assert.True(t, ok)
	assert.Equal(t, err, errorMsg.Err)
}

func TestSetInfo(t *testing.T) {
	info := "test info"
	help := "test help"
	cmd := commands.SetInfo(info, help)
	msg := cmd()

	infoMsg, ok := msg.(messages.Info)
	assert.True(t, ok)
	assert.Equal(t, info, infoMsg.Info)
	assert.Equal(t, help, infoMsg.Help)
}

func TestOpenConfigModel(t *testing.T) {
	cmd := commands.OpenConfigModel()
	msg := cmd()

	_, ok := msg.(messages.OpenConfigModel)
	assert.True(t, ok)
}

func TestCloseConfigModel(t *testing.T) {
	pathToKey := "test/path"
	addrToService := "test/service"
	cmd := commands.CloseConfigModel(pathToKey, addrToService)
	msg := cmd()

	closeConfigModelMsg, ok := msg.(messages.CloseConfigModel)
	assert.True(t, ok)
	assert.Equal(t, pathToKey, closeConfigModelMsg.PathToKey)
	assert.Equal(t, addrToService, closeConfigModelMsg.AddrToService)
}
