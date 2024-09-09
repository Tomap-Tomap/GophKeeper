//go:build unit

package columns

import (
	"context"
	"time"

	"github.com/Tomap-Tomap/GophKeeper/storage"
	"github.com/stretchr/testify/mock"
)

const (
	testID       = "TestID"
	testName     = "TestName"
	testMeta     = "TestMeta"
	testNumber   = "TestNumber"
	testCvc      = "TestCvc"
	testOwner    = "TestOwner"
	testExp      = "TestExp"
	testLogin    = "TestLogin"
	testPassword = "TestPassword"
	testText     = "TestText"
)

var testTime = time.Now()

type MockClient struct {
	mock.Mock
}

func (m *MockClient) GetAllPasswords(_ context.Context) ([]storage.Password, error) {
	args := m.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]storage.Password), args.Error(1)
}

func (m *MockClient) CreatePassword(_ context.Context, name, login, password, meta string) error {
	args := m.Called(name, login, password, meta)
	return args.Error(0)
}

func (m *MockClient) UpdatePassword(_ context.Context, id, name, login, password, meta string) error {
	args := m.Called(id, name, login, password, meta)
	return args.Error(0)
}

func (m *MockClient) DeletePassword(_ context.Context, id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockClient) GetAllBanks(_ context.Context) ([]storage.Bank, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]storage.Bank), args.Error(1)
}

func (m *MockClient) CreateBank(_ context.Context, name, number, cvc, owner, exp, meta string) error {
	args := m.Called(name, number, cvc, owner, exp, meta)
	return args.Error(0)
}

func (m *MockClient) UpdateBank(_ context.Context, id, name, number, cvc, owner, exp, meta string) error {
	args := m.Called(id, name, number, cvc, owner, exp, meta)
	return args.Error(0)
}

func (m *MockClient) DeleteBank(_ context.Context, id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockClient) GetAllTexts(_ context.Context) ([]storage.Text, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]storage.Text), args.Error(1)
}

func (m *MockClient) CreateText(_ context.Context, name, text, meta string) error {
	args := m.Called(name, text, meta)
	return args.Error(0)
}

func (m *MockClient) UpdateText(_ context.Context, id, name, text, meta string) error {
	args := m.Called(id, name, text, meta)
	return args.Error(0)
}

func (m *MockClient) DeleteText(_ context.Context, id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockClient) GetAllFiles(_ context.Context) ([]storage.File, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]storage.File), args.Error(1)
}

func (m *MockClient) CreateFile(_ context.Context, name, path, meta string) error {
	args := m.Called(name, path, meta)
	return args.Error(0)
}

func (m *MockClient) UpdateFile(_ context.Context, id, name, path, meta string) error {
	args := m.Called(id, name, path, meta)
	return args.Error(0)
}

func (m *MockClient) GetFile(_ context.Context, id, path string) error {
	args := m.Called(id, path)
	return args.Error(0)
}

func (m *MockClient) DeleteFile(_ context.Context, id string) error {
	args := m.Called(id)
	return args.Error(0)
}
