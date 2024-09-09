// Package storage implements methods for working with a PostgreSQL database.
package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Storage represents a structure for interacting with the database.
type Storage struct {
	conn *pgxpool.Pool
}

// NewStorage allocates and initializes a new Storage instance.
func NewStorage(ctx context.Context, DSN string) (*Storage, error) {
	conn, err := pgxpool.New(ctx, DSN)
	if err != nil {
		return nil, fmt.Errorf("create pgxpool: %w", err)
	}

	dbs := &Storage{conn: conn}

	return dbs, nil
}

// Close gracefully closes the database connection pool.
func (s *Storage) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}

// CreateUser adds a user record to the database.
func (s *Storage) CreateUser(ctx context.Context, login, loginHashed, salt, password string) (*User, error) {
	ud := &User{}

	err := pgx.BeginFunc(ctx, s.conn, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, queryInsertUser, login, password).Scan(ud)

		if err != nil {
			if IsUniqueViolation(err) {
				return fmt.Errorf("%s: %w", login, ErrUserAlreadyExists)
			}

			return fmt.Errorf("insert into users table login %s: %w", login, err)
		}

		err = tx.QueryRow(ctx, queryInsertSalt, loginHashed, salt).Scan(ud)

		if err != nil {
			return fmt.Errorf("insert into salts table login %s: %w", login, err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return ud, nil
}

// GetUser retrieves user data from the database.
func (s *Storage) GetUser(ctx context.Context, login, loginHashed string) (*User, error) {
	ud := &User{}

	err := s.conn.QueryRow(ctx, querySelectUser, login, loginHashed).Scan(ud)

	if err != nil {
		if IsNoRowError(err) {
			return nil, fmt.Errorf("%s: %w", login, ErrUserNotFound)
		}

		return nil, fmt.Errorf("get user %s: %w", login, err)
	}

	return ud, nil
}

// CreatePassword adds password data to the database.
func (s *Storage) CreatePassword(ctx context.Context, userID, name, login, password, meta string) (*Password, error) {
	pwd := &Password{}

	err := s.conn.QueryRow(ctx, queryInsertPassword, userID, name, login, password, meta).Scan(pwd)

	if err != nil {
		if IsForeignKeyViolation(err) {
			return nil, fmt.Errorf("%w: %s", ErrUserNotFound, userID)
		}

		return nil, fmt.Errorf("insert into passwords table name %s: %w", name, err)
	}

	return pwd, nil
}

// UpdatePassword updates password data in the database.
func (s *Storage) UpdatePassword(ctx context.Context, passwordID, userID, name, login, password, meta string) (*Password, error) {
	pwd := &Password{}

	err := s.conn.QueryRow(ctx, queryUpdatePassword, userID, name, login, password, meta, passwordID).Scan(pwd)

	if err != nil {
		switch {
		case IsForeignKeyViolation(err):
			return nil, fmt.Errorf("%s: %w", userID, ErrUserNotFound)
		case IsNoRowError(err):
			return nil, fmt.Errorf("%s: %w", passwordID, ErrPasswordNotFound)
		default:
			return nil, fmt.Errorf("update passwords table name %s: %w", name, err)
		}
	}

	return pwd, nil
}

// GetPassword returns password data from the database.
func (s *Storage) GetPassword(ctx context.Context, passwordID, userID string) (*Password, error) {
	pwd := &Password{}

	err := s.conn.QueryRow(ctx, querySelectPassword, passwordID, userID).Scan(pwd)

	if err != nil {
		if IsNoRowError(err) {
			return nil, fmt.Errorf("%s: %w", passwordID, ErrPasswordNotFound)
		}

		return nil, fmt.Errorf("get password id %s: %w", passwordID, err)
	}

	return pwd, nil
}

// GetAllPassword returns all passwords data from the database.
func (s *Storage) GetAllPassword(ctx context.Context, userID string) ([]Password, error) {
	pwds := make([]Password, 0)

	rows, err := s.conn.Query(ctx, querySelectPasswords, userID)

	if err != nil {
		return nil, fmt.Errorf("query execution from table passwords user_id %s: %w", userID, err)
	}

	defer rows.Close()

	for rows.Next() {
		var pwd Password
		err := rows.Scan(&pwd)

		if err != nil {
			return nil, fmt.Errorf("scanning the query string from passwords table user_id %s: %w", userID, err)
		}

		pwds = append(pwds, pwd)
	}

	return pwds, nil
}

// DeletePassword delete password data in the database.
func (s *Storage) DeletePassword(ctx context.Context, passwordID, userID string) error {
	file := &Password{}

	err := s.conn.QueryRow(ctx, queryDeletePassword, passwordID, userID).Scan(file)

	if err != nil {
		if IsNoRowError(err) {
			return fmt.Errorf("%s: %w", passwordID, ErrPasswordNotFound)
		}
		return fmt.Errorf("delete passwords %s: %w", passwordID, err)
	}

	return nil
}

// CreateFile adds file data to the database.
func (s *Storage) CreateFile(ctx context.Context, userID, name, pathToFile, meta string) (*File, error) {
	file := &File{}

	err := s.conn.QueryRow(ctx, queryInsertFile, userID, name, pathToFile, meta).Scan(file)

	if err != nil {
		if IsForeignKeyViolation(err) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("insert into files table name %s: %w", name, err)
	}

	return file, nil
}

// UpdateFile updates file data in the database.
func (s *Storage) UpdateFile(ctx context.Context, fileID, userID, name, pathToFile, meta string) (*File, error) {
	file := &File{}

	err := s.conn.QueryRow(ctx, queryUpdateFile, userID, name, pathToFile, meta, fileID).Scan(file)

	if err != nil {
		switch {
		case IsForeignKeyViolation(err):
			return nil, fmt.Errorf("%s: %w", userID, ErrUserNotFound)
		case IsNoRowError(err):
			return nil, fmt.Errorf("%s: %w", fileID, ErrFileNotFound)
		default:
			return nil, fmt.Errorf("update files table name %s: %w", name, err)
		}
	}

	return file, nil
}

// GetFile returns file data from the database.
func (s *Storage) GetFile(ctx context.Context, fileID, userID string) (*File, error) {
	file := &File{}

	err := s.conn.QueryRow(ctx, querySelectFile, fileID, userID).Scan(file)

	if err != nil {
		if IsNoRowError(err) {
			return nil, fmt.Errorf("%s: %w", fileID, ErrFileNotFound)
		}

		return nil, fmt.Errorf("get file id %s: %w", fileID, err)
	}

	return file, nil
}

// GetAllFiles returns all files data from the database.
func (s *Storage) GetAllFiles(ctx context.Context, userID string) ([]File, error) {
	files := make([]File, 0)

	rows, err := s.conn.Query(ctx, querySelectFiles, userID)

	if err != nil {
		return nil, fmt.Errorf("query execution from table files user_id %s: %w", userID, err)
	}

	defer rows.Close()

	for rows.Next() {
		var file File
		err := rows.Scan(&file)

		if err != nil {
			return nil, fmt.Errorf("scanning the query string from files table user_id %s: %w", userID, err)
		}

		files = append(files, file)
	}

	return files, nil
}

// DeleteFile delete file data in the database.
func (s *Storage) DeleteFile(ctx context.Context, fileID, userID string) (*File, error) {
	file := &File{}

	err := s.conn.QueryRow(ctx, queryDeleteFile, fileID, userID).Scan(file)

	if err != nil {
		if IsNoRowError(err) {
			return nil, fmt.Errorf("%s: %w", fileID, ErrFileNotFound)
		}
		return nil, fmt.Errorf("delete file id %s: %w", fileID, err)
	}

	return file, nil
}

// CreateBank adds bank data to the database.
func (s *Storage) CreateBank(ctx context.Context, userID, name, number, cvc, owner, exp, meta string) (*Bank, error) {
	bank := &Bank{}

	err := s.conn.QueryRow(ctx, queryInsertBank, userID, name, number, exp, cvc, owner, meta).Scan(bank)

	if err != nil {
		if IsForeignKeyViolation(err) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("insert into banks table name %s: %w", name, err)
	}

	return bank, nil
}

// UpdateBank updates bank data in the database.
func (s *Storage) UpdateBank(ctx context.Context, bankID, userID, name, number, cvc, owner, exp, meta string) (*Bank, error) {
	bank := &Bank{}

	err := s.conn.QueryRow(ctx, queryUpdateBank, userID, name, number, exp, cvc, owner, meta, bankID).Scan(bank)

	if err != nil {
		switch {
		case IsForeignKeyViolation(err):
			return nil, fmt.Errorf("%s: %w", userID, ErrUserNotFound)
		case IsNoRowError(err):
			return nil, fmt.Errorf("%s: %w", bankID, ErrBankNotFound)
		default:
			return nil, fmt.Errorf("update banks table name %s: %w", name, err)
		}
	}

	return bank, nil
}

// GetBank returns bank data from the database.
func (s *Storage) GetBank(ctx context.Context, bankID, userID string) (*Bank, error) {
	bank := &Bank{}

	err := s.conn.QueryRow(ctx, querySelectBank, bankID, userID).Scan(bank)

	if err != nil {
		if IsNoRowError(err) {
			return nil, fmt.Errorf("%s: %w", bankID, ErrBankNotFound)
		}

		return nil, fmt.Errorf("get bank id %s: %w", bankID, err)
	}

	return bank, nil
}

// GetAllBanks returns all banks data from the database.
func (s *Storage) GetAllBanks(ctx context.Context, userID string) ([]Bank, error) {
	banks := make([]Bank, 0)

	rows, err := s.conn.Query(ctx, querySelectBanks, userID)

	if err != nil {
		return nil, fmt.Errorf("query execution from table banks user_id %s: %w", userID, err)
	}

	defer rows.Close()

	for rows.Next() {
		var bank Bank
		err := rows.Scan(&bank)

		if err != nil {
			return nil, fmt.Errorf("scanning the query string from banks table user_id %s: %w", userID, err)
		}

		banks = append(banks, bank)
	}

	return banks, nil
}

// DeleteBank delete bank data in the database.
func (s *Storage) DeleteBank(ctx context.Context, bankID, userID string) error {
	bank := &Bank{}
	err := s.conn.QueryRow(ctx, queryDeleteBank, bankID, userID).Scan(bank)

	if err != nil {
		if IsNoRowError(err) {
			return fmt.Errorf("%s: %w", bankID, ErrBankNotFound)
		}

		return fmt.Errorf("delete bank %s: %w", bankID, err)
	}

	return nil
}

// CreateText adds text data to the database.
func (s *Storage) CreateText(ctx context.Context, userID, name, text, meta string) (*Text, error) {
	t := &Text{}

	err := s.conn.QueryRow(ctx, queryInsertText, userID, name, text, meta).Scan(t)

	if err != nil {
		if IsForeignKeyViolation(err) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("insert into texts table name %s: %w", name, err)
	}

	return t, nil
}

// UpdateText updates text data in the database.
func (s *Storage) UpdateText(ctx context.Context, textID, userID, name, text, meta string) (*Text, error) {

	t := &Text{}

	err := s.conn.QueryRow(ctx, queryUpdateText, userID, name, text, meta, textID).Scan(t)

	if err != nil {
		switch {
		case IsForeignKeyViolation(err):
			return nil, fmt.Errorf("%s: %w", userID, ErrUserNotFound)
		case IsNoRowError(err):
			return nil, fmt.Errorf("%s: %w", textID, ErrTextNotFound)
		default:
			return nil, fmt.Errorf("update texts table name %s: %w", name, err)
		}
	}

	return t, nil
}

// GetText returns text data from the database.
func (s *Storage) GetText(ctx context.Context, textID, userID string) (*Text, error) {
	t := &Text{}

	err := s.conn.QueryRow(ctx, querySelectText, textID, userID).Scan(t)

	if err != nil {
		if IsNoRowError(err) {
			return nil, fmt.Errorf("%s: %w", textID, ErrTextNotFound)
		}

		return nil, fmt.Errorf("get text id %s: %w", textID, err)
	}

	return t, nil
}

// GetAllTexts returns all texts data from the database.
func (s *Storage) GetAllTexts(ctx context.Context, userID string) ([]Text, error) {
	texts := make([]Text, 0)

	rows, err := s.conn.Query(ctx, querySelectTexts, userID)

	if err != nil {
		return nil, fmt.Errorf("query execution from table texts user_id %s: %w", userID, err)
	}

	defer rows.Close()

	for rows.Next() {
		var text Text
		err := rows.Scan(&text)

		if err != nil {
			return nil, fmt.Errorf("scanning the query string from texts table user_id %s: %w", userID, err)
		}

		texts = append(texts, text)
	}

	return texts, nil
}

// DeleteText delete text data in the database.
func (s *Storage) DeleteText(ctx context.Context, textID, userID string) error {
	text := &Text{}
	err := s.conn.QueryRow(ctx, queryDeleteText, textID, userID).Scan(text)

	if err != nil {
		if IsNoRowError(err) {
			return fmt.Errorf("%s: %w", textID, ErrTextNotFound)
		}

		return fmt.Errorf("delete text %s: %w", textID, err)
	}

	return nil
}
