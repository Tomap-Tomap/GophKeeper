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
	conn        *pgxpool.Pool
	retryPolicy retryPolicy
}

// NewStorage аллоцирует новый Storage
func NewStorage(ctx context.Context, DSN string) (*Storage, error) {
	conn, err := pgxpool.New(ctx, DSN)
	if err != nil {
		return nil, fmt.Errorf("create pgxpool: %w", err)
	}

	rp := retryPolicy{3, 1, 2}
	dbs := &Storage{conn: conn, retryPolicy: rp}

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
		err := retry(ctx, s.retryPolicy, func() error {
			return s.conn.QueryRow(ctx, queryUsers, login, password).Scan(ud)
		})

		if err != nil {
			return fmt.Errorf("insert into users table login %s: %w", login, err)
		}

		err = retry(ctx, s.retryPolicy, func() error {
			return s.conn.QueryRow(ctx, querySalts, loginHashed, salt).Scan(ud)
		})

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

	err := retry(ctx, s.retryPolicy, func() error {
		return s.conn.QueryRow(ctx, query, login, loginHashed).Scan(ud)
	})

	if err != nil {
		return nil, fmt.Errorf("get user %s: %w", login, err)
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

	err := retry(ctx, s.retryPolicy, func() error {
		return s.conn.QueryRow(ctx, query, userID, name, login, password, meta).Scan(pwd)
	})

	if err != nil {
		return nil, fmt.Errorf("insert into passwords table name %s: %w", name, err)
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

	err := retry(ctx, s.retryPolicy, func() error {
		return s.conn.QueryRow(ctx, query, passwordID).Scan(pwd)
	})

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

	err := retry(ctx, s.retryPolicy, func() error {
		rows, err := s.conn.Query(ctx, query, userID)

		if err != nil {
			return err
		}

		defer rows.Close()

		for rows.Next() {
			var pwd Password
			err := rows.Scan(&pwd)

			if err != nil {
				return err
			}

			pwds = append(pwds, pwd)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("get passwords user_id %s: %w", userID, err)
	}

	if len(pwds) == 0 {
		return nil, fmt.Errorf("user to user_id %s don't have passwords", userID)
	}

	return pwds, nil
}
