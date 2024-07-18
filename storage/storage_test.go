// Package storage определяет структуры и методы для работы с базой данных postgres
package storage

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testDSN       = "host=localhost port=5433 user=gophkeeper password=gophkeeper dbname=gophkeeper sslmode=disable"
	truncateQuery = "TRUNCATE TABLE users, banks, files, passwords, salts, texts;"
)

func TestNewStorage(t *testing.T) {
	t.Run("error test", func(t *testing.T) {
		_, err := NewStorage(context.Background(), "errorDSN")

		require.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		s, err := NewStorage(context.Background(), testDSN)

		require.NoError(t, err)
		defer s.Close()
		require.NotEmpty(t, s)
	})
}

func TestStorage_CreateUser(t *testing.T) {
	s, err := NewStorage(context.Background(), testDSN)
	require.NoError(t, err)
	defer s.Close()
	defer truncateTable(t, s)

	var (
		login       = "TestUser"
		password    = "TestPassword"
		gotLogin    string
		gotPassword string
	)

	err = s.CreateUser(context.Background(), login, password)
	require.NoError(t, err)

	err = s.conn.QueryRow(context.Background(), "SELECT login, password FROM users WHERE login = $1;", login).Scan(&gotLogin, &gotPassword)
	require.NoError(t, err)
	require.Equal(t, login, gotLogin)
	require.Equal(t, password, strings.TrimSpace(gotPassword))
}

func truncateTable(t *testing.T, s *Storage) {
	_, err := s.conn.Exec(context.Background(), truncateQuery)
	require.NoError(t, err)
}
