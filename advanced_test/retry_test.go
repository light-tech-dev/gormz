package advanced_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/light-tech-dev/gormz/advanced"
	"github.com/stretchr/testify/assert"
)

func TestRetry_Success(t *testing.T) {
	ctx := context.Background()
	var attempts int32

	cfg := advanced.DefaultRetryConfig()
	cfg.MaxAttempts = 3

	err := advanced.Retry(ctx, cfg, func() error {
		atomic.AddInt32(&attempts, 1)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	var attempts int32

	cfg := advanced.DefaultRetryConfig()
	cfg.MaxAttempts = 5
	cfg.InitialDelay = 10 * time.Millisecond

	err := advanced.Retry(ctx, cfg, func() error {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			return errors.New("deadlock")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestRetry_NonRetryable(t *testing.T) {
	ctx := context.Background()
	var attempts int32

	cfg := advanced.DefaultRetryConfig()
	cfg.MaxAttempts = 5
	cfg.RetryIf = func(err error) bool {
		return false
	}

	err := advanced.Retry(ctx, cfg, func() error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("fatal")
	})

	assert.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
}

func TestRetry_MaxAttempts(t *testing.T) {
	ctx := context.Background()
	var attempts int32

	cfg := advanced.DefaultRetryConfig()
	cfg.MaxAttempts = 3
	cfg.InitialDelay = 5 * time.Millisecond

	err := advanced.Retry(ctx, cfg, func() error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("deadlock")
	})

	assert.Error(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestRetry_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := advanced.Retry(ctx, advanced.DefaultRetryConfig(), func() error {
		return nil
	})

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestRetry_NilFunction(t *testing.T) {
	ctx := context.Background()
	err := advanced.Retry(ctx, advanced.DefaultRetryConfig(), nil)
	assert.Error(t, err)
}

func TestIsRetryableError(t *testing.T) {
	assert.True(t, advanced.IsRetryableError(errors.New("deadlock")))
	assert.True(t, advanced.IsRetryableError(errors.New("database is locked")))
	assert.True(t, advanced.IsRetryableError(errors.New("connection reset")))
	assert.False(t, advanced.IsRetryableError(errors.New("fatal error")))
	assert.False(t, advanced.IsRetryableError(nil))
}

func TestBackoff_Exponential(t *testing.T) {
	backoff := advanced.ExponentialBackoff(100*time.Millisecond, 10*time.Second)
	assert.Equal(t, 100*time.Millisecond, backoff(1))
	assert.Equal(t, 200*time.Millisecond, backoff(2))
	assert.Equal(t, 400*time.Millisecond, backoff(3))
	assert.Equal(t, 800*time.Millisecond, backoff(4))
}

func TestBackoff_Linear(t *testing.T) {
	backoff := advanced.LinearBackoff(100*time.Millisecond, 5*time.Second)
	assert.Equal(t, 100*time.Millisecond, backoff(1))
	assert.Equal(t, 200*time.Millisecond, backoff(2))
	assert.Equal(t, 300*time.Millisecond, backoff(3))
}

func TestBackoff_Constant(t *testing.T) {
	backoff := advanced.ConstantBackoff(500 * time.Millisecond)
	assert.Equal(t, 500*time.Millisecond, backoff(1))
	assert.Equal(t, 500*time.Millisecond, backoff(2))
	assert.Equal(t, 500*time.Millisecond, backoff(5))
}

func TestRetryWithBackoff_Success(t *testing.T) {
	ctx := context.Background()
	var attempts int32

	backoff := advanced.ExponentialBackoff(10*time.Millisecond, 100*time.Millisecond)

	err := advanced.RetryWithBackoff(ctx, 5, backoff, func() error {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			return errors.New("deadlock")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestRetryWithBackoff_NonRetryable(t *testing.T) {
	ctx := context.Background()
	var attempts int32

	backoff := advanced.ConstantBackoff(10 * time.Millisecond)

	err := advanced.RetryWithBackoff(ctx, 5, backoff, func() error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("fatal error")
	})

	assert.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
}