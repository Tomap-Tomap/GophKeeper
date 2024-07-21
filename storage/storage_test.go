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
		require.ErrorContains(t, err, "create pgxpool")
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

	t.Run("create user test", func(t *testing.T) {
		defer truncateTable(t, s)

		var (
			login       = "TestUser"
			password    = "TestPassword"
			salt        = "TestSalt"
			gotLogin    string
			gotPassword string
			gotSalt     string
			gotID       string
		)

		gotUD, err := s.CreateUser(context.Background(), login, login, salt, password)
		require.NoError(t, err)
		require.Equal(t, login, gotUD.Login)
		require.Equal(t, password, gotUD.Password)
		require.Equal(t, salt, gotUD.Salt)
		require.NotEmpty(t, gotUD.ID)

		err = s.conn.QueryRow(context.Background(), "SELECT id, login, password FROM users WHERE login = $1;", login).Scan(&gotID, &gotLogin, &gotPassword)
		require.NoError(t, err)
		require.Equal(t, login, gotLogin)
		require.Equal(t, password, strings.TrimSpace(gotPassword))
		require.Equal(t, gotUD.ID, gotID)

		err = s.conn.QueryRow(context.Background(), "SELECT login, salt FROM salts WHERE login = $1;", login).Scan(&gotLogin, &gotSalt)
		require.NoError(t, err)
		require.Equal(t, login, strings.TrimSpace(gotLogin))
		require.Equal(t, salt, strings.TrimSpace(gotSalt))
	})

	t.Run("create with empty login", func(t *testing.T) {
		defer truncateTable(t, s)

		gotUD, err := s.CreateUser(context.Background(), "", "", "TestSalt", "TestPassword")
		require.Error(t, err)
		require.ErrorContains(t, err, "insert into users table login")
		require.Nil(t, gotUD)
	})

	t.Run("create with empty password", func(t *testing.T) {
		defer truncateTable(t, s)

		gotUD, err := s.CreateUser(context.Background(), "TestLogin", "TestLogin", "TestSalt", "")
		require.Error(t, err)
		require.ErrorContains(t, err, "insert into users table login")
		require.Nil(t, gotUD)
	})

	t.Run("create dublicate login", func(t *testing.T) {
		defer truncateTable(t, s)

		_, err := s.CreateUser(context.Background(), "TestLogin", "TestLogin", "TestSalt", "TestPassword")
		require.NoError(t, err)

		gotUD, err := s.CreateUser(context.Background(), "TestLogin", "TestLogin", "TestSalt", "TestPassword")
		require.Error(t, err)
		require.ErrorContains(t, err, "insert into users table login")
		require.Nil(t, gotUD)
	})

	t.Run("create with empty salt", func(t *testing.T) {
		defer truncateTable(t, s)

		gotUD, err := s.CreateUser(context.Background(), "TestLogin", "TestLogin", "", "TestPassword")
		require.Error(t, err)
		require.ErrorContains(t, err, "insert into salts table login")
		require.Nil(t, gotUD)
	})

	t.Run("create with empty login hashed", func(t *testing.T) {
		defer truncateTable(t, s)

		gotUD, err := s.CreateUser(context.Background(), "TestLogin", "", "TestSalt", "TestPassword")
		require.Error(t, err)
		require.ErrorContains(t, err, "insert into salts table login")
		require.Nil(t, gotUD)
	})
}

func TestStorage_GetUser(t *testing.T) {
	s, err := NewStorage(context.Background(), testDSN)
	require.NoError(t, err)
	defer s.Close()

	t.Run("get user test", func(t *testing.T) {
		defer truncateTable(t, s)

		wantUD := &User{
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

		gotUD, err := s.GetUser(context.Background(), wantUD.Login, wantUD.Login)
		require.NoError(t, err)
		require.Equal(t, wantUD, gotUD)
	})

	t.Run("unknown user test", func(t *testing.T) {
		defer truncateTable(t, s)
		_, err := s.GetUser(context.Background(), "testUser", "testLogin")
		require.Error(t, err)
		require.ErrorContains(t, err, "get user")
	})
}

func truncateTable(t *testing.T, s *Storage) {
	_, err := s.conn.Exec(context.Background(), truncateQuery)
	require.NoError(t, err)
}

func TestStorage_CreatePassword(t *testing.T) {
	s, err := NewStorage(context.Background(), testDSN)
	require.NoError(t, err)
	defer s.Close()

	t.Run("positive test", func(t *testing.T) {
		defer truncateTable(t, s)

		u, err := s.CreateUser(context.Background(), "testUser", "testUser", "testSalt", "testPWD")
		require.NoError(t, err)

		wantPassword := Password{
			UserID:   u.ID,
			Name:     "PWDName",
			Login:    "PWDLogin",
			Password: "PWDPassword",
			Meta:     "PWDMeta",
		}

		gotPassword, err := s.CreatePassword(context.Background(), u.ID, wantPassword.Name, wantPassword.Login, wantPassword.Password, wantPassword.Meta)
		require.NoError(t, err)
		require.Equal(t, wantPassword.UserID, gotPassword.UserID)
		require.Equal(t, wantPassword.Name, gotPassword.Name)
		require.Equal(t, wantPassword.Login, gotPassword.Login)
		require.Equal(t, wantPassword.Password, gotPassword.Password)
		require.Equal(t, wantPassword.Meta, gotPassword.Meta)
		require.NotEmpty(t, gotPassword.ID)
	})

	t.Run("unknown user", func(t *testing.T) {
		defer truncateTable(t, s)

		gotPassword, err := s.CreatePassword(
			context.Background(),
			"00000000-0000-0000-0000-000000000000",
			"PWDName",
			"PWDLogin",
			"PWDPassword",
			"PWDMeta",
		)
		require.Error(t, err)
		require.ErrorContains(t, err, "insert into passwords table name PWDName")
		require.Nil(t, gotPassword)
	})
}

func TestStorage_GetPassword(t *testing.T) {
	s, err := NewStorage(context.Background(), testDSN)
	require.NoError(t, err)
	defer s.Close()

	t.Run("positive test", func(t *testing.T) {
		defer truncateTable(t, s)

		u, err := s.CreateUser(context.Background(), "testUser", "testUser", "testSalt", "testPWD")
		require.NoError(t, err)

		wantPassword, err := s.CreatePassword(context.Background(), u.ID, "PWDName", "PWDLogin", "PWDPassword", "PWDMeta")
		require.NoError(t, err)

		gotPassword, err := s.GetPassword(context.Background(), wantPassword.ID)
		require.NoError(t, err)
		require.Equal(t, wantPassword, gotPassword)
	})

	t.Run("unknown id", func(t *testing.T) {
		defer truncateTable(t, s)

		gotPassword, err := s.GetPassword(context.Background(), "00000000-0000-0000-0000-000000000000")
		require.Error(t, err)
		require.ErrorContains(t, err, "get password id 00000000-0000-0000-0000-000000000000")
		require.Nil(t, gotPassword)
	})
}

func TestStorage_GetAllPassword(t *testing.T) {
	s, err := NewStorage(context.Background(), testDSN)
	require.NoError(t, err)
	defer s.Close()

	t.Run("positive test", func(t *testing.T) {
		defer truncateTable(t, s)

		u, err := s.CreateUser(context.Background(), "testUser", "testUser", "testSalt", "testPWD")
		require.NoError(t, err)

		wantPassword1, err := s.CreatePassword(context.Background(), u.ID, "PWDName1", "PWDLogin1", "PWDPassword1", "PWDMeta1")
		require.NoError(t, err)

		wantPassword2, err := s.CreatePassword(context.Background(), u.ID, "PWDName2", "PWDLogin2", "PWDPassword2", "PWDMeta2")
		require.NoError(t, err)

		wantPWDs := []Password{*wantPassword1, *wantPassword2}

		gotPWDs, err := s.GetAllPassword(context.Background(), u.ID)
		require.NoError(t, err)
		require.Equal(t, wantPWDs, gotPWDs)
	})

	t.Run("unknown user_id", func(t *testing.T) {
		defer truncateTable(t, s)

		gotPWDs, err := s.GetAllPassword(context.Background(), "00000000-0000-0000-0000-000000000000")
		require.Error(t, err)
		require.Nil(t, gotPWDs)
		require.ErrorContains(t, err, "user to user_id 00000000-0000-0000-0000-000000000000 don't have passwords")
	})
}
