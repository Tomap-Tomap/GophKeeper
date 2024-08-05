//go:build unit

package client

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Tomap-Tomap/GophKeeper/crypto"
	"github.com/Tomap-Tomap/GophKeeper/proto"
	"github.com/Tomap-Tomap/GophKeeper/storage"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var errTest = errors.New("test")

type ClientTestSuite struct {
	suite.Suite

	testLogin            string
	testPassword         string
	testToken            string
	testPasswordID       string
	testName             string
	testMeta             string
	testUpdateAt         time.Time
	testBankID           string
	testCardNumber       string
	testCvc              string
	testOwner            string
	testExp              string
	testTextID           string
	testText             string
	testFileID           string
	testPathToFile       string
	testPathToFileForGet string

	testChunkSize int
	testNonceSize int

	testNonce    []byte
	testContent1 []byte
	testContent2 []byte

	serverMock           *GophKeeperServerMockedObject
	crypterMock          *CrypterMockedObject
	clienMock            *GophKeeperClientMockedObject
	createFileStreamMock *CreateFileClientMockedObject
	updateFileStreamMock *UpdateFileClientMockedObject
	getFileStreamMock    *GetFileClientMockedObject
	server               *grpc.Server
	client               *Client
}

func (suite *ClientTestSuite) SetupTest() {
	require := suite.Require()
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(err)

	server := grpc.NewServer()
	gpmo := new(GophKeeperServerMockedObject)
	cmo := new(CrypterMockedObject)
	gkcmo := new(GophKeeperClientMockedObject)
	cfcmo := new(CreateFileClientMockedObject)
	ufcmo := new(UpdateFileClientMockedObject)
	gfcmo := new(GetFileClientMockedObject)

	proto.RegisterGophKeeperServer(server, gpmo)

	go func() {
		if err := server.Serve(lis); err != nil && err != http.ErrServerClosed {
			require.FailNow(err.Error())
		}
	}()

	client, err := New(cmo, lis.Addr().String())
	require.NoError(err)

	suite.client = client
	suite.server = server
	suite.serverMock = gpmo
	suite.crypterMock = cmo
	suite.clienMock = gkcmo
	suite.createFileStreamMock = cfcmo
	suite.updateFileStreamMock = ufcmo
	suite.getFileStreamMock = gfcmo

	suite.testLogin = "testLogin"
	suite.testPassword = "testPassword"
	suite.testToken = "testToken"
	suite.testPasswordID = "testPasswordID"
	suite.testName = "testName"
	suite.testMeta = "testMeta"
	suite.testUpdateAt = time.Now().UTC()
	suite.testBankID = "testBankID"
	suite.testCardNumber = "testCardNumber"
	suite.testCvc = "testCVC"
	suite.testOwner = "testOwner"
	suite.testExp = "testExp"
	suite.testTextID = "testTextID"
	suite.testText = "testText"
	suite.testFileID = "testFileID"
	suite.testPathToFile = "./testdata/testfile"
	suite.testPathToFileForGet = "./testdata"
	suite.testChunkSize = 1024

	suite.testNonce = []byte("testNonce")
	suite.testNonceSize = len(suite.testNonce)
	suite.testContent1 = make([]byte, 0, 1024)
	suite.testContent2 = make([]byte, 0, 1024)

	for i := 0; i < 1024; i++ {
		suite.testContent1 = append(suite.testContent1, 1)
		suite.testContent2 = append(suite.testContent2, 1)
	}

	suite.client.ti.setToken(suite.testToken)
}

func (suite *ClientTestSuite) TearDownTest() {
	suite.server.Stop()

	err := suite.client.Close()
	suite.Require().NoError(err)
}

func (suite *ClientTestSuite) TearDownSubTest() {
	suite.serverMock.AssertExpectations(suite.T())
	suite.crypterMock.AssertExpectations(suite.T())
	suite.clienMock.AssertExpectations(suite.T())
	suite.createFileStreamMock.AssertExpectations(suite.T())
	suite.updateFileStreamMock.AssertExpectations(suite.T())
	suite.getFileStreamMock.AssertExpectations(suite.T())

	for len(suite.serverMock.ExpectedCalls) != 0 {
		suite.serverMock.ExpectedCalls[0].Unset()
	}

	for len(suite.crypterMock.ExpectedCalls) != 0 {
		suite.crypterMock.ExpectedCalls[0].Unset()
	}

	for len(suite.clienMock.ExpectedCalls) != 0 {
		suite.clienMock.ExpectedCalls[0].Unset()
	}

	for len(suite.createFileStreamMock.ExpectedCalls) != 0 {
		suite.createFileStreamMock.ExpectedCalls[0].Unset()
	}

	for len(suite.updateFileStreamMock.ExpectedCalls) != 0 {
		suite.updateFileStreamMock.ExpectedCalls[0].Unset()
	}

	for len(suite.getFileStreamMock.ExpectedCalls) != 0 {
		suite.getFileStreamMock.ExpectedCalls[0].Unset()
	}
}

func (suite *ClientTestSuite) TestRegister() {
	require := suite.Require()

	suite.Run("error test", func() {
		suite.client.ti.token = ""
		suite.serverMock.onRegister(
			&proto.RegisterRequest{
				Login:    suite.testLogin,
				Password: suite.testPassword,
			},
			nil,
			errTest,
		)

		err := suite.client.Register(context.Background(), suite.testLogin, suite.testPassword)
		require.Error(err)
		suite.Equal("", suite.client.ti.token)
	})

	suite.Run("positive test", func() {
		suite.serverMock.onRegister(
			&proto.RegisterRequest{
				Login:    suite.testLogin,
				Password: suite.testPassword,
			},
			&proto.RegisterResponse{
				Token: suite.testToken,
			},
			nil,
		)

		err := suite.client.Register(context.Background(), suite.testLogin, suite.testPassword)
		require.NoError(err)
		suite.Equal("Bearer "+suite.testToken, suite.client.ti.token)
	})
}

func (suite *ClientTestSuite) TestSignIn() {
	require := suite.Require()

	suite.Run("error test", func() {
		suite.client.ti.token = ""
		suite.serverMock.onAuth(
			&proto.AuthRequest{
				Login:    suite.testLogin,
				Password: suite.testPassword,
			},
			nil,
			errTest,
		)

		err := suite.client.SignIn(context.Background(), suite.testLogin, suite.testPassword)
		require.Error(err)
		suite.Equal("", suite.client.ti.token)
	})

	suite.Run("positive test", func() {
		suite.serverMock.onAuth(
			&proto.AuthRequest{
				Login:    suite.testLogin,
				Password: suite.testPassword,
			},
			&proto.AuthResponse{
				Token: suite.testToken,
			},
			nil,
		)

		err := suite.client.SignIn(context.Background(), suite.testLogin, suite.testPassword)
		require.NoError(err)
		suite.Equal("Bearer "+suite.testToken, suite.client.ti.token)
	})
}

func (suite *ClientTestSuite) TestGetAllPasswords() {
	require := suite.Require()

	res := &proto.Password{
		Id:       suite.testPasswordID,
		Name:     suite.testName,
		Login:    suite.testLogin,
		Password: suite.testPassword,
		Meta:     suite.testMeta,
		UpdateAt: timestamppb.New(suite.testUpdateAt),
	}

	reses := &proto.GetPasswordsResponse{
		Passwords: []*proto.Password{
			res,
			res,
		},
	}

	suite.Run("service error", func() {
		suite.serverMock.onGetPasswords(nil, errTest)

		pwds, err := suite.client.GetAllPasswords(context.Background())
		require.ErrorContains(err, "cannot get passwords")
		suite.Nil(pwds)
	})

	suite.Run("cannot open password data", func() {
		suite.serverMock.onGetPasswords(reses, nil)

		suite.crypterMock.onOpenStringWithoutNonce(suite.testName, "", errTest).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testLogin, "", errTest).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testPassword, "", errTest).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testMeta, "", errTest).Twice()

		pwds, err := suite.client.GetAllPasswords(context.Background())
		require.ErrorContains(err, "cannot open password data")
		require.ErrorContains(err, "cannot open name")
		require.ErrorContains(err, "cannot open login")
		require.ErrorContains(err, "cannot open password")
		require.ErrorContains(err, "cannot open meta")

		suite.Nil(pwds)
	})

	suite.Run("positive test", func() {
		wantPwd := storage.Password{
			ID:       suite.testPasswordID,
			Name:     suite.testName,
			Login:    suite.testLogin,
			Password: suite.testPassword,
			Meta:     suite.testMeta,
			UpdateAt: suite.testUpdateAt,
		}

		wantPwds := []storage.Password{
			wantPwd,
			wantPwd,
		}

		suite.serverMock.onGetPasswords(reses, nil)

		suite.crypterMock.onOpenStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onOpenStringWithoutNonce(suite.testLogin, suite.testLogin, nil)
		suite.crypterMock.onOpenStringWithoutNonce(suite.testPassword, suite.testPassword, nil)
		suite.crypterMock.onOpenStringWithoutNonce(suite.testMeta, res.Meta, nil)

		pwds, err := suite.client.GetAllPasswords(context.Background())
		require.NoError(err)
		suite.Equal(wantPwds, pwds)
	})
}

func (suite *ClientTestSuite) TestCreatePassword() {
	require := suite.Require()

	req := &proto.CreatePasswordRequest{
		Name:     suite.testName,
		Login:    suite.testLogin,
		Password: suite.testPassword,
		Meta:     suite.testMeta,
	}

	suite.Run("seal password error", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testLogin, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testPassword, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, "", errTest)

		err := suite.client.CreatePassword(context.Background(), suite.testName, suite.testLogin, suite.testPassword, suite.testMeta)
		require.ErrorContains(err, "cannot seal password")
		suite.ErrorContains(err, "cannot seal name")
		suite.ErrorContains(err, "cannot seal login")
		suite.ErrorContains(err, "cannot seal password")
		suite.ErrorContains(err, "cannot seal meta")
	})

	suite.Run("service error", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testLogin, suite.testLogin, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testPassword, suite.testPassword, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.serverMock.onCreatePassword(req, nil, errTest)

		err := suite.client.CreatePassword(context.Background(), suite.testName, suite.testLogin, suite.testPassword, suite.testMeta)
		require.ErrorContains(err, "cannot create password")
	})

	suite.Run("positive test", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testLogin, suite.testLogin, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testPassword, suite.testPassword, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.serverMock.onCreatePassword(req, nil, nil)

		err := suite.client.CreatePassword(context.Background(), suite.testName, suite.testLogin, suite.testPassword, suite.testMeta)
		require.NoError(err)
	})
}

func (suite *ClientTestSuite) TestUpdatePassword() {
	require := suite.Require()

	req := &proto.UpdatePasswordRequest{
		Id:       suite.testPasswordID,
		Name:     suite.testName,
		Login:    suite.testLogin,
		Password: suite.testPassword,
		Meta:     suite.testMeta,
	}

	suite.Run("seal password error", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testLogin, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testPassword, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, "", errTest)

		err := suite.client.UpdatePassword(context.Background(), suite.testPasswordID, suite.testName, suite.testLogin, suite.testPassword, suite.testMeta)
		require.ErrorContains(err, "cannot seal password")
		suite.ErrorContains(err, "cannot seal name")
		suite.ErrorContains(err, "cannot seal login")
		suite.ErrorContains(err, "cannot seal password")
		suite.ErrorContains(err, "cannot seal meta")
	})

	suite.Run("service error", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testLogin, suite.testLogin, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testPassword, suite.testPassword, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.serverMock.onUpdatePassword(req, nil, errTest)

		err := suite.client.UpdatePassword(context.Background(), suite.testPasswordID, suite.testName, suite.testLogin, suite.testPassword, suite.testMeta)
		require.ErrorContains(err, "cannot update password")
	})

	suite.Run("positive test", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testLogin, suite.testLogin, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testPassword, suite.testPassword, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.serverMock.onUpdatePassword(req, nil, nil)

		err := suite.client.UpdatePassword(context.Background(), suite.testPasswordID, suite.testName, suite.testLogin, suite.testPassword, suite.testMeta)
		require.NoError(err)
	})
}

func (suite *ClientTestSuite) TestDeletePassword() {
	require := suite.Require()

	req := &proto.DeletePasswordRequest{
		Id: suite.testPasswordID,
	}

	suite.Run("service error", func() {
		suite.serverMock.onDeletePassword(req, errTest)

		err := suite.client.DeletePassword(context.Background(), suite.testPasswordID)
		require.ErrorContains(err, "cannot delete password")
	})

	suite.Run("positive test", func() {
		suite.serverMock.onDeletePassword(req, nil)

		err := suite.client.DeletePassword(context.Background(), suite.testPasswordID)
		require.NoError(err)
	})
}

func (suite *ClientTestSuite) TestGetAllBanks() {
	require := suite.Require()

	res := &proto.Bank{
		Id:         suite.testBankID,
		Name:       suite.testName,
		CardNumber: suite.testCardNumber,
		Cvc:        suite.testCvc,
		Owner:      suite.testOwner,
		Exp:        suite.testExp,
		Meta:       suite.testMeta,
		UpdateAt:   timestamppb.New(suite.testUpdateAt),
	}

	reses := &proto.GetBanksResponse{
		Banks: []*proto.Bank{
			res,
			res,
		},
	}

	suite.Run("service error", func() {
		suite.serverMock.onGetBanks(nil, errTest)

		banks, err := suite.client.GetAllBanks(context.Background())
		require.ErrorContains(err, "cannot get banks")
		suite.Nil(banks)
	})

	suite.Run("cannot open bank data", func() {
		suite.serverMock.onGetBanks(reses, nil)

		suite.crypterMock.onOpenStringWithoutNonce(suite.testName, "", errTest).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testCardNumber, "", errTest).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testCvc, "", errTest).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testOwner, "", errTest).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testExp, "", errTest).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testMeta, "", errTest).Twice()

		banks, err := suite.client.GetAllBanks(context.Background())
		require.ErrorContains(err, "cannot open bank's data")
		suite.ErrorContains(err, "cannot open name")
		suite.ErrorContains(err, "cannot open card number")
		suite.ErrorContains(err, "cannot open cvc")
		suite.ErrorContains(err, "cannot open owner")
		suite.ErrorContains(err, "cannot open exp")
		suite.ErrorContains(err, "cannot open meta")

		suite.Nil(banks)
	})

	suite.Run("positive test", func() {
		wantBank := storage.Bank{
			ID:         suite.testBankID,
			Name:       suite.testName,
			CardNumber: suite.testCardNumber,
			CVC:        suite.testCvc,
			Owner:      suite.testOwner,
			Exp:        suite.testExp,
			Meta:       suite.testMeta,
			UpdateAt:   suite.testUpdateAt,
		}

		wantBanks := []storage.Bank{
			wantBank,
			wantBank,
		}

		suite.serverMock.onGetBanks(reses, nil)

		suite.crypterMock.onOpenStringWithoutNonce(suite.testName, suite.testName, nil).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testCardNumber, suite.testCardNumber, nil).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testCvc, suite.testCvc, nil).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testOwner, suite.testOwner, nil).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testExp, suite.testExp, nil).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testMeta, suite.testMeta, nil).Twice()

		banks, err := suite.client.GetAllBanks(context.Background())
		require.NoError(err)
		suite.Equal(wantBanks, banks)
	})
}

func (suite *ClientTestSuite) TestCreateBank() {
	require := suite.Require()

	req := &proto.CreateBankRequest{
		Name:       suite.testName,
		CardNumber: suite.testCardNumber,
		Cvc:        suite.testCvc,
		Owner:      suite.testOwner,
		Exp:        suite.testExp,
		Meta:       suite.testMeta,
	}

	suite.Run("seal bank error", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testCardNumber, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testCvc, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testOwner, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testExp, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, "", errTest)

		err := suite.client.CreateBank(context.Background(), suite.testName, suite.testCardNumber, suite.testCvc, suite.testOwner, suite.testExp, suite.testMeta)
		require.ErrorContains(err, "cannot seal bank")
		suite.ErrorContains(err, "cannot seal name")
		suite.ErrorContains(err, "cannot seal card number")
		suite.ErrorContains(err, "cannot seal cvc")
		suite.ErrorContains(err, "cannot seal owner")
		suite.ErrorContains(err, "cannot seal exp")
		suite.ErrorContains(err, "cannot seal meta")
	})

	suite.Run("service error", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testCardNumber, suite.testCardNumber, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testCvc, suite.testCvc, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testOwner, suite.testOwner, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testExp, suite.testExp, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.serverMock.onCreateBank(req, nil, errTest)

		err := suite.client.CreateBank(context.Background(), suite.testName, suite.testCardNumber, suite.testCvc, suite.testOwner, suite.testExp, suite.testMeta)
		require.ErrorContains(err, "cannot create bank")
	})

	suite.Run("positive test", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testCardNumber, suite.testCardNumber, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testCvc, suite.testCvc, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testOwner, suite.testOwner, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testExp, suite.testExp, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.serverMock.onCreateBank(req, nil, nil)

		err := suite.client.CreateBank(context.Background(), suite.testName, suite.testCardNumber, suite.testCvc, suite.testOwner, suite.testExp, suite.testMeta)
		require.NoError(err)
	})
}

func (suite *ClientTestSuite) TestUpdateBank() {
	require := suite.Require()

	req := &proto.UpdateBankRequest{
		Id:         suite.testBankID,
		Name:       suite.testName,
		CardNumber: suite.testCardNumber,
		Cvc:        suite.testCvc,
		Owner:      suite.testOwner,
		Exp:        suite.testExp,
		Meta:       suite.testMeta,
	}

	suite.Run("seal bank error", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testCardNumber, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testCvc, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testOwner, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testExp, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, "", errTest)

		err := suite.client.UpdateBank(context.Background(), suite.testBankID, suite.testName, suite.testCardNumber, suite.testCvc, suite.testOwner, suite.testExp, suite.testMeta)
		require.ErrorContains(err, "cannot seal bank")
		suite.ErrorContains(err, "cannot seal name")
		suite.ErrorContains(err, "cannot seal card number")
		suite.ErrorContains(err, "cannot seal cvc")
		suite.ErrorContains(err, "cannot seal owner")
		suite.ErrorContains(err, "cannot seal exp")
		suite.ErrorContains(err, "cannot seal meta")
	})

	suite.Run("service error", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testCardNumber, suite.testCardNumber, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testCvc, suite.testCvc, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testOwner, suite.testOwner, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testExp, suite.testExp, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.serverMock.onUpdateBank(req, nil, errTest)

		err := suite.client.UpdateBank(context.Background(), suite.testBankID, suite.testName, suite.testCardNumber, suite.testCvc, suite.testOwner, suite.testExp, suite.testMeta)
		require.ErrorContains(err, "cannot update bank")
	})

	suite.Run("positive test", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testCardNumber, suite.testCardNumber, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testCvc, suite.testCvc, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testOwner, suite.testOwner, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testExp, suite.testExp, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.serverMock.onUpdateBank(req, nil, nil)

		err := suite.client.UpdateBank(context.Background(), suite.testBankID, suite.testName, suite.testCardNumber, suite.testCvc, suite.testOwner, suite.testExp, suite.testMeta)
		require.NoError(err)
	})
}

func (suite *ClientTestSuite) TestDeleteBank() {
	require := suite.Require()

	req := &proto.DeleteBankRequest{
		Id: suite.testBankID,
	}

	suite.Run("service error", func() {
		suite.serverMock.onDeleteBank(req, errTest)

		err := suite.client.DeleteBank(context.Background(), suite.testBankID)
		require.ErrorContains(err, "cannot delete bank")
	})

	suite.Run("positive test", func() {
		suite.serverMock.onDeleteBank(req, nil)

		err := suite.client.DeleteBank(context.Background(), suite.testBankID)
		require.NoError(err)
	})
}

func (suite *ClientTestSuite) TestGetAllTexts() {
	require := suite.Require()

	res := &proto.Text{
		Id:       suite.testTextID,
		Name:     suite.testName,
		Text:     suite.testText,
		Meta:     suite.testMeta,
		UpdateAt: timestamppb.New(suite.testUpdateAt),
	}

	reses := &proto.GetTextsResponse{
		Texts: []*proto.Text{
			res,
			res,
		},
	}

	suite.Run("service error", func() {
		suite.serverMock.onGetTexts(nil, errTest)

		texts, err := suite.client.GetAllTexts(context.Background())
		require.ErrorContains(err, "cannot get texts")
		suite.Nil(texts)
	})

	suite.Run("cannot open text data", func() {
		suite.serverMock.onGetTexts(reses, nil)

		suite.crypterMock.onOpenStringWithoutNonce(suite.testName, "", errTest).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testText, "", errTest).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testMeta, "", errTest).Twice()

		texts, err := suite.client.GetAllTexts(context.Background())
		require.ErrorContains(err, "cannot open text data")
		require.ErrorContains(err, "cannot open name")
		require.ErrorContains(err, "cannot open text")
		require.ErrorContains(err, "cannot open meta")

		suite.Nil(texts)
	})

	suite.Run("positive test", func() {
		wantText := storage.Text{
			ID:       suite.testTextID,
			Name:     suite.testName,
			Text:     suite.testText,
			Meta:     suite.testMeta,
			UpdateAt: suite.testUpdateAt,
		}

		wantTexts := []storage.Text{
			wantText,
			wantText,
		}

		suite.serverMock.onGetTexts(reses, nil)

		suite.crypterMock.onOpenStringWithoutNonce(suite.testName, suite.testName, nil).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testText, suite.testText, nil).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testMeta, suite.testMeta, nil).Twice()

		texts, err := suite.client.GetAllTexts(context.Background())
		require.NoError(err)
		suite.Equal(wantTexts, texts)
	})
}

func (suite *ClientTestSuite) TestCreateText() {
	require := suite.Require()

	req := &proto.CreateTextRequest{
		Name: suite.testName,
		Text: suite.testText,
		Meta: suite.testMeta,
	}

	suite.Run("seal text error", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testText, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, "", errTest)

		err := suite.client.CreateText(context.Background(), suite.testName, suite.testText, suite.testMeta)
		require.ErrorContains(err, "cannot seal text")
		suite.ErrorContains(err, "cannot seal name")
		suite.ErrorContains(err, "cannot seal text")
		suite.ErrorContains(err, "cannot seal meta")
	})

	suite.Run("service error", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testText, suite.testText, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.serverMock.onCreateText(req, nil, errTest)

		err := suite.client.CreateText(context.Background(), suite.testName, suite.testText, suite.testMeta)
		require.ErrorContains(err, "cannot create text")
	})

	suite.Run("positive test", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testText, suite.testText, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.serverMock.onCreateText(req, nil, nil)

		err := suite.client.CreateText(context.Background(), suite.testName, suite.testText, suite.testMeta)
		require.NoError(err)
	})
}

func (suite *ClientTestSuite) TestUpdateText() {
	require := suite.Require()

	req := &proto.UpdateTextRequest{
		Id:   suite.testTextID,
		Name: suite.testName,
		Text: suite.testText,
		Meta: suite.testMeta,
	}

	suite.Run("seal text error", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testText, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, "", errTest)

		err := suite.client.UpdateText(context.Background(), suite.testTextID, suite.testName, suite.testText, suite.testMeta)
		require.ErrorContains(err, "cannot seal text")
		suite.ErrorContains(err, "cannot seal name")
		suite.ErrorContains(err, "cannot seal text")
		suite.ErrorContains(err, "cannot seal meta")
	})

	suite.Run("service error", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testText, suite.testText, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.serverMock.onUpdateText(req, nil, errTest)

		err := suite.client.UpdateText(context.Background(), suite.testTextID, suite.testName, suite.testText, suite.testMeta)
		require.ErrorContains(err, "cannot update text")
	})

	suite.Run("positive test", func() {
		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testText, suite.testText, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.serverMock.onUpdateText(req, nil, nil)

		err := suite.client.UpdateText(context.Background(), suite.testTextID, suite.testName, suite.testText, suite.testMeta)
		require.NoError(err)
	})
}

func (suite *ClientTestSuite) TestDeleteText() {
	require := suite.Require()

	req := &proto.DeleteTextRequest{
		Id: suite.testTextID,
	}

	suite.Run("service error", func() {
		suite.serverMock.onDeleteText(req, errTest)

		err := suite.client.DeleteText(context.Background(), suite.testTextID)
		require.ErrorContains(err, "cannot delete text")
	})

	suite.Run("positive test", func() {
		suite.serverMock.onDeleteText(req, nil)

		err := suite.client.DeleteText(context.Background(), suite.testTextID)
		require.NoError(err)
	})
}
func (suite *ClientTestSuite) TestGetAllFiles() {
	require := suite.Require()

	res := &proto.File{
		Id:       suite.testFileID,
		Name:     suite.testName,
		Meta:     suite.testMeta,
		UpdateAt: timestamppb.New(suite.testUpdateAt),
	}

	reses := &proto.GetFilesResponse{
		FileInfo: []*proto.File{
			res,
			res,
		},
	}

	suite.Run("service error", func() {
		suite.serverMock.onGetFiles(nil, errTest)

		files, err := suite.client.GetAllFiles(context.Background())
		require.ErrorContains(err, "cannot get files")
		suite.Nil(files)
	})

	suite.Run("cannot open file data", func() {
		suite.serverMock.onGetFiles(reses, nil)

		suite.crypterMock.onOpenStringWithoutNonce(suite.testName, "", errTest).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testMeta, "", errTest).Twice()

		files, err := suite.client.GetAllFiles(context.Background())
		require.ErrorContains(err, "cannot open file data")
		require.ErrorContains(err, "cannot open name")
		require.ErrorContains(err, "cannot open meta")

		suite.Nil(files)
	})

	suite.Run("positive test", func() {
		wantFile := storage.File{
			ID:       suite.testFileID,
			Name:     suite.testName,
			Meta:     suite.testMeta,
			UpdateAt: suite.testUpdateAt,
		}

		wantFiles := []storage.File{
			wantFile,
			wantFile,
		}

		suite.serverMock.onGetFiles(reses, nil)

		suite.crypterMock.onOpenStringWithoutNonce(suite.testName, suite.testName, nil).Twice()
		suite.crypterMock.onOpenStringWithoutNonce(suite.testMeta, suite.testMeta, nil).Twice()

		files, err := suite.client.GetAllFiles(context.Background())
		require.NoError(err)
		suite.Equal(wantFiles, files)
	})
}

func (suite *ClientTestSuite) TestCreateFile() {
	suite.client.grpc = suite.clienMock

	require := suite.Require()

	csReq := &proto.GetChunkSizeResponse{
		Size: uint64(suite.testChunkSize),
	}

	fiReq := &proto.CreateFileRequest{
		Data: &proto.CreateFileRequest_FileInfo{
			FileInfo: &proto.File{
				Name: suite.testName,
				Meta: suite.testMeta,
			},
		},
	}

	nonceReq := &proto.CreateFileRequest{
		Data: &proto.CreateFileRequest_Content{
			Content: suite.testNonce,
		},
	}

	contentReq1 := &proto.CreateFileRequest{
		Data: &proto.CreateFileRequest_Content{
			Content: suite.testContent1,
		},
	}

	contentReq2 := &proto.CreateFileRequest{
		Data: &proto.CreateFileRequest_Content{
			Content: suite.testContent2,
		},
	}

	suite.Run("cannot get chunk size", func() {
		suite.clienMock.onGetChunkSize(nil, errTest)

		err := suite.client.CreateFile(context.Background(), suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot get chunk size")
	})

	suite.Run("cannot seal file info", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, "", errTest)

		err := suite.client.CreateFile(context.Background(), suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot seal file info")
		suite.ErrorContains(err, "cannot seal name")
		suite.ErrorContains(err, "cannot seal meta")
	})

	suite.Run("cannot start creating a file stream", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onCreateFile(nil, errTest)
		err := suite.client.CreateFile(context.Background(), suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot start creating a file stream")
	})

	suite.Run("cannot send file info", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onCreateFile(suite.createFileStreamMock, nil)

		suite.createFileStreamMock.onSend(fiReq, errTest)

		err := suite.client.CreateFile(context.Background(), suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot send file info")
	})

	suite.Run("cannot open file by path", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onCreateFile(suite.createFileStreamMock, nil)

		suite.createFileStreamMock.onSend(fiReq, nil)

		err := suite.client.CreateFile(context.Background(), suite.testName, "errPath", suite.testMeta)
		require.ErrorContains(err, "cannot open file by path")
	})

	suite.Run("cannot generate nonce", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onCreateFile(suite.createFileStreamMock, nil)

		suite.createFileStreamMock.onSend(fiReq, nil)

		suite.crypterMock.onGenerateNonce(nil, errTest)

		err := suite.client.CreateFile(context.Background(), suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot generate nonce")
	})

	suite.Run("cannot generate nonce", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onCreateFile(suite.createFileStreamMock, nil)

		suite.createFileStreamMock.onSend(fiReq, nil)

		suite.crypterMock.onGenerateNonce(suite.testNonce, nil)

		suite.createFileStreamMock.onSend(nonceReq, errTest)

		err := suite.client.CreateFile(context.Background(), suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot send nonce")
	})

	suite.Run("cannot send file data", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onCreateFile(suite.createFileStreamMock, nil)

		suite.createFileStreamMock.onSend(fiReq, nil)

		suite.crypterMock.onGenerateNonce(suite.testNonce, nil)

		suite.createFileStreamMock.onSend(nonceReq, nil)
		suite.crypterMock.onSealBytes(suite.testContent1, suite.testNonce, suite.testContent1).Once()
		suite.createFileStreamMock.onSend(contentReq1, errTest).Once()

		err := suite.client.CreateFile(context.Background(), suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot send file data")
	})

	suite.Run("cannot close streaming", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onCreateFile(suite.createFileStreamMock, nil)

		suite.createFileStreamMock.onSend(fiReq, nil)

		suite.crypterMock.onGenerateNonce(suite.testNonce, nil)

		suite.createFileStreamMock.onSend(nonceReq, nil)
		suite.crypterMock.onSealBytes(suite.testContent1, suite.testNonce, suite.testContent1)
		suite.createFileStreamMock.onSend(contentReq1, nil)
		suite.crypterMock.onSealBytes(suite.testContent2, suite.testNonce, suite.testContent2)
		suite.createFileStreamMock.onSend(contentReq2, nil)

		suite.createFileStreamMock.onCloseAndRecv(nil, errTest)

		err := suite.client.CreateFile(context.Background(), suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot close streaming")
	})

	suite.Run("positive test", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onCreateFile(suite.createFileStreamMock, nil)

		suite.createFileStreamMock.onSend(fiReq, nil)

		suite.crypterMock.onGenerateNonce(suite.testNonce, nil)

		suite.createFileStreamMock.onSend(nonceReq, nil)
		suite.crypterMock.onSealBytes(suite.testContent1, suite.testNonce, suite.testContent1)
		suite.createFileStreamMock.onSend(contentReq1, nil)
		suite.crypterMock.onSealBytes(suite.testContent2, suite.testNonce, suite.testContent2)
		suite.createFileStreamMock.onSend(contentReq2, nil)

		suite.createFileStreamMock.onCloseAndRecv(nil, nil)

		err := suite.client.CreateFile(context.Background(), suite.testName, suite.testPathToFile, suite.testMeta)
		require.NoError(err)
	})
}

func (suite *ClientTestSuite) TestUpdateFile() {
	suite.client.grpc = suite.clienMock

	require := suite.Require()

	csReq := &proto.GetChunkSizeResponse{
		Size: uint64(suite.testChunkSize),
	}

	fiReq := &proto.UpdateFileRequest{
		Data: &proto.UpdateFileRequest_FileInfo{
			FileInfo: &proto.File{
				Id:   suite.testFileID,
				Name: suite.testName,
				Meta: suite.testMeta,
			},
		},
	}

	nonceReq := &proto.UpdateFileRequest{
		Data: &proto.UpdateFileRequest_Content{
			Content: suite.testNonce,
		},
	}

	contentReq1 := &proto.UpdateFileRequest{
		Data: &proto.UpdateFileRequest_Content{
			Content: suite.testContent1,
		},
	}

	contentReq2 := &proto.UpdateFileRequest{
		Data: &proto.UpdateFileRequest_Content{
			Content: suite.testContent2,
		},
	}

	suite.Run("cannot get chunk size", func() {
		suite.clienMock.onGetChunkSize(nil, errTest)

		err := suite.client.UpdateFile(context.Background(), suite.testFileID, suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot get chunk size")
	})

	suite.Run("cannot seal file info", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, "", errTest)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, "", errTest)

		err := suite.client.UpdateFile(context.Background(), suite.testFileID, suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot seal file info")
		suite.ErrorContains(err, "cannot seal name")
		suite.ErrorContains(err, "cannot seal meta")
	})

	suite.Run("cannot start updating a file stream", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onUpdateFile(nil, errTest)
		err := suite.client.UpdateFile(context.Background(), suite.testFileID, suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot start updating a file stream")
	})

	suite.Run("cannot send file info", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onUpdateFile(suite.updateFileStreamMock, nil)

		suite.updateFileStreamMock.onSend(fiReq, errTest)

		err := suite.client.UpdateFile(context.Background(), suite.testFileID, suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot send file info")
	})

	suite.Run("cannot open file by path", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onUpdateFile(suite.updateFileStreamMock, nil)

		suite.updateFileStreamMock.onSend(fiReq, nil)

		err := suite.client.UpdateFile(context.Background(), suite.testFileID, suite.testName, "errPath", suite.testMeta)
		require.ErrorContains(err, "cannot open file by path")
	})

	suite.Run("cannot generate nonce", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onUpdateFile(suite.updateFileStreamMock, nil)

		suite.updateFileStreamMock.onSend(fiReq, nil)

		suite.crypterMock.onGenerateNonce(nil, errTest)

		err := suite.client.UpdateFile(context.Background(), suite.testFileID, suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot generate nonce")
	})

	suite.Run("cannot send nonce", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onUpdateFile(suite.updateFileStreamMock, nil)

		suite.updateFileStreamMock.onSend(fiReq, nil)

		suite.crypterMock.onGenerateNonce(suite.testNonce, nil)

		suite.updateFileStreamMock.onSend(nonceReq, errTest)

		err := suite.client.UpdateFile(context.Background(), suite.testFileID, suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot send nonce")
	})

	suite.Run("cannot send file data", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onUpdateFile(suite.updateFileStreamMock, nil)

		suite.updateFileStreamMock.onSend(fiReq, nil)

		suite.crypterMock.onGenerateNonce(suite.testNonce, nil)

		suite.updateFileStreamMock.onSend(nonceReq, nil)
		suite.crypterMock.onSealBytes(suite.testContent1, suite.testNonce, suite.testContent1).Once()
		suite.updateFileStreamMock.onSend(contentReq1, errTest).Once()

		err := suite.client.UpdateFile(context.Background(), suite.testFileID, suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot send file data")
	})

	suite.Run("cannot close streaming", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onUpdateFile(suite.updateFileStreamMock, nil)

		suite.updateFileStreamMock.onSend(fiReq, nil)

		suite.crypterMock.onGenerateNonce(suite.testNonce, nil)

		suite.updateFileStreamMock.onSend(nonceReq, nil)
		suite.crypterMock.onSealBytes(suite.testContent1, suite.testNonce, suite.testContent1)
		suite.updateFileStreamMock.onSend(contentReq1, nil)
		suite.crypterMock.onSealBytes(suite.testContent2, suite.testNonce, suite.testContent2)
		suite.updateFileStreamMock.onSend(contentReq2, nil)

		suite.updateFileStreamMock.onCloseAndRecv(nil, errTest)

		err := suite.client.UpdateFile(context.Background(), suite.testFileID, suite.testName, suite.testPathToFile, suite.testMeta)
		require.ErrorContains(err, "cannot close streaming")
	})

	suite.Run("positive test", func() {
		suite.clienMock.onGetChunkSize(csReq, nil)

		suite.crypterMock.onSealStringWithoutNonce(suite.testName, suite.testName, nil)
		suite.crypterMock.onSealStringWithoutNonce(suite.testMeta, suite.testMeta, nil)

		suite.clienMock.onUpdateFile(suite.updateFileStreamMock, nil)

		suite.updateFileStreamMock.onSend(fiReq, nil)

		suite.crypterMock.onGenerateNonce(suite.testNonce, nil)

		suite.updateFileStreamMock.onSend(nonceReq, nil)
		suite.crypterMock.onSealBytes(suite.testContent1, suite.testNonce, suite.testContent1)
		suite.updateFileStreamMock.onSend(contentReq1, nil)
		suite.crypterMock.onSealBytes(suite.testContent2, suite.testNonce, suite.testContent2)
		suite.updateFileStreamMock.onSend(contentReq2, nil)

		suite.updateFileStreamMock.onCloseAndRecv(nil, nil)

		err := suite.client.UpdateFile(context.Background(), suite.testFileID, suite.testName, suite.testPathToFile, suite.testMeta)
		require.NoError(err)
	})
}

func (suite *ClientTestSuite) TestGetFile() {
	suite.client.grpc = suite.clienMock
	require := suite.Require()

	gfReq := &proto.GetFileRequest{
		Id: suite.testFileID,
	}

	contentRes1 := &proto.GetFileResponse{
		Data: &proto.GetFileResponse_Content{
			Content: append(suite.testNonce, suite.testContent1...),
		},
	}

	contentRes2 := &proto.GetFileResponse{
		Data: &proto.GetFileResponse_Content{
			Content: suite.testContent2,
		},
	}

	suite.Run("cannot get file stream", func() {
		suite.clienMock.onGetFile(gfReq, nil, errTest)

		err := suite.client.GetFile(context.Background(), suite.testFileID, "")
		require.ErrorContains(err, "cannot get file stream")
	})

	suite.Run("cannot get file info", func() {
		suite.clienMock.onGetFile(gfReq, suite.getFileStreamMock, nil)
		suite.getFileStreamMock.onRecv(nil, errTest)

		err := suite.client.GetFile(context.Background(), suite.testFileID, "")
		require.ErrorContains(err, "cannot get file info")
	})

	suite.Run("cannot create file", func() {
		suite.clienMock.onGetFile(gfReq, suite.getFileStreamMock, nil)
		suite.getFileStreamMock.onRecv(nil, nil)

		err := suite.client.GetFile(context.Background(), suite.testFileID, "!#$@")
		require.ErrorContains(err, "cannot create file")
	})

	suite.Run("cannot receive nonce", func() {
		suite.clienMock.onGetFile(gfReq, suite.getFileStreamMock, nil)

		suite.getFileStreamMock.onRecv(nil, nil).Once()

		suite.crypterMock.onNonceSize(suite.testNonceSize)

		suite.getFileStreamMock.onRecv(nil, errTest).Once()

		err := suite.client.GetFile(context.Background(), suite.testFileID, suite.testPathToFileForGet)
		require.ErrorContains(err, "cannot receive nonce")

		err = os.Remove(filepath.Join(suite.testPathToFileForGet, suite.testFileID))
		require.NoError(err)
	})

	suite.Run("cannot get nonce", func() {
		suite.clienMock.onGetFile(gfReq, suite.getFileStreamMock, nil)

		suite.getFileStreamMock.onRecv(nil, nil).Once()

		suite.crypterMock.onNonceSize(suite.testNonceSize)

		suite.getFileStreamMock.onRecv(contentRes1, nil).Once()

		suite.crypterMock.onGetNonceFromBytes(
			contentRes1.GetContent(),
			suite.testNonceSize,
			crypto.AtFront,
			nil,
			nil,
			0,
			errTest,
		)

		err := suite.client.GetFile(context.Background(), suite.testFileID, suite.testPathToFileForGet)
		require.ErrorContains(err, "cannot get nonce")

		err = os.Remove(filepath.Join(suite.testPathToFileForGet, suite.testFileID))
		require.NoError(err)
	})

	suite.Run("cannot open content #1", func() {
		suite.clienMock.onGetFile(gfReq, suite.getFileStreamMock, nil)

		suite.getFileStreamMock.onRecv(nil, nil).Once()

		suite.crypterMock.onNonceSize(suite.testNonceSize)

		suite.getFileStreamMock.onRecv(contentRes1, nil).Once()

		suite.crypterMock.onGetNonceFromBytes(
			contentRes1.GetContent(),
			suite.testNonceSize,
			crypto.AtFront,
			suite.testNonce,
			suite.testContent1,
			0,
			nil,
		)

		suite.crypterMock.onOpenBytes(suite.testContent1, suite.testNonce, nil, errTest)

		err := suite.client.GetFile(context.Background(), suite.testFileID, suite.testPathToFileForGet)
		require.ErrorContains(err, "cannot open content")

		err = os.Remove(filepath.Join(suite.testPathToFileForGet, suite.testFileID))
		require.NoError(err)
	})

	suite.Run("cannot get content", func() {
		suite.clienMock.onGetFile(gfReq, suite.getFileStreamMock, nil)

		suite.getFileStreamMock.onRecv(nil, nil).Once()

		suite.crypterMock.onNonceSize(suite.testNonceSize)

		suite.getFileStreamMock.onRecv(contentRes1, nil).Once()

		suite.crypterMock.onGetNonceFromBytes(
			contentRes1.GetContent(),
			suite.testNonceSize,
			crypto.AtFront,
			suite.testNonce,
			suite.testContent1,
			0,
			nil,
		)
		suite.crypterMock.onOpenBytes(suite.testContent1, suite.testNonce, suite.testContent1, nil)
		suite.getFileStreamMock.onRecv(nil, errTest).Once()

		err := suite.client.GetFile(context.Background(), suite.testFileID, suite.testPathToFileForGet)
		require.ErrorContains(err, "cannot get content")

		err = os.Remove(filepath.Join(suite.testPathToFileForGet, suite.testFileID))
		require.NoError(err)
	})

	suite.Run("cannot open content #2", func() {
		suite.clienMock.onGetFile(gfReq, suite.getFileStreamMock, nil)

		suite.getFileStreamMock.onRecv(nil, nil).Once()

		suite.crypterMock.onNonceSize(suite.testNonceSize)

		suite.getFileStreamMock.onRecv(contentRes1, nil).Once()

		suite.crypterMock.onGetNonceFromBytes(
			contentRes1.GetContent(),
			suite.testNonceSize,
			crypto.AtFront,
			suite.testNonce,
			suite.testContent1,
			0,
			nil,
		)

		suite.crypterMock.onOpenBytes(suite.testContent1, suite.testNonce, suite.testContent1, nil).Once()
		suite.getFileStreamMock.onRecv(contentRes2, nil).Once()
		suite.crypterMock.onOpenBytes(suite.testContent2, suite.testNonce, nil, errTest).Once()

		err := suite.client.GetFile(context.Background(), suite.testFileID, suite.testPathToFileForGet)
		require.ErrorContains(err, "cannot open content")

		err = os.Remove(filepath.Join(suite.testPathToFileForGet, suite.testFileID))
		require.NoError(err)
	})

	suite.Run("positive test", func() {
		suite.clienMock.onGetFile(gfReq, suite.getFileStreamMock, nil)

		suite.getFileStreamMock.onRecv(nil, nil).Once()

		suite.crypterMock.onNonceSize(suite.testNonceSize)

		suite.getFileStreamMock.onRecv(contentRes1, nil).Once()

		suite.crypterMock.onGetNonceFromBytes(
			contentRes1.GetContent(),
			suite.testNonceSize,
			crypto.AtFront,
			suite.testNonce,
			suite.testContent1,
			0,
			nil,
		)

		suite.crypterMock.onOpenBytes(suite.testContent1, suite.testNonce, suite.testContent1, nil).Once()
		suite.getFileStreamMock.onRecv(contentRes2, nil).Once()
		suite.crypterMock.onOpenBytes(suite.testContent2, suite.testNonce, suite.testContent2, nil).Once()
		suite.getFileStreamMock.onRecv(nil, io.EOF).Once()

		err := suite.client.GetFile(context.Background(), suite.testFileID, suite.testPathToFileForGet)
		require.NoError(err)

		file, err := os.Open(filepath.Join(suite.testPathToFileForGet, suite.testFileID))
		require.NoError(err)

		scanner := bufio.NewScanner(file)
		require.True(scanner.Scan())
		gotBytes := scanner.Bytes()
		require.Equal(append(suite.testContent1, suite.testContent2...), gotBytes)

		err = os.Remove(filepath.Join(suite.testPathToFileForGet, suite.testFileID))
		require.NoError(err)
	})
}

func (suite *ClientTestSuite) TestDeleteFile() {
	require := suite.Require()

	req := &proto.DeleteFileRequest{
		Id: suite.testFileID,
	}

	suite.Run("service error", func() {
		suite.serverMock.onDeleteFile(req, errTest)

		err := suite.client.DeleteFile(context.Background(), suite.testFileID)
		require.ErrorContains(err, "cannot delete file")
	})

	suite.Run("positive test", func() {
		suite.serverMock.onDeleteFile(req, nil)

		err := suite.client.DeleteFile(context.Background(), suite.testFileID)
		require.NoError(err)
	})
}

func TestClientTestSuite(t *testing.T) {
	f, err := os.Create("./testdata/testfile")
	require.NoError(t, err)
	defer func() {
		err := f.Close()
		require.NoError(t, err)
	}()
	defer func() {
		err := os.Remove("./testdata/testfile")
		require.NoError(t, err)
	}()

	w := bufio.NewWriter(f)

	for i := 0; i < 1024*2; i++ {
		err := w.WriteByte(1)
		require.NoError(t, err)
	}

	err = w.Flush()
	require.NoError(t, err)

	suite.Run(t, new(ClientTestSuite))
}
