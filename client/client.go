// Package client provides a gRPC client for interacting with the GophKeeper service.
// It includes functionalities for user registration, authentication, and CRUD operations
// for different types of data such as passwords, bank details, texts, and files.
// The package ensures data security by incorporating cryptographic operations for
// sealing and opening data before transmission and after reception.
package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Tomap-Tomap/GophKeeper/crypto"
	"github.com/Tomap-Tomap/GophKeeper/proto"
	"github.com/Tomap-Tomap/GophKeeper/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Crypter defines the interface for cryptographic operations.
type Crypter interface {
	SealStringWithoutNonce(str string) (string, error)
	OpenStringWithoutNonce(encryptStr string) (string, error)
	GenerateNonce() ([]byte, error)
	SealBytes(b, nonce []byte) []byte
	NonceSize() int
	GetNonceFromBytes(b []byte, nonceSize int, location crypto.NonceLocation) ([]byte, []byte, int, error)
	OpenBytes(enctyptB []byte, nonce []byte) ([]byte, error)
}

// Client represents the gRPC client for interacting with the GophKeeper service.
type Client struct {
	grpc    proto.GophKeeperClient
	conn    *grpc.ClientConn
	crypter Crypter
	ti      *tokenInterceptor
}

// New creates a new Client instance with the given Crypter and server address.
func New(crypter Crypter, addr string) (*Client, error) {
	ti := newTokenInterceptor()
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			ti.interceptorAddTokenUnary,
		),
		grpc.WithChainStreamInterceptor(
			ti.interceptorAddTokenStream,
		),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
	)

	if err != nil {
		return nil, fmt.Errorf("cannot create grpc client: %w", err)
	}

	return &Client{
		grpc:    proto.NewGophKeeperClient(conn),
		conn:    conn,
		crypter: crypter,
		ti:      ti,
	}, nil
}

// Close closes the gRPC client connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Register registers a new user with the given login and password.
func (c *Client) Register(ctx context.Context, login, password string) error {
	_, err := c.grpc.Register(ctx, &proto.RegisterRequest{
		Login:    login,
		Password: password,
	})

	return err
}

// SignIn authenticates a user with the given login and password.
func (c *Client) SignIn(ctx context.Context, login, password string) error {
	_, err := c.grpc.Auth(ctx, &proto.AuthRequest{
		Login:    login,
		Password: password,
	})

	return err
}

// GetAllPasswords retrieves all stored passwords.
func (c *Client) GetAllPasswords(ctx context.Context) ([]storage.Password, error) {
	res, err := c.grpc.GetPasswords(ctx, &emptypb.Empty{})

	if err != nil {
		return nil, fmt.Errorf("cannot get passwords: %w", err)
	}

	pwds := make([]storage.Password, 0, len(res.GetPasswords()))

	var errs error

	for _, v := range res.GetPasswords() {
		pwd, err := c.openPassword(v)

		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("cannot open password data: %w", err))
			continue
		}

		pwds = append(pwds, pwd)

	}

	if errs != nil {
		return nil, errs
	}

	return pwds, nil
}

// CreatePassword creates a new password entry with the given details.
func (c *Client) CreatePassword(ctx context.Context, name, login, password, meta string) error {
	sealPassword, err := c.sealPassword(name, login, password, meta)

	if err != nil {
		return fmt.Errorf("cannot seal password: %w", err)
	}

	_, err = c.grpc.CreatePassword(ctx, &proto.CreatePasswordRequest{
		Name:     sealPassword.Name,
		Login:    sealPassword.Login,
		Password: sealPassword.Password,
		Meta:     sealPassword.Meta,
	})

	if err != nil {
		return fmt.Errorf("cannot create password: %w", err)
	}

	return nil
}

// UpdatePassword updates an existing password entry with the given details.
func (c *Client) UpdatePassword(ctx context.Context, id, name, login, password, meta string) error {
	sealPassword, err := c.sealPassword(name, login, password, meta)

	if err != nil {
		return fmt.Errorf("cannot seal password: %w", err)
	}

	_, err = c.grpc.UpdatePassword(ctx, &proto.UpdatePasswordRequest{
		Id:       id,
		Name:     sealPassword.Name,
		Login:    sealPassword.Login,
		Password: sealPassword.Password,
		Meta:     sealPassword.Meta,
	})

	if err != nil {
		return fmt.Errorf("cannot update password: %w", err)
	}

	return nil
}

// DeletePassword deletes a password entry by its ID.
func (c *Client) DeletePassword(ctx context.Context, id string) error {
	_, err := c.grpc.DeletePassword(ctx, &proto.DeletePasswordRequest{
		Id: id,
	})

	if err != nil {
		return fmt.Errorf("cannot delete password: %w", err)
	}

	return nil
}

// GetAllBanks retrieves all stored bank details.
func (c *Client) GetAllBanks(ctx context.Context) ([]storage.Bank, error) {
	res, err := c.grpc.GetBanks(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("cannot get banks: %w", err)
	}

	banks := make([]storage.Bank, 0, len(res.GetBanks()))

	var errs error

	for _, v := range res.GetBanks() {
		bank, err := c.openBank(v)

		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("cannot open bank's data: %w", err))
		}

		banks = append(banks, bank)

	}

	if errs != nil {
		return nil, errs
	}

	return banks, nil
}

// CreateBank creates a new bank entry with the given details.
func (c *Client) CreateBank(ctx context.Context, name, number, cvc, owner, exp, meta string) error {
	sealBank, err := c.sealBank(name, number, cvc, owner, exp, meta)
	if err != nil {
		return fmt.Errorf("cannot seal bank data: %w", err)
	}

	_, err = c.grpc.CreateBank(ctx, &proto.CreateBankRequest{
		Name:       sealBank.Name,
		CardNumber: sealBank.CardNumber,
		Cvc:        sealBank.CVC,
		Owner:      sealBank.Owner,
		Exp:        sealBank.Exp,
		Meta:       sealBank.Meta,
	})

	if err != nil {
		return fmt.Errorf("cannot create bank: %w", err)
	}

	return nil
}

// UpdateBank updates an existing bank entry with the given details.
func (c *Client) UpdateBank(ctx context.Context, id, name, number, cvc, owner, exp, meta string) error {
	sealBank, err := c.sealBank(name, number, cvc, owner, exp, meta)
	if err != nil {
		return fmt.Errorf("cannot seal bank data: %w", err)
	}

	_, err = c.grpc.UpdateBank(ctx, &proto.UpdateBankRequest{
		Id:         id,
		Name:       sealBank.Name,
		CardNumber: sealBank.CardNumber,
		Cvc:        sealBank.CVC,
		Owner:      sealBank.Owner,
		Exp:        sealBank.Exp,
		Meta:       sealBank.Meta,
	})

	if err != nil {
		return fmt.Errorf("cannot update bank: %w", err)
	}

	return nil
}

// DeleteBank deletes a bank entry by its ID.
func (c *Client) DeleteBank(ctx context.Context, id string) error {
	_, err := c.grpc.DeleteBank(ctx, &proto.DeleteBankRequest{
		Id: id,
	})

	if err != nil {
		return fmt.Errorf("cannot delete bank: %w", err)
	}

	return nil
}

// GetAllTexts retrieves all stored text entries.
func (c *Client) GetAllTexts(ctx context.Context) ([]storage.Text, error) {
	res, err := c.grpc.GetTexts(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("cannot get texts: %w", err)
	}

	texts := make([]storage.Text, 0, len(res.GetTexts()))

	var errs error

	for _, v := range res.GetTexts() {
		text, err := c.openText(v)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("cannot open text data: %w", err))
		}
		texts = append(texts, text)
	}

	if errs != nil {
		return nil, errs
	}

	return texts, nil
}

// CreateText creates a new text entry with the given details.
func (c *Client) CreateText(ctx context.Context, name, text, meta string) error {
	sealText, err := c.sealText(name, text, meta)
	if err != nil {
		return fmt.Errorf("cannot seal text: %w", err)
	}

	_, err = c.grpc.CreateText(ctx, &proto.CreateTextRequest{
		Name: sealText.Name,
		Text: sealText.Text,
		Meta: sealText.Meta,
	})

	if err != nil {
		return fmt.Errorf("cannot create text: %w", err)
	}

	return nil
}

// UpdateText updates an existing text entry with the given details.
func (c *Client) UpdateText(ctx context.Context, id, name, text, meta string) error {
	sealText, err := c.sealText(name, text, meta)
	if err != nil {
		return fmt.Errorf("cannot seal text: %w", err)
	}

	_, err = c.grpc.UpdateText(ctx, &proto.UpdateTextRequest{
		Id:   id,
		Name: sealText.Name,
		Text: sealText.Text,
		Meta: sealText.Meta,
	})

	if err != nil {
		return fmt.Errorf("cannot update text: %w", err)
	}

	return nil
}

// DeleteText deletes a text entry by its ID.
func (c *Client) DeleteText(ctx context.Context, id string) error {
	_, err := c.grpc.DeleteText(ctx, &proto.DeleteTextRequest{
		Id: id,
	})

	if err != nil {
		return fmt.Errorf("cannot delete text: %w", err)
	}

	return nil
}

// GetAllFiles retrieves all stored file entries.
func (c *Client) GetAllFiles(ctx context.Context) ([]storage.File, error) {
	res, err := c.grpc.GetFiles(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("cannot get files: %w", err)
	}

	files := make([]storage.File, 0, len(res.GetFileInfo()))
	var errs error

	for _, v := range res.GetFileInfo() {
		file, err := c.openFile(v)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("cannot open file data: %w", err))
		}
		files = append(files, file)
	}

	if errs != nil {
		return nil, errs
	}

	return files, nil
}

// CreateFile creates a new file entry with the given details.
func (c *Client) CreateFile(ctx context.Context, name, pathToFile, meta string) error {
	chunkSize, err := c.grpc.GetChunkSize(ctx, &emptypb.Empty{})

	if err != nil {
		return fmt.Errorf("cannot get chunk size: %w", err)
	}

	fileInfo, err := c.sealFile(name, meta)

	if err != nil {
		return fmt.Errorf("cannot seal file info: %w", err)
	}

	stream, err := c.grpc.CreateFile(ctx)

	if err != nil {
		return fmt.Errorf("cannot start creating a file stream: %w", err)
	}

	err = stream.Send(
		&proto.CreateFileRequest{
			Data: &proto.CreateFileRequest_FileInfo{
				FileInfo: &proto.File{
					Name: fileInfo.Name,
					Meta: fileInfo.Meta,
				},
			},
		},
	)

	if err != nil {
		return fmt.Errorf("cannot send file info: %w", err)
	}

	file, err := os.Open(pathToFile)

	if err != nil {
		return fmt.Errorf("cannot open file by path %s: %w", pathToFile, err)
	}

	defer file.Close()

	buf := make([]byte, chunkSize.GetSize())

	nonce, err := c.crypter.GenerateNonce()

	if err != nil {
		return fmt.Errorf("cannot generate nonce: %w", err)
	}

	err = stream.Send(&proto.CreateFileRequest{
		Data: &proto.CreateFileRequest_Content{
			Content: nonce,
		},
	})

	if err != nil {
		return fmt.Errorf("cannot send nonce: %w", err)
	}

	for {
		n, err := file.Read(buf)

		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("cannot read file data: %w", err)
		}

		err = stream.Send(&proto.CreateFileRequest{
			Data: &proto.CreateFileRequest_Content{
				Content: c.crypter.SealBytes(buf[:n], nonce),
			},
		})

		if err != nil {
			return fmt.Errorf("cannot send file data: %w", err)
		}
	}

	_, err = stream.CloseAndRecv()

	if err != nil {
		return fmt.Errorf("cannot close streaming: %w", err)
	}

	return nil
}

// UpdateFile updates an existing file entry with the given details.
func (c *Client) UpdateFile(ctx context.Context, id, name, pathToFile, meta string) error {
	chunkSize, err := c.grpc.GetChunkSize(ctx, &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("cannot get chunk size: %w", err)
	}

	fileInfo, err := c.sealFile(name, meta)
	if err != nil {
		return fmt.Errorf("cannot seal file info: %w", err)
	}

	stream, err := c.grpc.UpdateFile(ctx)
	if err != nil {
		return fmt.Errorf("cannot start updating a file stream: %w", err)
	}

	err = stream.Send(
		&proto.UpdateFileRequest{
			Data: &proto.UpdateFileRequest_FileInfo{
				FileInfo: &proto.File{
					Id:   id,
					Name: fileInfo.Name,
					Meta: fileInfo.Meta,
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("cannot send file info: %w", err)
	}

	file, err := os.Open(pathToFile)
	if err != nil {
		return fmt.Errorf("cannot open file by path %s: %w", pathToFile, err)
	}
	defer file.Close()

	buf := make([]byte, chunkSize.GetSize())

	nonce, err := c.crypter.GenerateNonce()
	if err != nil {
		return fmt.Errorf("cannot generate nonce: %w", err)
	}

	err = stream.Send(&proto.UpdateFileRequest{
		Data: &proto.UpdateFileRequest_Content{
			Content: nonce,
		},
	})
	if err != nil {
		return fmt.Errorf("cannot send nonce: %w", err)
	}

	for {
		n, err := file.Read(buf)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("cannot read file data: %w", err)
		}

		err = stream.Send(&proto.UpdateFileRequest{
			Data: &proto.UpdateFileRequest_Content{
				Content: c.crypter.SealBytes(buf[:n], nonce),
			},
		})
		if err != nil {
			return fmt.Errorf("cannot send file data: %w", err)
		}
	}

	_, err = stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("cannot close streaming: %w", err)
	}

	return nil
}

// GetFile retrieves a file by its ID and saves it to the specified path.
func (c *Client) GetFile(ctx context.Context, id, pathToFile string) error {
	stream, err := c.grpc.GetFile(ctx, &proto.GetFileRequest{
		Id: id,
	})

	if err != nil {
		return fmt.Errorf("cannot get file stream: %w", err)
	}

	_, err = stream.Recv()

	if err != nil {
		return fmt.Errorf("cannot get file info: %w", err)
	}

	filePath := filepath.Join(pathToFile, id)
	file, err := os.Create(filePath)

	if err != nil {
		return fmt.Errorf("cannot create file %s: %w", filePath, err)
	}

	defer file.Close()

	w := bufio.NewWriter(file)

	nonce, err := c.receiveNonce(stream, w)

	if err != nil {
		return err
	}

	for {
		res, err := stream.Recv()

		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("cannot get content: %w", err)
		}

		openData, err := c.crypter.OpenBytes(res.GetContent(), nonce)

		if err != nil {
			return fmt.Errorf("cannot open content: %w", err)
		}

		_, err = w.Write(openData)

		if err != nil {
			return fmt.Errorf("cannot write content: %w", err)
		}
	}

	err = w.Flush()

	if err != nil {
		return fmt.Errorf("cannot flush content: %w", err)
	}

	return nil
}

// DeleteFile deletes a file entry by its ID.
func (c *Client) DeleteFile(ctx context.Context, id string) error {
	_, err := c.grpc.DeleteFile(ctx, &proto.DeleteFileRequest{
		Id: id,
	})

	if err != nil {
		return fmt.Errorf("cannot delete file: %w", err)
	}

	return nil
}

func (c *Client) openPassword(password *proto.Password) (retPassword storage.Password, retErr error) {
	name, err := c.crypter.OpenStringWithoutNonce(password.GetName())

	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot open name: %w", err))
	}

	login, err := c.crypter.OpenStringWithoutNonce(password.GetLogin())

	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot open login: %w", err))
	}

	pwd, err := c.crypter.OpenStringWithoutNonce(password.GetPassword())

	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot open password: %w", err))
	}

	meta, err := c.crypter.OpenStringWithoutNonce(password.GetMeta())

	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot open meta: %w", err))
	}

	if retErr != nil {
		return
	}

	retPassword = storage.Password{
		ID:       password.GetId(),
		Name:     name,
		Login:    login,
		Password: pwd,
		Meta:     meta,
		UpdateAt: password.GetUpdateAt().AsTime(),
	}

	return
}

func (c *Client) sealPassword(name, login, password, meta string) (retPassword *storage.Password, retErr error) {
	sName, err := c.crypter.SealStringWithoutNonce(name)

	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot seal name: %w", err))
	}

	sLogin, err := c.crypter.SealStringWithoutNonce(login)

	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot seal login: %w", err))
	}

	sPassword, err := c.crypter.SealStringWithoutNonce(password)

	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot seal password: %w", err))
	}

	sMeta, err := c.crypter.SealStringWithoutNonce(meta)

	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot seal meta: %w", err))
	}

	if retErr != nil {
		return
	}

	retPassword = &storage.Password{
		Name:     sName,
		Login:    sLogin,
		Password: sPassword,
		Meta:     sMeta,
	}

	return
}

func (c *Client) openBank(bank *proto.Bank) (retBank storage.Bank, retErr error) {
	name, err := c.crypter.OpenStringWithoutNonce(bank.GetName())
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot open name: %w", err))
	}

	cardNumber, err := c.crypter.OpenStringWithoutNonce(bank.GetCardNumber())
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot open card number: %w", err))
	}

	cvc, err := c.crypter.OpenStringWithoutNonce(bank.GetCvc())
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot open cvc: %w", err))
	}

	owner, err := c.crypter.OpenStringWithoutNonce(bank.GetOwner())
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot open owner: %w", err))
	}

	exp, err := c.crypter.OpenStringWithoutNonce(bank.GetExp())
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot open exp: %w", err))
	}

	meta, err := c.crypter.OpenStringWithoutNonce(bank.GetMeta())
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot open meta: %w", err))
	}

	if retErr != nil {
		return
	}

	retBank = storage.Bank{
		ID:         bank.GetId(),
		Name:       name,
		CardNumber: cardNumber,
		CVC:        cvc,
		Owner:      owner,
		Exp:        exp,
		Meta:       meta,
		UpdateAt:   bank.GetUpdateAt().AsTime(),
	}

	return
}

func (c *Client) sealBank(name, number, cvc, owner, exp, meta string) (retBank *storage.Bank, retErr error) {
	sName, err := c.crypter.SealStringWithoutNonce(name)
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot seal name: %w", err))
	}

	sNumber, err := c.crypter.SealStringWithoutNonce(number)
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot seal card number: %w", err))
	}

	sCVC, err := c.crypter.SealStringWithoutNonce(cvc)
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot seal cvc: %w", err))
	}

	sOwner, err := c.crypter.SealStringWithoutNonce(owner)
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot seal owner: %w", err))
	}

	sExp, err := c.crypter.SealStringWithoutNonce(exp)
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot seal exp: %w", err))
	}

	sMeta, err := c.crypter.SealStringWithoutNonce(meta)
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot seal meta: %w", err))
	}

	if retErr != nil {
		return
	}

	retBank = &storage.Bank{
		Name:       sName,
		CardNumber: sNumber,
		CVC:        sCVC,
		Owner:      sOwner,
		Exp:        sExp,
		Meta:       sMeta,
	}

	return
}

func (c *Client) openText(text *proto.Text) (retText storage.Text, retErr error) {
	name, err := c.crypter.OpenStringWithoutNonce(text.GetName())
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot open name: %w", err))
	}

	content, err := c.crypter.OpenStringWithoutNonce(text.GetText())
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot open text: %w", err))
	}

	meta, err := c.crypter.OpenStringWithoutNonce(text.GetMeta())
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot open meta: %w", err))
	}

	if retErr != nil {
		return
	}

	retText = storage.Text{
		ID:       text.GetId(),
		Name:     name,
		Text:     content,
		Meta:     meta,
		UpdateAt: text.GetUpdateAt().AsTime(),
	}

	return
}

func (c *Client) sealText(name, text, meta string) (retText *storage.Text, retErr error) {
	sName, err := c.crypter.SealStringWithoutNonce(name)
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot seal name: %w", err))
	}

	sText, err := c.crypter.SealStringWithoutNonce(text)
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot seal text: %w", err))
	}

	sMeta, err := c.crypter.SealStringWithoutNonce(meta)
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot seal meta: %w", err))
	}

	if retErr != nil {
		return
	}

	retText = &storage.Text{
		Name: sName,
		Text: sText,
		Meta: sMeta,
	}

	return
}

func (c *Client) openFile(file *proto.File) (retFile storage.File, retErr error) {
	name, err := c.crypter.OpenStringWithoutNonce(file.GetName())
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot open name: %w", err))
	}

	meta, err := c.crypter.OpenStringWithoutNonce(file.GetMeta())
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot open meta: %w", err))
	}

	if retErr != nil {
		return
	}

	retFile = storage.File{
		ID:       file.GetId(),
		Name:     name,
		Meta:     meta,
		UpdateAt: file.GetUpdateAt().AsTime(),
	}

	return
}

func (c *Client) sealFile(name, meta string) (retFile *storage.File, retErr error) {
	sName, err := c.crypter.SealStringWithoutNonce(name)
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot seal name: %w", err))
	}

	sMeta, err := c.crypter.SealStringWithoutNonce(meta)
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("cannot seal meta: %w", err))
	}

	if retErr != nil {
		return
	}

	retFile = &storage.File{
		Name: sName,
		Meta: sMeta,
	}

	return
}

func (c *Client) receiveNonce(stream proto.GophKeeper_GetFileClient, w *bufio.Writer) ([]byte, error) {
	var nonce []byte
	nonceSize := c.crypter.NonceSize()

	for {
		res, err := stream.Recv()

		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("cannot receive nonce: %w", err)
		}

		n, content, lastNonceSize, err := c.crypter.GetNonceFromBytes(res.GetContent(), nonceSize, crypto.AtFront)

		if err != nil {
			return nil, fmt.Errorf("cannot get nonce: %w", err)
		}

		nonce = append(nonce, n...)
		nonceSize = lastNonceSize

		if content != nil {
			openData, err := c.crypter.OpenBytes(content, nonce)

			if err != nil {
				return nil, fmt.Errorf("cannot open content: %w", err)
			}

			_, err = w.Write(openData)

			if err != nil {
				return nil, fmt.Errorf("cannot write content: %w", err)
			}
		}

		if nonceSize == 0 {
			break
		}
	}

	return nonce, nil
}
