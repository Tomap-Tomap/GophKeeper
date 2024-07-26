package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// RetryPolicy определяет политику повторов
type RetryPolicy struct {
	retryCount int
	duration   int
	increment  int
}

// NewRetryPolicy инициализирует новую политику повторов
func NewRetryPolicy(retryCount, duration, increment int) *RetryPolicy {
	return &RetryPolicy{retryCount, duration, increment}
}

// Retry выполняет запрос к БД с учетом retryPolicy при возникновении ошибки Class 08
func Retry(ctx context.Context, rp RetryPolicy, fn func() error) error {
	fnWithReturn := func() (struct{}, error) {
		return struct{}{}, fn()
	}

	_, err := Retry2(ctx, rp, fnWithReturn)
	return err
}

// Retry2 выполняет запрос к БД с учетом retryPolicy при возникновении ошибки Class 08
func Retry2[T any](ctx context.Context, rp RetryPolicy, fn func() (T, error)) (T, error) {
	if val1, err := fn(); err == nil || !isConnectionException(err) {
		return val1, err
	}

	var err error
	var ret1 T
	duration := rp.duration
	for i := 0; i < rp.retryCount; i++ {
		select {
		case <-time.NewTimer(time.Duration(duration) * time.Second).C:
			ret1, err = fn()
			if err == nil || !isConnectionException(err) {
				return ret1, err
			}
		case <-ctx.Done():
			return ret1, err
		}

		duration += rp.increment
	}

	return ret1, err
}

func isConnectionException(err error) bool {
	var tError *pgconn.PgError
	if errors.As(err, &tError) && pgerrcode.IsConnectionException(tError.Code) {
		return true
	}

	return false
}
