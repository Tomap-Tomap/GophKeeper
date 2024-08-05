package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// RetryPolicy defines the retry policy for database queries
type RetryPolicy struct {
	retryCount int
	duration   int
	increment  int
}

// NewRetryPolicy initializes a new retry policy in millisecinds.
func NewRetryPolicy(retryCount, duration, increment int) *RetryPolicy {
	return &RetryPolicy{retryCount, duration, increment}
}

// Retry executes a database query considering the retry policy in case of Class 08 errors
func Retry(ctx context.Context, rp RetryPolicy, fn func() error) error {
	fnWithReturn := func() (struct{}, error) {
		return struct{}{}, fn()
	}

	_, err := Retry2(ctx, rp, fnWithReturn)
	return err
}

// Retry2 executes a database query considering the retry policy in case of Class 08 errors
func Retry2[T any](ctx context.Context, rp RetryPolicy, fn func() (T, error)) (T, error) {
	if val1, err := fn(); err == nil || !isConnectionException(err) {
		return val1, err
	}

	var err error
	var ret1 T
	duration := rp.duration
	for i := 0; i < rp.retryCount; i++ {
		select {
		case <-time.After(time.Duration(duration) * time.Millisecond):
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
