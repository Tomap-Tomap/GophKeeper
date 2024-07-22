//go:build integration

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
		require.False(t, gotPassword.UpdateAt.IsZero())
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

func TestStorage_CreateFile(t *testing.T) {
	s, err := NewStorage(context.Background(), testDSN)
	require.NoError(t, err)
	defer s.Close()

	t.Run("positive test", func(t *testing.T) {
		defer truncateTable(t, s)

		u, err := s.CreateUser(context.Background(), "testUser", "testUser", "testSalt", "testPWD")
		require.NoError(t, err)

		wantFile := File{
			UserID:     u.ID,
			Name:       "FileName",
			PathToFile: "FilePath",
			Meta:       "FileMeta",
		}

		gotFile, err := s.CreateFile(context.Background(), u.ID, wantFile.Name, wantFile.PathToFile, wantFile.Meta)
		require.NoError(t, err)
		require.Equal(t, wantFile.UserID, gotFile.UserID)
		require.Equal(t, wantFile.Name, gotFile.Name)
		require.Equal(t, wantFile.PathToFile, gotFile.PathToFile)
		require.Equal(t, wantFile.Meta, gotFile.Meta)
		require.NotEmpty(t, gotFile.ID)
		require.False(t, gotFile.UpdateAt.IsZero())
	})

	t.Run("unknown user", func(t *testing.T) {
		defer truncateTable(t, s)

		gotPassword, err := s.CreateFile(
			context.Background(),
			"00000000-0000-0000-0000-000000000000",
			"FileName",
			"FilePath",
			"FileMeta",
		)
		require.Error(t, err)
		require.ErrorContains(t, err, "insert into files table name FileName")
		require.Nil(t, gotPassword)
	})
}

func TestStorage_GetFile(t *testing.T) {
	s, err := NewStorage(context.Background(), testDSN)
	require.NoError(t, err)
	defer s.Close()

	t.Run("positive test", func(t *testing.T) {
		defer truncateTable(t, s)

		u, err := s.CreateUser(context.Background(), "testUser", "testUser", "testSalt", "testPWD")
		require.NoError(t, err)

		wantFile, err := s.CreateFile(context.Background(), u.ID, "FileName", "FilePath", "PWDMeta")
		require.NoError(t, err)

		gotFile, err := s.GetFile(context.Background(), wantFile.ID)
		require.NoError(t, err)
		require.Equal(t, wantFile, gotFile)
	})

	t.Run("unknown id", func(t *testing.T) {
		defer truncateTable(t, s)

		gotPassword, err := s.GetFile(context.Background(), "00000000-0000-0000-0000-000000000000")
		require.Error(t, err)
		require.ErrorContains(t, err, "get file id 00000000-0000-0000-0000-000000000000")
		require.Nil(t, gotPassword)
	})
}

func TestStorage_GetAllFiles(t *testing.T) {
	s, err := NewStorage(context.Background(), testDSN)
	require.NoError(t, err)
	defer s.Close()

	t.Run("positive test", func(t *testing.T) {
		defer truncateTable(t, s)

		u, err := s.CreateUser(context.Background(), "testUser", "testUser", "testSalt", "testPWD")
		require.NoError(t, err)

		wantFile1, err := s.CreateFile(context.Background(), u.ID, "FileName1", "FilePath1", "FileMeta1")
		require.NoError(t, err)

		wantFile2, err := s.CreateFile(context.Background(), u.ID, "FileName2", "FilePath2", "FileMeta2")
		require.NoError(t, err)

		wantFiles := []File{*wantFile1, *wantFile2}

		gotFiles, err := s.GetAllFiles(context.Background(), u.ID)
		require.NoError(t, err)
		require.Equal(t, wantFiles, gotFiles)
	})

	t.Run("unknown user_id", func(t *testing.T) {
		defer truncateTable(t, s)

		gotFiles, err := s.GetAllFiles(context.Background(), "00000000-0000-0000-0000-000000000000")
		require.Error(t, err)
		require.Nil(t, gotFiles)
		require.ErrorContains(t, err, "user to user_id 00000000-0000-0000-0000-000000000000 don't have files")
	})
}

func TestStorage_CreateBank(t *testing.T) {
	s, err := NewStorage(context.Background(), testDSN)
	require.NoError(t, err)
	defer s.Close()

	t.Run("positive test", func(t *testing.T) {
		defer truncateTable(t, s)

		u, err := s.CreateUser(context.Background(), "testUser", "testUser", "testSalt", "testPWD")
		require.NoError(t, err)

		wantBank := Bank{
			UserID:    u.ID,
			Name:      "BankName",
			BanksData: "BankData",
			Meta:      "BankMeta",
		}

		gotBank, err := s.CreateBank(context.Background(), u.ID, wantBank.Name, wantBank.BanksData, wantBank.Meta)
		require.NoError(t, err)
		require.Equal(t, wantBank.UserID, gotBank.UserID)
		require.Equal(t, wantBank.Name, gotBank.Name)
		require.Equal(t, wantBank.BanksData, gotBank.BanksData)
		require.Equal(t, wantBank.Meta, gotBank.Meta)
		require.NotEmpty(t, gotBank.ID)
		require.False(t, gotBank.UpdateAt.IsZero())
	})

	t.Run("unknown user", func(t *testing.T) {
		defer truncateTable(t, s)

		gotPassword, err := s.CreateBank(
			context.Background(),
			"00000000-0000-0000-0000-000000000000",
			"BankName",
			"BankData",
			"BankMeta",
		)
		require.Error(t, err)
		require.ErrorContains(t, err, "insert into banks table name BankName")
		require.Nil(t, gotPassword)
	})
}

func TestStorage_GetBank(t *testing.T) {
	s, err := NewStorage(context.Background(), testDSN)
	require.NoError(t, err)
	defer s.Close()

	t.Run("positive test", func(t *testing.T) {
		defer truncateTable(t, s)

		u, err := s.CreateUser(context.Background(), "testUser", "testUser", "testSalt", "testPWD")
		require.NoError(t, err)

		wantBank, err := s.CreateBank(context.Background(), u.ID, "BankName", "BankData", "BankMeta")
		require.NoError(t, err)

		gotBank, err := s.GetBank(context.Background(), wantBank.ID)
		require.NoError(t, err)
		require.Equal(t, wantBank, gotBank)
	})

	t.Run("unknown id", func(t *testing.T) {
		defer truncateTable(t, s)

		gotBank, err := s.GetBank(context.Background(), "00000000-0000-0000-0000-000000000000")
		require.Error(t, err)
		require.ErrorContains(t, err, "get bank data id 00000000-0000-0000-0000-000000000000")
		require.Nil(t, gotBank)
	})
}

func TestStorage_GetAllBanks(t *testing.T) {
	s, err := NewStorage(context.Background(), testDSN)
	require.NoError(t, err)
	defer s.Close()

	t.Run("positive test", func(t *testing.T) {
		defer truncateTable(t, s)

		u, err := s.CreateUser(context.Background(), "testUser", "testUser", "testSalt", "testPWD")
		require.NoError(t, err)

		wantBank1, err := s.CreateBank(context.Background(), u.ID, "BankName1", "BankData1", "BankMeta1")
		require.NoError(t, err)

		wantBank2, err := s.CreateBank(context.Background(), u.ID, "BankName2", "BankData2", "BankMeta2")
		require.NoError(t, err)

		wantBanks := []Bank{*wantBank1, *wantBank2}

		gotBanks, err := s.GetAllBanks(context.Background(), u.ID)
		require.NoError(t, err)
		require.Equal(t, wantBanks, gotBanks)
	})

	t.Run("unknown user_id", func(t *testing.T) {
		defer truncateTable(t, s)

		gotBanks, err := s.GetAllBanks(context.Background(), "00000000-0000-0000-0000-000000000000")
		require.Error(t, err)
		require.Nil(t, gotBanks)
		require.ErrorContains(t, err, "user to user_id 00000000-0000-0000-0000-000000000000 don't have bank data")
	})
}

func TestStorage_CreateText(t *testing.T) {
	s, err := NewStorage(context.Background(), testDSN)
	require.NoError(t, err)
	defer s.Close()

	t.Run("positive test", func(t *testing.T) {
		defer truncateTable(t, s)

		u, err := s.CreateUser(context.Background(), "testUser", "testUser", "testSalt", "testPWD")
		require.NoError(t, err)

		wantText := Text{
			UserID: u.ID,
			Name:   "TextName",
			Text:   "TextData",
			Meta:   "TextMeta",
		}

		gotText, err := s.CreateText(context.Background(), u.ID, wantText.Name, wantText.Text, wantText.Meta)
		require.NoError(t, err)
		require.Equal(t, wantText.UserID, gotText.UserID)
		require.Equal(t, wantText.Name, gotText.Name)
		require.Equal(t, wantText.Text, gotText.Text)
		require.Equal(t, wantText.Meta, gotText.Meta)
		require.NotEmpty(t, gotText.ID)
		require.False(t, gotText.UpdateAt.IsZero())
	})

	t.Run("unknown user", func(t *testing.T) {
		defer truncateTable(t, s)

		gotPassword, err := s.CreateText(
			context.Background(),
			"00000000-0000-0000-0000-000000000000",
			"TextName",
			"TextData",
			"TextMeta",
		)
		require.Error(t, err)
		require.ErrorContains(t, err, "insert into texts table name TextName")
		require.Nil(t, gotPassword)
	})
}

func TestStorage_GetText(t *testing.T) {
	s, err := NewStorage(context.Background(), testDSN)
	require.NoError(t, err)
	defer s.Close()

	t.Run("positive test", func(t *testing.T) {
		defer truncateTable(t, s)

		u, err := s.CreateUser(context.Background(), "testUser", "testUser", "testSalt", "testPWD")
		require.NoError(t, err)

		wantText, err := s.CreateText(context.Background(), u.ID, "TextName", "TextData", "TextMeta")
		require.NoError(t, err)

		gotText, err := s.GetText(context.Background(), wantText.ID)
		require.NoError(t, err)
		require.Equal(t, wantText, gotText)
	})

	t.Run("unknown id", func(t *testing.T) {
		defer truncateTable(t, s)

		gotText, err := s.GetText(context.Background(), "00000000-0000-0000-0000-000000000000")
		require.Error(t, err)
		require.ErrorContains(t, err, "get text data id 00000000-0000-0000-0000-000000000000")
		require.Nil(t, gotText)
	})
}

func TestStorage_GetAllTexts(t *testing.T) {
	s, err := NewStorage(context.Background(), testDSN)
	require.NoError(t, err)
	defer s.Close()

	t.Run("positive test", func(t *testing.T) {
		defer truncateTable(t, s)

		u, err := s.CreateUser(context.Background(), "testUser", "testUser", "testSalt", "testPWD")
		require.NoError(t, err)

		wantText1, err := s.CreateText(context.Background(), u.ID, "TextName1", "TextData1", "TextMeta1")
		require.NoError(t, err)

		wantText2, err := s.CreateText(context.Background(), u.ID, "TextName2", "TextData2", "TextMeta2")
		require.NoError(t, err)

		wantTexts := []Text{*wantText1, *wantText2}

		gotTexts, err := s.GetAllTexts(context.Background(), u.ID)
		require.NoError(t, err)
		require.Equal(t, wantTexts, gotTexts)
	})

	t.Run("unknown user_id", func(t *testing.T) {
		defer truncateTable(t, s)

		gotTexts, err := s.GetAllTexts(context.Background(), "00000000-0000-0000-0000-000000000000")
		require.Error(t, err)
		require.Nil(t, gotTexts)
		require.ErrorContains(t, err, "user to user_id 00000000-0000-0000-0000-000000000000 don't have text data")
	})
}
