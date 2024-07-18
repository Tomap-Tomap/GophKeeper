// Package storage определяет структуры и методы для работы с базой данных postgres
package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserData прдеставлявет собой структуру данных о пользователе
type UserData struct {
	ID       pgtype.UUID
	Login    string
	Password string
	Salt     string
}

// ScanRow необходим для реализации интерфейса pgx.RowScanner
func (u *UserData) ScanRow(rows pgx.Rows) error {
	values, err := rows.Values()
	if err != nil {
		return err
	}

	for i := range values {
		switch strings.ToLower(rows.FieldDescriptions()[i].Name) {
		case "id":
			u.ID.Bytes = values[i].([16]byte)
			u.ID.Valid = true
		case "login":
			u.Login = values[i].(string)
		case "password":
			u.Password = strings.TrimSpace(values[i].(string))
		case "salt":
			u.Salt = strings.TrimSpace(values[i].(string))
		}
	}

	return nil
}

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
func (s *Storage) CreateUser(ctx context.Context, login, password string) error {
	query := `
		INSERT INTO users (login, password) VALUES ($1, $2);
	`

	_, err := retry2(ctx, s.retryPolicy, func() (pgconn.CommandTag, error) {
		return s.conn.Exec(ctx, query, login, password)
	})

	return err
}

// GetUserData возвращает даные о пользователе БД
func (s *Storage) GetUserData(ctx context.Context, login, loginHash string) (*UserData, error) {
	query := `
		SELECT u.id, u.login, u.password, s.salt
		FROM users u, salts s
		WHERE u.login = $1 AND s.login = $2;
	`

	ud := &UserData{}

	err := retry(ctx, s.retryPolicy, func() error {
		return s.conn.QueryRow(ctx, query, login, loginHash).Scan(ud)
	})

	if err != nil {
		return nil, fmt.Errorf("get user %s: %w", login, err)
	}

	return ud, nil
}
