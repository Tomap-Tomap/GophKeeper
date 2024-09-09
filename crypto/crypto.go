// Package crypto implements AES encryption and decryption for strings and bytes.
// It leverages Go's crypto/aes and crypto/cipher packages to provide secure encoding and decoding.
// The package supports adding the nonce at the beginning or end of the data for flexibility.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// NonceLocation defines a type for enumerating Nonce locations.
type NonceLocation int

// Enumeration of Nonce locations.
const (
	AtEnd   NonceLocation = iota // Nonce is located at the end of the data.
	AtFront                      // Nonce is located at the beginning of the data.
)

// ErrUnknownLocation is returned when an unknown NonceLocation is selected.
var ErrUnknownLocation = errors.New("unknown nonce location selected")

// Crypter provides AES encryption and decryption for strings and bytes.
// It wraps the cipher.AEAD interface.
type Crypter struct {
	aesgcm cipher.AEAD
}

// NewCrypter creates a new AES key, saves it in the specified folder,
// and returns a new Crypter instance along with the path to the AES key.
func NewCrypter(keySize int, folderToSave string) (crypter *Crypter, pathToKey string, err error) {
	b := make([]byte, keySize)

	_, err = rand.Read(b)

	if err != nil {
		return nil, "", fmt.Errorf("cannot generate secret key: %w", err)
	}

	pathToKey = filepath.Join(folderToSave, "key.aes")
	file, err := os.Create(pathToKey)

	if err != nil {
		return nil, "", fmt.Errorf("cannot create key file: %w", err)
	}

	defer func() {
		err = errors.Join(err, file.Close())
	}()

	_, err = file.Write(b)

	if err != nil {
		return nil, "", fmt.Errorf("cannot write secret key: %w", err)
	}

	aesgcm, err := getAEAD(b)

	if err != nil {
		return nil, "", fmt.Errorf("cannot create AEAD: %w", err)
	}

	return &Crypter{
		aesgcm: aesgcm,
	}, pathToKey, nil
}

// NewCrypterByFile creates a new Crypter instance using an AES key stored in the provided file path.
func NewCrypterByFile(pathToKey string) (crypter *Crypter, err error) {
	file, err := os.Open(pathToKey)

	if err != nil {
		return nil, fmt.Errorf("cannot open secret key: %w", err)
	}

	defer func() {
		err = errors.Join(err, file.Close())
	}()

	key, err := io.ReadAll(file)

	if err != nil {
		return nil, fmt.Errorf("cannot read secret key: %w", err)
	}

	aesgcm, err := getAEAD(key)

	if err != nil {
		return nil, fmt.Errorf("cannot create AEAD: %w", err)
	}

	return &Crypter{
		aesgcm: aesgcm,
	}, nil
}

// GenerateNonce generates a new nonce for encryption and decryption.
func (c *Crypter) GenerateNonce() ([]byte, error) {
	nonce := make([]byte, c.aesgcm.NonceSize())
	_, err := rand.Read(nonce)

	if err != nil {
		return nil, fmt.Errorf("cannot generate nonce: %w", err)
	}

	return nonce, nil
}

// NonceSize returns the size of the nonce required for Seal and Open operations.
func (c *Crypter) NonceSize() int {
	return c.aesgcm.NonceSize()
}

// SealString encrypts and authenticates the given string and returns the encrypted string.
func (c *Crypter) SealString(str string, nonce []byte) (string, error) {
	return hex.EncodeToString(c.aesgcm.Seal(nil, nonce, []byte(str), nil)), nil
}

// SealStringWithoutNonce encrypts the string and appends the nonce at the end, returning the result.
func (c *Crypter) SealStringWithoutNonce(str string) (string, error) {
	nonce, err := c.GenerateNonce()

	if err != nil {
		return "", fmt.Errorf("cannot generate nonce: %w", err)
	}

	enc, err := c.SealString(str, nonce)

	if err != nil {
		return "", fmt.Errorf("cannot seal data: %w", err)
	}

	res, err := c.AddNonceInString(enc, nonce, AtEnd)

	if err != nil {
		return "", fmt.Errorf("cannot add nonce to end result: %w", err)
	}

	return res, nil
}

// SealBytes encrypts and authenticates the given byte slice, returning the encrypted bytes.
func (c *Crypter) SealBytes(b, nonce []byte) []byte {
	return c.aesgcm.Seal(nil, nonce, b, nil)
}

// OpenString decrypts and authenticates the given encrypted string, returning the original string.
func (c *Crypter) OpenString(encryptStr string, nonce []byte) (string, error) {
	d, err := hex.DecodeString(encryptStr)

	if err != nil {
		return "", fmt.Errorf("cannot decode string: %w", err)
	}

	src, err := c.aesgcm.Open(nil, nonce, d, nil)

	if err != nil {
		return "", fmt.Errorf("cannot open string: %w", err)
	}

	return string(src), nil
}

// OpenStringWithoutNonce decrypts and authenticates data with the nonce at the end, returning the original string.
func (c *Crypter) OpenStringWithoutNonce(encryptStr string) (string, error) {
	data, nonce, err := c.GetNonceFromString(encryptStr, AtEnd)

	if err != nil {
		return "", fmt.Errorf("cannot get nonce at the end of data: %w", err)
	}

	dec, err := c.OpenString(data, nonce)

	if err != nil {
		return "", fmt.Errorf("cannot open data: %w", err)
	}

	return dec, nil
}

// OpenBytes decrypts and authenticates the given encrypted bytes, returning the original bytes.
func (c *Crypter) OpenBytes(enctyptB []byte, nonce []byte) ([]byte, error) {
	src, err := c.aesgcm.Open(nil, nonce, enctyptB, nil)

	if err != nil {
		return nil, fmt.Errorf("cannot open bytes: %w", err)
	}

	return src, nil
}

// AddNonceInString appends or prepends the nonce to the string depending on the specified location.
func (c *Crypter) AddNonceInString(str string, nonce []byte, location NonceLocation) (string, error) {
	d, err := hex.DecodeString(str)

	if err != nil {
		return "", fmt.Errorf("cannot decode string: %w", err)
	}

	switch location {
	case AtFront:
		res := append(nonce, d...)
		return hex.EncodeToString(res), nil
	case AtEnd:
		res := append(d, nonce...)
		return hex.EncodeToString(res), nil
	}

	return "", ErrUnknownLocation
}

// AddNonceInBytes appends or prepends the nonce to the byte slice depending on the specified location.
func (c *Crypter) AddNonceInBytes(b []byte, nonce []byte, location NonceLocation) ([]byte, error) {
	switch location {
	case AtFront:
		res := append(nonce, b...)
		return res, nil
	case AtEnd:
		res := append(b, nonce...)
		return res, nil
	}

	return nil, ErrUnknownLocation
}

// GetNonceFromString extracts the nonce from the string and returns the string without the nonce along with the nonce itself.
func (c *Crypter) GetNonceFromString(str string, location NonceLocation) (string, []byte, error) {
	d, err := hex.DecodeString(str)

	if err != nil {
		return "", nil, fmt.Errorf("cannot decode string: %w", err)
	}

	if len(d) < c.aesgcm.NonceSize() {
		return "", nil, errors.New("data too short to contain nonce")
	}

	switch location {
	case AtFront:
		nonce := d[:c.aesgcm.NonceSize()]
		d = d[c.aesgcm.NonceSize():]
		return hex.EncodeToString(d), nonce, nil
	case AtEnd:
		nonce := d[len(d)-c.aesgcm.NonceSize():]
		d = d[:len(d)-c.aesgcm.NonceSize()]

		return hex.EncodeToString(d), nonce, nil
	}

	return "", nil, ErrUnknownLocation
}

// GetNonceFromBytes extracts the nonce from the byte slice and returns the nonce, the bytes without the nonce,
// and the number of remaining bytes required to complete the nonce if the original slice was too short.
func (c *Crypter) GetNonceFromBytes(b []byte, nonceSize int, location NonceLocation) ([]byte, []byte, int, error) {
	if len(b) == nonceSize {
		return b, nil, 0, nil
	}

	if len(b) < nonceSize {
		return b, nil, nonceSize - len(b), nil
	}

	switch location {
	case AtFront:
		nonce := b[:nonceSize]
		b = b[nonceSize:]

		return nonce, b, 0, nil
	case AtEnd:
		nonce := b[len(b)-nonceSize:]
		b = b[:len(b)-nonceSize]

		return nonce, b, 0, nil
	}

	return nil, nil, 0, ErrUnknownLocation
}

func getAEAD(key []byte) (cipher.AEAD, error) {
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	return aesgcm, nil
}
