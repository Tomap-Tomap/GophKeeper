//go:build unit

package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	proto "github.com/Tomap-Tomap/GophKeeper/proto/gophkeeper/v1"
	"github.com/Tomap-Tomap/GophKeeper/storage"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const emptyString = ""

var testError = errors.New("test")

type HandlersTestSuite struct {
	suite.Suite
	handler              *GophKeeperHandler
	storageMock          *StorageMockedObject
	hasherMock           *HasherMockedObject
	tokenerMock          *TokenerMockedObject
	fileStoreMock        *FileStoreMockedObject
	streamCreateFileMock *GophKeeper_CreateFileServerMockedObject
	streamUpdateFileMock *GophKeeper_UpdateFileServerMockedObject
	streamGetFileMock    *GophKeeper_GetFileServerMockedObject

	testIncomingContext context.Context

	testUpdateAt time.Time

	testLogin          string
	testPassword       string
	testHashedLogin    string
	testSalt           string
	testHashedPassword string
	testToken          string
	testName           string
	testMeta           string
	testUserID         string
	testPasswordID     string
	testCardNumber     string
	testCvc            string
	testOwner          string
	testExp            string
	testBankID         string
	testText           string
	testTextID         string
	testFileID         string

	testBatch1 []byte
	testBatch2 []byte

	testSaltLength int
}

func (suite *HandlersTestSuite) SetupTest() {
	suite.storageMock = new(StorageMockedObject)
	suite.hasherMock = new(HasherMockedObject)
	suite.tokenerMock = new(TokenerMockedObject)
	suite.fileStoreMock = new(FileStoreMockedObject)
	suite.streamCreateFileMock = new(GophKeeper_CreateFileServerMockedObject)
	suite.streamUpdateFileMock = new(GophKeeper_UpdateFileServerMockedObject)
	suite.streamGetFileMock = new(GophKeeper_GetFileServerMockedObject)

	suite.handler = NewGophKeeperHandler(
		suite.storageMock,
		suite.hasherMock,
		suite.tokenerMock,
		suite.fileStoreMock,
		*storage.NewRetryPolicy(3, 5, 3),
		75,
	)

	suite.testUpdateAt = time.Now()

	suite.testLogin = "testLogin"
	suite.testPassword = "testPassword"
	suite.testHashedLogin = "testHashedLogin"
	suite.testSalt = "testSalt"
	suite.testHashedPassword = "testHashedPassword"
	suite.testToken = "testToken"
	suite.testName = "testName"
	suite.testMeta = "testMeta"
	suite.testUserID = "testUserID"
	suite.testPasswordID = "testPasswordID"
	suite.testCardNumber = "testCardNumber"
	suite.testCvc = "testCvc"
	suite.testOwner = "testOwner"
	suite.testExp = "testExp"
	suite.testBankID = "testBankID"
	suite.testText = "testText"
	suite.testTextID = "testTextID"
	suite.testFileID = "testFileID"

	suite.testBatch1 = []byte{1, 2, 3, 4}
	suite.testBatch2 = []byte{5, 6, 7, 8}

	suite.testIncomingContext = metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(userIDHeader, suite.testUserID),
	)

	suite.testSaltLength = 75
}

func (suite *HandlersTestSuite) TearDownSubTest() {
	suite.storageMock.AssertExpectations(suite.T())
	suite.hasherMock.AssertExpectations(suite.T())
	suite.tokenerMock.AssertExpectations(suite.T())
	suite.fileStoreMock.AssertExpectations(suite.T())
	suite.streamCreateFileMock.AssertExpectations(suite.T())
	suite.streamUpdateFileMock.AssertExpectations(suite.T())
	suite.streamGetFileMock.AssertExpectations(suite.T())

	for len(suite.storageMock.ExpectedCalls) != 0 {
		suite.storageMock.ExpectedCalls[0].Unset()
	}

	for len(suite.hasherMock.ExpectedCalls) != 0 {
		suite.hasherMock.ExpectedCalls[0].Unset()
	}

	for len(suite.tokenerMock.ExpectedCalls) != 0 {
		suite.tokenerMock.ExpectedCalls[0].Unset()
	}

	for len(suite.fileStoreMock.ExpectedCalls) != 0 {
		suite.fileStoreMock.ExpectedCalls[0].Unset()
	}

	for len(suite.streamCreateFileMock.ExpectedCalls) != 0 {
		suite.streamCreateFileMock.ExpectedCalls[0].Unset()
	}

	for len(suite.streamUpdateFileMock.ExpectedCalls) != 0 {
		suite.streamUpdateFileMock.ExpectedCalls[0].Unset()
	}

	for len(suite.streamGetFileMock.ExpectedCalls) != 0 {
		suite.streamGetFileMock.ExpectedCalls[0].Unset()
	}
}

func (suite *HandlersTestSuite) TestRegister() {
	require := suite.Require()

	positiveReq := &proto.RegisterRequest{
		Login:    suite.testLogin,
		Password: suite.testPassword,
	}

	suite.Run("generate salt error", func() {
		suite.hasherMock.onGenerateHash(suite.testLogin, suite.testHashedLogin)
		suite.hasherMock.onGenerateSalt(suite.testSaltLength, emptyString, testError)

		res, err := suite.handler.Register(context.Background(), positiveReq)
		require.ErrorContains(err, "generate salt")
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("generate hash error", func() {
		suite.hasherMock.onGenerateHash(suite.testLogin, suite.testHashedLogin)
		suite.hasherMock.onGenerateSalt(suite.testSaltLength, suite.testSalt, nil)
		suite.hasherMock.onGenerateHashWithSalt(suite.testPassword, suite.testSalt, emptyString, testError)

		res, err := suite.handler.Register(context.Background(), positiveReq)
		require.ErrorContains(err, "generate hash")
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("user already exists error", func() {
		suite.hasherMock.onGenerateHash(suite.testLogin, suite.testHashedLogin)
		suite.hasherMock.onGenerateSalt(suite.testSaltLength, suite.testSalt, nil)
		suite.hasherMock.onGenerateHashWithSalt(suite.testPassword, suite.testSalt, suite.testHashedPassword, nil)
		suite.storageMock.onCreateUser(suite.testLogin, suite.testHashedLogin, suite.testSalt, suite.testHashedPassword, nil, storage.ErrUserAlreadyExists)

		res, err := suite.handler.Register(context.Background(), positiveReq)
		require.ErrorContains(err, fmt.Sprintf("user %s already exists", suite.testLogin))
		require.Equal(status.Code(err), codes.AlreadyExists)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.hasherMock.onGenerateHash(suite.testLogin, suite.testHashedLogin)
		suite.hasherMock.onGenerateSalt(suite.testSaltLength, suite.testSalt, nil)
		suite.hasherMock.onGenerateHashWithSalt(suite.testPassword, suite.testSalt, suite.testHashedPassword, nil)
		suite.storageMock.onCreateUser(suite.testLogin, suite.testHashedLogin, suite.testSalt, suite.testHashedPassword, nil, testError)

		res, err := suite.handler.Register(context.Background(), positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("get suite.testToken error", func() {
		suite.hasherMock.onGenerateHash(suite.testLogin, suite.testHashedLogin)
		suite.hasherMock.onGenerateSalt(suite.testSaltLength, suite.testSalt, nil)
		suite.hasherMock.onGenerateHashWithSalt(suite.testPassword, suite.testSalt, suite.testHashedPassword, nil)
		suite.storageMock.onCreateUser(suite.testLogin, suite.testHashedLogin, suite.testSalt, suite.testHashedPassword, &storage.User{}, nil)
		suite.tokenerMock.onGetToken(mock.Anything, emptyString, testError)

		res, err := suite.handler.Register(context.Background(), positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		suite.hasherMock.onGenerateHash(suite.testLogin, suite.testHashedLogin)
		suite.hasherMock.onGenerateSalt(suite.testSaltLength, suite.testSalt, nil)
		suite.hasherMock.onGenerateHashWithSalt(suite.testPassword, suite.testSalt, suite.testHashedPassword, nil)
		suite.storageMock.onCreateUser(suite.testLogin, suite.testHashedLogin, suite.testSalt, suite.testHashedPassword, &storage.User{}, nil)
		suite.tokenerMock.onGetToken(mock.Anything, suite.testToken, nil)

		res, err := suite.handler.Register(context.Background(), positiveReq)
		suite.Require().NoError(err)
		suite.Require().Equal(suite.testToken, res.Token)
	})
}

func (suite *HandlersTestSuite) TestAuth() {
	require := suite.Require()

	positiveReq := &proto.AuthRequest{
		Login:    suite.testLogin,
		Password: suite.testPassword,
	}

	suite.Run("database error", func() {
		suite.hasherMock.onGenerateHash(suite.testLogin, suite.testHashedLogin)
		suite.storageMock.onGetUser(suite.testLogin, suite.testHashedLogin, nil, testError)

		res, err := suite.handler.Auth(context.Background(), positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("user not found", func() {
		suite.hasherMock.onGenerateHash(suite.testLogin, suite.testHashedLogin)
		suite.storageMock.onGetUser(suite.testLogin, suite.testHashedLogin, nil, storage.ErrUserNotFound)

		res, err := suite.handler.Auth(context.Background(), positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown user %s", suite.testLogin))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("hash error", func() {
		suite.hasherMock.onGenerateHash(suite.testLogin, suite.testHashedLogin)
		suite.storageMock.onGetUser(suite.testLogin, suite.testHashedLogin, &storage.User{Salt: suite.testSalt}, nil)
		suite.hasherMock.onGenerateHashWithSalt(suite.testPassword, suite.testSalt, emptyString, testError)

		res, err := suite.handler.Auth(context.Background(), positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("invalid password", func() {
		invalidHash := "invalidHash"
		suite.hasherMock.onGenerateHash(suite.testLogin, suite.testHashedLogin)
		suite.storageMock.onGetUser(suite.testLogin, suite.testHashedLogin, &storage.User{Salt: suite.testSalt, Password: suite.testHashedPassword}, nil)
		suite.hasherMock.onGenerateHashWithSalt(suite.testPassword, suite.testSalt, invalidHash, nil)

		res, err := suite.handler.Auth(context.Background(), positiveReq)
		require.ErrorContains(err, "invalid password")
		require.Equal(status.Code(err), codes.PermissionDenied)
		require.Nil(res)
	})

	suite.Run("get suite.testToken error", func() {
		suite.hasherMock.onGenerateHash(suite.testLogin, suite.testHashedLogin)
		suite.storageMock.onGetUser(suite.testLogin, suite.testHashedLogin, &storage.User{Salt: suite.testSalt, Password: suite.testHashedPassword}, nil)
		suite.hasherMock.onGenerateHashWithSalt(suite.testPassword, suite.testSalt, suite.testHashedPassword, nil)
		suite.tokenerMock.onGetToken(mock.Anything, emptyString, testError)

		res, err := suite.handler.Auth(context.Background(), positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		suite.hasherMock.onGenerateHash(suite.testLogin, suite.testHashedLogin)
		suite.storageMock.onGetUser(suite.testLogin, suite.testHashedLogin, &storage.User{Salt: suite.testSalt, Password: suite.testHashedPassword}, nil)
		suite.hasherMock.onGenerateHashWithSalt(suite.testPassword, suite.testSalt, suite.testHashedPassword, nil)
		suite.tokenerMock.onGetToken(mock.Anything, suite.testToken, nil)

		res, err := suite.handler.Auth(context.Background(), positiveReq)
		suite.Require().NoError(err)
		suite.Require().Equal(suite.testToken, res.Token)
	})
}

func (suite *HandlersTestSuite) TestCreatePassword() {
	require := suite.Require()

	positiveReq := &proto.CreatePasswordRequest{
		Name:     suite.testName,
		Login:    suite.testLogin,
		Password: suite.testPassword,
		Meta:     suite.testMeta,
	}

	suite.Run("unauthenticated", func() {
		req := &proto.CreatePasswordRequest{
			Name:     "",
			Login:    "",
			Password: "",
			Meta:     "",
		}

		res, err := suite.handler.CreatePassword(context.Background(), req)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onCreatePassword(suite.testUserID, suite.testName, suite.testLogin, suite.testPassword, suite.testMeta, nil, testError)

		res, err := suite.handler.CreatePassword(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown UserID error", func() {
		suite.storageMock.onCreatePassword(suite.testUserID, suite.testName, suite.testLogin, suite.testPassword, suite.testMeta, nil, storage.ErrUserNotFound)

		res, err := suite.handler.CreatePassword(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown UserID %s", suite.testUserID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		suite.storageMock.onCreatePassword(suite.testUserID, suite.testName, suite.testLogin, suite.testPassword, suite.testMeta, &storage.Password{ID: suite.testPasswordID}, nil)

		res, err := suite.handler.CreatePassword(suite.testIncomingContext, positiveReq)
		suite.Require().NoError(err)
		suite.Require().Equal(suite.testPasswordID, res.GetId())
	})
}

func (suite *HandlersTestSuite) TestUpdatePassword() {
	require := suite.Require()

	positiveReq := &proto.UpdatePasswordRequest{
		Id:       suite.testPasswordID,
		Name:     suite.testName,
		Login:    suite.testLogin,
		Password: suite.testPassword,
		Meta:     suite.testMeta,
	}

	suite.Run("unauthenticated", func() {
		req := &proto.UpdatePasswordRequest{
			Id:       "",
			Name:     "",
			Login:    "",
			Password: "",
			Meta:     "",
		}

		res, err := suite.handler.UpdatePassword(context.Background(), req)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onUpdatePassword(suite.testPasswordID, suite.testUserID, suite.testName, suite.testLogin, suite.testPassword, suite.testMeta, nil, testError)

		res, err := suite.handler.UpdatePassword(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown UserID error", func() {
		suite.storageMock.onUpdatePassword(suite.testPasswordID, suite.testUserID, suite.testName, suite.testLogin, suite.testPassword, suite.testMeta, nil, storage.ErrUserNotFound)

		res, err := suite.handler.UpdatePassword(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown UserID %s", suite.testUserID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("unknown PasswordID error", func() {
		suite.storageMock.onUpdatePassword(suite.testPasswordID, suite.testUserID, suite.testName, suite.testLogin, suite.testPassword, suite.testMeta, nil, storage.ErrPasswordNotFound)

		res, err := suite.handler.UpdatePassword(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown PasswordID %s", suite.testPasswordID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		suite.storageMock.onUpdatePassword(suite.testPasswordID, suite.testUserID, suite.testName, suite.testLogin, suite.testPassword, suite.testMeta, &storage.Password{ID: suite.testPasswordID}, nil)

		res, err := suite.handler.UpdatePassword(suite.testIncomingContext, positiveReq)
		suite.Require().NoError(err)
		suite.Require().Equal(suite.testPasswordID, res.GetId())
	})
}

func (suite *HandlersTestSuite) TestGetPassword() {
	require := suite.Require()

	positiveReq := &proto.GetPasswordRequest{
		Id: suite.testPasswordID,
	}

	suite.Run("unauthenticated", func() {
		res, err := suite.handler.GetPassword(context.Background(), positiveReq)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("empty PasswordID", func() {
		req := &proto.GetPasswordRequest{
			Id: "",
		}
		res, err := suite.handler.GetPassword(suite.testIncomingContext, req)
		require.ErrorContains(err, "empty PasswordID")
		require.Equal(status.Code(err), codes.InvalidArgument)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onGetPassword(suite.testPasswordID, suite.testUserID, nil, testError)

		res, err := suite.handler.GetPassword(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown PasswordID error", func() {
		suite.storageMock.onGetPassword(suite.testPasswordID, suite.testUserID, nil, storage.ErrPasswordNotFound)

		res, err := suite.handler.GetPassword(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown PasswordID %s", suite.testPasswordID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		suite.storageMock.onGetPassword(suite.testPasswordID, suite.testUserID, &storage.Password{
			ID:       suite.testPasswordID,
			Name:     suite.testName,
			Login:    suite.testLogin,
			Password: suite.testPassword,
			Meta:     suite.testMeta,
			UpdateAt: suite.testUpdateAt,
		}, nil)

		res, err := suite.handler.GetPassword(suite.testIncomingContext, positiveReq)
		suite.Require().NoError(err)
		suite.Require().Equal(&proto.GetPasswordResponse{
			Password: &proto.Password{
				Id:       suite.testPasswordID,
				Name:     suite.testName,
				Login:    suite.testLogin,
				Password: suite.testPassword,
				Meta:     suite.testMeta,
				UpdateAt: timestamppb.New(suite.testUpdateAt),
			},
		}, res)
	})
}

func (suite *HandlersTestSuite) TestGetPasswords() {
	require := suite.Require()

	suite.Run("unauthenticated", func() {
		res, err := suite.handler.GetPasswords(context.Background(), &proto.GetPasswordsRequest{})
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onGetAllPassword(suite.testUserID, nil, testError)

		res, err := suite.handler.GetPasswords(suite.testIncomingContext, &proto.GetPasswordsRequest{})
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown UserID error", func() {
		suite.storageMock.onGetAllPassword(suite.testUserID, nil, storage.ErrUserNotFound)

		res, err := suite.handler.GetPasswords(suite.testIncomingContext, &proto.GetPasswordsRequest{})
		require.ErrorContains(err, fmt.Sprintf("unknown UserID %s", suite.testUserID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		passwords := []storage.Password{
			{
				ID:       suite.testPasswordID,
				Name:     suite.testName,
				Login:    suite.testLogin,
				Password: suite.testPassword,
				Meta:     suite.testMeta,
				UpdateAt: suite.testUpdateAt,
			},
			{
				ID:       "anotherPasswordID",
				Name:     "anotherName",
				Login:    "anotherLogin",
				Password: "anotherPassword",
				Meta:     "anotherMeta",
				UpdateAt: suite.testUpdateAt,
			},
		}
		suite.storageMock.onGetAllPassword(suite.testUserID, passwords, nil)

		res, err := suite.handler.GetPasswords(suite.testIncomingContext, &proto.GetPasswordsRequest{})
		suite.Require().NoError(err)
		suite.Require().Equal(&proto.GetPasswordsResponse{
			Passwords: []*proto.Password{
				{
					Id:       suite.testPasswordID,
					Name:     suite.testName,
					Login:    suite.testLogin,
					Password: suite.testPassword,
					Meta:     suite.testMeta,
					UpdateAt: timestamppb.New(suite.testUpdateAt),
				},
				{
					Id:       "anotherPasswordID",
					Name:     "anotherName",
					Login:    "anotherLogin",
					Password: "anotherPassword",
					Meta:     "anotherMeta",
					UpdateAt: timestamppb.New(suite.testUpdateAt),
				},
			},
		}, res)
	})
}

func (suite *HandlersTestSuite) TestDeletePassword() {
	require := suite.Require()

	positiveReq := &proto.DeletePasswordRequest{
		Id: suite.testPasswordID,
	}

	suite.Run("unauthenticated", func() {
		req := &proto.DeletePasswordRequest{
			Id: "",
		}

		res, err := suite.handler.DeletePassword(context.Background(), req)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onDeletePassword(suite.testPasswordID, suite.testUserID, testError)

		res, err := suite.handler.DeletePassword(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown PasswordID error", func() {
		suite.storageMock.onDeletePassword(suite.testPasswordID, suite.testUserID, storage.ErrPasswordNotFound)

		res, err := suite.handler.DeletePassword(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown PasswordID %s", suite.testPasswordID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		suite.storageMock.onDeletePassword(suite.testPasswordID, suite.testUserID, nil)

		res, err := suite.handler.DeletePassword(suite.testIncomingContext, positiveReq)
		suite.Require().NoError(err)
		suite.Require().Nil(res)
	})
}

func (suite *HandlersTestSuite) TestCreateBank() {
	require := suite.Require()

	positiveReq := &proto.CreateBankRequest{
		Name:       suite.testName,
		CardNumber: suite.testCardNumber,
		Cvc:        suite.testCvc,
		Owner:      suite.testOwner,
		Exp:        suite.testExp,
		Meta:       suite.testMeta,
	}

	suite.Run("unauthenticated", func() {
		req := &proto.CreateBankRequest{
			Name:       "",
			CardNumber: "",
			Cvc:        "",
			Owner:      "",
			Exp:        "",
			Meta:       "",
		}

		res, err := suite.handler.CreateBank(context.Background(), req)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onCreateBank(suite.testUserID, suite.testName, suite.testCardNumber, suite.testCvc, suite.testOwner, suite.testExp, suite.testMeta, nil, testError)

		res, err := suite.handler.CreateBank(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown UserID error", func() {
		suite.storageMock.onCreateBank(suite.testUserID, suite.testName, suite.testCardNumber, suite.testCvc, suite.testOwner, suite.testExp, suite.testMeta, nil, storage.ErrUserNotFound)

		res, err := suite.handler.CreateBank(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown UserID %s", suite.testUserID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		suite.storageMock.onCreateBank(suite.testUserID, suite.testName, suite.testCardNumber, suite.testCvc, suite.testOwner, suite.testExp, suite.testMeta, &storage.Bank{ID: suite.testBankID}, nil)

		res, err := suite.handler.CreateBank(suite.testIncomingContext, positiveReq)
		suite.Require().NoError(err)
		suite.Require().Equal(suite.testBankID, res.GetId())
	})
}

func (suite *HandlersTestSuite) TestUpdateBank() {
	require := suite.Require()

	positiveReq := &proto.UpdateBankRequest{
		Id:         suite.testBankID,
		Name:       suite.testName,
		CardNumber: suite.testCardNumber,
		Cvc:        suite.testCvc,
		Owner:      suite.testOwner,
		Exp:        suite.testExp,
		Meta:       suite.testMeta,
	}

	suite.Run("unauthenticated", func() {
		req := &proto.UpdateBankRequest{
			Id:         "",
			Name:       "",
			CardNumber: "",
			Cvc:        "",
			Owner:      "",
			Exp:        "",
			Meta:       "",
		}

		res, err := suite.handler.UpdateBank(context.Background(), req)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onUpdateBank(suite.testBankID, suite.testUserID, suite.testName, suite.testCardNumber, suite.testCvc, suite.testOwner, suite.testExp, suite.testMeta, nil, testError)

		res, err := suite.handler.UpdateBank(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown UserID error", func() {
		suite.storageMock.onUpdateBank(suite.testBankID, suite.testUserID, suite.testName, suite.testCardNumber, suite.testCvc, suite.testOwner, suite.testExp, suite.testMeta, nil, storage.ErrUserNotFound)

		res, err := suite.handler.UpdateBank(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown UserID %s", suite.testUserID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("unknown BankID error", func() {
		suite.storageMock.onUpdateBank(suite.testBankID, suite.testUserID, suite.testName, suite.testCardNumber, suite.testCvc, suite.testOwner, suite.testExp, suite.testMeta, nil, storage.ErrBankNotFound)

		res, err := suite.handler.UpdateBank(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown BankID %s", suite.testBankID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		suite.storageMock.onUpdateBank(suite.testBankID, suite.testUserID, suite.testName, suite.testCardNumber, suite.testCvc, suite.testOwner, suite.testExp, suite.testMeta, &storage.Bank{ID: suite.testBankID}, nil)

		res, err := suite.handler.UpdateBank(suite.testIncomingContext, positiveReq)
		suite.Require().NoError(err)
		suite.Require().Equal(suite.testBankID, res.GetId())
	})
}

func (suite *HandlersTestSuite) TestGetBank() {
	require := suite.Require()

	positiveReq := &proto.GetBankRequest{
		Id: suite.testBankID,
	}

	suite.Run("unauthenticated", func() {
		res, err := suite.handler.GetBank(context.Background(), positiveReq)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("empty BankID", func() {
		req := &proto.GetBankRequest{
			Id: "",
		}

		res, err := suite.handler.GetBank(suite.testIncomingContext, req)
		require.ErrorContains(err, "empty BankID")
		require.Equal(status.Code(err), codes.InvalidArgument)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onGetBank(suite.testBankID, suite.testUserID, nil, testError)

		res, err := suite.handler.GetBank(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown BankID error", func() {
		suite.storageMock.onGetBank(suite.testBankID, suite.testUserID, nil, storage.ErrBankNotFound)

		res, err := suite.handler.GetBank(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown BankID %s", suite.testBankID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		suite.storageMock.onGetBank(suite.testBankID, suite.testUserID, &storage.Bank{
			ID:         suite.testBankID,
			Name:       suite.testName,
			CardNumber: suite.testCardNumber,
			CVC:        suite.testCvc,
			Owner:      suite.testOwner,
			Exp:        suite.testExp,
			Meta:       suite.testMeta,
			UpdateAt:   suite.testUpdateAt,
		}, nil)

		res, err := suite.handler.GetBank(suite.testIncomingContext, positiveReq)
		suite.Require().NoError(err)
		suite.Require().Equal(&proto.GetBankResponse{
			Bank: &proto.Bank{
				Id:         suite.testBankID,
				Name:       suite.testName,
				CardNumber: suite.testCardNumber,
				Cvc:        suite.testCvc,
				Owner:      suite.testOwner,
				Exp:        suite.testExp,
				Meta:       suite.testMeta,
				UpdateAt:   timestamppb.New(suite.testUpdateAt),
			},
		}, res)
	})
}

func (suite *HandlersTestSuite) TestGetBanks() {
	require := suite.Require()

	suite.Run("unauthenticated", func() {
		res, err := suite.handler.GetBanks(context.Background(), &proto.GetBanksRequest{})
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onGetAllBanks(suite.testUserID, nil, testError)

		res, err := suite.handler.GetBanks(suite.testIncomingContext, &proto.GetBanksRequest{})
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown UserID error", func() {
		suite.storageMock.onGetAllBanks(suite.testUserID, nil, storage.ErrUserNotFound)

		res, err := suite.handler.GetBanks(suite.testIncomingContext, &proto.GetBanksRequest{})
		require.ErrorContains(err, fmt.Sprintf("unknown UserID %s", suite.testUserID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		banks := []storage.Bank{
			{
				ID:         suite.testBankID,
				Name:       suite.testName,
				CardNumber: suite.testCardNumber,
				CVC:        suite.testCvc,
				Owner:      suite.testOwner,
				Exp:        suite.testExp,
				Meta:       suite.testMeta,
				UpdateAt:   suite.testUpdateAt,
			},
			{
				ID:         "anotherBankID",
				Name:       "anotherName",
				CardNumber: "anotherCardNumber",
				CVC:        "anotherCvc",
				Owner:      "anotherOwner",
				Exp:        "anotherExp",
				Meta:       "anotherMeta",
				UpdateAt:   suite.testUpdateAt,
			},
		}
		suite.storageMock.onGetAllBanks(suite.testUserID, banks, nil)

		res, err := suite.handler.GetBanks(suite.testIncomingContext, &proto.GetBanksRequest{})
		suite.Require().NoError(err)
		suite.Require().Equal(&proto.GetBanksResponse{
			Banks: []*proto.Bank{
				{
					Id:         suite.testBankID,
					Name:       suite.testName,
					CardNumber: suite.testCardNumber,
					Cvc:        suite.testCvc,
					Owner:      suite.testOwner,
					Exp:        suite.testExp,
					Meta:       suite.testMeta,
					UpdateAt:   timestamppb.New(suite.testUpdateAt),
				},
				{
					Id:         "anotherBankID",
					Name:       "anotherName",
					CardNumber: "anotherCardNumber",
					Cvc:        "anotherCvc",
					Owner:      "anotherOwner",
					Exp:        "anotherExp",
					Meta:       "anotherMeta",
					UpdateAt:   timestamppb.New(suite.testUpdateAt),
				},
			},
		}, res)
	})
}

func (suite *HandlersTestSuite) TestDeleteBank() {
	require := suite.Require()

	positiveReq := &proto.DeleteBankRequest{
		Id: suite.testBankID,
	}

	suite.Run("unauthenticated", func() {
		req := &proto.DeleteBankRequest{
			Id: "",
		}

		res, err := suite.handler.DeleteBank(context.Background(), req)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onDeleteBank(suite.testBankID, suite.testUserID, testError)

		res, err := suite.handler.DeleteBank(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown BankID error", func() {
		suite.storageMock.onDeleteBank(suite.testBankID, suite.testUserID, storage.ErrBankNotFound)

		res, err := suite.handler.DeleteBank(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown BankID %s", suite.testBankID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		suite.storageMock.onDeleteBank(suite.testBankID, suite.testUserID, nil)

		res, err := suite.handler.DeleteBank(suite.testIncomingContext, positiveReq)
		suite.Require().NoError(err)
		suite.Require().Nil(res)
	})
}

func (suite *HandlersTestSuite) TestCreateText() {
	require := suite.Require()

	positiveReq := &proto.CreateTextRequest{
		Name: suite.testName,
		Text: suite.testText,
		Meta: suite.testMeta,
	}

	suite.Run("unauthenticated", func() {
		req := &proto.CreateTextRequest{
			Name: "",
			Text: "",
			Meta: "",
		}

		res, err := suite.handler.CreateText(context.Background(), req)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onCreateText(suite.testUserID, suite.testName, suite.testText, suite.testMeta, nil, testError)

		res, err := suite.handler.CreateText(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown UserID error", func() {
		suite.storageMock.onCreateText(suite.testUserID, suite.testName, suite.testText, suite.testMeta, nil, storage.ErrUserNotFound)

		res, err := suite.handler.CreateText(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown UserID %s", suite.testUserID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		suite.storageMock.onCreateText(suite.testUserID, suite.testName, suite.testText, suite.testMeta, &storage.Text{ID: suite.testTextID}, nil)

		res, err := suite.handler.CreateText(suite.testIncomingContext, positiveReq)
		suite.Require().NoError(err)
		suite.Require().Equal(suite.testTextID, res.GetId())
	})
}

func (suite *HandlersTestSuite) TestUpdateText() {
	require := suite.Require()

	positiveReq := &proto.UpdateTextRequest{
		Id:   suite.testTextID,
		Name: suite.testName,
		Text: suite.testText,
		Meta: suite.testMeta,
	}

	suite.Run("unauthenticated", func() {
		req := &proto.UpdateTextRequest{
			Id:   "",
			Name: "",
			Text: "",
			Meta: "",
		}

		res, err := suite.handler.UpdateText(context.Background(), req)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onUpdateText(suite.testTextID, suite.testUserID, suite.testName, suite.testText, suite.testMeta, nil, testError)

		res, err := suite.handler.UpdateText(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown UserID error", func() {
		suite.storageMock.onUpdateText(suite.testTextID, suite.testUserID, suite.testName, suite.testText, suite.testMeta, nil, storage.ErrUserNotFound)

		res, err := suite.handler.UpdateText(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown UserID %s", suite.testUserID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("unknown TextID error", func() {
		suite.storageMock.onUpdateText(suite.testTextID, suite.testUserID, suite.testName, suite.testText, suite.testMeta, nil, storage.ErrTextNotFound)

		res, err := suite.handler.UpdateText(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown TextID %s", suite.testTextID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		suite.storageMock.onUpdateText(suite.testTextID, suite.testUserID, suite.testName, suite.testText, suite.testMeta, &storage.Text{ID: suite.testTextID}, nil)

		res, err := suite.handler.UpdateText(suite.testIncomingContext, positiveReq)
		suite.Require().NoError(err)
		suite.Require().Equal(suite.testTextID, res.GetId())
	})
}

func (suite *HandlersTestSuite) TestGetText() {
	require := suite.Require()

	positiveReq := &proto.GetTextRequest{
		Id: suite.testTextID,
	}

	suite.Run("unauthenticated", func() {
		res, err := suite.handler.GetText(context.Background(), positiveReq)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("empty TextID", func() {
		req := &proto.GetTextRequest{
			Id: "",
		}

		res, err := suite.handler.GetText(suite.testIncomingContext, req)
		require.ErrorContains(err, "empty TextID")
		require.Equal(status.Code(err), codes.InvalidArgument)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onGetText(suite.testTextID, suite.testUserID, nil, testError)

		res, err := suite.handler.GetText(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown TextID error", func() {
		suite.storageMock.onGetText(suite.testTextID, suite.testUserID, nil, storage.ErrTextNotFound)

		res, err := suite.handler.GetText(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown TextID %s", suite.testTextID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		suite.storageMock.onGetText(suite.testTextID, suite.testUserID, &storage.Text{
			ID:       suite.testTextID,
			Name:     suite.testName,
			Text:     suite.testText,
			Meta:     suite.testMeta,
			UpdateAt: suite.testUpdateAt,
		}, nil)

		res, err := suite.handler.GetText(suite.testIncomingContext, positiveReq)
		suite.Require().NoError(err)
		suite.Require().Equal(&proto.GetTextResponse{
			Text: &proto.Text{
				Id:       suite.testTextID,
				Name:     suite.testName,
				Text:     suite.testText,
				Meta:     suite.testMeta,
				UpdateAt: timestamppb.New(suite.testUpdateAt),
			},
		}, res)
	})
}

func (suite *HandlersTestSuite) TestGetTexts() {
	require := suite.Require()

	suite.Run("unauthenticated", func() {
		res, err := suite.handler.GetTexts(context.Background(), &proto.GetTextsRequest{})
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onGetAllTexts(suite.testUserID, nil, testError)

		res, err := suite.handler.GetTexts(suite.testIncomingContext, &proto.GetTextsRequest{})
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown UserID error", func() {
		suite.storageMock.onGetAllTexts(suite.testUserID, nil, storage.ErrUserNotFound)

		res, err := suite.handler.GetTexts(suite.testIncomingContext, &proto.GetTextsRequest{})
		require.ErrorContains(err, fmt.Sprintf("unknown UserID %s", suite.testUserID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		texts := []storage.Text{
			{
				ID:       suite.testTextID,
				Name:     suite.testName,
				Text:     suite.testText,
				Meta:     suite.testMeta,
				UpdateAt: suite.testUpdateAt,
			},
			{
				ID:       "anotherTextID",
				Name:     "anotherName",
				Text:     "anotherText",
				Meta:     "anotherMeta",
				UpdateAt: suite.testUpdateAt,
			},
		}
		suite.storageMock.onGetAllTexts(suite.testUserID, texts, nil)

		res, err := suite.handler.GetTexts(suite.testIncomingContext, &proto.GetTextsRequest{})
		suite.Require().NoError(err)
		suite.Require().Equal(&proto.GetTextsResponse{
			Texts: []*proto.Text{
				{
					Id:       suite.testTextID,
					Name:     suite.testName,
					Text:     suite.testText,
					Meta:     suite.testMeta,
					UpdateAt: timestamppb.New(suite.testUpdateAt),
				},
				{
					Id:       "anotherTextID",
					Name:     "anotherName",
					Text:     "anotherText",
					Meta:     "anotherMeta",
					UpdateAt: timestamppb.New(suite.testUpdateAt),
				},
			},
		}, res)
	})
}

func (suite *HandlersTestSuite) TestDeleteText() {
	require := suite.Require()

	positiveReq := &proto.DeleteTextRequest{
		Id: suite.testTextID,
	}

	suite.Run("unauthenticated", func() {
		req := &proto.DeleteTextRequest{
			Id: "",
		}

		res, err := suite.handler.DeleteText(context.Background(), req)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onDeleteText(suite.testTextID, suite.testUserID, testError)

		res, err := suite.handler.DeleteText(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown TextID error", func() {
		suite.storageMock.onDeleteText(suite.testTextID, suite.testUserID, storage.ErrTextNotFound)

		res, err := suite.handler.DeleteText(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown TextID %s", suite.testTextID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		suite.storageMock.onDeleteText(suite.testTextID, suite.testUserID, nil)

		res, err := suite.handler.DeleteText(suite.testIncomingContext, positiveReq)
		suite.Require().NoError(err)
		suite.Require().Nil(res)
	})
}

func (suite *HandlersTestSuite) TestCreateFile() {
	require := suite.Require()

	fileInfo := &proto.File{
		Name: suite.testName,
		Meta: suite.testMeta,
	}
	positiveReq := []*proto.CreateFileRequest{
		{
			Data: &proto.CreateFileRequest_FileInfo{
				FileInfo: fileInfo,
			},
		},
		{
			Data: &proto.CreateFileRequest_Content{
				Content: suite.testBatch1,
			},
		},
		{
			Data: &proto.CreateFileRequest_Content{
				Content: suite.testBatch2,
			},
		},
	}

	suite.Run("unauthenticated", func() {
		suite.streamCreateFileMock.onContext(context.Background())

		err := suite.handler.CreateFile(suite.streamCreateFileMock)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
	})

	suite.Run("cannot receive file info", func() {
		suite.streamCreateFileMock.onContext(suite.testIncomingContext)
		suite.streamCreateFileMock.onRecvWithOnce(nil, testError)

		err := suite.handler.CreateFile(suite.streamCreateFileMock)
		require.ErrorContains(err, "cannot receive file info")
		require.Equal(status.Code(err), codes.Unknown)
	})

	suite.Run("create DB file error", func() {
		suite.streamCreateFileMock.onContext(suite.testIncomingContext)
		suite.streamCreateFileMock.onRecvWithOnce(positiveReq[0], nil)
		suite.fileStoreMock.onCreateDBFile(mock.Anything, nil, testError)

		err := suite.handler.CreateFile(suite.streamCreateFileMock)
		require.Error(err)
		require.Equal(status.Code(err), codes.Internal)
	})

	suite.Run("cannot receive content", func() {
		suite.streamCreateFileMock.onContext(suite.testIncomingContext)
		suite.streamCreateFileMock.onRecvWithOnce(positiveReq[0], nil)
		suite.streamCreateFileMock.onRecvWithOnce(nil, testError)

		dbfmo := new(DBFilerMockedObject)
		dbfmo.onClose(nil)
		defer dbfmo.AssertExpectations(suite.T())

		suite.fileStoreMock.onCreateDBFile(mock.Anything, dbfmo, nil)

		err := suite.handler.CreateFile(suite.streamCreateFileMock)
		require.ErrorContains(err, "cannot receive content")
		require.Equal(status.Code(err), codes.Unknown)
	})

	suite.Run("write chunk error", func() {
		suite.streamCreateFileMock.onContext(suite.testIncomingContext)
		suite.streamCreateFileMock.onRecvWithOnce(positiveReq[0], nil)
		suite.streamCreateFileMock.onRecvWithOnce(positiveReq[1], nil)

		dbfmo := new(DBFilerMockedObject)
		dbfmo.onClose(nil)
		dbfmo.onWriteOnce(suite.testBatch1, 0, testError)
		defer dbfmo.AssertExpectations(suite.T())

		suite.fileStoreMock.onCreateDBFile(mock.Anything, dbfmo, nil)

		err := suite.handler.CreateFile(suite.streamCreateFileMock)
		require.Error(err)
		require.Equal(status.Code(err), codes.Internal)
	})

	suite.Run("database error", func() {
		suite.streamCreateFileMock.onContext(suite.testIncomingContext)
		suite.streamCreateFileMock.onRecvWithOnce(positiveReq[0], nil)
		suite.streamCreateFileMock.onRecvWithOnce(positiveReq[1], nil)
		suite.streamCreateFileMock.onRecvWithOnce(positiveReq[2], nil)
		suite.streamCreateFileMock.onRecvWithOnce(nil, io.EOF)

		dbfmo := new(DBFilerMockedObject)
		dbfmo.onClose(nil)
		dbfmo.onWriteOnce(suite.testBatch1, len(suite.testBatch1), nil)
		dbfmo.onWriteOnce(suite.testBatch2, len(suite.testBatch2), nil)
		defer dbfmo.AssertExpectations(suite.T())

		suite.storageMock.onCreateFile(suite.testUserID, suite.testName, mock.Anything, suite.testMeta, nil, testError)

		suite.fileStoreMock.onCreateDBFile(mock.Anything, dbfmo, nil)

		err := suite.handler.CreateFile(suite.streamCreateFileMock)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
	})

	suite.Run("unknown UserID error", func() {
		suite.streamCreateFileMock.onContext(suite.testIncomingContext)
		suite.streamCreateFileMock.onRecvWithOnce(positiveReq[0], nil)
		suite.streamCreateFileMock.onRecvWithOnce(positiveReq[1], nil)
		suite.streamCreateFileMock.onRecvWithOnce(positiveReq[2], nil)
		suite.streamCreateFileMock.onRecvWithOnce(nil, io.EOF)

		dbfmo := new(DBFilerMockedObject)
		dbfmo.onClose(nil)
		dbfmo.onWriteOnce(suite.testBatch1, len(suite.testBatch1), nil)
		dbfmo.onWriteOnce(suite.testBatch2, len(suite.testBatch2), nil)
		defer dbfmo.AssertExpectations(suite.T())

		suite.storageMock.onCreateFile(suite.testUserID, suite.testName, mock.Anything, suite.testMeta, nil, storage.ErrUserNotFound)

		suite.fileStoreMock.onCreateDBFile(mock.Anything, dbfmo, nil)

		err := suite.handler.CreateFile(suite.streamCreateFileMock)
		require.ErrorContains(err, fmt.Sprintf("unknown UserID %s", suite.testUserID))
		require.Equal(status.Code(err), codes.Unknown)
	})

	suite.Run("positive test", func() {
		suite.streamCreateFileMock.onContext(suite.testIncomingContext)
		suite.streamCreateFileMock.onRecvWithOnce(positiveReq[0], nil)
		suite.streamCreateFileMock.onRecvWithOnce(positiveReq[1], nil)
		suite.streamCreateFileMock.onRecvWithOnce(positiveReq[2], nil)
		suite.streamCreateFileMock.onRecvWithOnce(nil, io.EOF)
		suite.streamCreateFileMock.onSendAndClose(&proto.CreateFileResponse{Id: suite.testFileID}, nil)

		dbfmo := new(DBFilerMockedObject)
		dbfmo.onClose(nil)
		dbfmo.onWriteOnce(suite.testBatch1, len(suite.testBatch1), nil)
		dbfmo.onWriteOnce(suite.testBatch2, len(suite.testBatch2), nil)
		defer dbfmo.AssertExpectations(suite.T())

		suite.storageMock.onCreateFile(suite.testUserID, suite.testName, mock.Anything, suite.testMeta, &storage.File{ID: suite.testFileID}, nil)

		suite.fileStoreMock.onCreateDBFile(mock.Anything, dbfmo, nil)

		err := suite.handler.CreateFile(suite.streamCreateFileMock)
		require.NoError(err)
	})
}

func (suite *HandlersTestSuite) TestUpdateFile() {
	require := suite.Require()

	fileInfo := &proto.File{
		Id:   suite.testFileID,
		Name: suite.testName,
		Meta: suite.testMeta,
	}
	positiveReq := []*proto.UpdateFileRequest{
		{
			Data: &proto.UpdateFileRequest_FileInfo{
				FileInfo: fileInfo,
			},
		},
		{
			Data: &proto.UpdateFileRequest_Content{
				Content: suite.testBatch1,
			},
		},
		{
			Data: &proto.UpdateFileRequest_Content{
				Content: suite.testBatch2,
			},
		},
	}

	suite.Run("unauthenticated", func() {
		suite.streamUpdateFileMock.onContext(context.Background())

		err := suite.handler.UpdateFile(suite.streamUpdateFileMock)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
	})

	suite.Run("cannot receive file info", func() {
		suite.streamUpdateFileMock.onContext(suite.testIncomingContext)
		suite.streamUpdateFileMock.onRecvWithOnce(nil, testError)

		err := suite.handler.UpdateFile(suite.streamUpdateFileMock)
		require.ErrorContains(err, "cannot receive file info")
		require.Equal(status.Code(err), codes.Unknown)
	})

	suite.Run("create DB file error", func() {
		suite.streamUpdateFileMock.onContext(suite.testIncomingContext)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[0], nil)
		suite.fileStoreMock.onCreateDBFile(mock.Anything, nil, testError)

		err := suite.handler.UpdateFile(suite.streamUpdateFileMock)
		require.Error(err)
		require.Equal(status.Code(err), codes.Internal)
	})

	suite.Run("cannot receive content", func() {
		suite.streamUpdateFileMock.onContext(suite.testIncomingContext)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[0], nil)
		suite.streamUpdateFileMock.onRecvWithOnce(nil, testError)

		dbfmo := new(DBFilerMockedObject)
		dbfmo.onClose(nil)
		defer dbfmo.AssertExpectations(suite.T())

		suite.fileStoreMock.onCreateDBFile(mock.Anything, dbfmo, nil)

		err := suite.handler.UpdateFile(suite.streamUpdateFileMock)
		require.ErrorContains(err, "cannot receive content")
		require.Equal(status.Code(err), codes.Unknown)
	})

	suite.Run("write chunk error", func() {
		suite.streamUpdateFileMock.onContext(suite.testIncomingContext)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[0], nil)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[1], nil)

		dbfmo := new(DBFilerMockedObject)
		dbfmo.onClose(nil)
		dbfmo.onWriteOnce(suite.testBatch1, 0, testError)
		defer dbfmo.AssertExpectations(suite.T())

		suite.fileStoreMock.onCreateDBFile(mock.Anything, dbfmo, nil)

		err := suite.handler.UpdateFile(suite.streamUpdateFileMock)
		require.Error(err)
		require.Equal(status.Code(err), codes.Internal)
	})

	suite.Run("database error", func() {
		suite.streamUpdateFileMock.onContext(suite.testIncomingContext)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[0], nil)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[1], nil)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[2], nil)
		suite.streamUpdateFileMock.onRecvWithOnce(nil, io.EOF)

		dbfmo := new(DBFilerMockedObject)
		dbfmo.onClose(nil)
		dbfmo.onWriteOnce(suite.testBatch1, len(suite.testBatch1), nil)
		dbfmo.onWriteOnce(suite.testBatch2, len(suite.testBatch2), nil)
		defer dbfmo.AssertExpectations(suite.T())

		suite.storageMock.onUpdateFile(suite.testFileID, suite.testUserID, suite.testName, mock.Anything, suite.testMeta, nil, testError)

		suite.fileStoreMock.onCreateDBFile(mock.Anything, dbfmo, nil)

		err := suite.handler.UpdateFile(suite.streamUpdateFileMock)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
	})

	suite.Run("unknown UserID error", func() {
		suite.streamUpdateFileMock.onContext(suite.testIncomingContext)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[0], nil)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[1], nil)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[2], nil)
		suite.streamUpdateFileMock.onRecvWithOnce(nil, io.EOF)

		dbfmo := new(DBFilerMockedObject)
		dbfmo.onClose(nil)
		dbfmo.onWriteOnce(suite.testBatch1, len(suite.testBatch1), nil)
		dbfmo.onWriteOnce(suite.testBatch2, len(suite.testBatch2), nil)
		defer dbfmo.AssertExpectations(suite.T())

		suite.storageMock.onUpdateFile(suite.testFileID, suite.testUserID, suite.testName, mock.Anything, suite.testMeta, nil, storage.ErrUserNotFound)

		suite.fileStoreMock.onCreateDBFile(mock.Anything, dbfmo, nil)

		err := suite.handler.UpdateFile(suite.streamUpdateFileMock)
		require.ErrorContains(err, fmt.Sprintf("unknown UserID %s", suite.testUserID))
		require.Equal(status.Code(err), codes.Unknown)
	})

	suite.Run("delete file error", func() {
		suite.streamUpdateFileMock.onContext(suite.testIncomingContext)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[0], nil)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[1], nil)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[2], nil)
		suite.streamUpdateFileMock.onRecvWithOnce(nil, io.EOF)

		dbfmo := new(DBFilerMockedObject)
		dbfmo.onClose(nil)
		dbfmo.onWriteOnce(suite.testBatch1, len(suite.testBatch1), nil)
		dbfmo.onWriteOnce(suite.testBatch2, len(suite.testBatch2), nil)
		defer dbfmo.AssertExpectations(suite.T())

		suite.storageMock.onUpdateFile(suite.testFileID, suite.testUserID, suite.testName, mock.Anything, suite.testMeta, &storage.File{
			ID:         suite.testFileID,
			PathToFile: mock.Anything,
		}, nil)

		suite.fileStoreMock.onCreateDBFile(mock.Anything, dbfmo, nil)
		suite.fileStoreMock.onDeleteDBFile(mock.Anything, testError)

		err := suite.handler.UpdateFile(suite.streamUpdateFileMock)
		require.Error(err)
		require.Equal(status.Code(err), codes.Internal)
	})

	suite.Run("positive test", func() {
		suite.streamUpdateFileMock.onContext(suite.testIncomingContext)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[0], nil)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[1], nil)
		suite.streamUpdateFileMock.onRecvWithOnce(positiveReq[2], nil)
		suite.streamUpdateFileMock.onRecvWithOnce(nil, io.EOF)
		suite.streamUpdateFileMock.onSendAndClose(&proto.UpdateFileResponse{Id: suite.testFileID}, nil)

		dbfmo := new(DBFilerMockedObject)
		dbfmo.onClose(nil)
		dbfmo.onWriteOnce(suite.testBatch1, len(suite.testBatch1), nil)
		dbfmo.onWriteOnce(suite.testBatch2, len(suite.testBatch2), nil)
		defer dbfmo.AssertExpectations(suite.T())

		suite.storageMock.onUpdateFile(suite.testFileID, suite.testUserID, suite.testName, mock.Anything, suite.testMeta, &storage.File{
			ID:         suite.testFileID,
			PathToFile: mock.Anything,
		}, nil)

		suite.fileStoreMock.onCreateDBFile(mock.Anything, dbfmo, nil)
		suite.fileStoreMock.onDeleteDBFile(mock.Anything, nil)

		err := suite.handler.UpdateFile(suite.streamUpdateFileMock)
		require.NoError(err)
	})
}

func (suite *HandlersTestSuite) TestGetFile() {
	require := suite.Require()

	positiveReq := &proto.GetFileRequest{
		Id: suite.testFileID,
	}

	fileInfo := &proto.GetFileResponse{
		Data: &proto.GetFileResponse_FileInfo{
			FileInfo: &proto.File{
				Id:       suite.testFileID,
				Name:     suite.testName,
				Meta:     suite.testMeta,
				UpdateAt: timestamppb.New(suite.testUpdateAt),
			},
		},
	}

	content1 := &proto.GetFileResponse{
		Data: &proto.GetFileResponse_Content{
			Content: suite.testBatch1,
		},
	}

	content2 := &proto.GetFileResponse{
		Data: &proto.GetFileResponse_Content{
			Content: suite.testBatch2,
		},
	}

	fileReq := &storage.File{
		ID:         suite.testFileID,
		UserID:     suite.testUserID,
		Name:       suite.testName,
		PathToFile: mock.Anything,
		Meta:       suite.testMeta,
		UpdateAt:   suite.testUpdateAt,
	}

	suite.Run("unauthenticated", func() {
		suite.streamGetFileMock.onContext(context.Background())

		err := suite.handler.GetFile(positiveReq, suite.streamGetFileMock)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
	})

	suite.Run("empty FileID", func() {
		req := &proto.GetFileRequest{
			Id: "",
		}
		suite.streamGetFileMock.onContext(suite.testIncomingContext)

		err := suite.handler.GetFile(req, suite.streamGetFileMock)
		require.ErrorContains(err, "empty FileID")
		require.Equal(status.Code(err), codes.InvalidArgument)
	})

	suite.Run("database error", func() {
		suite.streamGetFileMock.onContext(suite.testIncomingContext)
		suite.storageMock.onGetFile(suite.testFileID, suite.testUserID, nil, testError)

		err := suite.handler.GetFile(positiveReq, suite.streamGetFileMock)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
	})

	suite.Run("unknown FileID error", func() {
		suite.streamGetFileMock.onContext(suite.testIncomingContext)
		suite.storageMock.onGetFile(suite.testFileID, suite.testUserID, nil, storage.ErrFileNotFound)

		err := suite.handler.GetFile(positiveReq, suite.streamGetFileMock)
		require.ErrorContains(err, fmt.Sprintf("unknown FileID %s", suite.testFileID))
		require.Equal(status.Code(err), codes.Unknown)
	})

	suite.Run("send file info error", func() {
		suite.streamGetFileMock.onContext(suite.testIncomingContext)
		suite.streamGetFileMock.onSendOnce(fileInfo, testError)

		suite.storageMock.onGetFile(suite.testFileID, suite.testUserID, fileReq, nil)

		err := suite.handler.GetFile(positiveReq, suite.streamGetFileMock)
		require.Error(err)
		require.Equal(status.Code(err), codes.Internal)
	})

	suite.Run("get DB file error", func() {
		suite.streamGetFileMock.onContext(suite.testIncomingContext)
		suite.streamGetFileMock.onSendOnce(fileInfo, nil)

		suite.storageMock.onGetFile(suite.testFileID, suite.testUserID, fileReq, nil)

		suite.fileStoreMock.onGetDBFile(mock.Anything, nil, testError)

		err := suite.handler.GetFile(positiveReq, suite.streamGetFileMock)
		require.Error(err)
		require.Equal(status.Code(err), codes.Internal)
	})

	suite.Run("get chunk error", func() {
		suite.streamGetFileMock.onContext(suite.testIncomingContext)
		suite.streamGetFileMock.onSendOnce(fileInfo, nil)

		suite.storageMock.onGetFile(suite.testFileID, suite.testUserID, fileReq, nil)

		dbfmo := new(DBFilerMockedObject)
		dbfmo.onClose(nil)
		dbfmo.onGetChunkOnce(nil, testError)
		defer dbfmo.AssertExpectations(suite.T())

		suite.fileStoreMock.onGetDBFile(mock.Anything, dbfmo, nil)

		err := suite.handler.GetFile(positiveReq, suite.streamGetFileMock)
		require.Error(err)
		require.Equal(status.Code(err), codes.Internal)
	})

	suite.Run("send error", func() {
		suite.streamGetFileMock.onContext(suite.testIncomingContext)
		suite.streamGetFileMock.onSendOnce(fileInfo, nil)
		suite.streamGetFileMock.onSendOnce(content1, testError)

		suite.storageMock.onGetFile(suite.testFileID, suite.testUserID, fileReq, nil)

		dbfmo := new(DBFilerMockedObject)
		dbfmo.onClose(nil)
		dbfmo.onGetChunkOnce(suite.testBatch1, nil)
		defer dbfmo.AssertExpectations(suite.T())

		suite.fileStoreMock.onGetDBFile(mock.Anything, dbfmo, nil)

		err := suite.handler.GetFile(positiveReq, suite.streamGetFileMock)
		require.Error(err)
		require.Equal(status.Code(err), codes.Internal)
	})

	suite.Run("positive test", func() {
		suite.streamGetFileMock.onContext(suite.testIncomingContext)
		suite.streamGetFileMock.onSendOnce(fileInfo, nil)
		suite.streamGetFileMock.onSendOnce(content1, nil)
		suite.streamGetFileMock.onSendOnce(content2, nil)

		suite.storageMock.onGetFile(suite.testFileID, suite.testUserID, fileReq, nil)

		dbfmo := new(DBFilerMockedObject)
		dbfmo.onClose(nil)
		dbfmo.onGetChunkOnce(suite.testBatch1, nil)
		dbfmo.onGetChunkOnce(suite.testBatch2, nil)
		dbfmo.onGetChunkOnce(nil, io.EOF)
		defer dbfmo.AssertExpectations(suite.T())

		suite.fileStoreMock.onGetDBFile(mock.Anything, dbfmo, nil)

		err := suite.handler.GetFile(positiveReq, suite.streamGetFileMock)
		require.NoError(err)
	})
}

func (suite *HandlersTestSuite) TestGetFiles() {
	require := suite.Require()

	suite.Run("unauthenticated", func() {
		res, err := suite.handler.GetFiles(context.Background(), &proto.GetFilesRequest{})
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onGetAllFiles(suite.testUserID, nil, testError)

		res, err := suite.handler.GetFiles(suite.testIncomingContext, &proto.GetFilesRequest{})
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown UserID error", func() {
		suite.storageMock.onGetAllFiles(suite.testUserID, nil, storage.ErrUserNotFound)

		res, err := suite.handler.GetFiles(suite.testIncomingContext, &proto.GetFilesRequest{})
		require.ErrorContains(err, fmt.Sprintf("unknown UserID %s", suite.testUserID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		files := []storage.File{
			{
				ID:       suite.testFileID,
				Name:     suite.testName,
				Meta:     suite.testMeta,
				UpdateAt: suite.testUpdateAt,
			},
			{
				ID:       "anotherFileID",
				Name:     "anotherName",
				Meta:     "anotherMeta",
				UpdateAt: suite.testUpdateAt,
			},
		}
		suite.storageMock.onGetAllFiles(suite.testUserID, files, nil)

		res, err := suite.handler.GetFiles(suite.testIncomingContext, &proto.GetFilesRequest{})
		suite.Require().NoError(err)
		suite.Require().Equal(&proto.GetFilesResponse{
			FileInfo: []*proto.File{
				{
					Id:       suite.testFileID,
					Name:     suite.testName,
					Meta:     suite.testMeta,
					UpdateAt: timestamppb.New(suite.testUpdateAt),
				},
				{
					Id:       "anotherFileID",
					Name:     "anotherName",
					Meta:     "anotherMeta",
					UpdateAt: timestamppb.New(suite.testUpdateAt),
				},
			},
		}, res)
	})
}

func (suite *HandlersTestSuite) TestDeleteFile() {
	require := suite.Require()

	positiveReq := &proto.DeleteFileRequest{
		Id: suite.testFileID,
	}

	file := &storage.File{
		ID:         suite.testFileID,
		UserID:     suite.testUserID,
		Name:       suite.testName,
		PathToFile: mock.Anything,
		Meta:       suite.testMeta,
		UpdateAt:   suite.testUpdateAt,
	}

	suite.Run("unauthenticated", func() {
		req := &proto.DeleteFileRequest{
			Id: "",
		}

		res, err := suite.handler.DeleteFile(context.Background(), req)
		require.Error(err)
		require.Equal(status.Code(err), codes.Unauthenticated)
		require.Nil(res)
	})

	suite.Run("database error", func() {
		suite.storageMock.onDeleteFile(suite.testFileID, suite.testUserID, nil, testError)

		res, err := suite.handler.DeleteFile(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, testError.Error())
		require.Equal(status.Code(err), codes.Internal)
		require.Nil(res)
	})

	suite.Run("unknown FileID error", func() {
		suite.storageMock.onDeleteFile(suite.testFileID, suite.testUserID, nil, storage.ErrFileNotFound)

		res, err := suite.handler.DeleteFile(suite.testIncomingContext, positiveReq)
		require.ErrorContains(err, fmt.Sprintf("unknown FileID %s", suite.testFileID))
		require.Equal(status.Code(err), codes.Unknown)
		require.Nil(res)
	})

	suite.Run("delete file error", func() {
		suite.storageMock.onDeleteFile(suite.testFileID, suite.testUserID, file, nil)

		suite.fileStoreMock.onDeleteDBFile(mock.Anything, testError)

		res, err := suite.handler.DeleteFile(suite.testIncomingContext, positiveReq)
		require.Error(err)
		require.Nil(res)
	})

	suite.Run("positive test", func() {
		suite.storageMock.onDeleteFile(suite.testFileID, suite.testUserID, file, nil)

		suite.fileStoreMock.onDeleteDBFile(mock.Anything, nil)

		res, err := suite.handler.DeleteFile(suite.testIncomingContext, positiveReq)
		require.NoError(err)
		require.Nil(res)
	})
}

func (suite *HandlersTestSuite) TestGetChunkSize() {
	suite.Run("positive test", func() {
		suite.fileStoreMock.onGetChunkSize(1024)

		res, err := suite.handler.GetChunkSize(nil, nil)
		suite.Require().NoError(err)
		suite.Equal(&proto.GetChunkSizeResponse{
			Size: 1024,
		}, res)
	})
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}
