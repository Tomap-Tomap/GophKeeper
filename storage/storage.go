// Package storage определяет структуры и методы для работы с базой данных postgres
package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Storage представляет собой структуры для взаимодействия с БД
type Storage struct {
	conn *pgxpool.Pool
}

// NewStorage аллоцирует новый Storage
func NewStorage(ctx context.Context, DSN string) (*Storage, error) {
	conn, err := pgxpool.New(ctx, DSN)
	if err != nil {
		return nil, fmt.Errorf("create pgxpool: %w", err)
	}

	dbs := &Storage{conn: conn}

	return dbs, nil
}

// Close closes DBStorage
func (s *Storage) Close() {
	s.conn.Close()
}

// CreateUser добавляет запись по пользователю в БД
func (s *Storage) CreateUser(ctx context.Context, login, loginHashed, salt, password string) (*User, error) {
	queryUsers := `
		WITH t AS (
			INSERT INTO users (login, password) VALUES ($1, $2)
			RETURNING *
		)
		SELECT id, login, password FROM t;
	`

	querySalts := `
		WITH t AS (
			INSERT INTO salts (login, salt) VALUES ($1, $2)
			RETURNING *
		)
		SELECT salt FROM t;
	`

	ud := &User{}

	err := pgx.BeginFunc(ctx, s.conn, func(tx pgx.Tx) error {
		err := s.conn.QueryRow(ctx, queryUsers, login, password).Scan(ud)

		if err != nil {
			switch IsUniqueViolation(err) {
			case true:
				return fmt.Errorf("user %s already exists: %w", login, err)
			default:
				return fmt.Errorf("insert into users table login %s: %w", login, err)
			}
		}

		err = s.conn.QueryRow(ctx, querySalts, loginHashed, salt).Scan(ud)

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

// GetUser возвращает даные о пользователе БД
func (s *Storage) GetUser(ctx context.Context, login, loginHashed string) (*User, error) {
	query := `
		SELECT u.id, u.login, u.password, s.salt
		FROM users u, salts s
		WHERE u.login = $1 AND s.login = $2;
	`

	ud := &User{}

	err := s.conn.QueryRow(ctx, query, login, loginHashed).Scan(ud)

	if err != nil {
		switch {
		case IsNowRowError(err):
			return nil, fmt.Errorf("unknown user %s: %w", login, err)
		default:
			return nil, fmt.Errorf("get user %s: %w", login, err)
		}
	}

	return ud, nil
}

// CreatePassword добавляет данные по паролю в БД
func (s *Storage) CreatePassword(ctx context.Context, userID, name, login, password, meta string) (*Password, error) {
	query := `
		WITH t AS (
			INSERT INTO passwords (user_id, name, login, password, meta) VALUES ($1, $2, $3, $4, $5)
			RETURNING *
		)
		SELECT * FROM t;
	`

	pwd := &Password{}

	err := s.conn.QueryRow(ctx, query, userID, name, login, password, meta).Scan(pwd)

	if err != nil {
		switch {
		case IsForeignKeyViolation(err):
			return nil, fmt.Errorf("unknown user %s: %w", userID, err)
		default:
			return nil, fmt.Errorf("insert into passwords table name %s: %w", name, err)
		}
	}

	return pwd, nil
}

// GetPassword возращает данные по сохраненному паспорту
func (s *Storage) GetPassword(ctx context.Context, passwordID string) (*Password, error) {
	query := `
		SELECT *
		FROM passwords
		WHERE id = $1;
	`

	pwd := &Password{}

	err := s.conn.QueryRow(ctx, query, passwordID).Scan(pwd)

	if err != nil {
		return nil, fmt.Errorf("get password id %s: %w", passwordID, err)
	}

	return pwd, nil
}

// GetAllPassword возращает все данные по сохраненным паспортам
func (s *Storage) GetAllPassword(ctx context.Context, userID string) ([]Password, error) {
	query := `
		SELECT *
		FROM passwords
		WHERE user_id = $1;
	`

	pwds := make([]Password, 0)

	rows, err := s.conn.Query(ctx, query, userID)

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

	if len(pwds) == 0 {
		return nil, fmt.Errorf("user to user_id %s don't have passwords", userID)
	}

	return pwds, nil
}

// CreateFile добавляет данные о файле в БД
func (s *Storage) CreateFile(ctx context.Context, userID, name, pathToFile, meta string) (*File, error) {
	query := `
		WITH t AS (
			INSERT INTO files (user_id, name, pathtofile, meta) VALUES ($1, $2, $3, $4)
			RETURNING *
		)
		SELECT * FROM t;
	`

	file := &File{}

	err := s.conn.QueryRow(ctx, query, userID, name, pathToFile, meta).Scan(file)

	if err != nil {
		return nil, fmt.Errorf("insert into files table name %s: %w", name, err)
	}

	return file, nil
}

// GetFile возращает данные по сохраненному файлу
func (s *Storage) GetFile(ctx context.Context, fileID string) (*File, error) {
	query := `
		SELECT *
		FROM files
		WHERE id = $1;
	`

	file := &File{}

	err := s.conn.QueryRow(ctx, query, fileID).Scan(file)

	if err != nil {
		return nil, fmt.Errorf("get file id %s: %w", fileID, err)
	}

	return file, nil
}

// GetAllFiles возращает все данные по сохраненным файлам
func (s *Storage) GetAllFiles(ctx context.Context, userID string) ([]File, error) {
	query := `
		SELECT *
		FROM files
		WHERE user_id = $1;
	`

	files := make([]File, 0)

	rows, err := s.conn.Query(ctx, query, userID)

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

	if len(files) == 0 {
		return nil, fmt.Errorf("user to user_id %s don't have files", userID)
	}

	return files, nil
}

// CreateBank добавляет данные о банковской информации
func (s *Storage) CreateBank(ctx context.Context, userID, name, banksData, meta string) (*Bank, error) {
	query := `
		WITH t AS (
			INSERT INTO banks (user_id, name, banksdata, meta) VALUES ($1, $2, $3, $4)
			RETURNING *
		)
		SELECT * FROM t;
	`

	bank := &Bank{}

	err := s.conn.QueryRow(ctx, query, userID, name, banksData, meta).Scan(bank)

	if err != nil {
		return nil, fmt.Errorf("insert into banks table name %s: %w", name, err)
	}

	return bank, nil
}

// GetBank возращает данные по сохраненной банковской информации
func (s *Storage) GetBank(ctx context.Context, bankID string) (*Bank, error) {
	query := `
		SELECT *
		FROM banks
		WHERE id = $1;
	`

	bank := &Bank{}

	err := s.conn.QueryRow(ctx, query, bankID).Scan(bank)

	if err != nil {
		return nil, fmt.Errorf("get bank data id %s: %w", bankID, err)
	}

	return bank, nil
}

// GetAllBanks возращает все данные по сохраненным банковским информациям
func (s *Storage) GetAllBanks(ctx context.Context, userID string) ([]Bank, error) {
	query := `
		SELECT *
		FROM banks
		WHERE user_id = $1;
	`

	banks := make([]Bank, 0)

	rows, err := s.conn.Query(ctx, query, userID)

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

	if len(banks) == 0 {
		return nil, fmt.Errorf("user to user_id %s don't have bank data", userID)
	}

	return banks, nil
}

// CreateText добавляет данные о тексте
func (s *Storage) CreateText(ctx context.Context, userID, name, text, meta string) (*Text, error) {
	query := `
		WITH t AS (
			INSERT INTO texts (user_id, name, text, meta) VALUES ($1, $2, $3, $4)
			RETURNING *
		)
		SELECT * FROM t;
	`

	t := &Text{}

	err := s.conn.QueryRow(ctx, query, userID, name, text, meta).Scan(t)

	if err != nil {
		return nil, fmt.Errorf("insert into texts table name %s: %w", name, err)
	}

	return t, nil
}

// GetText возращает данные по сохраненному тексту
func (s *Storage) GetText(ctx context.Context, textID string) (*Text, error) {
	query := `
		SELECT *
		FROM texts
		WHERE id = $1;
	`

	t := &Text{}

	err := s.conn.QueryRow(ctx, query, textID).Scan(t)

	if err != nil {
		return nil, fmt.Errorf("get text data id %s: %w", textID, err)
	}

	return t, nil
}

// GetAllTexts возращает все данные по сохраненным текстовым данным
func (s *Storage) GetAllTexts(ctx context.Context, userID string) ([]Text, error) {
	query := `
		SELECT *
		FROM texts
		WHERE user_id = $1;
	`

	texts := make([]Text, 0)

	rows, err := s.conn.Query(ctx, query, userID)

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

	if len(texts) == 0 {
		return nil, fmt.Errorf("user to user_id %s don't have text data", userID)
	}

	return texts, nil
}
