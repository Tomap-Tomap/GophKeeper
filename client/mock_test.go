//go:build unit

package client

import (
	"context"

	"github.com/Tomap-Tomap/GophKeeper/crypto"
	"github.com/Tomap-Tomap/GophKeeper/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type GophKeeperServerMockedObject struct {
	proto.UnimplementedGophKeeperServer
	mock.Mock
}

func (m *GophKeeperServerMockedObject) Register(_ context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.RegisterResponse), args.Error(1)
}

func (m *GophKeeperServerMockedObject) onRegister(req *proto.RegisterRequest, retRes *proto.RegisterResponse, retErr error) {
	m.On("Register", req).Return(retRes, retErr)
}

func (m *GophKeeperServerMockedObject) Auth(_ context.Context, req *proto.AuthRequest) (*proto.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.AuthResponse), args.Error(1)
}

func (m *GophKeeperServerMockedObject) onAuth(req *proto.AuthRequest, retRes *proto.AuthResponse, retErr error) {
	m.On("Auth", req).Return(retRes, retErr)
}

func (m *GophKeeperServerMockedObject) GetChunkSize(_ context.Context, _ *empty.Empty) (*proto.GetChunkSizeResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.GetChunkSizeResponse), args.Error(1)
}

func (m *GophKeeperServerMockedObject) onGetChunkSize(retRes *proto.GetChunkSizeResponse, retErr error) {
	m.On("GetChunkSize").Return(retRes, retErr)
}

func (m *GophKeeperServerMockedObject) GetPasswords(_ context.Context, _ *empty.Empty) (*proto.GetPasswordsResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.GetPasswordsResponse), args.Error(1)
}

func (m *GophKeeperServerMockedObject) onGetPasswords(retRes *proto.GetPasswordsResponse, retErr error) {
	m.On("GetPasswords").Return(retRes, retErr)
}

func (m *GophKeeperServerMockedObject) CreatePassword(_ context.Context, req *proto.CreatePasswordRequest) (*proto.CreatePasswordResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.CreatePasswordResponse), args.Error(1)
}

func (m *GophKeeperServerMockedObject) onCreatePassword(req *proto.CreatePasswordRequest, retRes *proto.CreatePasswordResponse, retErr error) {
	m.On("CreatePassword", req).Return(retRes, retErr)
}

func (m *GophKeeperServerMockedObject) UpdatePassword(_ context.Context, req *proto.UpdatePasswordRequest) (*proto.UpdatePasswordResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.UpdatePasswordResponse), args.Error(1)
}

func (m *GophKeeperServerMockedObject) onUpdatePassword(req *proto.UpdatePasswordRequest, retRes *proto.UpdatePasswordResponse, retErr error) {
	m.On("UpdatePassword", req).Return(retRes, retErr)
}

func (m *GophKeeperServerMockedObject) DeletePassword(_ context.Context, req *proto.DeletePasswordRequest) (*empty.Empty, error) {
	args := m.Called(req)

	return nil, args.Error(1)
}

func (m *GophKeeperServerMockedObject) onDeletePassword(req *proto.DeletePasswordRequest, retErr error) {
	m.On("DeletePassword", req).Return(nil, retErr)
}

func (m *GophKeeperServerMockedObject) GetBanks(_ context.Context, _ *empty.Empty) (*proto.GetBanksResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.GetBanksResponse), args.Error(1)
}

func (m *GophKeeperServerMockedObject) onGetBanks(retRes *proto.GetBanksResponse, retErr error) {
	m.On("GetBanks").Return(retRes, retErr)
}

func (m *GophKeeperServerMockedObject) CreateBank(_ context.Context, req *proto.CreateBankRequest) (*proto.CreateBankResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.CreateBankResponse), args.Error(1)
}

func (m *GophKeeperServerMockedObject) onCreateBank(req *proto.CreateBankRequest, retRes *proto.CreateBankResponse, retErr error) {
	m.On("CreateBank", req).Return(retRes, retErr)
}

func (m *GophKeeperServerMockedObject) UpdateBank(_ context.Context, req *proto.UpdateBankRequest) (*proto.UpdateBankResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.UpdateBankResponse), args.Error(1)
}

func (m *GophKeeperServerMockedObject) onUpdateBank(req *proto.UpdateBankRequest, retRes *proto.UpdateBankResponse, retErr error) {
	m.On("UpdateBank", req).Return(retRes, retErr)
}

func (m *GophKeeperServerMockedObject) DeleteBank(_ context.Context, req *proto.DeleteBankRequest) (*empty.Empty, error) {
	args := m.Called(req)

	return nil, args.Error(1)
}

func (m *GophKeeperServerMockedObject) onDeleteBank(req *proto.DeleteBankRequest, retErr error) {
	m.On("DeleteBank", req).Return(nil, retErr)
}

func (m *GophKeeperServerMockedObject) GetTexts(_ context.Context, _ *empty.Empty) (*proto.GetTextsResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.GetTextsResponse), args.Error(1)
}

func (m *GophKeeperServerMockedObject) onGetTexts(retRes *proto.GetTextsResponse, retErr error) {
	m.On("GetTexts").Return(retRes, retErr)
}

func (m *GophKeeperServerMockedObject) CreateText(_ context.Context, req *proto.CreateTextRequest) (*proto.CreateTextResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.CreateTextResponse), args.Error(1)
}

func (m *GophKeeperServerMockedObject) onCreateText(req *proto.CreateTextRequest, retRes *proto.CreateTextResponse, retErr error) {
	m.On("CreateText", req).Return(retRes, retErr)
}

func (m *GophKeeperServerMockedObject) UpdateText(_ context.Context, req *proto.UpdateTextRequest) (*proto.UpdateTextResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.UpdateTextResponse), args.Error(1)
}

func (m *GophKeeperServerMockedObject) onUpdateText(req *proto.UpdateTextRequest, retRes *proto.UpdateTextResponse, retErr error) {
	m.On("UpdateText", req).Return(retRes, retErr)
}

func (m *GophKeeperServerMockedObject) DeleteText(_ context.Context, req *proto.DeleteTextRequest) (*empty.Empty, error) {
	args := m.Called(req)

	return nil, args.Error(1)
}

func (m *GophKeeperServerMockedObject) onDeleteText(req *proto.DeleteTextRequest, retErr error) {
	m.On("DeleteText", req).Return(nil, retErr)
}

func (m *GophKeeperServerMockedObject) GetFiles(_ context.Context, _ *empty.Empty) (*proto.GetFilesResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.GetFilesResponse), args.Error(1)
}

func (m *GophKeeperServerMockedObject) onGetFiles(retRes *proto.GetFilesResponse, retErr error) {
	m.On("GetFiles").Return(retRes, retErr)
}

func (m *GophKeeperServerMockedObject) CreateFile(_ proto.GophKeeper_CreateFileServer) error {
	args := m.Called()

	return args.Error(0)
}

func (m *GophKeeperServerMockedObject) onCreateFile(retErr error) {
	m.On("CreateFile").Return(retErr)
}

func (m *GophKeeperServerMockedObject) UpdateFile(stream proto.GophKeeper_UpdateFileServer) error {
	args := m.Called(stream)

	return args.Error(1)
}

func (m *GophKeeperServerMockedObject) GetFile(req *proto.GetFileRequest, stream proto.GophKeeper_GetFileServer) error {
	args := m.Called(req, stream)

	return args.Error(1)
}

func (m *GophKeeperServerMockedObject) DeleteFile(_ context.Context, req *proto.DeleteFileRequest) (*empty.Empty, error) {
	args := m.Called(req)

	return nil, args.Error(1)
}

func (m *GophKeeperServerMockedObject) onDeleteFile(req *proto.DeleteFileRequest, retErr error) {
	m.On("DeleteFile", req).Return(nil, retErr)
}

type CrypterMockedObject struct {
	mock.Mock
}

func (m *CrypterMockedObject) SealStringWithoutNonce(str string) (string, error) {
	args := m.Called(str)

	return args.String(0), args.Error(1)
}

func (m *CrypterMockedObject) onSealStringWithoutNonce(str, retStr string, retError error) {
	m.On("SealStringWithoutNonce", str).Return(retStr, retError)
}

func (m *CrypterMockedObject) OpenStringWithoutNonce(encryptStr string) (string, error) {
	args := m.Called(encryptStr)

	return args.String(0), args.Error(1)
}

func (m *CrypterMockedObject) onOpenStringWithoutNonce(encryptStr, retStr string, retError error) *mock.Call {
	return m.On("OpenStringWithoutNonce", encryptStr).Return(retStr, retError)
}

func (m *CrypterMockedObject) GenerateNonce() ([]byte, error) {
	args := m.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]byte), args.Error(1)
}

func (m *CrypterMockedObject) onGenerateNonce(retBytes []byte, retErr error) {
	m.On("GenerateNonce").Return(retBytes, retErr)
}

func (m *CrypterMockedObject) SealBytes(b, nonce []byte) []byte {
	args := m.Called(b, nonce)

	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).([]byte)
}

func (m *CrypterMockedObject) onSealBytes(b, nonce, retb []byte) *mock.Call {
	return m.On("SealBytes", b, nonce).Return(retb)
}

func (m *CrypterMockedObject) NonceSize() int {
	args := m.Called()

	return args.Int(0)
}

func (m *CrypterMockedObject) onNonceSize(retSize int) {
	m.On("NonceSize").Return(retSize)
}

func (m *CrypterMockedObject) GetNonceFromBytes(b []byte, nonceSize int, location crypto.NonceLocation) ([]byte, []byte, int, error) {
	args := m.Called(b, nonceSize, location)

	if args.Get(0) == nil && args.Get(1) == nil {
		return nil, nil, args.Int(2), args.Error(3)
	}

	if args.Get(0) == nil {
		return nil, args.Get(1).([]byte), args.Int(2), args.Error(3)
	}

	if args.Get(1) == nil {
		return args.Get(0).([]byte), nil, args.Int(2), args.Error(3)
	}

	return args.Get(0).([]byte), args.Get(1).([]byte), args.Int(2), args.Error(3)
}

func (m *CrypterMockedObject) onGetNonceFromBytes(b []byte, nonceSize int, location crypto.NonceLocation, retNonce, retContent []byte, retSize int, retErr error) {
	m.On("GetNonceFromBytes", b, nonceSize, location).Return(retNonce, retContent, retSize, retErr)
}

func (m *CrypterMockedObject) OpenBytes(enctyptB []byte, nonce []byte) ([]byte, error) {
	args := m.Called(enctyptB, nonce)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]byte), args.Error(1)
}

func (m *CrypterMockedObject) onOpenBytes(enctyptB, nonce, retBytes []byte, retErr error) *mock.Call {
	return m.On("OpenBytes", enctyptB, nonce).Return(retBytes, retErr)
}

type CreateFileClientMockedObject struct {
	mock.Mock
	grpc.ClientStream
}

func (m *CreateFileClientMockedObject) Send(req *proto.CreateFileRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *CreateFileClientMockedObject) onSend(req *proto.CreateFileRequest, retErr error) *mock.Call {
	return m.On("Send", req).Return(retErr)
}

func (m *CreateFileClientMockedObject) CloseAndRecv() (*proto.CreateFileResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.CreateFileResponse), args.Error(1)
}

func (m *CreateFileClientMockedObject) onCloseAndRecv(retRes *proto.CreateFileResponse, retErr error) {
	m.On("CloseAndRecv").Return(retRes, retErr)
}

type GophKeeperClientMockedObject struct {
	mock.Mock
	proto.GophKeeperClient
}

func (m *GophKeeperClientMockedObject) GetChunkSize(_ context.Context, _ *empty.Empty, _ ...grpc.CallOption) (*proto.GetChunkSizeResponse, error) {
	args := m.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.GetChunkSizeResponse), args.Error(1)
}

func (m *GophKeeperClientMockedObject) onGetChunkSize(retRes *proto.GetChunkSizeResponse, retErr error) {
	m.On("GetChunkSize").Return(retRes, retErr)
}

func (m *GophKeeperClientMockedObject) CreateFile(_ context.Context, _ ...grpc.CallOption) (proto.GophKeeper_CreateFileClient, error) {
	args := m.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(proto.GophKeeper_CreateFileClient), args.Error(1)
}

func (m *GophKeeperClientMockedObject) onCreateFile(retStream proto.GophKeeper_CreateFileClient, retErr error) {
	m.On("CreateFile").Return(retStream, retErr)
}

func (m *GophKeeperClientMockedObject) UpdateFile(_ context.Context, _ ...grpc.CallOption) (proto.GophKeeper_UpdateFileClient, error) {
	args := m.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(proto.GophKeeper_UpdateFileClient), args.Error(1)
}

func (m *GophKeeperClientMockedObject) onUpdateFile(retStream proto.GophKeeper_UpdateFileClient, retErr error) {
	m.On("UpdateFile").Return(retStream, retErr)
}

func (m *GophKeeperClientMockedObject) GetFile(_ context.Context, req *proto.GetFileRequest, _ ...grpc.CallOption) (proto.GophKeeper_GetFileClient, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(proto.GophKeeper_GetFileClient), args.Error(1)
}

func (m *GophKeeperClientMockedObject) onGetFile(req *proto.GetFileRequest, retStream proto.GophKeeper_GetFileClient, retErr error) *mock.Call {
	return m.On("GetFile", req, mock.Anything).Return(retStream, retErr)
}

type UpdateFileClientMockedObject struct {
	mock.Mock
	grpc.ClientStream
}

func (m *UpdateFileClientMockedObject) Send(req *proto.UpdateFileRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *UpdateFileClientMockedObject) onSend(req *proto.UpdateFileRequest, retErr error) *mock.Call {
	return m.On("Send", req).Return(retErr)
}

func (m *UpdateFileClientMockedObject) CloseAndRecv() (*proto.UpdateFileResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.UpdateFileResponse), args.Error(1)
}

func (m *UpdateFileClientMockedObject) onCloseAndRecv(retRes *proto.UpdateFileResponse, retErr error) *mock.Call {
	return m.On("CloseAndRecv").Return(retRes, retErr)
}

type GetFileClientMockedObject struct {
	mock.Mock
	grpc.ClientStream
}

func (m *GetFileClientMockedObject) Recv() (*proto.GetFileResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.GetFileResponse), args.Error(1)
}

func (m *GetFileClientMockedObject) onRecv(retRes *proto.GetFileResponse, retErr error) *mock.Call {
	return m.On("Recv").Return(retRes, retErr)
}
