// Package hasher определяет методы и структуры для генерации хэшей и солей
package hasher

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Hasher структура для работы с хэшами
type Hasher struct {
}

// NewHasher аллоцирует новый Hasher
func NewHasher() *Hasher {
	return &Hasher{}
}

// GenerateSalt возвращает соль
func (h *Hasher) GenerateSalt(len int) (string, error) {
	salt := make([]byte, len)

	_, err := rand.Read(salt)

	if err != nil {
		return "", fmt.Errorf("connot read random data: %w", err)
	}

	return hex.EncodeToString(salt), nil
}

// GetHash генерирует хэш для строки
func (h *Hasher) GetHash(str string) string {
	hash := sha256.New()

	hash.Write([]byte(str))
	dst := hash.Sum(nil)

	return hex.EncodeToString(dst)
}

// GetHashWithSalt генерирует хэш для строки с солью
func (h *Hasher) GetHashWithSalt(str, salt string) (string, error) {
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
