//go:build unit

package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/Tomap-Tomap/GophKeeper/proto"
	"github.com/Tomap-Tomap/GophKeeper/storage"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func RuneGPRCTestServer(t *testing.T, smo *StorageMockedObject, hmo *HasherMockedObject, tmo *TokenerMockerdObject, fsmo *FileStoreMockerObject) (addr string, stopFunc func()) {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	s := grpc.NewServer()

	proto.RegisterGophKeeperServer(s, NewGophKeeperHandler(smo, hmo, tmo, fsmo))

	go func() {
		if err := s.Serve(lis); err != nil {
			require.FailNow(t, err.Error())
		}
	}()

	addr = lis.Addr().String()
	stopFunc = s.Stop

	return
}

func CreateGRPCTestClient(t *testing.T, addr string) (client proto.GophKeeperClient, closeFunc func() error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	require.NoError(t, err)

	closeFunc = conn.Close
	client = proto.NewGophKeeperClient(conn)

	return
}

type SuiteGK struct {
	suite.Suite

	smo        *StorageMockedObject
	hmo        *HasherMockedObject
	tmo        *TokenerMockerdObject
	fsmo       *FileStoreMockerObject
	client     proto.GophKeeperClient
	stopClient func() error
	stopServer func()

	wantUser          *storage.User
	wantPassword      *storage.Password
	wantProtoPassword *proto.Password
	wantBank          *storage.Bank
	wantProtoBank     *proto.Bank
	wantText          *storage.Text
	wantProtoText     *proto.Text
	wantFile          *storage.File
	wantProtoFile     *proto.File

	testBuffer bytes.Buffer
	testBatch1 []byte
	testBatch2 []byte

	testUserID     string
	testPasswordID string
	testBankID     string
	testTextID     string
	testFileID     string

	testPassword   string
	testLogin      string
	testLoginHash  string
	testSalt       string
	testHash       string
	testToken      string
	testName       string
	testMeta       string
	testBankData   string
	testText       string
	testPathToFile string
}

func (s *SuiteGK) SetupSuite() {
	smo := new(StorageMockedObject)
	hmo := new(HasherMockedObject)
	tmo := new(TokenerMockerdObject)
	fsmo := new(FileStoreMockerObject)

	addr, stopServer := RuneGPRCTestServer(s.T(), smo, hmo, tmo, fsmo)

	client, closeClient := CreateGRPCTestClient(s.T(), addr)

	s.smo = smo
	s.hmo = hmo
	s.tmo = tmo
	s.fsmo = fsmo
	s.client = client
	s.stopClient = closeClient
	s.stopServer = stopServer

	s.testUserID = "TestUserID"
	s.testPasswordID = "TestPasswordID"
	s.testBankID = "TestBankID"
	s.testTextID = "TestTextID"
	s.testFileID = "TestFileID"

	s.testPassword = "TestPassword"
	s.testLogin = "TestLogin"
	s.testLoginHash = "TestLoginHash"
	s.testSalt = "TestSalt"
	s.testHash = "TestHash"
	s.testToken = "TestToken"
	s.testName = "TestName"
	s.testMeta = "TestMeta"
	s.testBankData = "TestBankData"
	s.testText = "TestText"
	s.testPathToFile = "TestPathToFile"

	s.wantUser = &storage.User{
		ID:       s.testUserID,
		Login:    s.testLogin,
		Password: s.testHash,
		Salt:     s.testSalt,
	}

	updateAt := time.Now()
	s.wantPassword = &storage.Password{
		ID:       s.testPasswordID,
		UserID:   s.testUserID,
		Name:     s.testName,
		Login:    s.testLogin,
		Password: s.testPassword,
		Meta:     s.testMeta,
		UpdateAt: updateAt,
	}
	s.wantProtoPassword = &proto.Password{
		Id:       s.testPasswordID,
		UserID:   s.testUserID,
		Name:     s.testName,
		Login:    s.testLogin,
		Password: s.testPassword,
		Meta:     s.testMeta,
		UpdateAt: timestamppb.New(updateAt),
	}

	s.wantBank = &storage.Bank{
		ID:        s.testBankID,
		UserID:    s.testUserID,
		Name:      s.testName,
		BanksData: s.testBankData,
		Meta:      s.testMeta,
		UpdateAt:  updateAt,
	}
	s.wantProtoBank = &proto.Bank{
		Id:        s.testBankID,
		UserID:    s.testUserID,
		Name:      s.testName,
		BanksData: s.testBankData,
		Meta:      s.testMeta,
		UpdateAt:  timestamppb.New(updateAt),
	}

	s.wantText = &storage.Text{
		ID:       s.testTextID,
		UserID:   s.testUserID,
		Name:     s.testName,
		Text:     s.testText,
		Meta:     s.testMeta,
		UpdateAt: updateAt,
	}
	s.wantProtoText = &proto.Text{
		Id:       s.testTextID,
		UserID:   s.testUserID,
		Name:     s.testName,
		Text:     s.testText,
		Meta:     s.testMeta,
		UpdateAt: timestamppb.New(updateAt),
	}

	s.wantFile = &storage.File{
		ID:         s.testFileID,
		UserID:     s.testUserID,
		Name:       s.testName,
		PathToFile: s.testPathToFile,
		Meta:       s.testMeta,
		UpdateAt:   updateAt,
	}
	s.wantProtoFile = &proto.File{
		Id:       s.testFileID,
		UserID:   s.testUserID,
		Name:     s.testName,
		Meta:     s.testMeta,
		UpdateAt: timestamppb.New(updateAt),
	}

	s.testBatch1 = make([]byte, 0, 64)

	for i := 0; i < 64; i++ {
		s.testBatch1 = append(s.testBatch1, byte(i))
		s.testBatch2 = append(s.testBatch2, byte(i))
	}

	_, err := s.testBuffer.Write(s.testBatch1)
	s.Require().NoError(err)
	_, err = s.testBuffer.Write(s.testBatch2)
	s.Require().NoError(err)
}

func (s *SuiteGK) TearDownSuite() {
	s.stopServer()

	err := s.stopClient()
	s.NoError(err)
}

func (s *SuiteGK) TearDownSubTest() {
	s.hmo.AssertExpectations(s.T())
	s.smo.AssertExpectations(s.T())
	s.tmo.AssertExpectations(s.T())
	s.fsmo.AssertExpectations(s.T())

	for len(s.hmo.ExpectedCalls) != 0 {
		s.hmo.ExpectedCalls[0].Unset()
	}

	for len(s.smo.ExpectedCalls) != 0 {
		s.smo.ExpectedCalls[0].Unset()
	}

	for len(s.tmo.ExpectedCalls) != 0 {
		s.tmo.ExpectedCalls[0].Unset()
	}

	for len(s.fsmo.ExpectedCalls) != 0 {
		s.fsmo.ExpectedCalls[0].Unset()
	}
}

func (s *SuiteGK) TestEmptyArguments() {
	for _, t := range []struct {
		name, login, password string
		err                   []string
	}{
		{"empty login error", "", s.testPassword, []string{"empty login"}},
		{"empty password error", s.testLogin, "", []string{"empty password"}},
		{"empty login and password error", "", "", []string{"empty login", "empty password"}},
	} {
		s.Run(t.name+" registry", func() {
			res, err := s.client.Register(context.Background(), &proto.RegisterRequest{
				Login:    t.login,
				Password: t.password,
			})

			s.Require().ErrorContains(err, t.err[0])

			if len(t.err) > 1 {
				s.Require().ErrorContains(err, t.err[1])
			}

			s.Equal(status.Code(err), codes.InvalidArgument)
			s.Nil(res)
		})
		s.Run(t.name+" auth", func() {
			res, err := s.client.Auth(context.Background(), &proto.AuthRequest{
				Login:    t.login,
				Password: t.password,
			})

			s.Require().ErrorContains(err, t.err[0])

			if len(t.err) > 1 {
				s.Require().ErrorContains(err, t.err[1])
			}

			s.Equal(status.Code(err), codes.InvalidArgument)
			s.Nil(res)
		})
	}

	for _, t := range []struct {
		name, spesialError string
		spesialMethod      func() (any, error)
		methods            []func() (any, error)
	}{
		{
			name:         "password empty errors",
			spesialError: "empty PasswordID",
			spesialMethod: func() (any, error) {
				return s.client.GetPassword(context.Background(), &proto.GetPasswordRequest{})
			},
			methods: []func() (any, error){
				func() (any, error) {
					return s.client.CreatePassword(context.Background(), &proto.CreatePasswordRequest{})
				},
				func() (any, error) {
					return s.client.GetPasswords(context.Background(), &proto.GetPasswordsRequest{})
				},
			},
		},
		{
			name:         "bank empty erros",
			spesialError: "empty BankID",
			spesialMethod: func() (any, error) {
				return s.client.GetBank(context.Background(), &proto.GetBankRequest{})
			},
			methods: []func() (any, error){
				func() (any, error) {
					return s.client.CreateBank(context.Background(), &proto.CreateBankRequest{})
				},
				func() (any, error) {
					return s.client.GetBanks(context.Background(), &proto.GetBanksRequest{})
				},
			},
		},
		{
			name:         "text empty erros",
			spesialError: "empty TextID",
			spesialMethod: func() (any, error) {
				return s.client.GetText(context.Background(), &proto.GetTextRequest{})
			},
			methods: []func() (any, error){
				func() (any, error) {
					return s.client.CreateText(context.Background(), &proto.CreateTextRequest{})
				},
				func() (any, error) {
					return s.client.GetTexts(context.Background(), &proto.GetTextsRequest{})
				},
			},
		},
		{
			name:         "file empty erros",
			spesialError: "empty FileID",
			spesialMethod: func() (any, error) {
				stream, err := s.client.GetFile(context.Background(), &proto.GetFileRequest{})
				s.Require().NoError(err)

				return stream.Recv()
			},
			methods: []func() (any, error){
				func() (any, error) {
					stream, err := s.client.CreateFile(context.Background())
					s.Require().NoError(err)
					err = stream.Send(&proto.CreateFileRequest{Data: &proto.CreateFileRequest_FileInfo{}})
					s.Require().NoError(err)

					return stream.CloseAndRecv()
				},
				func() (any, error) {
					stream, err := s.client.GetFiles(context.Background(), &proto.GetFilesRequest{})
					s.Require().NoError(err)

					return stream.Recv()
				},
			},
		},
	} {
		s.Run(t.name, func() {
			for _, val := range t.methods {
				res, err := val()

				s.Require().ErrorContains(err, "empty UserID")
				s.Equal(status.Code(err), codes.InvalidArgument)
				s.Nil(res)
			}

			res, err := t.spesialMethod()

			s.Require().ErrorContains(err, t.spesialError)
			s.Equal(status.Code(err), codes.InvalidArgument)
			s.Nil(res)
		})
	}
}

func (s *SuiteGK) TestRegisterErrors() {
	for _, t := range []struct {
		name      string
		err       string
		code      codes.Code
		setupMock func()
	}{
		{
			name: "login hash error",
			err:  "generate hash",
			code: codes.Internal,
			setupMock: func() {
				s.hmo.On("GetHash", s.testLogin).Return("", errors.New("hash login error"))
			},
		},
		{
			name: "salt error",
			err:  "generate salt",
			code: codes.Internal,
			setupMock: func() {
				s.hmo.On("GetHash", s.testLogin).Return(s.testLoginHash, nil)
				s.hmo.On("GenerateSalt").Return("", errors.New("salt error"))
			},
		},
		{
			name: "hash error",
			err:  "generate hash",
			code: codes.Internal,
			setupMock: func() {
				s.hmo.On("GetHash", s.testLogin).Return(s.testLoginHash, nil)
				s.hmo.On("GenerateSalt").Return(s.testSalt, nil)
				s.hmo.On("GetHashWithSalt", s.testPassword, s.testSalt).Return("", errors.New("hash error"))
			},
		},
		{
			name: "db connection error",
			err:  fmt.Sprintf("create user %s", s.testLogin),
			code: codes.Internal,
			setupMock: func() {
				s.hmo.On("GetHash", s.testLogin).Return(s.testLoginHash, nil)
				s.hmo.On("GenerateSalt").Return(s.testSalt, nil)
				s.hmo.On("GetHashWithSalt", s.testPassword, s.testSalt).Return(s.testHash, nil)

				s.smo.On("CreateUser", s.testLogin, s.testLoginHash, s.testSalt, s.testHash).
					Return(nil, &pgconn.PgError{Code: "08000"})
			},
		},
		{
			name: "user alredy exist error",
			err:  fmt.Sprintf("user %s already exists", s.testLogin),
			code: codes.AlreadyExists,
			setupMock: func() {
				s.hmo.On("GetHash", s.testLogin).Return(s.testLoginHash, nil)
				s.hmo.On("GenerateSalt").Return(s.testSalt, nil)
				s.hmo.On("GetHashWithSalt", s.testPassword, s.testSalt).Return(s.testHash, nil)

				s.smo.On("CreateUser", s.testLogin, s.testLoginHash, s.testSalt, s.testHash).
					Return(nil, &pgconn.PgError{Code: "23505"})
			},
		},
		{
			name: "tokener error",
			err:  fmt.Sprintf("gen token for user %s", s.testLogin),
			code: codes.Internal,
			setupMock: func() {
				s.hmo.On("GetHash", s.testLogin).Return(s.testLoginHash, nil)
				s.hmo.On("GenerateSalt").Return(s.testSalt, nil)
				s.hmo.On("GetHashWithSalt", s.testPassword, s.testSalt).Return(s.testHash, nil)

				s.smo.On("CreateUser", s.testLogin, s.testLoginHash, s.testSalt, s.testHash).
					Return(s.wantUser, nil)

				s.tmo.On("GetToken", mock.Anything).Return("", errors.New("token error"))
			},
		},
	} {
		s.Run(t.name, func() {
			t.setupMock()

			res, err := s.client.Register(context.Background(), &proto.RegisterRequest{
				Login:    s.testLogin,
				Password: s.testPassword,
			})

			s.Require().ErrorContains(err, t.err)
			s.Equal(status.Code(err), t.code)
			s.Nil(res)
		})
	}
}

func (s *SuiteGK) TestAuthErrors() {
	for _, t := range []struct {
		name      string
		err       string
		code      codes.Code
		setupMock func()
	}{
		{
			name: "login hash error",
			err:  "generate hash",
			code: codes.Internal,
			setupMock: func() {
				s.hmo.On("GetHash", s.testLogin).Return("", errors.New("hash login error"))
			},
		},
		{
			name: "db connection error",
			err:  fmt.Sprintf("get user %s", s.testLogin),
			code: codes.Internal,
			setupMock: func() {
				s.hmo.On("GetHash", s.testLogin).Return(s.testLoginHash, nil)

				s.smo.On("GetUser", s.testLogin, s.testLoginHash).
					Return(nil, &pgconn.PgError{Code: "08000"})
			},
		},
		{
			name: "user unknown error",
			err:  fmt.Sprintf("unknown user %s", s.testLogin),
			code: codes.Unknown,
			setupMock: func() {
				s.hmo.On("GetHash", s.testLogin).Return(s.testLoginHash, nil)

				s.smo.On("GetUser", s.testLogin, s.testLoginHash).
					Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "hash error",
			err:  "generate hash",
			code: codes.Internal,
			setupMock: func() {
				s.hmo.On("GetHash", s.testLogin).Return(s.testLoginHash, nil)
				s.hmo.On("GetHashWithSalt", s.testPassword, s.testSalt).Return("", errors.New("hash error"))

				s.smo.On("GetUser", s.testLogin, s.testLoginHash).
					Return(s.wantUser, nil)
			},
		},
		{
			name: "invalid password error",
			err:  "invalid password",
			code: codes.PermissionDenied,
			setupMock: func() {
				s.hmo.On("GetHash", s.testLogin).Return(s.testLoginHash, nil)
				s.hmo.On("GetHashWithSalt", s.testPassword, s.testSalt).Return("invalidHash", nil)

				s.smo.On("GetUser", s.testLogin, s.testLoginHash).
					Return(s.wantUser, nil)
			},
		},
		{
			name: "tokener error",
			err:  fmt.Sprintf("gen token for user %s", s.testLogin),
			code: codes.Internal,
			setupMock: func() {
				s.hmo.On("GetHash", s.testLogin).Return(s.testLoginHash, nil)
				s.hmo.On("GetHashWithSalt", s.testPassword, s.testSalt).Return(s.testHash, nil)

				s.smo.On("GetUser", s.testLogin, s.testLoginHash).
					Return(s.wantUser, nil)

				s.tmo.On("GetToken", mock.Anything).Return("", errors.New("token error"))
			},
		},
	} {
		s.Run(t.name, func() {
			t.setupMock()

			res, err := s.client.Auth(context.Background(), &proto.AuthRequest{
				Login:    s.testLogin,
				Password: s.testPassword,
			})

			s.Require().ErrorContains(err, t.err)
			s.Equal(status.Code(err), t.code)
			s.Nil(res)
		})
	}
}

func (s *SuiteGK) TestCreatePasswordErrors() {
	for _, t := range []struct {
		name      string
		err       string
		code      codes.Code
		setupMock func()
	}{
		{
			name: "db connection error",
			err:  fmt.Sprintf("create password for user %s", s.testUserID),
			code: codes.Internal,
			setupMock: func() {
				s.smo.On(
					"CreatePassword",
					s.testUserID,
					s.testName,
					s.testLogin,
					s.testPassword,
					s.testMeta,
				).Return(nil, &pgconn.PgError{Code: "08000"})
			},
		},
		{
			name: "unknown UserID error",
			err:  fmt.Sprintf("unknown UserID %s", s.testUserID),
			code: codes.Unknown,
			setupMock: func() {
				s.smo.On(
					"CreatePassword",
					s.testUserID,
					s.testName,
					s.testLogin,
					s.testPassword,
					s.testMeta,
				).Return(nil, &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
		},
	} {
		s.Run(t.name, func() {
			t.setupMock()

			res, err := s.client.CreatePassword(context.Background(), &proto.CreatePasswordRequest{
				UserID:   s.testUserID,
				Name:     s.testName,
				Login:    s.testLogin,
				Password: s.testPassword,
				Meta:     s.testMeta,
			})

			s.Require().ErrorContains(err, t.err)
			s.Equal(status.Code(err), t.code)
			s.Nil(res)
		})
	}
}

func (s *SuiteGK) TestGetPasswordErrors() {
	for _, t := range []struct {
		name      string
		err       string
		code      codes.Code
		setupMock func()
	}{
		{
			name: "db connection error",
			err:  fmt.Sprintf("get password %s", s.testPasswordID),
			code: codes.Internal,
			setupMock: func() {
				s.smo.On(
					"GetPassword",
					s.testPasswordID,
				).Return(nil, &pgconn.PgError{Code: "08000"})
			},
		},
		{
			name: "unknown PasswordID error",
			err:  fmt.Sprintf("unknown PasswordID %s", s.testPasswordID),
			code: codes.Unknown,
			setupMock: func() {
				s.smo.On(
					"GetPassword",
					s.testPasswordID,
				).Return(nil, &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
		},
	} {
		s.Run(t.name, func() {
			t.setupMock()

			res, err := s.client.GetPassword(context.Background(), &proto.GetPasswordRequest{
				Id: s.testPasswordID,
			})

			s.Require().ErrorContains(err, t.err)
			s.Equal(status.Code(err), t.code)
			s.Nil(res)
		})
	}
}

func (s *SuiteGK) TestGetPasswordsErrors() {
	for _, t := range []struct {
		name      string
		err       string
		code      codes.Code
		setupMock func()
	}{
		{
			name: "db connection error",
			err:  fmt.Sprintf("get passwords %s", s.testUserID),
			code: codes.Internal,
			setupMock: func() {
				s.smo.On(
					"GetAllPassword",
					s.testUserID,
				).Return(nil, &pgconn.PgError{Code: "08000"})
			},
		},
		{
			name: "unknown UserID error",
			err:  fmt.Sprintf("unknown UserID %s", s.testUserID),
			code: codes.Unknown,
			setupMock: func() {
				s.smo.On(
					"GetAllPassword",
					s.testUserID,
				).Return(nil, &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
		},
	} {
		s.Run(t.name, func() {
			t.setupMock()

			res, err := s.client.GetPasswords(context.Background(), &proto.GetPasswordsRequest{
				UserID: s.testUserID,
			})

			s.Require().ErrorContains(err, t.err)
			s.Equal(status.Code(err), t.code)
			s.Nil(res)
		})
	}
}

func (s *SuiteGK) TestCreateBankErrors() {
	for _, t := range []struct {
		name      string
		err       string
		code      codes.Code
		setupMock func()
	}{
		{
			name: "db connection error",
			err:  fmt.Sprintf("create bank data for user %s", s.testUserID),
			code: codes.Internal,
			setupMock: func() {
				s.smo.On(
					"CreateBank",
					s.testUserID,
					s.testName,
					s.testBankData,
					s.testMeta,
				).Return(nil, &pgconn.PgError{Code: "08000"})
			},
		},
		{
			name: "unknown UserID error",
			err:  fmt.Sprintf("unknown UserID %s", s.testUserID),
			code: codes.Unknown,
			setupMock: func() {
				s.smo.On(
					"CreateBank",
					s.testUserID,
					s.testName,
					s.testBankData,
					s.testMeta,
				).Return(nil, &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
		},
	} {
		s.Run(t.name, func() {
			t.setupMock()

			res, err := s.client.CreateBank(context.Background(), &proto.CreateBankRequest{
				UserID:    s.testUserID,
				Name:      s.testName,
				BanksData: s.testBankData,
				Meta:      s.testMeta,
			})

			s.Require().ErrorContains(err, t.err)
			s.Equal(status.Code(err), t.code)
			s.Nil(res)
		})
	}
}

func (s *SuiteGK) TestGetBankErrors() {
	for _, t := range []struct {
		name      string
		err       string
		code      codes.Code
		setupMock func()
	}{
		{
			name: "db connection error",
			err:  fmt.Sprintf("get bank data %s", s.testBankID),
			code: codes.Internal,
			setupMock: func() {
				s.smo.On(
					"GetBank",
					s.testBankID,
				).Return(nil, &pgconn.PgError{Code: "08000"})
			},
		},
		{
			name: "unknown BankID error",
			err:  fmt.Sprintf("unknown BankID %s", s.testBankID),
			code: codes.Unknown,
			setupMock: func() {
				s.smo.On(
					"GetBank",
					s.testBankID,
				).Return(nil, &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
		},
	} {
		s.Run(t.name, func() {
			t.setupMock()

			res, err := s.client.GetBank(context.Background(), &proto.GetBankRequest{
				Id: s.testBankID,
			})

			s.Require().ErrorContains(err, t.err)
			s.Equal(status.Code(err), t.code)
			s.Nil(res)
		})
	}
}

func (s *SuiteGK) TestGetBanksErrors() {
	for _, t := range []struct {
		name      string
		err       string
		code      codes.Code
		setupMock func()
	}{
		{
			name: "db connection error",
			err:  fmt.Sprintf("get banks %s", s.testUserID),
			code: codes.Internal,
			setupMock: func() {
				s.smo.On(
					"GetAllBanks",
					s.testUserID,
				).Return(nil, &pgconn.PgError{Code: "08000"})
			},
		},
		{
			name: "unknown UserID error",
			err:  fmt.Sprintf("unknown UserID %s", s.testUserID),
			code: codes.Unknown,
			setupMock: func() {
				s.smo.On(
					"GetAllBanks",
					s.testUserID,
				).Return(nil, &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
		},
	} {
		s.Run(t.name, func() {
			t.setupMock()

			res, err := s.client.GetBanks(context.Background(), &proto.GetBanksRequest{
				UserID: s.testUserID,
			})

			s.Require().ErrorContains(err, t.err)
			s.Equal(status.Code(err), t.code)
			s.Nil(res)
		})
	}
}

func (s *SuiteGK) TestCreateTextErrors() {
	for _, t := range []struct {
		name      string
		err       string
		code      codes.Code
		setupMock func()
	}{
		{
			name: "db connection error",
			err:  fmt.Sprintf("create text for user %s", s.testUserID),
			code: codes.Internal,
			setupMock: func() {
				s.smo.On(
					"CreateText",
					s.testUserID,
					s.testName,
					s.testText,
					s.testMeta,
				).Return(nil, &pgconn.PgError{Code: "08000"})
			},
		},
		{
			name: "unknown UserID error",
			err:  fmt.Sprintf("unknown UserID %s", s.testUserID),
			code: codes.Unknown,
			setupMock: func() {
				s.smo.On(
					"CreateText",
					s.testUserID,
					s.testName,
					s.testText,
					s.testMeta,
				).Return(nil, &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
		},
	} {
		s.Run(t.name, func() {
			t.setupMock()

			res, err := s.client.CreateText(context.Background(), &proto.CreateTextRequest{
				UserID: s.testUserID,
				Name:   s.testName,
				Text:   s.testText,
				Meta:   s.testMeta,
			})

			s.Require().ErrorContains(err, t.err)
			s.Equal(status.Code(err), t.code)
			s.Nil(res)
		})
	}
}

func (s *SuiteGK) TestGetTextErrors() {
	for _, t := range []struct {
		name      string
		err       string
		code      codes.Code
		setupMock func()
	}{
		{
			name: "db connection error",
			err:  fmt.Sprintf("get text %s", s.testTextID),
			code: codes.Internal,
			setupMock: func() {
				s.smo.On(
					"GetText",
					s.testTextID,
				).Return(nil, &pgconn.PgError{Code: "08000"})
			},
		},
		{
			name: "unknown TextID error",
			err:  fmt.Sprintf("unknown TextID %s", s.testTextID),
			code: codes.Unknown,
			setupMock: func() {
				s.smo.On(
					"GetText",
					s.testTextID,
				).Return(nil, &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
		},
	} {
		s.Run(t.name, func() {
			t.setupMock()

			res, err := s.client.GetText(context.Background(), &proto.GetTextRequest{
				Id: s.testTextID,
			})

			s.Require().ErrorContains(err, t.err)
			s.Equal(status.Code(err), t.code)
			s.Nil(res)
		})
	}
}

func (s *SuiteGK) TestGetTextsErrors() {
	for _, t := range []struct {
		name      string
		err       string
		code      codes.Code
		setupMock func()
	}{
		{
			name: "db connection error",
			err:  fmt.Sprintf("get texts %s", s.testUserID),
			code: codes.Internal,
			setupMock: func() {
				s.smo.On(
					"GetAllTexts",
					s.testUserID,
				).Return(nil, &pgconn.PgError{Code: "08000"})
			},
		},
		{
			name: "unknown UserID error",
			err:  fmt.Sprintf("unknown UserID %s", s.testUserID),
			code: codes.Unknown,
			setupMock: func() {
				s.smo.On(
					"GetAllTexts",
					s.testUserID,
				).Return(nil, &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
		},
	} {
		s.Run(t.name, func() {
			t.setupMock()

			res, err := s.client.GetTexts(context.Background(), &proto.GetTextsRequest{
				UserID: s.testUserID,
			})

			s.Require().ErrorContains(err, t.err)
			s.Equal(status.Code(err), t.code)
			s.Nil(res)
		})
	}
}

func (s *SuiteGK) TestCreateFileErrors() {
	for _, t := range []struct {
		name      string
		err       string
		code      codes.Code
		setupMock func()
	}{
		{
			name: "file save error",
			err:  fmt.Sprintf("save file for user %s", s.testUserID),
			code: codes.Internal,
			setupMock: func() {
				s.fsmo.On("Save", s.testBuffer).Return("", errors.New("save file error"))
			},
		},
		{
			name: "db connection error",
			err:  fmt.Sprintf("create file for user %s", s.testUserID),
			code: codes.Internal,
			setupMock: func() {
				s.fsmo.On("Save", s.testBuffer).Return(s.testPathToFile, nil)

				s.smo.On("CreateFile", s.testUserID, s.testName, s.testPathToFile, s.testMeta).
					Return(nil, &pgconn.PgError{Code: "08000"})
			},
		},
		{
			name: "unknown UserID error",
			err:  fmt.Sprintf("unknown UserID %s", s.testUserID),
			code: codes.Unknown,
			setupMock: func() {
				s.fsmo.On("Save", s.testBuffer).Return(s.testPathToFile, nil)

				s.smo.On("CreateFile", s.testUserID, s.testName, s.testPathToFile, s.testMeta).
					Return(nil, &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
		},
	} {
		s.Run(t.name, func() {
			t.setupMock()

			stream, err := s.client.CreateFile(context.Background())
			s.Require().NoError(err)

			err = stream.Send(&proto.CreateFileRequest{
				Data: &proto.CreateFileRequest_FileInfo{
					FileInfo: &proto.File{
						UserID: s.testUserID,
						Name:   s.testName,
						Meta:   s.testMeta,
					},
				},
			})
			s.Require().NoError(err)

			err = stream.Send(&proto.CreateFileRequest{
				Data: &proto.CreateFileRequest_Content{
					Content: s.testBatch1,
				},
			})
			s.Require().NoError(err)

			err = stream.Send(&proto.CreateFileRequest{
				Data: &proto.CreateFileRequest_Content{
					Content: s.testBatch2,
				},
			})
			s.Require().NoError(err)

			res, err := stream.CloseAndRecv()

			s.Require().ErrorContains(err, t.err)
			s.Equal(status.Code(err), t.code)
			s.Nil(res)
		})
	}
}

func (s *SuiteGK) TestGetFileErrors() {
	for _, t := range []struct {
		name      string
		err       string
		code      codes.Code
		setupMock func() DBFiler
	}{
		{
			name: "db connection error",
			err:  fmt.Sprintf("get file %s", s.testFileID),
			code: codes.Internal,
			setupMock: func() DBFiler {
				s.smo.On(
					"GetFile",
					s.testFileID,
				).Return(nil, &pgconn.PgError{Code: "08000"})

				return nil
			},
		},
		{
			name: "unknown FileID error",
			err:  fmt.Sprintf("unknown FileID %s", s.testFileID),
			code: codes.Unknown,
			setupMock: func() DBFiler {
				s.smo.On(
					"GetFile",
					s.testFileID,
				).Return(nil, &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

				return nil
			},
		},
		{
			name: "get db filer error",
			err:  fmt.Sprintf("get file %s", s.testFileID),
			code: codes.Internal,
			setupMock: func() DBFiler {
				s.smo.On(
					"GetFile",
					s.testFileID,
				).Return(s.wantFile, nil)

				s.fsmo.On("GetDBFiler", s.testPathToFile).Return(nil, errors.New("test"))

				return nil
			},
		},
		{
			name: "get chunk error",
			err:  fmt.Sprintf("get file %s", s.testFileID),
			code: codes.Internal,
			setupMock: func() DBFiler {
				s.smo.On(
					"GetFile",
					s.testFileID,
				).Return(s.wantFile, nil)

				fmo := new(DBFilerMocketObject)
				s.fsmo.On("GetDBFiler", s.testPathToFile).Return(fmo, nil)

				fmo.On("GetChunck").Return(nil, errors.New("test error"))
				fmo.On("Close")

				return fmo
			},
		},
	} {
		s.Run(t.name, func() {
			fmo := t.setupMock()

			stream, err := s.client.GetFile(context.Background(), &proto.GetFileRequest{
				Id: s.testFileID,
			})
			s.Require().NoError(err)

			fi, err := stream.Recv()

			if err != nil {
				s.Require().ErrorContains(err, t.err)
				s.Equal(status.Code(err), t.code)
				s.Nil(fi)
				return
			}

			content, err := stream.Recv()

			s.Require().ErrorContains(err, t.err)
			s.Equal(status.Code(err), t.code)
			s.Nil(content)

			if fmo, ok := fmo.(*DBFilerMocketObject); ok {
				fmo.AssertExpectations(s.T())
			}
		})
	}
}

func (s *SuiteGK) TestGetFilesErrors() {
	for _, t := range []struct {
		name      string
		err       string
		code      codes.Code
		setupMock func() DBFiler
	}{
		{
			name: "db connection error",
			err:  fmt.Sprintf("get files %s", s.testUserID),
			code: codes.Internal,
			setupMock: func() DBFiler {
				s.smo.On(
					"GetAllFiles",
					s.testUserID,
				).Return(nil, &pgconn.PgError{Code: "08000"})

				return nil
			},
		},
		{
			name: "unknown UserID error",
			err:  fmt.Sprintf("unknown UserID %s", s.testUserID),
			code: codes.Unknown,
			setupMock: func() DBFiler {
				s.smo.On(
					"GetAllFiles",
					s.testUserID,
				).Return(nil, &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

				return nil
			},
		},
		{
			name: "get db filer error",
			err:  fmt.Sprintf("get files %s", s.testUserID),
			code: codes.Internal,
			setupMock: func() DBFiler {
				s.smo.On(
					"GetAllFiles",
					s.testUserID,
				).Return([]storage.File{*s.wantFile}, nil)

				s.fsmo.On("GetDBFiler", s.testPathToFile).Return(nil, errors.New("test"))

				return nil
			},
		},
		{
			name: "get chunk error",
			err:  fmt.Sprintf("get files %s", s.testUserID),
			code: codes.Internal,
			setupMock: func() DBFiler {
				s.smo.On(
					"GetAllFiles",
					s.testUserID,
				).Return([]storage.File{*s.wantFile}, nil)

				fmo := new(DBFilerMocketObject)
				s.fsmo.On("GetDBFiler", s.testPathToFile).Return(fmo, nil)

				fmo.On("GetChunck").Return(nil, errors.New("test error"))
				fmo.On("Close")

				return fmo
			},
		},
	} {
		s.Run(t.name, func() {
			fmo := t.setupMock()

			stream, err := s.client.GetFiles(context.Background(), &proto.GetFilesRequest{
				UserID: s.testUserID,
			})
			s.Require().NoError(err)

			fi, err := stream.Recv()

			if err != nil {
				s.Require().ErrorContains(err, t.err)
				s.Equal(status.Code(err), t.code)
				s.Nil(fi)
				return
			}

			content, err := stream.Recv()

			s.Require().ErrorContains(err, t.err)
			s.Equal(status.Code(err), t.code)
			s.Nil(content)

			if fmo, ok := fmo.(*DBFilerMocketObject); ok {
				fmo.AssertExpectations(s.T())
			}
		})
	}
}

func (s *SuiteGK) TestPositive() {
	s.Run("test register", func() {
		s.hmo.On("GetHash", s.testLogin).Return(s.testLoginHash, nil)
		s.hmo.On("GenerateSalt").Return(s.testSalt, nil)
		s.hmo.On("GetHashWithSalt", s.testPassword, s.testSalt).Return(s.testHash, nil)

		s.smo.On("CreateUser", s.testLogin, s.testLoginHash, s.testSalt, s.testHash).
			Return(s.wantUser, nil)

		s.tmo.On("GetToken", mock.Anything).Return(s.testToken, nil)

		res, err := s.client.Register(context.Background(), &proto.RegisterRequest{
			Login:    s.testLogin,
			Password: s.testPassword,
		})

		s.Require().NoError(err)
		s.Equal(s.testToken, res.GetToken())
	})

	s.Run("test auth", func() {
		s.hmo.On("GetHash", s.testLogin).Return(s.testLoginHash, nil)
		s.hmo.On("GetHashWithSalt", s.testPassword, s.testSalt).Return(s.testHash, nil)

		s.smo.On("GetUser", s.testLogin, s.testLoginHash).
			Return(s.wantUser, nil)

		s.tmo.On("GetToken", mock.Anything).Return(s.testToken, nil)

		res, err := s.client.Auth(context.Background(), &proto.AuthRequest{
			Login:    s.testLogin,
			Password: s.testPassword,
		})

		s.Require().NoError(err)
		s.Equal(s.testToken, res.GetToken())
	})

	s.Run("test create password", func() {
		s.smo.On(
			"CreatePassword",
			s.testUserID,
			s.testName,
			s.testLogin,
			s.testPassword,
			s.testMeta,
		).Return(s.wantPassword, nil)

		res, err := s.client.CreatePassword(context.Background(), &proto.CreatePasswordRequest{
			UserID:   s.testUserID,
			Name:     s.testName,
			Login:    s.testLogin,
			Password: s.testPassword,
			Meta:     s.testMeta,
		})
		s.Require().NoError(err)
		s.Equal(s.testPasswordID, res.GetId())
	})

	s.Run("test get password", func() {
		s.smo.On("GetPassword", s.testPasswordID).Return(s.wantPassword, nil)

		res, err := s.client.GetPassword(context.Background(), &proto.GetPasswordRequest{
			Id: s.testPasswordID,
		})
		s.Require().NoError(err)
		s.Equal(s.wantProtoPassword, res.Password)
	})

	s.Run("test get passwords", func() {
		s.smo.On("GetAllPassword", s.testUserID).
			Return([]storage.Password{*s.wantPassword, *s.wantPassword}, nil)

		res, err := s.client.GetPasswords(context.Background(), &proto.GetPasswordsRequest{
			UserID: s.testUserID,
		})
		s.Require().NoError(err)
		s.Equal(
			[]*proto.Password{
				s.wantProtoPassword,
				s.wantProtoPassword,
			}, res.Passwords)
	})

	s.Run("test create bank", func() {
		s.smo.On(
			"CreateBank",
			s.testUserID,
			s.testName,
			s.testBankData,
			s.testMeta,
		).
			Return(s.wantBank, nil)

		res, err := s.client.CreateBank(context.Background(), &proto.CreateBankRequest{
			UserID:    s.testUserID,
			Name:      s.testName,
			BanksData: s.testBankData,
			Meta:      s.testMeta,
		})
		s.Require().NoError(err)
		s.Equal(s.testBankID, res.GetId())
	})

	s.Run("test get bank", func() {
		s.smo.On("GetBank", s.testBankID).Return(s.wantBank, nil)

		res, err := s.client.GetBank(context.Background(), &proto.GetBankRequest{
			Id: s.testBankID,
		})
		s.Require().NoError(err)
		s.Equal(s.wantProtoBank, res.Bank)
	})

	s.Run("test get banks", func() {
		s.smo.On("GetAllBanks", s.testUserID).
			Return([]storage.Bank{*s.wantBank, *s.wantBank}, nil)

		res, err := s.client.GetBanks(context.Background(), &proto.GetBanksRequest{
			UserID: s.testUserID,
		})
		s.Require().NoError(err)
		s.Equal(
			[]*proto.Bank{
				s.wantProtoBank,
				s.wantProtoBank,
			}, res.Banks)
	})

	s.Run("test create text", func() {
		s.smo.On(
			"CreateText",
			s.testUserID,
			s.testName,
			s.testText,
			s.testMeta,
		).Return(s.wantText, nil)

		res, err := s.client.CreateText(context.Background(), &proto.CreateTextRequest{
			UserID: s.testUserID,
			Name:   s.testName,
			Text:   s.testText,
			Meta:   s.testMeta,
		})
		s.Require().NoError(err)
		s.Equal(s.testTextID, res.GetId())
	})

	s.Run("test get text", func() {
		s.smo.On("GetText", s.testTextID).Return(s.wantText, nil)

		res, err := s.client.GetText(context.Background(), &proto.GetTextRequest{
			Id: s.testTextID,
		})
		s.Require().NoError(err)
		s.Equal(s.wantProtoText, res.Text)
	})

	s.Run("test get texts", func() {
		s.smo.On("GetAllTexts", s.testUserID).
			Return([]storage.Text{*s.wantText, *s.wantText}, nil)

		res, err := s.client.GetTexts(context.Background(), &proto.GetTextsRequest{
			UserID: s.testUserID,
		})
		s.Require().NoError(err)
		s.Equal(
			[]*proto.Text{
				s.wantProtoText,
				s.wantProtoText,
			}, res.Texts)
	})

	s.Run("test create file", func() {
		s.fsmo.On("Save", s.testBuffer).Return(s.testPathToFile, nil)

		s.smo.On("CreateFile", s.testUserID, s.testName, s.testPathToFile, s.testMeta).
			Return(s.wantFile, nil)

		stream, err := s.client.CreateFile(context.Background())
		s.Require().NoError(err)

		err = stream.Send(&proto.CreateFileRequest{
			Data: &proto.CreateFileRequest_FileInfo{
				FileInfo: &proto.File{
					UserID: s.testUserID,
					Name:   s.testName,
					Meta:   s.testMeta,
				},
			},
		})
		s.Require().NoError(err)

		err = stream.Send(&proto.CreateFileRequest{
			Data: &proto.CreateFileRequest_Content{
				Content: s.testBatch1,
			},
		})
		s.Require().NoError(err)

		err = stream.Send(&proto.CreateFileRequest{
			Data: &proto.CreateFileRequest_Content{
				Content: s.testBatch2,
			},
		})
		s.Require().NoError(err)

		res, err := stream.CloseAndRecv()
		s.Require().NoError(err)
		s.Equal(s.testFileID, res.GetId())
	})

	s.Run("test get file", func() {
		s.smo.On("GetFile", s.testFileID).Return(s.wantFile, nil)

		fmo := new(DBFilerMocketObject)
		s.fsmo.On("GetDBFiler", s.testPathToFile).Return(fmo, nil)

		fmo.On("GetChunck").Return(s.testBatch1, nil).Once()
		fmo.On("GetChunck").Return(s.testBatch2, nil).Once()
		fmo.On("GetChunck").Return(nil, io.EOF).Once()
		fmo.On("Close")

		stream, err := s.client.GetFile(context.Background(), &proto.GetFileRequest{
			Id: s.testFileID,
		})
		s.Require().NoError(err)

		fi, err := stream.Recv()
		s.Require().NoError(err)
		s.Equal(s.wantProtoFile, fi.GetFileInfo())

		content, err := stream.Recv()
		s.Require().NoError(err)
		s.Equal(s.testBatch1, content.GetContent())

		content, err = stream.Recv()
		s.Require().NoError(err)
		s.Equal(s.testBatch2, content.GetContent())

		fmo.AssertExpectations(s.T())
	})

	s.Run("test get files", func() {
		s.smo.On("GetAllFiles", s.testUserID).Return(
			[]storage.File{
				*s.wantFile,
				*s.wantFile,
			}, nil)

		fmo := new(DBFilerMocketObject)
		s.fsmo.On("GetDBFiler", s.testPathToFile).Return(fmo, nil).Once()
		s.fsmo.On("GetDBFiler", s.testPathToFile).Return(fmo, nil).Once()

		fmo.On("GetChunck").Return(s.testBatch1, nil).Once()
		fmo.On("GetChunck").Return(s.testBatch2, nil).Once()
		fmo.On("GetChunck").Return(nil, io.EOF).Once()
		fmo.On("Close").Once()

		fmo.On("GetChunck").Return(s.testBatch1, nil).Once()
		fmo.On("GetChunck").Return(s.testBatch2, nil).Once()
		fmo.On("GetChunck").Return(nil, io.EOF).Once()
		fmo.On("Close").Once()

		stream, err := s.client.GetFiles(context.Background(), &proto.GetFilesRequest{
			UserID: s.testUserID,
		})
		s.Require().NoError(err)

		fi, err := stream.Recv()
		s.Require().NoError(err)
		s.Equal(s.wantProtoFile, fi.GetFileInfo())

		content, err := stream.Recv()
		s.Require().NoError(err)
		s.Equal(s.testBatch1, content.GetContent())

		content, err = stream.Recv()
		s.Require().NoError(err)
		s.Equal(s.testBatch2, content.GetContent())

		fi, err = stream.Recv()
		s.Require().NoError(err)
		s.Equal(s.wantProtoFile, fi.GetFileInfo())

		content, err = stream.Recv()
		s.Require().NoError(err)
		s.Equal(s.testBatch1, content.GetContent())

		content, err = stream.Recv()
		s.Require().NoError(err)
		s.Equal(s.testBatch2, content.GetContent())

		fmo.AssertExpectations(s.T())
	})
}

func TestSuiteGK(t *testing.T) {
	suiteGK := new(SuiteGK)
	suite.Run(t, suiteGK)
}
