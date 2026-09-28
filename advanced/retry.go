package advanced

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/light-tech-dev/gormz"
	"gorm.io/gorm"
)

// ═══════════════════════════════════════════════
// Retry — إعادة المحاولة
// ═══════════════════════════════════════════════

// RetryConfig إعدادات إعادة المحاولة.
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       float64
	RetryIf      func(error) bool
}

// DefaultRetryConfig إعدادات افتراضية.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.1,
		RetryIf:      IsRetryableError,
	}
}

// Retry ينفّذ fn مع إعادة المحاولة.
func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	if fn == nil {
		return fmt.Errorf("gormz/advanced: nil retry function")
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}
	if cfg.Multiplier <= 0 {
		cfg.Multiplier = 1.0
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var lastErr error
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if cfg.RetryIf != nil && !cfg.RetryIf(err) {
			return err
		}

		if attempt == cfg.MaxAttempts {
			break
		}

		wait := delay
		if cfg.Jitter > 0 {
			jitter := time.Duration(float64(delay) * cfg.Jitter * (rand.Float64()*2 - 1))
			wait += jitter
			if wait < 0 {
				wait = 0
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}

		delay = time.Duration(float64(delay) * cfg.Multiplier)
		if cfg.MaxDelay > 0 && delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}

	return fmt.Errorf("gormz/advanced: max attempts (%d) reached: %w", cfg.MaxAttempts, lastErr)
}

// RetryDB ينفّذ عملية DB مع retry.
func RetryDB(ctx context.Context, fn func(db *gorm.DB) error) error {
	if fn == nil {
		return fmt.Errorf("gormz/advanced: nil retry function")
	}
	cfg := DefaultRetryConfig()
	return Retry(ctx, cfg, func() error {
		return fn(DBFromContextOrGlobal(ctx))
	})
}

// DBFromContextOrGlobal يستخرج DB من context أو يستخدم global.
func DBFromContextOrGlobal(ctx context.Context) *gorm.DB {
	if ctx != nil {
		if i, ok := gormz.DBFromContext(ctx); ok {
			return i.DB()
		}
	}
	return gormz.DB()
}

// ═══════════════════════════════════════════════
// Error Detection
// ═══════════════════════════════════════════════

// IsRetryableError يفحص إذا كان الخطأ قابلًا لإعادة المحاولة.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "deadlock") ||
		strings.Contains(msg, "lock wait timeout") ||
		strings.Contains(msg, "could not serialize") ||
		strings.Contains(msg, "serialization failure") ||
		strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "too many connections") ||
		strings.Contains(msg, "server closed the connection") ||
		strings.Contains(msg, "i/o timeout")
}

// ═══════════════════════════════════════════════
// Backoff Strategies
// ═══════════════════════════════════════════════

// BackoffStrategy استراتيجية التراجع.
type BackoffStrategy func(attempt int) time.Duration

// ExponentialBackoff تراجع أسي.
func ExponentialBackoff(initial, max time.Duration) BackoffStrategy {
	if initial <= 0 {
		initial = 100 * time.Millisecond
	}
	return func(attempt int) time.Duration {
		if attempt < 1 {
			attempt = 1
		}
		delay := time.Duration(float64(initial) * math.Pow(2, float64(attempt-1)))
		if max > 0 && delay > max {
			delay = max
		}
		return delay
	}
}

// LinearBackoff تراجع خطي.
func LinearBackoff(step, max time.Duration) BackoffStrategy {
	if step <= 0 {
		step = 100 * time.Millisecond
	}
	return func(attempt int) time.Duration {
		if attempt < 1 {
			attempt = 1
		}
		delay := step * time.Duration(attempt)
		if max > 0 && delay > max {
			delay = max
		}
		return delay
	}
}

// ConstantBackoff تراجع ثابت.
func ConstantBackoff(delay time.Duration) BackoffStrategy {
	if delay <= 0 {
		delay = 100 * time.Millisecond
	}
	return func(attempt int) time.Duration {
		return delay
	}
}

// RetryWithBackoff ينفّذ fn مع إعادة المحاولة باستخدام BackoffStrategy.
func RetryWithBackoff(
	ctx context.Context,
	maxAttempts int,
	backoff BackoffStrategy,
	fn func() error,
) error {
	if fn == nil {
		return fmt.Errorf("gormz/advanced: nil retry function")
	}
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	if backoff == nil {
		backoff = ConstantBackoff(100 * time.Millisecond)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		if !IsRetryableError(err) {
			return err
		}

		if attempt == maxAttempts {
			break
		}

		wait := backoff(attempt)
		if wait > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}
	}

	return fmt.Errorf("gormz/advanced: max attempts (%d) reached: %w", maxAttempts, lastErr)
}