// Package hasher defines methods and structures for generating hashes and salts.
package hasher

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Hasher provides methods to work with hashing functionalities.
type Hasher struct {
}

// NewHasher allocates and returns a new Hasher instance.
func NewHasher() *Hasher {
	return &Hasher{}
}

// GenerateSalt generates and returns a salt of the specified length.
func (h *Hasher) GenerateSalt(length int) (string, error) {
	salt := make([]byte, length)

	_, err := rand.Read(salt)

	if err != nil {
		return "", fmt.Errorf("connot read random data: %w", err)
	}

	return hex.EncodeToString(salt), nil
}

// GenerateHash generates a hash for the given string.
func (h *Hasher) GenerateHash(str string) string {
	hash := sha256.Sum256([]byte(str))

	return hex.EncodeToString(hash[:])
}

// GenerateHashWithSalt generates a hash for the given string combined with the provided salt.
func (h *Hasher) GenerateHashWithSalt(str, salt string) (string, error) {
	decodeSalt, err := hex.DecodeString(salt)

	if err != nil {
		return "", fmt.Errorf("cannot decode salt: %w", err)
	}

	hash := sha256.New()
	data := append(decodeSalt, str...)
	hash.Write(data)
	dst := hash.Sum(nil)

	return hex.EncodeToString(dst), nil
}
