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

func TestStorage_GetUserData(t *testing.T) {
	s, err := NewStorage(context.Background(), testDSN)
	require.NoError(t, err)
	defer s.Close()
	defer truncateTable(t, s)

	wantUD := &UserData{
		Login:    "TestLogin",
		Password: "TestPassword",
		Salt:     "TestSalt",
	}

	_, err = s.conn.Exec(context.Background(), "INSERT INTO users (login, password) VALUES ($1, $2);", wantUD.Login, wantUD.Password)
	require.NoError(t, err)
	err = s.conn.QueryRow(context.Background(), "SELECT id FROM users WHERE login = $1;", wantUD.Login).Scan(&wantUD.ID)
	require.NoError(t, err)
	_, err = s.conn.Exec(context.Background(), "INSERT INTO salts (login, salt) VALUES ($1, $2);", wantUD.Login, wantUD.Salt)
	require.NoError(t, err)

	gotUD, err := s.GetUserData(context.Background(), wantUD.Login, wantUD.Login)
	require.NoError(t, err)
	require.Equal(t, wantUD, gotUD)
}

func truncateTable(t *testing.T, s *Storage) {
	_, err := s.conn.Exec(context.Background(), truncateQuery)
	require.NoError(t, err)
}
