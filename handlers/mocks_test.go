//go:build unit

package handlers

import (
	"bytes"
	"context"

	"github.com/Tomap-Tomap/GophKeeper/storage"
	"github.com/stretchr/testify/mock"
)

type HasherMockedObject struct {
	mock.Mock
}

func (h *HasherMockedObject) GenerateSalt() (string, error) {
	args := h.Called()

	return args.String(0), args.Error(1)
}

func (h *HasherMockedObject) GetHash(str string) (string, error) {
	args := h.Called(str)

	return args.String(0), args.Error(1)
}

func (h *HasherMockedObject) GetHashWithSalt(str, salt string) (string, error) {
	args := h.Called(str, salt)

	return args.String(0), args.Error(1)
}

type StorageMockedObject struct {
	mock.Mock
}

func (sm *StorageMockedObject) CreateUser(_ context.Context, login, loginHashed, salt, password string) (*storage.User, error) {
	args := sm.Called(login, loginHashed, salt, password)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.User), args.Error(1)
}

func (sm *StorageMockedObject) GetUser(_ context.Context, login, loginHashed string) (*storage.User, error) {
	args := sm.Called(login, loginHashed)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.User), args.Error(1)
}

func (sm *StorageMockedObject) CreatePassword(_ context.Context, userID, name, login, password, meta string) (*storage.Password, error) {
	args := sm.Called(userID, name, login, password, meta)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.Password), args.Error(1)
}

func (sm *StorageMockedObject) GetPassword(_ context.Context, passwordID string) (*storage.Password, error) {
	args := sm.Called(passwordID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.Password), args.Error(1)
}

func (sm *StorageMockedObject) GetAllPassword(_ context.Context, userID string) ([]storage.Password, error) {
	args := sm.Called(userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]storage.Password), args.Error(1)
}

func (sm *StorageMockedObject) CreateFile(_ context.Context, userID, name, pathToFile, meta string) (*storage.File, error) {
	args := sm.Called(userID, name, pathToFile, meta)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.File), args.Error(1)
}

func (sm *StorageMockedObject) GetFile(_ context.Context, fileID string) (*storage.File, error) {
	args := sm.Called(fileID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.File), args.Error(1)
}

func (sm *StorageMockedObject) GetAllFiles(_ context.Context, userID string) ([]storage.File, error) {
	args := sm.Called(userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]storage.File), args.Error(1)
}

func (sm *StorageMockedObject) CreateBank(_ context.Context, userID, name, banksData, meta string) (*storage.Bank, error) {
	args := sm.Called(userID, name, banksData, meta)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.Bank), args.Error(1)
}

func (sm *StorageMockedObject) GetBank(_ context.Context, bankID string) (*storage.Bank, error) {
	args := sm.Called(bankID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.Bank), args.Error(1)
}

func (sm *StorageMockedObject) GetAllBanks(_ context.Context, userID string) ([]storage.Bank, error) {
	args := sm.Called(userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]storage.Bank), args.Error(1)
}

func (sm *StorageMockedObject) CreateText(_ context.Context, userID, name, text, meta string) (*storage.Text, error) {
	args := sm.Called(userID, name, text, meta)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.Text), args.Error(1)
}

func (sm *StorageMockedObject) GetText(_ context.Context, textID string) (*storage.Text, error) {
	args := sm.Called(textID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.Text), args.Error(1)
}

func (sm *StorageMockedObject) GetAllTexts(_ context.Context, userID string) ([]storage.Text, error) {
	args := sm.Called(userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]storage.Text), args.Error(1)
}

type TokenerMockerdObject struct {
	mock.Mock
}

func (t *TokenerMockerdObject) GetToken(sub string) (string, error) {
	args := t.Called(sub)

	return args.String(0), args.Error(1)
}

type ImageStoreMockerObject struct {
	mock.Mock
}

func (is *ImageStoreMockerObject) Save(content bytes.Buffer) (string, error) {
	args := is.Called(content)

	return args.String(0), args.Error(1)
}

func (is *ImageStoreMockerObject) GetDBFiler(pathToFile string) (DBFiler, error) {
	args := is.Called(pathToFile)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(DBFiler), args.Error(1)
}

type DBFilerMocketObject struct {
	mock.Mock
}

func (dbf *DBFilerMocketObject) GetChunck() ([]byte, error) {
	args := dbf.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]byte), args.Error(1)
}

func (dbf *DBFilerMocketObject) Close() {
	dbf.Called()
	return
}
