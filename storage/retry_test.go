//go:build unit

package storage

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func Test_isConnectionException(t *testing.T) {
	t.Run("is connection exception", func(t *testing.T) {
		require.True(t, isConnectionException(&pgconn.PgError{Code: "08000"}))
	})

	t.Run("isn't connection exception", func(t *testing.T) {
		require.False(t, isConnectionException(&pgconn.PgError{Code: "02000"}))
	})

	t.Run("random error", func(t *testing.T) {
		require.False(t, isConnectionException(errors.New("test")))
	})
}

func Test_retry2(t *testing.T) {
	t.Run("test no error", func(t *testing.T) {
		t.Parallel()
		rp := RetryPolicy{
			retryCount: 3,
			duration:   1,
			increment:  2,
		}
		got, err := Retry2(context.Background(), rp, func() (int, error) {
			return 0, nil
		})

		require.NoError(t, err)
		require.Equal(t, 0, got)
	})

	t.Run("test error", func(t *testing.T) {
		t.Parallel()
		rp := RetryPolicy{
			retryCount: 3,
			duration:   1,
			increment:  2,
		}
		_, err := Retry2(context.Background(), rp, func() (*int, error) {
			return nil, &pgconn.PgError{Code: "02000"}
		})

		require.Error(t, err)
	})

	t.Run("test error connection", func(t *testing.T) {
		t.Parallel()
		rp := RetryPolicy{
			retryCount: 3,
			duration:   1,
			increment:  2,
		}
		_, err := Retry2(context.Background(), rp, func() (*int, error) {
			return nil, &pgconn.PgError{Code: "08000"}
		})

		require.Error(t, err)
	})

	t.Run("test error resolved", func(t *testing.T) {
		t.Parallel()
		rp := RetryPolicy{
			retryCount: 3,
			duration:   1,
			increment:  2,
		}

		var errConn error = &pgconn.PgError{Code: "08000"}
		var mu sync.RWMutex

		go func() {
			time.Sleep(5 * time.Second)
			mu.Lock()
			defer mu.Unlock()
			errConn = nil
		}()

		_, err := Retry2(context.Background(), rp, func() (*int, error) {
			mu.RLock()
			defer mu.RUnlock()
			return nil, errConn
		})

		require.NoError(t, err)
	})

	t.Run("test context done", func(t *testing.T) {
		t.Parallel()
		rp := RetryPolicy{
			retryCount: 3,
			duration:   1,
			increment:  2,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, err := Retry2(ctx, rp, func() (*int, error) {
			return nil, &pgconn.PgError{Code: "08000"}
		})

		require.Error(t, err)
	})
}

func Test_retry(t *testing.T) {
	t.Run("test no error", func(t *testing.T) {
		t.Parallel()
		rp := RetryPolicy{
			retryCount: 3,
			duration:   1,
			increment:  2,
		}
		err := Retry(context.Background(), rp, func() error {
			return nil
		})

		require.NoError(t, err)
	})

	t.Run("test error", func(t *testing.T) {
		t.Parallel()
		rp := RetryPolicy{
			retryCount: 3,
			duration:   1,
			increment:  2,
		}
		err := Retry(context.Background(), rp, func() error {
			return &pgconn.PgError{Code: "02000"}
		})

		require.Error(t, err)
	})

	t.Run("test error connection", func(t *testing.T) {
		t.Parallel()
		rp := RetryPolicy{
			retryCount: 3,
			duration:   1,
			increment:  2,
		}
		err := Retry(context.Background(), rp, func() error {
			return &pgconn.PgError{Code: "08000"}
		})

		require.Error(t, err)
	})

	t.Run("test error resolved", func(t *testing.T) {
		t.Parallel()
		rp := RetryPolicy{
			retryCount: 3,
			duration:   1,
			increment:  2,
		}

		var errConn error = &pgconn.PgError{Code: "08000"}
		var mu sync.RWMutex

		go func() {
			time.Sleep(5 * time.Second)
			mu.Lock()
			defer mu.Unlock()
			errConn = nil
		}()

		err := Retry(context.Background(), rp, func() error {
			mu.RLock()
			defer mu.RUnlock()
			return errConn
		})

		require.NoError(t, err)
	})

	t.Run("test context done", func(t *testing.T) {
		t.Parallel()
		rp := RetryPolicy{
			retryCount: 3,
			duration:   1,
			increment:  2,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := Retry(ctx, rp, func() error {
			return &pgconn.PgError{Code: "08000"}
		})

		require.Error(t, err)
	})
}
