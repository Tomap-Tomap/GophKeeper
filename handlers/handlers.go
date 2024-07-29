// Package handlers определяет методы и структуры для работы grpc сервера
package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/Tomap-Tomap/GophKeeper/proto"
	"github.com/Tomap-Tomap/GophKeeper/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Storage описывает методы для работы с неким хранилищем
type Storage interface {
	CreateUser(ctx context.Context, login, loginHashed, salt, password string) (*storage.User, error)
	GetUser(ctx context.Context, login, loginHashed string) (*storage.User, error)

	CreatePassword(ctx context.Context, userID, name, login, password, meta string) (*storage.Password, error)
	GetPassword(ctx context.Context, passwordID string) (*storage.Password, error)
	GetAllPassword(ctx context.Context, userID string) ([]storage.Password, error)

	CreateFile(ctx context.Context, userID, name, pathToFile, meta string) (*storage.File, error)
	GetFile(ctx context.Context, fileID string) (*storage.File, error)
	GetAllFiles(ctx context.Context, userID string) ([]storage.File, error)

	CreateBank(ctx context.Context, userID, name, banksData, meta string) (*storage.Bank, error)
	GetBank(ctx context.Context, bankID string) (*storage.Bank, error)
	GetAllBanks(ctx context.Context, userID string) ([]storage.Bank, error)

	CreateText(ctx context.Context, userID, name, text, meta string) (*storage.Text, error)
	GetText(ctx context.Context, textID string) (*storage.Text, error)
	GetAllTexts(ctx context.Context, userID string) ([]storage.Text, error)
}

// Hasher описывает методы для генерации хэшей
type Hasher interface {
	GenerateSalt() (string, error)
	GetHash(str string) (string, error)
	GetHashWithSalt(str, salt string) (string, error)
}

// Tokener описывает методы для генерации токенов
type Tokener interface {
	GetToken(sub string) (string, error)
}

// FileStore описывает методы для сохранения файла на сервере
type FileStore interface {
	Save(content bytes.Buffer) (string, error)
	GetDBFiler(pathToFile string) (DBFiler, error)
}

// DBFiler интерфейс описывающий объекты для чтения файлов
type DBFiler interface {
	GetChunck() ([]byte, error)
	Close()
}

// GophKeeperHandler структура реализующая работу grpc сервера
type GophKeeperHandler struct {
	proto.UnimplementedGophKeeperServer

	rp storage.RetryPolicy
	s  Storage
	h  Hasher
	t  Tokener
	fs FileStore
}

// NewGophKeeperHandler иницирует структуру GophKeeperHandler
func NewGophKeeperHandler(s Storage, h Hasher, t Tokener, fs FileStore) *GophKeeperHandler {
	return &GophKeeperHandler{
		s:  s,
		h:  h,
		rp: *storage.NewRetryPolicy(1, 1, 1),
		t:  t,
		fs: fs,
	}
}

// Register обработчик регистрации пользователя
func (gk *GophKeeperHandler) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	login := strings.TrimSpace(req.GetLogin())

	var err error

	if login == "" {
		err = errors.Join(status.Error(codes.InvalidArgument, "empty login"))
	}

	password := strings.TrimSpace(req.GetPassword())

	if password == "" {
		err = errors.Join(err, status.Error(codes.InvalidArgument, "empty password"))
	}

	if err != nil {
		return nil, err
	}

	loginHash, err := gk.h.GetHash(login)

	if err != nil {
		return nil, status.Error(codes.Internal, "generate hash")
	}

	salt, err := gk.h.GenerateSalt()

	if err != nil {
		return nil, status.Error(codes.Internal, "generate salt")
	}

	passwordHash, err := gk.h.GetHashWithSalt(password, salt)

	if err != nil {
		return nil, status.Error(codes.Internal, "generate hash")
	}

	user, err := storage.Retry2(ctx, gk.rp, func() (*storage.User, error) {
		return gk.s.CreateUser(ctx, login, loginHash, salt, passwordHash)
	})

	if err != nil {
		switch {
		case storage.IsUniqueViolation(err):
			return nil, status.Errorf(codes.AlreadyExists, "user %s already exists", login)
		default:
			return nil, status.Errorf(codes.Internal, "create user %s", login)
		}
	}

	token, err := gk.t.GetToken(user.ID)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "gen token for user %s", login)
	}

	return &proto.RegisterResponse{Token: token}, nil
}

// Auth обработчик аунтефикации пользователя
func (gk *GophKeeperHandler) Auth(ctx context.Context, req *proto.AuthRequest) (*proto.AuthResponse, error) {
	login := strings.TrimSpace(req.GetLogin())

	var err error

	if login == "" {
		err = errors.Join(status.Error(codes.InvalidArgument, "empty login"))
	}

	password := strings.TrimSpace(req.GetPassword())

	if password == "" {
		err = errors.Join(err, status.Error(codes.InvalidArgument, "empty password"))
	}

	if err != nil {
		return nil, err
	}

	loginHash, err := gk.h.GetHash(login)

	if err != nil {
		return nil, status.Error(codes.Internal, "generate hash")
	}

	user, err := storage.Retry2(ctx, gk.rp, func() (*storage.User, error) {
		return gk.s.GetUser(ctx, login, loginHash)
	})

	if err != nil {
		switch {
		case storage.IsNowRowError(err):
			return nil, status.Errorf(codes.Unknown, "unknown user %s", login)
		default:
			return nil, status.Errorf(codes.Internal, "get user %s", login)
		}
	}

	hash, err := gk.h.GetHashWithSalt(password, user.Salt)

	if err != nil {
		return nil, status.Error(codes.Internal, "generate hash")
	}

	if hash != user.Password {
		return nil, status.Error(codes.PermissionDenied, "invalid password")
	}

	token, err := gk.t.GetToken(user.ID)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "gen token for user %s", login)
	}

	return &proto.AuthResponse{Token: token}, nil
}

// CreatePassword обработчик добавления нового сохраненного пароля пользователя
func (gk *GophKeeperHandler) CreatePassword(ctx context.Context, req *proto.CreatePasswordRequest) (*proto.CreatePasswordResponse, error) {
	userID := strings.TrimSpace(req.GetUserID())

	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty UserID")
	}

	pwd, err := storage.Retry2(ctx, gk.rp, func() (*storage.Password, error) {
		return gk.s.CreatePassword(
			ctx,
			userID,
			req.Name,
			req.Login,
			req.Password,
			req.Meta,
		)
	})

	if err != nil {
		switch {
		case storage.IsForeignKeyViolation(err):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return nil, status.Errorf(codes.Internal, "create password for user %s", userID)
		}
	}

	return &proto.CreatePasswordResponse{Id: pwd.ID}, nil
}

// GetPassword обработчик получения данных пароля пользователя
func (gk *GophKeeperHandler) GetPassword(ctx context.Context, req *proto.GetPasswordRequest) (*proto.GetPasswordResponse, error) {
	passwordID := strings.TrimSpace(req.GetId())

	if passwordID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty PasswordID")
	}

	pwd, err := storage.Retry2(ctx, gk.rp, func() (*storage.Password, error) {
		return gk.s.GetPassword(
			ctx,
			passwordID,
		)
	})

	if err != nil {
		switch {
		case storage.IsForeignKeyViolation(err):
			return nil, status.Errorf(codes.Unknown, "unknown PasswordID %s", passwordID)
		default:
			return nil, status.Errorf(codes.Internal, "get password %s", passwordID)
		}
	}

	return &proto.GetPasswordResponse{
		Password: &proto.Password{
			Id:       pwd.ID,
			UserID:   pwd.UserID,
			Name:     pwd.Name,
			Login:    pwd.Login,
			Password: pwd.Password,
			Meta:     pwd.Meta,
			UpdateAt: timestamppb.New(pwd.UpdateAt),
		},
	}, nil
}

// GetPasswords обработчик получения всех паролей пользователя
func (gk *GophKeeperHandler) GetPasswords(ctx context.Context, req *proto.GetPasswordsRequest) (*proto.GetPasswordsResponse, error) {
	userID := strings.TrimSpace(req.GetUserID())

	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty UserID")
	}

	pwds, err := storage.Retry2(ctx, gk.rp, func() ([]storage.Password, error) {
		return gk.s.GetAllPassword(
			ctx,
			userID,
		)
	})

	if err != nil {
		switch {
		case storage.IsForeignKeyViolation(err):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return nil, status.Errorf(codes.Internal, "get passwords %s", userID)
		}
	}

	protoPWDs := make([]*proto.Password, 0, len(pwds))

	for _, val := range pwds {
		protoPWDs = append(protoPWDs, &proto.Password{
			Id:       val.ID,
			UserID:   val.UserID,
			Name:     val.Name,
			Login:    val.Login,
			Password: val.Password,
			Meta:     val.Meta,
			UpdateAt: timestamppb.New(val.UpdateAt),
		})
	}

	return &proto.GetPasswordsResponse{
		Passwords: protoPWDs,
	}, nil
}

// CreateBank обработчик добавления новой банковской информации
func (gk *GophKeeperHandler) CreateBank(ctx context.Context, req *proto.CreateBankRequest) (*proto.CreateBankResponse, error) {
	userID := strings.TrimSpace(req.GetUserID())

	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty UserID")
	}

	bank, err := storage.Retry2(ctx, gk.rp, func() (*storage.Bank, error) {
		return gk.s.CreateBank(
			ctx,
			userID,
			req.Name,
			req.BanksData,
			req.Meta,
		)
	})

	if err != nil {
		switch {
		case storage.IsForeignKeyViolation(err):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return nil, status.Errorf(codes.Internal, "create bank data for user %s", userID)
		}
	}

	return &proto.CreateBankResponse{Id: bank.ID}, nil
}

// GetBank обработчик получения банковских данных пользователя
func (gk *GophKeeperHandler) GetBank(ctx context.Context, req *proto.GetBankRequest) (*proto.GetBankResponse, error) {
	bankID := strings.TrimSpace(req.GetId())

	if bankID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty BankID")
	}

	bank, err := storage.Retry2(ctx, gk.rp, func() (*storage.Bank, error) {
		return gk.s.GetBank(
			ctx,
			bankID,
		)
	})

	if err != nil {
		switch {
		case storage.IsForeignKeyViolation(err):
			return nil, status.Errorf(codes.Unknown, "unknown BankID %s", bankID)
		default:
			return nil, status.Errorf(codes.Internal, "get bank data %s", bankID)
		}
	}

	return &proto.GetBankResponse{
		Bank: &proto.Bank{
			Id:        bank.ID,
			UserID:    bank.UserID,
			Name:      bank.Name,
			BanksData: bank.BanksData,
			Meta:      bank.Meta,
			UpdateAt:  timestamppb.New(bank.UpdateAt),
		},
	}, nil
}

// GetBanks обработчик получения всех баковских данных
func (gk *GophKeeperHandler) GetBanks(ctx context.Context, req *proto.GetBanksRequest) (*proto.GetBanksResponse, error) {
	userID := strings.TrimSpace(req.GetUserID())

	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty UserID")
	}

	banks, err := storage.Retry2(ctx, gk.rp, func() ([]storage.Bank, error) {
		return gk.s.GetAllBanks(
			ctx,
			userID,
		)
	})

	if err != nil {
		switch {
		case storage.IsForeignKeyViolation(err):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return nil, status.Errorf(codes.Internal, "get banks %s", userID)
		}
	}

	protoBanks := make([]*proto.Bank, 0, len(banks))

	for _, val := range banks {
		protoBanks = append(protoBanks, &proto.Bank{
			Id:        val.ID,
			UserID:    val.UserID,
			Name:      val.Name,
			BanksData: val.BanksData,
			Meta:      val.Meta,
			UpdateAt:  timestamppb.New(val.UpdateAt),
		})
	}

	return &proto.GetBanksResponse{
		Banks: protoBanks,
	}, nil
}

// CreateText обработчик добавления новой банковской информации
func (gk *GophKeeperHandler) CreateText(ctx context.Context, req *proto.CreateTextRequest) (*proto.CreateTextResponse, error) {
	userID := strings.TrimSpace(req.GetUserID())

	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty UserID")
	}

	text, err := storage.Retry2(ctx, gk.rp, func() (*storage.Text, error) {
		return gk.s.CreateText(
			ctx,
			userID,
			req.Name,
			req.Text,
			req.Meta,
		)
	})

	if err != nil {
		switch {
		case storage.IsForeignKeyViolation(err):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return nil, status.Errorf(codes.Internal, "create text for user %s", userID)
		}
	}

	return &proto.CreateTextResponse{Id: text.ID}, nil
}

// GetText обработчик получения банковских данных пользователя
func (gk *GophKeeperHandler) GetText(ctx context.Context, req *proto.GetTextRequest) (*proto.GetTextResponse, error) {
	textID := strings.TrimSpace(req.GetId())

	if textID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty TextID")
	}

	text, err := storage.Retry2(ctx, gk.rp, func() (*storage.Text, error) {
		return gk.s.GetText(
			ctx,
			textID,
		)
	})

	if err != nil {
		switch {
		case storage.IsForeignKeyViolation(err):
			return nil, status.Errorf(codes.Unknown, "unknown TextID %s", textID)
		default:
			return nil, status.Errorf(codes.Internal, "get text %s", textID)
		}
	}

	return &proto.GetTextResponse{
		Text: &proto.Text{
			Id:       text.ID,
			UserID:   text.UserID,
			Name:     text.Name,
			Text:     text.Text,
			Meta:     text.Meta,
			UpdateAt: timestamppb.New(text.UpdateAt),
		},
	}, nil
}

// GetTexts обработчик получения всех баковских данных
func (gk *GophKeeperHandler) GetTexts(ctx context.Context, req *proto.GetTextsRequest) (*proto.GetTextsResponse, error) {
	userID := strings.TrimSpace(req.GetUserID())

	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty UserID")
	}

	texts, err := storage.Retry2(ctx, gk.rp, func() ([]storage.Text, error) {
		return gk.s.GetAllTexts(
			ctx,
			userID,
		)
	})

	if err != nil {
		switch {
		case storage.IsForeignKeyViolation(err):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return nil, status.Errorf(codes.Internal, "get texts %s", userID)
		}
	}

	protoTexts := make([]*proto.Text, 0, len(texts))

	for _, val := range texts {
		protoTexts = append(protoTexts, &proto.Text{
			Id:       val.ID,
			UserID:   val.UserID,
			Name:     val.Name,
			Text:     val.Text,
			Meta:     val.Meta,
			UpdateAt: timestamppb.New(val.UpdateAt),
		})
	}

	return &proto.GetTextsResponse{
		Texts: protoTexts,
	}, nil
}

// CreateFile обработчик загрузки файла
func (gk *GophKeeperHandler) CreateFile(stream proto.GophKeeper_CreateFileServer) error {
	req, err := stream.Recv()

	if err != nil {
		return status.Error(codes.Unknown, "cannot receive file info")
	}

	userID := strings.TrimSpace(req.GetFileInfo().GetUserID())

	if userID == "" {
		return status.Error(codes.InvalidArgument, "empty UserID")
	}

	name := req.GetFileInfo().Name
	meta := req.GetFileInfo().Meta

	fileData := bytes.Buffer{}

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Unknown, "cannot receive content")
		}

		_, err = fileData.Write(req.GetContent())

		if err != nil {
			return status.Errorf(codes.Internal, "write content on buffer: %s", err.Error())
		}
	}

	pathToFile, err := gk.fs.Save(fileData)

	if err != nil {
		return status.Errorf(codes.Internal, "save file for user %s", userID)
	}

	file, err := storage.Retry2(stream.Context(), gk.rp, func() (*storage.File, error) {
		return gk.s.CreateFile(
			stream.Context(),
			userID,
			name,
			pathToFile,
			meta,
		)
	})

	if err != nil {
		switch {
		case storage.IsForeignKeyViolation(err):
			return status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return status.Errorf(codes.Internal, "create file for user %s", userID)
		}
	}

	return stream.SendAndClose(&proto.CreateFileResponse{
		Id: file.ID,
	})
}

// GetFile обработчик получения файла
func (gk *GophKeeperHandler) GetFile(req *proto.GetFileRequest, stream proto.GophKeeper_GetFileServer) error {
	fileID := strings.TrimSpace(req.GetId())

	if fileID == "" {
		return status.Error(codes.InvalidArgument, "empty FileID")
	}

	file, err := storage.Retry2(stream.Context(), gk.rp, func() (*storage.File, error) {
		return gk.s.GetFile(
			stream.Context(),
			fileID,
		)
	})

	if err != nil {
		switch {
		case storage.IsForeignKeyViolation(err):
			return status.Errorf(codes.Unknown, "unknown FileID %s", fileID)
		default:
			return status.Errorf(codes.Internal, "get file %s", fileID)
		}
	}

	err = stream.Send(&proto.GetFileResponse{
		Data: &proto.GetFileResponse_FileInfo{
			FileInfo: &proto.File{
				Id:       file.ID,
				UserID:   file.UserID,
				Name:     file.Name,
				Meta:     file.Meta,
				UpdateAt: timestamppb.New(file.UpdateAt),
			},
		},
	})

	if err != nil {
		return status.Errorf(codes.Internal, "get file %s: %s", fileID, err.Error())
	}

	filer, err := gk.fs.GetDBFiler(file.PathToFile)

	if err != nil {
		return status.Errorf(codes.Internal, "get file %s: %s", fileID, err.Error())
	}
	defer filer.Close()

	for {
		content, err := filer.GetChunck()

		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "get file %s: %s", fileID, err.Error())
		}

		err = stream.Send(&proto.GetFileResponse{
			Data: &proto.GetFileResponse_Content{
				Content: content,
			},
		})

		if err != nil {
			return status.Errorf(codes.Internal, "get file %s: %s", fileID, err.Error())
		}
	}

	return nil
}

// GetFiles обработчик получения всех файлов пользователя
func (gk *GophKeeperHandler) GetFiles(req *proto.GetFilesRequest, stream proto.GophKeeper_GetFilesServer) error {
	userID := strings.TrimSpace(req.GetUserID())

	if userID == "" {
		return status.Error(codes.InvalidArgument, "empty UserID")
	}

	files, err := storage.Retry2(stream.Context(), gk.rp, func() ([]storage.File, error) {
		return gk.s.GetAllFiles(
			stream.Context(),
			userID,
		)
	})

	if err != nil {
		switch {
		case storage.IsForeignKeyViolation(err):
			return status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return status.Errorf(codes.Internal, "get files %s", userID)
		}
	}

	for _, file := range files {
		err := stream.Send(&proto.GetFilesResponse{
			Data: &proto.GetFilesResponse_FileInfo{
				FileInfo: &proto.File{
					Id:       file.ID,
					UserID:   file.UserID,
					Name:     file.Name,
					Meta:     file.Meta,
					UpdateAt: timestamppb.New(file.UpdateAt),
				},
			},
		})

		if err != nil {
			return status.Errorf(codes.Internal, "get files %s: %s", userID, err.Error())
		}

		filer, err := gk.fs.GetDBFiler(file.PathToFile)

		if err != nil {
			return status.Errorf(codes.Internal, "get files %s: %s", userID, err.Error())
		}
		defer filer.Close()

		for {
			content, err := filer.GetChunck()

			if err == io.EOF {
				break
			}
			if err != nil {
				return status.Errorf(codes.Internal, "get files %s: %s", userID, err.Error())
			}

			err = stream.Send(&proto.GetFilesResponse{
				Data: &proto.GetFilesResponse_Content{
					Content: content,
				},
			})

			if err != nil {
				return status.Errorf(codes.Internal, "get files %s: %s", userID, err.Error())
			}
		}
	}

	return nil
}
