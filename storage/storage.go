// Package storage определяет структуры и методы для работы с базой данных postgres
package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
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
func (s *Storage) CreateUser(ctx context.Context, login, password string) error {
	query := `
		INSERT INTO users (login, password) VALUES ($1, $2);
	`

	_, err := retry2(ctx, s.retryPolicy, func() (pgconn.CommandTag, error) {
		return s.conn.Exec(ctx, query, login, password)
	})

	return err
}
