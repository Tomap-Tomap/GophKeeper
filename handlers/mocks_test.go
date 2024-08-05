//go:build unit

package handlers

import (
	"context"

	"github.com/Tomap-Tomap/GophKeeper/proto"
	"github.com/Tomap-Tomap/GophKeeper/storage"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type HasherMockedObject struct {
	mock.Mock
}

func (h *HasherMockedObject) GenerateSalt(length int) (string, error) {
	args := h.Called(length)

	return args.String(0), args.Error(1)
}

func (h *HasherMockedObject) onGenerateSalt(length int, retSalt string, retErr error) {
	h.On("GenerateSalt", length).Return(retSalt, retErr)
}

func (h *HasherMockedObject) GenerateHash(str string) string {
	args := h.Called(str)

	return args.String(0)
}

func (h *HasherMockedObject) onGenerateHash(str, retHash string) {
	h.On("GenerateHash", str).Return(retHash)
}

func (h *HasherMockedObject) GenerateHashWithSalt(str, salt string) (string, error) {
	args := h.Called(str, salt)

	return args.String(0), args.Error(1)
}

func (h *HasherMockedObject) onGenerateHashWithSalt(str, salt, retHash string, retErr error) {
	h.On("GenerateHashWithSalt", str, salt).Return(retHash, retErr)
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

func (sm *StorageMockedObject) onCreateUser(login, loginHashed, salt, password string, retUser *storage.User, retErr error) {
	sm.On("CreateUser", login, loginHashed, salt, password).Return(retUser, retErr)
}

func (sm *StorageMockedObject) GetUser(_ context.Context, login, loginHashed string) (*storage.User, error) {
	args := sm.Called(login, loginHashed)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.User), args.Error(1)
}

func (sm *StorageMockedObject) onGetUser(login, loginHashed string, retUser *storage.User, retErr error) {
	sm.On("GetUser", login, loginHashed).Return(retUser, retErr)
}

func (sm *StorageMockedObject) CreatePassword(_ context.Context, userID, name, login, password, meta string) (*storage.Password, error) {
	args := sm.Called(userID, name, login, password, meta)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.Password), args.Error(1)
}

func (sm *StorageMockedObject) onCreatePassword(userID, name, login, password, meta string, retPassword *storage.Password, retErr error) {
	sm.On("CreatePassword", userID, name, login, password, meta).Return(retPassword, retErr)
}

func (sm *StorageMockedObject) UpdatePassword(_ context.Context, id, userID, name, login, password, meta string) (*storage.Password, error) {
	args := sm.Called(id, userID, name, login, password, meta)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.Password), args.Error(1)
}

func (sm *StorageMockedObject) onUpdatePassword(id, userID, name, login, password, meta string, retPassword *storage.Password, retErr error) {
	sm.On("UpdatePassword", id, userID, name, login, password, meta).Return(retPassword, retErr)
}

func (sm *StorageMockedObject) GetPassword(_ context.Context, passwordID, userID string) (*storage.Password, error) {
	args := sm.Called(passwordID, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.Password), args.Error(1)
}

func (sm *StorageMockedObject) onGetPassword(passwordID, userID string, retPassword *storage.Password, retErr error) {
	sm.On("GetPassword", passwordID, userID).Return(retPassword, retErr)
}

func (sm *StorageMockedObject) GetAllPassword(_ context.Context, userID string) ([]storage.Password, error) {
	args := sm.Called(userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]storage.Password), args.Error(1)
}

func (sm *StorageMockedObject) onGetAllPassword(userID string, retPasswords []storage.Password, retErr error) {
	sm.On("GetAllPassword", userID).Return(retPasswords, retErr)
}

func (sm *StorageMockedObject) DeletePassword(_ context.Context, id, userID string) error {
	args := sm.Called(id, userID)

	return args.Error(0)
}

func (sm *StorageMockedObject) onDeletePassword(id, userID string, retErr error) {
	sm.On("DeletePassword", id, userID).Return(retErr)
}

func (sm *StorageMockedObject) CreateFile(_ context.Context, userID, name, pathToFile, meta string) (*storage.File, error) {
	args := sm.Called(userID, name, pathToFile, meta)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.File), args.Error(1)
}

func (sm *StorageMockedObject) UpdateFile(_ context.Context, id, userID, name, pathToFile, meta string) (*storage.File, error) {
	args := sm.Called(id, userID, name, pathToFile, meta)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.File), args.Error(1)
}

func (sm *StorageMockedObject) GetFile(_ context.Context, fileID, userID string) (*storage.File, error) {
	args := sm.Called(fileID, userID)

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

func (sm *StorageMockedObject) DeleteFile(_ context.Context, id, userID string) (*storage.File, error) {
	args := sm.Called(id, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.File), args.Error(1)
}

func (sm *StorageMockedObject) onCreateFile(userID, name, pathToFile, meta string, retFile *storage.File, retErr error) {
	sm.On("CreateFile", userID, name, pathToFile, meta).Return(retFile, retErr)
}

func (sm *StorageMockedObject) onUpdateFile(fileID, userID, name, pathToFile, meta string, retFile *storage.File, retErr error) {
	sm.On("UpdateFile", fileID, userID, name, pathToFile, meta).Return(retFile, retErr)
}

func (sm *StorageMockedObject) onGetFile(fileID, userID string, retFile *storage.File, retErr error) {
	sm.On("GetFile", fileID, userID).Return(retFile, retErr)
}

func (sm *StorageMockedObject) onGetAllFiles(userID string, retFiles []storage.File, retErr error) {
	sm.On("GetAllFiles", userID).Return(retFiles, retErr)
}

func (sm *StorageMockedObject) onDeleteFile(fileID, userID string, retFile *storage.File, retErr error) {
	sm.On("DeleteFile", fileID, userID).Return(retFile, retErr)
}

func (sm *StorageMockedObject) CreateBank(_ context.Context, userID, name, number, cvc, owner, exp, meta string) (*storage.Bank, error) {
	args := sm.Called(userID, name, number, cvc, owner, exp, meta)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.Bank), args.Error(1)
}

func (sm *StorageMockedObject) UpdateBank(_ context.Context, id, userID, name, number, cvc, owner, exp, meta string) (*storage.Bank, error) {
	args := sm.Called(id, userID, name, number, cvc, owner, exp, meta)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.Bank), args.Error(1)
}

func (sm *StorageMockedObject) GetBank(_ context.Context, bankID, userID string) (*storage.Bank, error) {
	args := sm.Called(bankID, userID)

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

func (sm *StorageMockedObject) DeleteBank(_ context.Context, id, userID string) error {
	args := sm.Called(id, userID)

	return args.Error(0)
}

func (sm *StorageMockedObject) onCreateBank(userID, name, number, cvc, owner, exp, meta string, retBank *storage.Bank, retErr error) {
	sm.On("CreateBank", userID, name, number, cvc, owner, exp, meta).Return(retBank, retErr)
}

func (sm *StorageMockedObject) onUpdateBank(bankID, userID, name, number, cvc, owner, exp, meta string, retBank *storage.Bank, retErr error) {
	sm.On("UpdateBank", bankID, userID, name, number, cvc, owner, exp, meta).Return(retBank, retErr)
}

func (sm *StorageMockedObject) onGetBank(bankID, userID string, retBank *storage.Bank, retErr error) {
	sm.On("GetBank", bankID, userID).Return(retBank, retErr)
}

func (sm *StorageMockedObject) onGetAllBanks(userID string, retBanks []storage.Bank, retErr error) {
	sm.On("GetAllBanks", userID).Return(retBanks, retErr)
}

func (sm *StorageMockedObject) onDeleteBank(bankID, userID string, retErr error) {
	sm.On("DeleteBank", bankID, userID).Return(retErr)
}

func (sm *StorageMockedObject) CreateText(_ context.Context, userID, name, text, meta string) (*storage.Text, error) {
	args := sm.Called(userID, name, text, meta)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.Text), args.Error(1)
}

func (sm *StorageMockedObject) UpdateText(_ context.Context, id, userID, name, text, meta string) (*storage.Text, error) {
	args := sm.Called(id, userID, name, text, meta)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.Text), args.Error(1)
}

func (sm *StorageMockedObject) GetText(_ context.Context, textID, userID string) (*storage.Text, error) {
	args := sm.Called(textID, userID)

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

func (sm *StorageMockedObject) DeleteText(_ context.Context, id, userID string) error {
	args := sm.Called(id, userID)

	return args.Error(0)
}

func (sm *StorageMockedObject) onCreateText(userID, name, text, meta string, retText *storage.Text, retErr error) {
	sm.On("CreateText", userID, name, text, meta).Return(retText, retErr)
}

func (sm *StorageMockedObject) onUpdateText(textID, userID, name, text, meta string, retText *storage.Text, retErr error) {
	sm.On("UpdateText", textID, userID, name, text, meta).Return(retText, retErr)
}

func (sm *StorageMockedObject) onGetText(textID, userID string, retText *storage.Text, retErr error) {
	sm.On("GetText", textID, userID).Return(retText, retErr)
}

func (sm *StorageMockedObject) onGetAllTexts(userID string, retTexts []storage.Text, retErr error) {
	sm.On("GetAllTexts", userID).Return(retTexts, retErr)
}

func (sm *StorageMockedObject) onDeleteText(textID, userID string, retErr error) {
	sm.On("DeleteText", textID, userID).Return(retErr)
}

type TokenerMockedObject struct {
	mock.Mock
}

func (t *TokenerMockedObject) GetToken(sub string) (string, error) {
	args := t.Called(sub)

	return args.String(0), args.Error(1)
}

func (t *TokenerMockedObject) onGetToken(sub, retToken string, retErr error) {
	t.On("GetToken", sub).Return(retToken, retErr)
}

type FileStoreMockedObject struct {
	mock.Mock
}

func (fs *FileStoreMockedObject) CreateDBFile(fileName string) (storage.DBFiler, error) {
	args := fs.Called(fileName)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(storage.DBFiler), args.Error(1)
}

func (fs *FileStoreMockedObject) GetDBFile(fileName string) (storage.DBFiler, error) {
	args := fs.Called(fileName)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(storage.DBFiler), args.Error(1)
}

func (fs *FileStoreMockedObject) DeleteDBFile(fileName string) error {
	args := fs.Called(fileName)

	return args.Error(0)
}

func (fs *FileStoreMockedObject) onCreateDBFile(fileName string, retDBFiler storage.DBFiler, retErr error) {
	fs.On("CreateDBFile", fileName).Return(retDBFiler, retErr)
}

func (fs *FileStoreMockedObject) onGetDBFile(fileName string, retDBFiler storage.DBFiler, retErr error) {
	fs.On("GetDBFile", fileName).Return(retDBFiler, retErr)
}

func (fs *FileStoreMockedObject) onDeleteDBFile(fileName string, retErr error) {
	fs.On("DeleteDBFile", fileName).Return(retErr)
}

func (fs *FileStoreMockedObject) GetChunkSize() int {
	args := fs.Called()

	return args.Int(0)
}

func (fs *FileStoreMockedObject) onGetChunkSize(retSize int) {
	fs.On("GetChunkSize").Return(retSize)
}

type DBFilerMockedObject struct {
	mock.Mock
}

func (dbf *DBFilerMockedObject) Write(chunk []byte) (int, error) {
	args := dbf.Called(chunk)
	return args.Int(0), args.Error(1)
}

func (dbf *DBFilerMockedObject) GetChunk() ([]byte, error) {
	args := dbf.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (dbf *DBFilerMockedObject) Close() error {
	args := dbf.Called()
	return args.Error(0)
}

func (dbf *DBFilerMockedObject) onWriteOnce(chunk []byte, retInt int, retErr error) {
	dbf.On("Write", chunk).Return(retInt, retErr).Once()
}

func (dbf *DBFilerMockedObject) onGetChunkOnce(retChunk []byte, retErr error) {
	dbf.On("GetChunk").Return(retChunk, retErr).Once()
}

func (dbf *DBFilerMockedObject) onClose(retErr error) {
	dbf.On("Close").Return(retErr)
}

type GophKeeper_CreateFileServerMockedObject struct {
	mock.Mock
	grpc.ServerStream
}

func (m *GophKeeper_CreateFileServerMockedObject) SendAndClose(res *proto.CreateFileResponse) error {
	args := m.Called(res)
	return args.Error(0)
}

func (m *GophKeeper_CreateFileServerMockedObject) Recv() (*proto.CreateFileRequest, error) {
	args := m.Called()
	return args.Get(0).(*proto.CreateFileRequest), args.Error(1)
}

func (m *GophKeeper_CreateFileServerMockedObject) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *GophKeeper_CreateFileServerMockedObject) onSendAndClose(res *proto.CreateFileResponse, retErr error) {
	m.On("SendAndClose", res).Return(retErr)
}

func (m *GophKeeper_CreateFileServerMockedObject) onRecvWithOnce(retReq *proto.CreateFileRequest, retErr error) {
	m.On("Recv").Return(retReq, retErr).Once()
}

func (m *GophKeeper_CreateFileServerMockedObject) onContext(retCtx context.Context) {
	m.On("Context").Return(retCtx)
}

type GophKeeper_UpdateFileServerMockedObject struct {
	mock.Mock
	grpc.ServerStream
}

func (m *GophKeeper_UpdateFileServerMockedObject) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *GophKeeper_UpdateFileServerMockedObject) SendAndClose(res *proto.UpdateFileResponse) error {
	args := m.Called(res)
	return args.Error(0)
}

func (m *GophKeeper_UpdateFileServerMockedObject) Recv() (*proto.UpdateFileRequest, error) {
	args := m.Called()
	return args.Get(0).(*proto.UpdateFileRequest), args.Error(1)
}

func (m *GophKeeper_UpdateFileServerMockedObject) onContext(ctx context.Context) {
	m.On("Context").Return(ctx)
}

func (m *GophKeeper_UpdateFileServerMockedObject) onSendAndClose(res *proto.UpdateFileResponse, err error) {
	m.On("SendAndClose", res).Return(err)
}

func (m *GophKeeper_UpdateFileServerMockedObject) onRecvWithOnce(req *proto.UpdateFileRequest, err error) {
	m.On("Recv").Return(req, err).Once()
}

type GophKeeper_GetFileServerMockedObject struct {
	mock.Mock
	grpc.ServerStream
}

func (m *GophKeeper_GetFileServerMockedObject) Send(res *proto.GetFileResponse) error {
	args := m.Called(res)
	return args.Error(0)
}

func (m *GophKeeper_GetFileServerMockedObject) Recv() (*proto.GetFileRequest, error) {
	args := m.Called()
	return args.Get(0).(*proto.GetFileRequest), args.Error(1)
}

func (m *GophKeeper_GetFileServerMockedObject) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *GophKeeper_GetFileServerMockedObject) onSendOnce(res *proto.GetFileResponse, retErr error) {
	m.On("Send", res).Return(retErr).Once()
}

func (m *GophKeeper_GetFileServerMockedObject) onRecvWithOnce(retReq *proto.GetFileRequest, retErr error) {
	m.On("Recv").Return(retReq, retErr).Once()
}

func (m *GophKeeper_GetFileServerMockedObject) onContext(retCtx context.Context) {
	m.On("Context").Return(retCtx)
}
