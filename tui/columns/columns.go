// Package columns provides various column implementations for different types of data
// that can be displayed in a tabular format within the terminal user interface (TUI).
// Each column type corresponds to a specific data model and includes methods for
// managing and displaying the data, as well as handling user interactions.
package columns

import (
	"context"

	"github.com/Tomap-Tomap/GophKeeper/storage"
)

// Client interface abstracts the methods required for interacting with different storage entities.
type Client interface {
	GetAllPasswords(ctx context.Context) ([]storage.Password, error)
	CreatePassword(ctx context.Context, name, login, password, meta string) error
	UpdatePassword(ctx context.Context, id, name, login, password, meta string) error
	DeletePassword(ctx context.Context, id string) error

	GetAllBanks(ctx context.Context) ([]storage.Bank, error)
	CreateBank(ctx context.Context, name, number, cvc, owner, exp, meta string) error
	UpdateBank(ctx context.Context, id, name, number, cvc, owner, exp, meta string) error
	DeleteBank(ctx context.Context, id string) error

	GetAllTexts(ctx context.Context) ([]storage.Text, error)
	CreateText(ctx context.Context, name, text, meta string) error
	UpdateText(ctx context.Context, id, name, text, meta string) error
	DeleteText(ctx context.Context, id string) error

	GetAllFiles(ctx context.Context) ([]storage.File, error)
	CreateFile(ctx context.Context, name, path, meta string) error
	UpdateFile(ctx context.Context, id, name, path, meta string) error
	GetFile(ctx context.Context, id, path string) error
	DeleteFile(ctx context.Context, id string) error
}

type column struct {
	name string
	idx  int
}
