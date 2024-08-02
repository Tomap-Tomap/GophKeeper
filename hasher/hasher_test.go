//go:build unit

package hasher

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasher_GenerateSalt(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		hasher := NewHasher()

		salt, err := hasher.GenerateSalt(75)
		require.NoError(t, err)
		assert.NotEmpty(t, salt)
	})
}

func TestHasher_GetHash(t *testing.T) {
	hasher := NewHasher()

	testStr := "Password"
	hashStr := hasher.GetHash(testStr)

	require.NotEmpty(t, hashStr)
	assert.NotEqual(t, testStr, hashStr)
}

func TestHasher_GetHashWithSalt(t *testing.T) {
	t.Run("decode error", func(t *testing.T) {
		hasher := NewHasher()
		got, err := hasher.GetHashWithSalt("", "asd")
		require.ErrorContains(t, err, "cannot decode salt")
		assert.Empty(t, got)
	})

	t.Run("positive test", func(t *testing.T) {
		salt := make([]byte, 10)

		_, err := rand.Read(salt)
		require.NoError(t, err)

		hasher := NewHasher()
		got, err := hasher.GetHashWithSalt("", hex.EncodeToString(salt))
		require.NoError(t, err)
		assert.NotEmpty(t, got)
	})
}
