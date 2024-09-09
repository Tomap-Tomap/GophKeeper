// Package handlers defines methods and structures for a grpc server
package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	proto "github.com/Tomap-Tomap/GophKeeper/proto/gophkeeper/v1"
	"github.com/Tomap-Tomap/GophKeeper/storage"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const userIDHeader = "user_id"

// Storage defines methods for interacting with a storage system.
type Storage interface {
	CreateUser(ctx context.Context, login, loginHashed, salt, password string) (*storage.User, error)
	GetUser(ctx context.Context, login, loginHashed string) (*storage.User, error)

	CreatePassword(ctx context.Context, userID, name, login, password, meta string) (*storage.Password, error)
	UpdatePassword(ctx context.Context, id, userID, name, login, password, meta string) (*storage.Password, error)
	GetPassword(ctx context.Context, passwordID, userID string) (*storage.Password, error)
	GetAllPassword(ctx context.Context, userID string) ([]storage.Password, error)
	DeletePassword(ctx context.Context, passwordID, userID string) error

	CreateFile(ctx context.Context, userID, name, pathToFile, meta string) (*storage.File, error)
	UpdateFile(ctx context.Context, id, userID, name, pathToFile, meta string) (*storage.File, error)
	GetFile(ctx context.Context, fileID, userID string) (*storage.File, error)
	GetAllFiles(ctx context.Context, userID string) ([]storage.File, error)
	DeleteFile(ctx context.Context, fileID, userID string) (*storage.File, error)

	CreateBank(ctx context.Context, userID, name, number, cvc, owner, exp, meta string) (*storage.Bank, error)
	UpdateBank(ctx context.Context, id, userID, name, number, cvc, owner, exp, meta string) (*storage.Bank, error)
	GetBank(ctx context.Context, bankID, userID string) (*storage.Bank, error)
	GetAllBanks(ctx context.Context, userID string) ([]storage.Bank, error)
	DeleteBank(ctx context.Context, bankID, userID string) error

	CreateText(ctx context.Context, userID, name, text, meta string) (*storage.Text, error)
	UpdateText(ctx context.Context, id, userID, name, text, meta string) (*storage.Text, error)
	GetText(ctx context.Context, textID, userID string) (*storage.Text, error)
	GetAllTexts(ctx context.Context, userID string) ([]storage.Text, error)
	DeleteText(ctx context.Context, textID, userID string) error
}

// Hasher interface defines methods for generating salts and hashes.
type Hasher interface {
	GenerateSalt(len int) (string, error)
	GenerateHash(str string) string
	GenerateHashWithSalt(str, salt string) (string, error)
}

// Tokener describes methods for generating tokens.
type Tokener interface {
	GetToken(sub string) (string, error)
}

// FileStore defines the interface for interacting with database files.
type FileStore interface {
	CreateDBFile(fileName string) (storage.DBFiler, error)
	GetDBFile(fileName string) (storage.DBFiler, error)
	DeleteDBFile(fileName string) error
	GetChunkSize() int
}

// GophKeeperHandler is a struct that implements the functionality of a gRPC server.
type GophKeeperHandler struct {
	proto.UnimplementedGophKeeperServiceServer

	rp storage.RetryPolicy
	s  Storage
	h  Hasher
	t  Tokener
	fs FileStore

	saltLenght int
}

// NewGophKeeperHandler initializes a GophKeeperHandler structure.
func NewGophKeeperHandler(s Storage, h Hasher, t Tokener, fs FileStore, rp storage.RetryPolicy, saltLength int) *GophKeeperHandler {

	return &GophKeeperHandler{
		s:          s,
		h:          h,
		rp:         rp,
		t:          t,
		fs:         fs,
		saltLenght: saltLength,
	}
}

// Register handles the creation of a new user. It validates the login and password,
// generates a salt and hash for the password, and stores the user information in the storage.
func (gk *GophKeeperHandler) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	login := strings.TrimSpace(req.GetLogin())
	password := strings.TrimSpace(req.GetPassword())

	loginHash := gk.h.GenerateHash(login)

	salt, err := gk.h.GenerateSalt(gk.saltLenght)

	if err != nil {
		return nil, status.Error(codes.Internal, "generate salt")
	}

	passwordHash, err := gk.h.GenerateHashWithSalt(password, salt)

	if err != nil {
		return nil, status.Error(codes.Internal, "generate hash")
	}

	user, err := storage.Retry2(ctx, gk.rp, func() (*storage.User, error) {
		return gk.s.CreateUser(ctx, login, loginHash, salt, passwordHash)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserAlreadyExists):
			return nil, status.Errorf(codes.AlreadyExists, "user %s already exists", login)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	token, err := gk.t.GetToken(user.ID)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.RegisterResponse{Token: token}, nil
}

// Auth handles the retrieval of a user by their login and password. It validates the login and password,
// generates a hash with the stored salt, and retrieves the user information from the storage.
func (gk *GophKeeperHandler) Auth(ctx context.Context, req *proto.AuthRequest) (*proto.AuthResponse, error) {
	login := strings.TrimSpace(req.GetLogin())
	password := strings.TrimSpace(req.GetPassword())

	loginHash := gk.h.GenerateHash(login)

	user, err := storage.Retry2(ctx, gk.rp, func() (*storage.User, error) {
		return gk.s.GetUser(ctx, login, loginHash)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown user %s", login)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	hash, err := gk.h.GenerateHashWithSalt(password, user.Salt)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if hash != user.Password {
		return nil, status.Error(codes.PermissionDenied, "invalid password")
	}

	token, err := gk.t.GetToken(user.ID)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.AuthResponse{Token: token}, nil
}

// GetChunkSize is a gRPC handler that returns the chunk size used by the file storage system.
func (gk *GophKeeperHandler) GetChunkSize(_ context.Context, _ *proto.GetChunkSizeRequest) (*proto.GetChunkSizeResponse, error) {
	return &proto.GetChunkSizeResponse{
		Size: uint64(gk.fs.GetChunkSize()),
	}, nil
}

// CreatePassword handles the creation of a new password entry for a user. It retrieves the user ID from the context,
// and stores the password information in the storage.
func (gk *GophKeeperHandler) CreatePassword(ctx context.Context, req *proto.CreatePasswordRequest) (*proto.CreatePasswordResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
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
		case errors.Is(err, storage.ErrUserNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &proto.CreatePasswordResponse{Id: pwd.ID}, nil
}

// UpdatePassword handles the updating of an existing password entry for a user.
func (gk *GophKeeperHandler) UpdatePassword(ctx context.Context, req *proto.UpdatePasswordRequest) (*proto.UpdatePasswordResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	pwd, err := storage.Retry2(ctx, gk.rp, func() (*storage.Password, error) {
		return gk.s.UpdatePassword(
			ctx,
			req.Id,
			userID,
			req.Name,
			req.Login,
			req.Password,
			req.Meta,
		)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		case errors.Is(err, storage.ErrPasswordNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown PasswordID %s", req.Id)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &proto.UpdatePasswordResponse{Id: pwd.ID}, nil
}

// GetPassword handles the retrieval of a password entry by its ID. It retrieves the user ID from the context,
// and fetches the password information from the storage.
func (gk *GophKeeperHandler) GetPassword(ctx context.Context, req *proto.GetPasswordRequest) (*proto.GetPasswordResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	passwordID := strings.TrimSpace(req.GetId())

	if passwordID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty PasswordID")
	}

	pwd, err := storage.Retry2(ctx, gk.rp, func() (*storage.Password, error) {
		return gk.s.GetPassword(
			ctx,
			passwordID,
			userID,
		)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrPasswordNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown PasswordID %s", passwordID)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &proto.GetPasswordResponse{
		Password: &proto.Password{
			Id:       pwd.ID,
			Name:     pwd.Name,
			Login:    pwd.Login,
			Password: pwd.Password,
			Meta:     pwd.Meta,
			UpdateAt: timestamppb.New(pwd.UpdateAt),
		},
	}, nil
}

// GetPasswords handles the retrieval of all password entries for a user. It retrieves the user ID from the context,
// and fetches all password information from the storage.
func (gk *GophKeeperHandler) GetPasswords(ctx context.Context, _ *proto.GetPasswordsRequest) (*proto.GetPasswordsResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	pwds, err := storage.Retry2(ctx, gk.rp, func() ([]storage.Password, error) {
		return gk.s.GetAllPassword(
			ctx,
			userID,
		)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	protoPWDs := make([]*proto.Password, 0, len(pwds))

	for _, val := range pwds {
		protoPWDs = append(protoPWDs, &proto.Password{
			Id:       val.ID,
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

// DeletePassword handles the deletion of a password entry by its ID. It retrieves the user ID from the context,
// and deletes the password information from the storage.
func (gk *GophKeeperHandler) DeletePassword(ctx context.Context, req *proto.DeletePasswordRequest) (*proto.DeletePasswordResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = storage.Retry(ctx, gk.rp, func() error {
		return gk.s.DeletePassword(
			ctx,
			req.Id,
			userID,
		)
	})

	if err != nil {
		if errors.Is(err, storage.ErrPasswordNotFound) {
			return nil, status.Errorf(codes.Unknown, "unknown PasswordID %s", req.Id)
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

// CreateBank handles the creation of a new bank entry for a user. It retrieves the user ID from the context,
// and stores the bank information in the storage.
func (gk *GophKeeperHandler) CreateBank(ctx context.Context, req *proto.CreateBankRequest) (*proto.CreateBankResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	bank, err := storage.Retry2(ctx, gk.rp, func() (*storage.Bank, error) {
		return gk.s.CreateBank(
			ctx,
			userID,
			req.Name,
			req.CardNumber,
			req.Cvc,
			req.Owner,
			req.Exp,
			req.Meta,
		)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &proto.CreateBankResponse{Id: bank.ID}, nil
}

// UpdateBank handles the updating of a bank entry for a user. It retrieves the user ID from the context,
// and updates the bank information in the storage.
func (gk *GophKeeperHandler) UpdateBank(ctx context.Context, req *proto.UpdateBankRequest) (*proto.UpdateBankResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	bank, err := storage.Retry2(ctx, gk.rp, func() (*storage.Bank, error) {
		return gk.s.UpdateBank(
			ctx,
			req.Id,
			userID,
			req.Name,
			req.CardNumber,
			req.Cvc,
			req.Owner,
			req.Exp,
			req.Meta,
		)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		case errors.Is(err, storage.ErrBankNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown BankID %s", req.Id)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &proto.UpdateBankResponse{Id: bank.ID}, nil
}

// GetBank handles the retrieval of a bank entry by its ID. It retrieves the user ID from the context,
// and fetches the bank information from the storage.
func (gk *GophKeeperHandler) GetBank(ctx context.Context, req *proto.GetBankRequest) (*proto.GetBankResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	bankID := strings.TrimSpace(req.GetId())

	if bankID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty BankID")
	}

	bank, err := storage.Retry2(ctx, gk.rp, func() (*storage.Bank, error) {
		return gk.s.GetBank(
			ctx,
			bankID,
			userID,
		)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrBankNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown BankID %s", bankID)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &proto.GetBankResponse{
		Bank: &proto.Bank{
			Id:         bank.ID,
			Name:       bank.Name,
			CardNumber: bank.CardNumber,
			Cvc:        bank.CVC,
			Owner:      bank.Owner,
			Exp:        bank.Exp,
			Meta:       bank.Meta,
			UpdateAt:   timestamppb.New(bank.UpdateAt),
		},
	}, nil
}

// GetBanks handles the retrieval of all bank entries for a user. It retrieves the user ID from the context,
// and fetches all bank information from the storage.
func (gk *GophKeeperHandler) GetBanks(ctx context.Context, _ *proto.GetBanksRequest) (*proto.GetBanksResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	banks, err := storage.Retry2(ctx, gk.rp, func() ([]storage.Bank, error) {
		return gk.s.GetAllBanks(
			ctx,
			userID,
		)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	protoBanks := make([]*proto.Bank, 0, len(banks))

	for _, val := range banks {
		protoBanks = append(protoBanks, &proto.Bank{
			Id:         val.ID,
			Name:       val.Name,
			CardNumber: val.CardNumber,
			Cvc:        val.CVC,
			Owner:      val.Owner,
			Exp:        val.Exp,
			Meta:       val.Meta,
			UpdateAt:   timestamppb.New(val.UpdateAt),
		})
	}

	return &proto.GetBanksResponse{
		Banks: protoBanks,
	}, nil
}

// DeleteBank handles the deletion of a bank entry by its ID. It retrieves the user ID from the context,
// and deletes the bank information from the storage.
func (gk *GophKeeperHandler) DeleteBank(ctx context.Context, req *proto.DeleteBankRequest) (*proto.DeleteBankResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = storage.Retry(ctx, gk.rp, func() error {
		return gk.s.DeleteBank(
			ctx,
			req.Id,
			userID,
		)
	})

	if err != nil {
		if errors.Is(err, storage.ErrBankNotFound) {
			return nil, status.Errorf(codes.Unknown, "unknown BankID %s", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

// CreateText handles the creation of a new text entry for a user. It retrieves the user ID from the context,
// and stores the text information in the storage.
func (gk *GophKeeperHandler) CreateText(ctx context.Context, req *proto.CreateTextRequest) (*proto.CreateTextResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
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
		case errors.Is(err, storage.ErrUserNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &proto.CreateTextResponse{Id: text.ID}, nil
}

// UpdateText handles the updating of a text entry for a user. It retrieves the user ID from the context,
// and updates the text information in the storage.
func (gk *GophKeeperHandler) UpdateText(ctx context.Context, req *proto.UpdateTextRequest) (*proto.UpdateTextResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	text, err := storage.Retry2(ctx, gk.rp, func() (*storage.Text, error) {
		return gk.s.UpdateText(
			ctx,
			req.Id,
			userID,
			req.Name,
			req.Text,
			req.Meta,
		)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		case errors.Is(err, storage.ErrTextNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown TextID %s", req.Id)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &proto.UpdateTextResponse{Id: text.ID}, nil
}

// GetText handles the retrieval of a text entry by its ID. It retrieves the user ID from the context,
// and fetches the text information from the storage.
func (gk *GophKeeperHandler) GetText(ctx context.Context, req *proto.GetTextRequest) (*proto.GetTextResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	textID := strings.TrimSpace(req.GetId())

	if textID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty TextID")
	}

	text, err := storage.Retry2(ctx, gk.rp, func() (*storage.Text, error) {
		return gk.s.GetText(
			ctx,
			textID,
			userID,
		)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrTextNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown TextID %s", textID)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &proto.GetTextResponse{
		Text: &proto.Text{
			Id:       text.ID,
			Name:     text.Name,
			Text:     text.Text,
			Meta:     text.Meta,
			UpdateAt: timestamppb.New(text.UpdateAt),
		},
	}, nil
}

// GetTexts handles the retrieval of all text entries for a user. It retrieves the user ID from the context,
// and fetches all text information from the storage.
func (gk *GophKeeperHandler) GetTexts(ctx context.Context, _ *proto.GetTextsRequest) (*proto.GetTextsResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	texts, err := storage.Retry2(ctx, gk.rp, func() ([]storage.Text, error) {
		return gk.s.GetAllTexts(
			ctx,
			userID,
		)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	protoTexts := make([]*proto.Text, 0, len(texts))

	for _, val := range texts {
		protoTexts = append(protoTexts, &proto.Text{
			Id:       val.ID,
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

// DeleteText handles the deletion of a text entry by its ID. It retrieves the user ID from the context,
// and deletes the text information from the storage.
func (gk *GophKeeperHandler) DeleteText(ctx context.Context, req *proto.DeleteTextRequest) (*proto.DeleteTextResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = storage.Retry(ctx, gk.rp, func() error {
		return gk.s.DeleteText(
			ctx,
			req.Id,
			userID,
		)
	})

	if err != nil {
		if errors.Is(err, storage.ErrTextNotFound) {
			return nil, status.Errorf(codes.Unknown, "unknown TextID %s", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

// CreateFile handles the uploading of a file for a user. It retrieves the user ID from the context,
// and stores the file information in the storage.
func (gk *GophKeeperHandler) CreateFile(stream proto.GophKeeperService_CreateFileServer) (err error) {
	userID, err := getUserIDFromContext(stream.Context())

	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}

	req, err := stream.Recv()

	if err != nil {
		return status.Error(codes.Unknown, "cannot receive file info")
	}

	name := req.GetFileInfo().Name
	meta := req.GetFileInfo().Meta

	fileName, err := uuid.NewRandom()

	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	dbf, err := gk.fs.CreateDBFile(fileName.String())

	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	defer func() {
		err = errors.Join(err, dbf.Close())
	}()

	for {
		req, err := stream.Recv()

		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return status.Error(codes.Unknown, "cannot receive content")
		}

		_, err = dbf.Write(req.GetContent())

		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	file, err := storage.Retry2(stream.Context(), gk.rp, func() (*storage.File, error) {
		return gk.s.CreateFile(
			stream.Context(),
			userID,
			name,
			fileName.String(),
			meta,
		)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			return status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return status.Error(codes.Internal, err.Error())
		}
	}

	return stream.SendAndClose(&proto.CreateFileResponse{
		Id: file.ID,
	})
}

// UpdateFile handles the updating of a file for a user. It retrieves the user ID from the context,
// and updates the file information in the storage.
func (gk *GophKeeperHandler) UpdateFile(stream proto.GophKeeperService_UpdateFileServer) (err error) {
	userID, err := getUserIDFromContext(stream.Context())

	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}

	req, err := stream.Recv()

	if err != nil {
		return status.Error(codes.Unknown, "cannot receive file info")
	}

	id := req.GetFileInfo().Id
	name := req.GetFileInfo().Name
	meta := req.GetFileInfo().Meta

	fileName, err := uuid.NewRandom()

	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	dbf, err := gk.fs.CreateDBFile(fileName.String())

	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	defer func() {
		err = errors.Join(err, dbf.Close())
	}()

	for {
		req, err := stream.Recv()

		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return status.Error(codes.Unknown, "cannot receive content")
		}

		_, err = dbf.Write(req.GetContent())

		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	file, err := storage.Retry2(stream.Context(), gk.rp, func() (*storage.File, error) {
		return gk.s.UpdateFile(
			stream.Context(),
			id,
			userID,
			name,
			fileName.String(),
			meta,
		)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			return status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return status.Error(codes.Internal, err.Error())
		}
	}

	err = gk.fs.DeleteDBFile(file.PathToFile)

	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return stream.SendAndClose(&proto.UpdateFileResponse{
		Id: file.ID,
	})
}

// GetFile handles the retrieval of a file by its ID. It retrieves the user ID from the context,
// and fetches the file information from the storage.
func (gk *GophKeeperHandler) GetFile(req *proto.GetFileRequest, stream proto.GophKeeperService_GetFileServer) (err error) {
	userID, err := getUserIDFromContext(stream.Context())

	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}

	fileID := strings.TrimSpace(req.GetId())

	if fileID == "" {
		return status.Error(codes.InvalidArgument, "empty FileID")
	}

	file, err := storage.Retry2(stream.Context(), gk.rp, func() (*storage.File, error) {
		return gk.s.GetFile(
			stream.Context(),
			fileID,
			userID,
		)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrFileNotFound):
			return status.Errorf(codes.Unknown, "unknown FileID %s", fileID)
		default:
			return status.Error(codes.Internal, err.Error())
		}
	}

	err = stream.Send(&proto.GetFileResponse{
		Data: &proto.GetFileResponse_FileInfo{
			FileInfo: &proto.File{
				Id:       file.ID,
				Name:     file.Name,
				Meta:     file.Meta,
				UpdateAt: timestamppb.New(file.UpdateAt),
			},
		},
	})

	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	filer, err := gk.fs.GetDBFile(file.PathToFile)

	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	defer func() {
		err = errors.Join(err, filer.Close())
	}()

	for {
		content, err := filer.GetChunk()

		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		err = stream.Send(&proto.GetFileResponse{
			Data: &proto.GetFileResponse_Content{
				Content: content,
			},
		})

		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}

// GetFiles handles the retrieval of all files for a user. It retrieves the user ID from the context,
// and fetches all file information from the storage.
func (gk *GophKeeperHandler) GetFiles(ctx context.Context, _ *proto.GetFilesRequest) (*proto.GetFilesResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	files, err := storage.Retry2(ctx, gk.rp, func() ([]storage.File, error) {
		return gk.s.GetAllFiles(
			ctx,
			userID,
		)
	})

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			return nil, status.Errorf(codes.Unknown, "unknown UserID %s", userID)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	protoFiles := make([]*proto.File, 0, len(files))

	for _, val := range files {
		protoFiles = append(protoFiles, &proto.File{
			Id:       val.ID,
			Name:     val.Name,
			Meta:     val.Meta,
			UpdateAt: timestamppb.New(val.UpdateAt),
		})
	}

	return &proto.GetFilesResponse{
		FileInfo: protoFiles,
	}, nil
}

// DeleteFile handles the deletion of a file by its ID. It retrieves the user ID from the context,
// and deletes the file information from the storage.
func (gk *GophKeeperHandler) DeleteFile(ctx context.Context, req *proto.DeleteFileRequest) (*proto.DeleteFileResponse, error) {
	userID, err := getUserIDFromContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	file, err := storage.Retry2(ctx, gk.rp, func() (*storage.File, error) {
		return gk.s.DeleteFile(
			ctx,
			req.Id,
			userID,
		)
	})

	if err != nil {
		if errors.Is(err, storage.ErrFileNotFound) {
			return nil, status.Errorf(codes.Unknown, "unknown FileID %s", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = gk.fs.DeleteDBFile(file.PathToFile)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

func loginPasswordValidate(login, password string) error {
	var err error

	if login == "" {
		err = errors.Join(errors.New("empty login"))
	}

	if password == "" {
		err = errors.Join(err, errors.New("empty password"))
	}

	if err != nil {
		return err
	}
	return nil
}

func getUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return "", errors.New("missing metadata")
	}

	uid := md.Get(userIDHeader)

	if len(uid) == 0 {
		return "", fmt.Errorf("missing %s", userIDHeader)
	}

	return uid[0], nil
}
