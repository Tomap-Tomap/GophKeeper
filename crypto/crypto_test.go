//go:build unit

package crypto

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CryptoTestSuite struct {
	suite.Suite
	crypter         *Crypter
	tempDir         string
	errDir          string
	keyPath         string
	testKey         []byte
	testNonce       []byte
	testNonceForErr []byte
	testMessage     string
	testBytes       []byte
}

func (suite *CryptoTestSuite) SetupSuite() {
	require := suite.Require()

	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(err)
	suite.tempDir = tempDir
	suite.errDir = "№*?;"
	suite.testMessage = "hello world"
	suite.testBytes = []byte("hello world")
	suite.testKey = make([]byte, 32) // AES-256
	_, err = rand.Read(suite.testKey)
	require.NoError(err)

	suite.testNonce = make([]byte, 12) // GCM standard nonce length
	_, err = rand.Read(suite.testNonce)
	require.NoError(err)

	suite.testNonceForErr = make([]byte, 12) // GCM standard nonce length
	_, err = rand.Read(suite.testNonceForErr)
	require.NoError(err)

	crypter, keyPath, err := NewCrypter(len(suite.testKey), suite.tempDir)
	require.NoError(err)

	suite.keyPath = keyPath
	suite.crypter = crypter
}

func (suite *CryptoTestSuite) TearDownSuite() {
	err := os.Remove(suite.keyPath)
	suite.Require().NoError(err)
}

func (suite *CryptoTestSuite) TestNewCrypter() {
	assert := suite.Require()

	suite.Run("positive test", func() {
		newCrypter, _, err := NewCrypter(32, suite.tempDir)
		assert.NoError(err)
		assert.NotNil(newCrypter)
	})

	suite.Run("cannot create key file", func() {
		newCrypter, _, err := NewCrypter(32, suite.errDir)
		assert.ErrorContains(err, "cannot create key file")
		assert.Nil(newCrypter)
	})

	suite.Run("cannot create AEAD", func() {
		tempDir, err := os.MkdirTemp("", "testErr")
		assert.NoError(err)
		newCrypter, _, err := NewCrypter(0, tempDir)
		assert.ErrorContains(err, "cannot create AEAD")
		assert.Nil(newCrypter)
		err = os.Remove(filepath.Join(tempDir, "key.aes"))
		assert.NoError(err)
	})
}

func (suite *CryptoTestSuite) TestNewCrypterByFile() {
	assert := suite.Require()

	suite.Run("positive test", func() {
		newCrypter, err := NewCrypterByFile(suite.keyPath)
		assert.NoError(err)
		assert.NotNil(newCrypter)
	})

	suite.Run("cannot open secret key", func() {
		newCrypter, err := NewCrypterByFile(suite.errDir)
		assert.ErrorContains(err, "cannot open secret key")
		assert.Nil(newCrypter)
	})

}

func (suite *CryptoTestSuite) TestGenerateNonce() {
	assert := suite.Require()
	nonce, err := suite.crypter.GenerateNonce()
	assert.NoError(err)
	assert.NotEmpty(nonce)
}

func (suite *CryptoTestSuite) TestNonceSize() {
	assert := suite.Require()
	size := suite.crypter.NonceSize()
	assert.Equal(suite.crypter.aesgcm.NonceSize(), size)
}

func (suite *CryptoTestSuite) TestSealString() {
	assert := suite.Require()

	encrypted, err := suite.crypter.SealString(suite.testMessage, suite.testNonce)
	assert.NoError(err)
	assert.NotEmpty(encrypted)
}

func (suite *CryptoTestSuite) TestSealStringWithoutNonce() {
	assert := suite.Require()
	encrypted, err := suite.crypter.SealStringWithoutNonce(suite.testMessage)
	assert.NoError(err)
	assert.NotEmpty(encrypted)
}

func (suite *CryptoTestSuite) TestOpenString() {
	assert := suite.Require()

	suite.Run("positive test", func() {
		encrypted, err := suite.crypter.SealString(suite.testMessage, suite.testNonce)
		assert.NoError(err)

		decrypted, err := suite.crypter.OpenString(encrypted, suite.testNonce)
		assert.NoError(err)
		assert.Equal(suite.testMessage, decrypted)
	})

	suite.Run("cannot decode string", func() {
		decrypted, err := suite.crypter.OpenString("#@!", suite.testNonce)
		assert.ErrorContains(err, "cannot decode string")
		assert.Equal("", decrypted)
	})

	suite.Run("cannot open string", func() {
		encrypted, err := suite.crypter.SealString(suite.testMessage, suite.testNonce)
		assert.NoError(err)

		decrypted, err := suite.crypter.OpenString(encrypted, suite.testNonceForErr)
		assert.ErrorContains(err, "cannot open string")
		assert.Equal("", decrypted)
	})
}

func (suite *CryptoTestSuite) TestOpenStringWithoutNonce() {
	assert := suite.Require()
	encrypted, err := suite.crypter.SealStringWithoutNonce(suite.testMessage)
	assert.NoError(err)

	decrypted, err := suite.crypter.OpenStringWithoutNonce(encrypted)
	assert.NoError(err)
	assert.Equal(suite.testMessage, decrypted)
}

func (suite *CryptoTestSuite) TestAddNonceInString() {
	assert := suite.Require()

	suite.Run("test at end", func() {
		encrypted, err := suite.crypter.SealString(suite.testMessage, suite.testNonce)
		assert.NoError(err)

		withNonce, err := suite.crypter.AddNonceInString(encrypted, suite.testNonce, AtEnd)
		assert.NoError(err)
		assert.NotEmpty(withNonce)
	})

	suite.Run("test at front", func() {
		encrypted, err := suite.crypter.SealString(suite.testMessage, suite.testNonce)
		assert.NoError(err)

		withNonce, err := suite.crypter.AddNonceInString(encrypted, suite.testNonce, AtFront)
		assert.NoError(err)
		assert.NotEmpty(withNonce)
	})

	suite.Run("cannot decode string", func() {
		withNonce, err := suite.crypter.AddNonceInString("$@!", suite.testNonce, AtFront)
		assert.ErrorContains(err, "cannot decode string")
		assert.Empty(withNonce)
	})

	suite.Run("ErrUnknownLocation", func() {
		encrypted, err := suite.crypter.SealString(suite.testMessage, suite.testNonce)
		assert.NoError(err)

		withNonce, err := suite.crypter.AddNonceInString(encrypted, suite.testNonce, 3)
		assert.ErrorIs(err, ErrUnknownLocation)
		assert.Empty(withNonce)
	})
}

func (suite *CryptoTestSuite) TestAddNonceInBytes() {
	assert := suite.Require()

	suite.Run("test at end", func() {
		encrypted := suite.crypter.SealBytes(suite.testBytes, suite.testNonce)

		withNonce, err := suite.crypter.AddNonceInBytes(encrypted, suite.testNonce, AtEnd)
		assert.NoError(err)
		assert.NotNil(withNonce)
	})

	suite.Run("test at front", func() {
		encrypted := suite.crypter.SealBytes(suite.testBytes, suite.testNonce)

		withNonce, err := suite.crypter.AddNonceInBytes(encrypted, suite.testNonce, AtFront)
		assert.NoError(err)
		assert.NotNil(withNonce)
	})

	suite.Run("ErrUnknownLocation", func() {
		encrypted := suite.crypter.SealBytes(suite.testBytes, suite.testNonce)

		withNonce, err := suite.crypter.AddNonceInBytes(encrypted, suite.testNonce, 3)
		assert.ErrorIs(err, ErrUnknownLocation)
		assert.Nil(withNonce)
	})
}

func (suite *CryptoTestSuite) TestGetNonceFromString() {
	assert := suite.Require()

	suite.Run("test at end", func() {
		encrypted, err := suite.crypter.SealString(suite.testMessage, suite.testNonce)
		assert.NoError(err)

		withNonce, err := suite.crypter.AddNonceInString(encrypted, suite.testNonce, AtEnd)
		assert.NoError(err)

		data, nonce, err := suite.crypter.GetNonceFromString(withNonce, AtEnd)
		assert.NoError(err)
		assert.Equal(encrypted, data)
		assert.Equal(suite.testNonce, nonce)
	})

	suite.Run("test at front", func() {
		encrypted, err := suite.crypter.SealString(suite.testMessage, suite.testNonce)
		assert.NoError(err)

		withNonce, err := suite.crypter.AddNonceInString(encrypted, suite.testNonce, AtFront)
		assert.NoError(err)

		data, nonce, err := suite.crypter.GetNonceFromString(withNonce, AtFront)
		assert.NoError(err)
		assert.Equal(encrypted, data)
		assert.Equal(suite.testNonce, nonce)
	})

	suite.Run("cannot decode string", func() {
		data, nonce, err := suite.crypter.GetNonceFromString("!#%", AtFront)
		assert.ErrorContains(err, "cannot decode string")
		assert.Empty(data)
		assert.Nil(nonce)
	})

	suite.Run("data too short to contain nonce", func() {
		data, nonce, err := suite.crypter.GetNonceFromString("687a", AtFront)
		assert.ErrorContains(err, "data too short to contain nonce")
		assert.Empty(data)
		assert.Nil(nonce)
	})

	suite.Run("ErrUnknownLocation", func() {
		encrypted, err := suite.crypter.SealString(suite.testMessage, suite.testNonce)
		assert.NoError(err)

		withNonce, err := suite.crypter.AddNonceInString(encrypted, suite.testNonce, AtFront)
		assert.NoError(err)

		data, nonce, err := suite.crypter.GetNonceFromString(withNonce, 3)
		assert.ErrorIs(err, ErrUnknownLocation)
		assert.Empty(data)
		assert.Nil(nonce)
	})
}

func (suite *CryptoTestSuite) TestGetNonceFromBytes() {
	assert := suite.Require()

	suite.Run("test at end", func() {
		encrypted := suite.crypter.SealBytes(suite.testBytes, suite.testNonce)

		withNonce, err := suite.crypter.AddNonceInBytes(encrypted, suite.testNonce, AtEnd)
		assert.NoError(err)

		nonce, data, length, err := suite.crypter.GetNonceFromBytes(withNonce, suite.crypter.NonceSize(), AtEnd)
		assert.NoError(err)
		assert.Equal(suite.testNonce, nonce)
		assert.Equal(encrypted, data)
		assert.Equal(0, length)
	})

	suite.Run("test at front", func() {
		encrypted := suite.crypter.SealBytes(suite.testBytes, suite.testNonce)

		withNonce, err := suite.crypter.AddNonceInBytes(encrypted, suite.testNonce, AtFront)
		assert.NoError(err)

		nonce, data, length, err := suite.crypter.GetNonceFromBytes(withNonce, suite.crypter.NonceSize(), AtFront)
		assert.NoError(err)
		assert.Equal(suite.testNonce, nonce)
		assert.Equal(encrypted, data)
		assert.Equal(0, length)
	})

	suite.Run("test len b equal nonceSize", func() {
		nonce, data, length, err := suite.crypter.GetNonceFromBytes(suite.testNonce, suite.crypter.NonceSize(), AtFront)
		assert.NoError(err)
		assert.Equal(suite.testNonce, nonce)
		assert.Nil(data)
		assert.Equal(0, length)
	})

	suite.Run("test len b less nonceSize", func() {
		testB := suite.testNonce[:len(suite.testNonce)-1]
		expLen := 1
		nonce, data, length, err := suite.crypter.GetNonceFromBytes(testB, suite.crypter.NonceSize(), AtFront)
		assert.NoError(err)
		assert.Equal(testB, nonce)
		assert.Nil(data)
		assert.Equal(expLen, length)
	})

	suite.Run("ErrUnknownLocation", func() {
		encrypted := suite.crypter.SealBytes(suite.testBytes, suite.testNonce)

		withNonce, err := suite.crypter.AddNonceInBytes(encrypted, suite.testNonce, AtEnd)
		assert.NoError(err)

		nonce, data, length, err := suite.crypter.GetNonceFromBytes(withNonce, suite.crypter.NonceSize(), 3)
		assert.ErrorIs(err, ErrUnknownLocation)
		assert.Nil(nonce)
		assert.Nil(data)
		assert.Equal(0, length)
	})
}

func (suite *CryptoTestSuite) TestOpenBytes() {
	assert := suite.Require()

	suite.Run("positive test", func() {
		encrypted := suite.crypter.SealBytes([]byte(suite.testMessage), suite.testNonce)
		decrypted, err := suite.crypter.OpenBytes(encrypted, suite.testNonce)

		assert.NoError(err)
		assert.Equal([]byte(suite.testMessage), decrypted)
	})

	suite.Run("cannot open bytes", func() {
		encrypted := suite.crypter.SealBytes([]byte(suite.testMessage), suite.testNonce)
		decrypted, err := suite.crypter.OpenBytes(encrypted, suite.testNonceForErr)

		assert.ErrorContains(err, "cannot open bytes")
		assert.Nil(decrypted)
	})

}

func TestCryptoTestSuite(t *testing.T) {
	suite.Run(t, new(CryptoTestSuite))
}
