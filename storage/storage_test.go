//go:build integration

// Package storage определяет структуры и методы для работы с базой данных postgres
package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	emptyUUID = "00000000-0000-0000-0000-000000000000"
)

func cleanupTables(s *Storage) error {
	_, err := s.conn.Exec(context.Background(), "TRUNCATE TABLE users, banks, files, passwords, salts, texts;")
	if err != nil {
		return err
	}

	return nil
}

type StorageTestSuite struct {
	suite.Suite

	testStorage *Storage

	testLogin     string
	testLoginHash string
	testSalt      string
	testPassword  string
	testUserID    string

	testPWDName     string
	testPWDLogin    string
	testPWDPassword string
	testPWDMeta     string
	testPWDID       string
	testPWDUploadAt time.Time

	testFileName     string
	testFilePath     string
	testFileMeta     string
	testFileID       string
	testFileUploadAt time.Time

	testBankID       string
	testBankName     string
	testBankNumber   string
	testBankCVC      string
	testBankOwner    string
	testBankExp      string
	testBankMeta     string
	testBankUploadAt time.Time

	testTextID       string
	testTextName     string
	testTextText     string
	testTextMeta     string
	testTextUploadAt time.Time
}

func (suite *StorageTestSuite) SetupSuite() {
	require := suite.Require()
	dsn := os.Getenv("TEST_DSN")
	require.NotEmpty(dsn, "TEST_DSN environment variable is not set")
	s, err := NewStorage(context.Background(), dsn)
	require.NoError(err)
	suite.testStorage = s

	suite.testLogin = "TestLogin"
	suite.testLoginHash = "TestLoginHash"
	suite.testPassword = "TestPassword"
	suite.testSalt = "TestSalt"

	gotUD, err := suite.testStorage.CreateUser(
		context.Background(),
		suite.testLogin,
		suite.testLoginHash,
		suite.testSalt,
		suite.testPassword,
	)
	require.NoError(err)
	require.NotNil(gotUD)

	suite.testUserID = gotUD.ID

	suite.testPWDName = "TestPWDName"
	suite.testPWDLogin = "testPWDLogin"
	suite.testPWDPassword = "testPWDPassword"
	suite.testPWDMeta = "testPWDMeta"

	gotPD, err := suite.testStorage.CreatePassword(
		context.Background(),
		suite.testUserID, suite.testPWDName,
		suite.testPWDLogin,
		suite.testPWDPassword,
		suite.testPWDMeta,
	)
	require.NoError(err)
	require.NotNil(gotPD)

	suite.testPWDID = gotPD.ID
	suite.testPWDUploadAt = gotPD.UpdateAt

	suite.testFileName = "TestFileName"
	suite.testFilePath = "testFilePath"
	suite.testFileMeta = "testFileMeta"

	gotFile, err := suite.testStorage.CreateFile(
		context.Background(),
		suite.testUserID,
		suite.testFileName,
		suite.testFilePath,
		suite.testFileMeta,
	)
	require.NoError(err)
	require.NotNil(gotFile)

	suite.testFileID = gotFile.ID
	suite.testFileUploadAt = gotFile.UpdateAt

	suite.testBankName = "testBankName"
	suite.testBankNumber = "testBankNumber"
	suite.testBankCVC = "testBankCVC"
	suite.testBankOwner = "testBankOwner"
	suite.testBankExp = "testBankExp"
	suite.testBankMeta = "testBankMeta"

	gotBank, err := suite.testStorage.CreateBank(
		context.Background(),
		suite.testUserID,
		suite.testBankName,
		suite.testBankNumber,
		suite.testBankCVC,
		suite.testBankOwner,
		suite.testBankExp,
		suite.testBankMeta,
	)
	require.NoError(err)
	require.NotNil(gotBank)

	suite.testBankID = gotBank.ID
	suite.testBankUploadAt = gotBank.UpdateAt

	suite.testTextName = "testTextName"
	suite.testTextText = "testTextText"
	suite.testTextMeta = "testTextMeta"

	gotText, err := suite.testStorage.CreateText(
		context.Background(),
		suite.testUserID,
		suite.testTextName,
		suite.testTextText,
		suite.testTextMeta,
	)

	require.NoError(err)
	require.NotNil(gotText)

	suite.testTextID = gotText.ID
	suite.testTextUploadAt = gotText.UpdateAt
}

func (suite *StorageTestSuite) TearDownSuite() {
	err := cleanupTables(suite.testStorage)
	suite.Require().NoError(err)
	suite.testStorage.Close()
}

func (suite *StorageTestSuite) TestNewStorage() {
	_, err := NewStorage(context.Background(), "errorDSN")

	suite.Require().ErrorContains(err, "create pgxpool")
}

func (suite *StorageTestSuite) TestCreateUser() {
	require := suite.Require()
	suite.Run("create with empty login", func() {
		gotUD, err := suite.testStorage.CreateUser(context.Background(), "", "Test", "Test", "Test")
		require.ErrorContains(err, "insert into users table login")
		require.Nil(gotUD)
	})

	suite.Run("create with empty password", func() {
		gotUD, err := suite.testStorage.CreateUser(context.Background(), "Test", "Test", "Test", "")
		require.ErrorContains(err, "insert into users table login")
		require.Nil(gotUD)
	})

	suite.Run("create with empty salt", func() {
		gotUD, err := suite.testStorage.CreateUser(context.Background(), "Test", "Test", "", "Test")
		require.ErrorContains(err, "insert into salts table login")
		require.Nil(gotUD)
	})

	suite.Run("create with empty login hashed", func() {
		gotUD, err := suite.testStorage.CreateUser(context.Background(), "Test", "", "Test", "Test")
		require.ErrorContains(err, "insert into salts table login")
		require.Nil(gotUD)
	})

	suite.Run("create dublicate login", func() {
		gotUD, err := suite.testStorage.CreateUser(context.Background(), suite.testLogin, suite.testLoginHash, suite.testSalt, suite.testPassword)
		require.ErrorIs(err, ErrUserAlreadyExists)
		require.Nil(gotUD)
	})
}

func (suite *StorageTestSuite) TestGetUser() {
	require := suite.Require()
	suite.Run("get user test", func() {
		gotUD, err := suite.testStorage.GetUser(context.Background(), suite.testLogin, suite.testLoginHash)

		require.NoError(err)
		require.Equal(&User{
			ID:       suite.testUserID,
			Login:    suite.testLogin,
			Password: suite.testPassword,
			Salt:     suite.testSalt,
		}, gotUD)
	})

	suite.Run("unknown user test", func() {
		_, err := suite.testStorage.GetUser(context.Background(), "Test", "Test")
		require.ErrorIs(err, ErrUserNotFound)
	})
}

func (suite *StorageTestSuite) TestCreatePassword() {
	suite.Run("unknown user", func() {
		gotPassword, err := suite.testStorage.CreatePassword(
			context.Background(),
			emptyUUID,
			"Test",
			"Test",
			"Test",
			"Test",
		)
		suite.Require().ErrorIs(err, ErrUserNotFound)
		suite.Require().Nil(gotPassword)
	})
}

func (suite *StorageTestSuite) TestUpdatePassword() {
	require := suite.Require()

	suite.Run("update existing password", func() {
		newName := "NewTestName"
		newLogin := "NewTestLogin"
		newPassword := "NewTestPassword"
		newMeta := "NewTestMeta"

		gotPWD, err := suite.testStorage.UpdatePassword(
			context.Background(),
			suite.testPWDID,
			suite.testUserID,
			newName,
			newLogin,
			newPassword,
			newMeta,
		)
		require.NoError(err)

		updatedPassword, err := suite.testStorage.GetPassword(context.Background(), suite.testPWDID, suite.testUserID)
		require.NoError(err)

		require.Equal(gotPWD, updatedPassword)
	})

	suite.Run("update non-existing password", func() {
		gotPWD, err := suite.testStorage.UpdatePassword(
			context.Background(),
			emptyUUID,
			suite.testUserID,
			"Test",
			"Test",
			"Test",
			"Test",
		)
		require.ErrorIs(err, ErrPasswordNotFound)
		require.Nil(gotPWD)
	})

	suite.Run("update non-existing user", func() {
		gotPWD, err := suite.testStorage.UpdatePassword(
			context.Background(),
			suite.testPWDID,
			emptyUUID,
			"Test",
			"Test",
			"Test",
			"Test",
		)
		require.ErrorIs(err, ErrUserNotFound)
		require.Nil(gotPWD)
	})
}

func (suite *StorageTestSuite) TestGetPassword() {
	require := suite.Require()

	suite.Run("positive test", func() {
		gotPassword, err := suite.testStorage.GetPassword(context.Background(), suite.testPWDID, suite.testUserID)
		require.NoError(err)
		require.Equal(&Password{
			ID:       suite.testPWDID,
			UserID:   suite.testUserID,
			Name:     suite.testPWDName,
			Login:    suite.testPWDLogin,
			Password: suite.testPWDPassword,
			Meta:     suite.testPWDMeta,
			UpdateAt: suite.testPWDUploadAt,
		}, gotPassword)
	})

	suite.Run("unknown id", func() {
		gotPassword, err := suite.testStorage.GetPassword(context.Background(), emptyUUID, suite.testUserID)
		require.ErrorIs(err, ErrPasswordNotFound)
		require.Nil(gotPassword)
	})
}

func (suite *StorageTestSuite) TestCreateFile() {
	require := suite.Require()
	suite.Run("unknown user", func() {
		gotPassword, err := suite.testStorage.CreateFile(
			context.Background(),
			emptyUUID,
			"Test",
			"Test",
			"Test",
		)
		require.ErrorIs(err, ErrUserNotFound)
		require.Nil(gotPassword)
	})
}

func (suite *StorageTestSuite) TestGetAllPassword() {
	require := suite.Require()

	suite.Run("positive test", func() {
		gotPWDs, err := suite.testStorage.GetAllPassword(context.Background(), suite.testUserID)
		require.NoError(err)
		require.Equal([]Password{
			{
				ID:       suite.testPWDID,
				UserID:   suite.testUserID,
				Name:     suite.testPWDName,
				Login:    suite.testPWDLogin,
				Password: suite.testPWDPassword,
				Meta:     suite.testPWDMeta,
				UpdateAt: suite.testPWDUploadAt,
			},
		}, gotPWDs)
	})

	suite.Run("unknown user_id", func() {
		gotPWDs, err := suite.testStorage.GetAllPassword(context.Background(), emptyUUID)
		require.NoError(err)
		require.Empty(gotPWDs)
	})
}

func (suite *StorageTestSuite) TestDeletePassword() {
	require := suite.Require()

	suite.Run("delete existing password", func() {
		newPassword, err := suite.testStorage.CreatePassword(
			context.Background(),
			suite.testUserID,
			"DeleteTestName",
			"DeleteTestLogin",
			"DeleteTestPassword",
			"DeleteTestMeta",
		)
		require.NoError(err)
		require.NotNil(newPassword)

		err = suite.testStorage.DeletePassword(context.Background(), newPassword.ID, newPassword.UserID)
		require.NoError(err)

		deletedPassword, err := suite.testStorage.GetPassword(context.Background(), newPassword.ID, suite.testUserID)
		require.ErrorIs(err, ErrPasswordNotFound)
		require.Nil(deletedPassword)
	})

	suite.Run("delete non-existing passwword", func() {
		err := suite.testStorage.DeletePassword(context.Background(), emptyUUID, suite.testUserID)
		require.ErrorIs(err, ErrPasswordNotFound)
	})
}

func (suite *StorageTestSuite) TestGetFile() {
	require := suite.Require()

	suite.Run("positive test", func() {
		gotFile, err := suite.testStorage.GetFile(context.Background(), suite.testFileID, suite.testUserID)
		require.NoError(err)
		require.Equal(&File{
			ID:         suite.testFileID,
			UserID:     suite.testUserID,
			Name:       suite.testFileName,
			PathToFile: suite.testFilePath,
			Meta:       suite.testFileMeta,
			UpdateAt:   suite.testFileUploadAt,
		}, gotFile)
	})

	suite.Run("unknown id", func() {
		gotPassword, err := suite.testStorage.GetFile(context.Background(), emptyUUID, suite.testUserID)
		require.ErrorIs(err, ErrFileNotFound)
		require.Nil(gotPassword)
	})
}

func (suite *StorageTestSuite) TestUpdateFile() {
	require := suite.Require()

	suite.Run("update existing file", func() {
		newName := "NewTestFileName"
		newPath := "NewTestFilePath"
		newMeta := "NewTestFileMeta"

		gotFile, err := suite.testStorage.UpdateFile(
			context.Background(),
			suite.testFileID,
			suite.testUserID,
			newName,
			newPath,
			newMeta,
		)
		require.NoError(err)

		updatedFile, err := suite.testStorage.GetFile(context.Background(), suite.testFileID, suite.testUserID)
		require.NoError(err)

		require.Equal(&File{
			ID:         gotFile.ID,
			UserID:     gotFile.UserID,
			Name:       gotFile.Name,
			PathToFile: newPath,
			Meta:       gotFile.Meta,
			UpdateAt:   gotFile.UpdateAt,
		}, updatedFile)
	})

	suite.Run("update non-existing file", func() {
		gotFile, err := suite.testStorage.UpdateFile(
			context.Background(),
			emptyUUID,
			suite.testUserID,
			"Test",
			"Test",
			"Test",
		)
		require.ErrorIs(err, ErrFileNotFound)
		require.Nil(gotFile)
	})

	suite.Run("update non-existing user", func() {
		gotFile, err := suite.testStorage.UpdateFile(
			context.Background(),
			suite.testFileID,
			emptyUUID,
			"Test",
			"Test",
			"Test",
		)
		require.ErrorIs(err, ErrUserNotFound)
		require.Nil(gotFile)
	})
}

func (suite *StorageTestSuite) TestGetAllFiles() {
	require := suite.Require()

	suite.Run("positive test", func() {
		gotFiles, err := suite.testStorage.GetAllFiles(context.Background(), suite.testUserID)
		require.NoError(err)
		require.Equal([]File{
			{
				ID:         suite.testFileID,
				UserID:     suite.testUserID,
				Name:       suite.testFileName,
				PathToFile: suite.testFilePath,
				Meta:       suite.testFileMeta,
				UpdateAt:   suite.testFileUploadAt,
			},
		}, gotFiles)
	})

	suite.Run("unknown user_id", func() {
		gotFiles, err := suite.testStorage.GetAllFiles(context.Background(), emptyUUID)
		require.NoError(err)
		require.Empty(gotFiles)
	})
}

func (suite *StorageTestSuite) TestDeleteFile() {
	require := suite.Require()

	suite.Run("delete existing file", func() {
		newFile, err := suite.testStorage.CreateFile(
			context.Background(),
			suite.testUserID,
			"DeleteTestFileName",
			"DeleteTestFilePath",
			"DeleteTestFileMeta",
		)
		require.NoError(err)
		require.NotNil(newFile)

		file, err := suite.testStorage.DeleteFile(context.Background(), newFile.ID, newFile.UserID)
		require.NoError(err)
		require.Equal(&File{
			ID:         newFile.ID,
			UserID:     suite.testUserID,
			Name:       "DeleteTestFileName",
			PathToFile: "DeleteTestFilePath",
			Meta:       "DeleteTestFileMeta",
			UpdateAt:   newFile.UpdateAt,
		}, file)

		deletedFile, err := suite.testStorage.GetFile(context.Background(), newFile.ID, suite.testUserID)
		require.ErrorIs(err, ErrFileNotFound)
		require.Nil(deletedFile)
	})

	suite.Run("delete non-existing file", func() {
		file, err := suite.testStorage.DeleteFile(context.Background(), emptyUUID, suite.testUserID)
		require.ErrorIs(err, ErrFileNotFound)
		require.Nil(file)
	})
}

func (suite *StorageTestSuite) TestCreateBank() {
	require := suite.Require()

	suite.Run("unknown user", func() {
		gotPassword, err := suite.testStorage.CreateBank(
			context.Background(),
			emptyUUID,
			"Test",
			"Test",
			"Test",
			"Test",
			"Test",
			"Test",
		)
		require.ErrorIs(err, ErrUserNotFound)
		require.Nil(gotPassword)
	})
}

func (suite *StorageTestSuite) TestUpdateBank() {
	require := suite.Require()

	suite.Run("update existing bank", func() {
		newName := "NewTestBankName"
		newNumber := "NewTestBankNumber"
		newCVC := "NewTestBankCVC"
		newOwner := "NewTestBankOwner"
		newExp := "NewTestBankExp"
		newMeta := "NewTestBankMeta"

		gotBank, err := suite.testStorage.UpdateBank(
			context.Background(),
			suite.testBankID,
			suite.testUserID,
			newName,
			newNumber,
			newCVC,
			newOwner,
			newExp,
			newMeta,
		)
		require.NoError(err)

		updatedBank, err := suite.testStorage.GetBank(context.Background(), suite.testBankID, suite.testUserID)
		require.NoError(err)

		require.Equal(&Bank{
			ID:         gotBank.ID,
			UserID:     gotBank.UserID,
			Name:       newName,
			CardNumber: newNumber,
			CVC:        newCVC,
			Owner:      newOwner,
			Exp:        newExp,
			Meta:       newMeta,
			UpdateAt:   gotBank.UpdateAt,
		}, updatedBank)
	})

	suite.Run("update non-existing bank", func() {
		gotBank, err := suite.testStorage.UpdateBank(
			context.Background(),
			emptyUUID,
			suite.testUserID,
			"Test",
			"Test",
			"Test",
			"Test",
			"Test",
			"Test",
		)
		require.ErrorIs(err, ErrBankNotFound)
		require.Nil(gotBank)
	})

	suite.Run("update non-existing user", func() {
		gotBank, err := suite.testStorage.UpdateBank(
			context.Background(),
			suite.testBankID,
			emptyUUID,
			"Test",
			"Test",
			"Test",
			"Test",
			"Test",
			"Test",
		)
		require.ErrorIs(err, ErrUserNotFound)
		require.Nil(gotBank)
	})
}

func (suite *StorageTestSuite) TestGetBank() {
	require := suite.Require()

	suite.Run("positive test", func() {
		gotBank, err := suite.testStorage.GetBank(context.Background(), suite.testBankID, suite.testUserID)
		require.NoError(err)
		require.Equal(&Bank{
			ID:         suite.testBankID,
			UserID:     suite.testUserID,
			Name:       suite.testBankName,
			CardNumber: suite.testBankNumber,
			CVC:        suite.testBankCVC,
			Owner:      suite.testBankOwner,
			Meta:       suite.testBankMeta,
			Exp:        suite.testBankExp,
			UpdateAt:   suite.testBankUploadAt,
		}, gotBank)
	})

	suite.Run("unknown id", func() {
		gotBank, err := suite.testStorage.GetBank(context.Background(), emptyUUID, suite.testUserID)
		require.ErrorIs(err, ErrBankNotFound)
		require.Nil(gotBank)
	})
}

func (suite *StorageTestSuite) TestGetAllBanks() {
	require := suite.Require()

	suite.Run("positive test", func() {
		gotBanks, err := suite.testStorage.GetAllBanks(context.Background(), suite.testUserID)
		require.NoError(err)
		require.Equal([]Bank{
			{
				ID:         suite.testBankID,
				UserID:     suite.testUserID,
				Name:       suite.testBankName,
				CardNumber: suite.testBankNumber,
				CVC:        suite.testBankCVC,
				Owner:      suite.testBankOwner,
				Meta:       suite.testBankMeta,
				Exp:        suite.testBankExp,
				UpdateAt:   suite.testBankUploadAt,
			},
		}, gotBanks)
	})

	suite.Run("unknown user_id", func() {
		gotBanks, err := suite.testStorage.GetAllBanks(context.Background(), emptyUUID)
		require.NoError(err)
		require.Empty(gotBanks)
	})
}

func (suite *StorageTestSuite) TestDeleteBank() {
	require := suite.Require()

	suite.Run("delete existing bank", func() {
		newBank, err := suite.testStorage.CreateBank(
			context.Background(),
			suite.testUserID,
			"DeleteTestBankName",
			"DeleteTestBankNumber",
			"DeleteTestBankCVC",
			"DeleteTestBankOwner",
			"DeleteTestBankExp",
			"DeleteTestBankMeta",
		)
		require.NoError(err)
		require.NotNil(newBank)

		err = suite.testStorage.DeleteBank(context.Background(), newBank.ID, newBank.UserID)
		require.NoError(err)

		deletedBank, err := suite.testStorage.GetBank(context.Background(), newBank.ID, suite.testUserID)
		require.ErrorIs(err, ErrBankNotFound)
		require.Nil(deletedBank)
	})

	suite.Run("delete non-existing bank", func() {
		err := suite.testStorage.DeleteBank(context.Background(), emptyUUID, suite.testUserID)
		require.ErrorIs(err, ErrBankNotFound)
	})
}

func (suite *StorageTestSuite) TestCreateText() {
	require := suite.Require()

	suite.Run("unknown user", func() {
		gotPassword, err := suite.testStorage.CreateText(
			context.Background(),
			emptyUUID,
			"Test",
			"Test",
			"Test",
		)
		require.ErrorIs(err, ErrUserNotFound)
		require.Nil(gotPassword)
	})
}

func (suite *StorageTestSuite) TestUpdateText() {
	require := suite.Require()

	suite.Run("update existing text", func() {
		newName := "NewTestTextName"
		newText := "NewTestTextData"
		newMeta := "NewTestTextMeta"

		gotText, err := suite.testStorage.UpdateText(
			context.Background(),
			suite.testTextID,
			suite.testUserID,
			newName,
			newText,
			newMeta,
		)
		require.NoError(err)

		updatedText, err := suite.testStorage.GetText(context.Background(), suite.testTextID, suite.testUserID)
		require.NoError(err)

		require.Equal(&Text{
			ID:       gotText.ID,
			UserID:   gotText.UserID,
			Name:     newName,
			Text:     newText,
			Meta:     newMeta,
			UpdateAt: gotText.UpdateAt,
		}, updatedText)
	})

	suite.Run("update non-existing text", func() {
		gotText, err := suite.testStorage.UpdateText(
			context.Background(),
			emptyUUID,
			suite.testUserID,
			"Test",
			"Test",
			"Test",
		)
		require.ErrorIs(err, ErrTextNotFound)
		require.Nil(gotText)
	})

	suite.Run("update non-existing user", func() {
		gotText, err := suite.testStorage.UpdateText(
			context.Background(),
			suite.testTextID,
			emptyUUID,
			"Test",
			"Test",
			"Test",
		)
		require.ErrorIs(err, ErrUserNotFound)
		require.Nil(gotText)
	})
}

func (suite *StorageTestSuite) TestGetText() {
	require := suite.Require()

	suite.Run("positive test", func() {
		gotText, err := suite.testStorage.GetText(context.Background(), suite.testTextID, suite.testUserID)
		require.NoError(err)
		require.Equal(&Text{
			ID:       suite.testTextID,
			UserID:   suite.testUserID,
			Name:     suite.testTextName,
			Text:     suite.testTextText,
			Meta:     suite.testTextMeta,
			UpdateAt: suite.testTextUploadAt,
		}, gotText)
	})

	suite.Run("unknown id", func() {
		gotText, err := suite.testStorage.GetText(context.Background(), emptyUUID, suite.testUserID)
		require.ErrorIs(err, ErrTextNotFound)
		require.Nil(gotText)
	})
}

func (suite *StorageTestSuite) TestGetAllTexts() {
	require := suite.Require()
	suite.Run("positive test", func() {
		gotTexts, err := suite.testStorage.GetAllTexts(context.Background(), suite.testUserID)
		require.NoError(err)
		require.Equal([]Text{
			{
				ID:       suite.testTextID,
				UserID:   suite.testUserID,
				Name:     suite.testTextName,
				Text:     suite.testTextText,
				Meta:     suite.testTextMeta,
				UpdateAt: suite.testTextUploadAt,
			},
		}, gotTexts)
	})

	suite.Run("unknown user_id", func() {
		gotTexts, err := suite.testStorage.GetAllTexts(context.Background(), emptyUUID)
		require.NoError(err)
		require.Empty(gotTexts)
	})
}

func (suite *StorageTestSuite) TestDeleteText() {
	require := suite.Require()

	suite.Run("delete existing text", func() {
		newText, err := suite.testStorage.CreateText(
			context.Background(),
			suite.testUserID,
			"DeleteTestTextName",
			"DeleteTestTextData",
			"DeleteTestTextMeta",
		)
		require.NoError(err)
		require.NotNil(newText)

		err = suite.testStorage.DeleteText(context.Background(), newText.ID, newText.UserID)
		require.NoError(err)

		deletedText, err := suite.testStorage.GetText(context.Background(), newText.ID, suite.testUserID)
		require.ErrorIs(err, ErrTextNotFound)
		require.Nil(deletedText)
	})

	suite.Run("delete non-existing text", func() {
		err := suite.testStorage.DeleteText(context.Background(), emptyUUID, suite.testUserID)
		require.ErrorIs(err, ErrTextNotFound)
	})
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}
