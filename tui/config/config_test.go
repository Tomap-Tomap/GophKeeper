//go:build unit

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testKey    = "key.aes"
	testAddr   = "localhost:0000"
	configPath = "testdata"
)

func TestNew(t *testing.T) {
	t.Run("test not exist", func(t *testing.T) {
		notExistPath := "notexist"
		cfg, err := New(notExistPath)

		require.NoError(t, err)
		assert.Equal(t, &Config{folderPath: notExistPath}, cfg)
	})

	t.Run("test error", func(t *testing.T) {
		errorPath := "errortestdata"
		cfg, err := New(errorPath)

		require.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("positive test", func(t *testing.T) {
		wantCfg := &Config{
			folderPath:      configPath,
			PathToSecretKey: testKey,
			AddrToService:   testAddr,
		}

		cfg, err := New(configPath)

		require.NoError(t, err)
		assert.Equal(t, wantCfg, cfg)
	})
}

func TestConfig_Save(t *testing.T) {
	t.Run("cannot create config file", func(t *testing.T) {
		cfg := &Config{
			folderPath: "errorpath",
		}

		err := cfg.Save()
		assert.ErrorContains(t, err, "cannot create config file")
	})

	t.Run("positive test", func(t *testing.T) {
		cfg := &Config{
			PathToSecretKey: testKey,
			AddrToService:   testAddr,
			folderPath:      t.TempDir(),
		}

		err := cfg.Save()
		require.NoError(t, err)

		file, err := os.Open(filepath.Join(cfg.folderPath, configName))
		require.NoError(t, err)
		require.NoError(t, file.Close())
	})
}
