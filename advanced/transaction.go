package advanced

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/light-tech-dev/gormz"
	"gorm.io/gorm"
)

// ═══════════════════════════════════════════════
// Transaction — معاملات متقدمة
// ═══════════════════════════════════════════════

// TxConfig إعدادات المعاملة.
type TxConfig struct {
	// Isolation مستوى العزل.
	Isolation Isolation

	// ReadOnly معاملة للقراءة فقط.
	ReadOnly bool

	// Timeout مهلة المعاملة.
	Timeout time.Duration

	// Retries عدد المحاولات عند الفشل (deadlock).
	Retries int

	// RetryDelay التأخير بين المحاولات.
	RetryDelay time.Duration
}

// DefaultTxConfig إعدادات افتراضية.
func DefaultTxConfig() TxConfig {
	return TxConfig{
		Isolation:  IsolationDefault,
		Retries:    3,
		RetryDelay: 100 * time.Millisecond,
	}
}

// Isolation مستوى العزل.
type Isolation string

const (
	IsolationDefault         Isolation = ""
	IsolationReadUncommitted Isolation = "READ UNCOMMITTED"
	IsolationReadCommitted   Isolation = "READ COMMITTED"
	IsolationRepeatableRead  Isolation = "REPEATABLE READ"
	IsolationSerializable    Isolation = "SERIALIZABLE"
)

// ═══════════════════════════════════════════════
// Savepoint name validation
// ═══════════════════════════════════════════════

var savepointNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func validateSavepointName(name string) error {
	if name == "" {
		return errors.New("gormz/advanced: empty savepoint name")
	}
	if len(name) > 64 {
		return errors.New("gormz/advanced: savepoint name too long")
	}
	if !savepointNameRegex.MatchString(name) {
		return fmt.Errorf("gormz/advanced: invalid savepoint name %q", name)
	}
	return nil
}

// ═══════════════════════════════════════════════
// Tx — يمثل معاملة
// ═══════════════════════════════════════════════

// Tx يمثل معاملة مع API gormz.
type Tx struct {
	tx     *gorm.DB
	ctx    context.Context
	config TxConfig
}

// Begin يبدأ معاملة.
//
//	tx, err := advanced.Begin(ctx, advanced.DefaultTxConfig())
//	if err != nil {
//	    return err
//	}
//	defer tx.RollbackIfActive()
//
//	if err := advanced.Query[User](tx).Create(&user); err != nil {
//	    return err
//	}
//
//	return tx.Commit()
func Begin(ctx context.Context, cfg TxConfig) (*Tx, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	db := gormz.DB().WithContext(ctx)

	tx := db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &Tx{
		tx:     tx,
		ctx:    ctx,
		config: cfg,
	}, nil
}

// DB يرجّع *gorm.DB الأساسي.
func (t *Tx) DB() *gorm.DB {
	return t.tx
}

// Instance يرجّع Instance على هذه المعاملة.
//
// استخدمه مع gormz.NewWith:
//
//	users, _ := gormz.NewWith[User](tx.Instance()).All()
func (t *Tx) Instance() *gormz.Instance {
	if t.tx == nil {
		panic("gormz/advanced: transaction already finished")
	}
	return gormz.NewInstance(t.tx)
}

// Context يرجّع context المعاملة.
func (t *Tx) Context() context.Context {
	return t.ctx
}

// Config يرجّع إعدادات المعاملة.
func (t *Tx) Config() TxConfig {
	return t.config
}

// IsActive يفحص إذا كانت المعاملة نشطة.
func (t *Tx) IsActive() bool {
	return t.tx != nil
}

// Commit ينفّذ commit.
func (t *Tx) Commit() error {
	if t.tx == nil {
		return errors.New("gormz/advanced: transaction already finished")
	}
	err := t.tx.Commit().Error
	t.tx = nil
	return err
}

// Rollback يتراجع.
func (t *Tx) Rollback() error {
	if t.tx == nil {
		return errors.New("gormz/advanced: transaction already finished")
	}
	err := t.tx.Rollback().Error
	t.tx = nil
	return err
}

// RollbackIfActive يتراجع إذا لم تُنتهِ المعاملة.
//
// استخدمه في defer:
//
//	defer tx.RollbackIfActive()
func (t *Tx) RollbackIfActive() {
	if t.tx != nil {
		_ = t.tx.Rollback()
		t.tx = nil
	}
}

// ═══════════════════════════════════════════════
// QuerySet Constructors — Standalone Functions
// ═══════════════════════════════════════════════

// Query ينشئ QuerySet على المعاملة.
//
//	advanced.Query[User](tx).Create(&user)
//
// ملاحظة: هذه دالة وليست method لأن Go لا يسمح بـ generic methods.
func Query[T any](t *Tx) *gormz.QuerySet[T] {
	if t == nil || t.tx == nil {
		panic("gormz/advanced: transaction already finished")
	}
	return gormz.NewWith[T](gormz.NewInstance(t.tx))
}

// QueryWithContext مثل Query لكن مع context.
func QueryWithContext[T any](t *Tx, ctx context.Context) *gormz.QuerySet[T] {
	if t == nil || t.tx == nil {
		panic("gormz/advanced: transaction already finished")
	}
	q := gormz.NewWith[T](gormz.NewInstance(t.tx))
	if ctx != nil {
		return q.WithContext(ctx)
	}
	return q
}

// ═══════════════════════════════════════════════
// WithTransaction — معاملة مع callback
// ═══════════════════════════════════════════════

// WithTransaction ينفّذ fn داخل معاملة.
// يتعامل مع panic + retry تلقائيًا.
//
//	err := advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
//	    func(tx *advanced.Tx) error {
//	        if err := advanced.Query[User](tx).Create(&user); err != nil {
//	            return err
//	        }
//	        return advanced.Query[Order](tx).Create(&order)
//	    })
func WithTransaction(ctx context.Context, cfg TxConfig, fn func(tx *Tx) error) error {
	if fn == nil {
		return errors.New("gormz/advanced: nil transaction callback")
	}

	var lastErr error
	attempts := cfg.Retries + 1
	if attempts < 1 {
		attempts = 1
	}

	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 && cfg.RetryDelay > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(cfg.RetryDelay * time.Duration(attempt)):
			}
		}

		err := runTxOnce(ctx, cfg, fn)
		if err == nil {
			return nil
		}

		lastErr = err

		// هل نعيد المحاولة؟
		if !isRetryableError(err) {
			return err
		}
	}

	return fmt.Errorf("gormz/advanced: transaction failed after %d attempts: %w", attempts, lastErr)
}

// runTxOnce ينفّذ معاملة واحدة.
func runTxOnce(ctx context.Context, cfg TxConfig, fn func(tx *Tx) error) (err error) {
	tx, err := Begin(ctx, cfg)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.RollbackIfActive()
			err = fmt.Errorf("gormz/advanced: panic in transaction: %v", r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.RollbackIfActive()
		return err
	}

	return tx.Commit()
}

// ═══════════════════════════════════════════════
// Error Detection
// ═══════════════════════════════════════════════

// isRetryableError يفحص إذا كان الخطأ قابلًا لإعادة المحاولة.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "deadlock") ||
		strings.Contains(msg, "lock wait timeout") ||
		strings.Contains(msg, "could not serialize") ||
		strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "busy")
}

// ═══════════════════════════════════════════════
// Nested Transactions (Savepoints)
// ═══════════════════════════════════════════════

// Nested ينفّذ معاملة متداخلة (savepoint).
//
//	advanced.WithTransaction(ctx, cfg, func(tx *advanced.Tx) error {
//	    return tx.Nested(func(inner *advanced.Tx) error {
//	        // savepoint
//	        return nil
//	    })
//	})
func (t *Tx) Nested(fn func(tx *Tx) error) error {
	if t.tx == nil {
		return errors.New("gormz/advanced: transaction not active")
	}
	if fn == nil {
		return errors.New("gormz/advanced: nil nested callback")
	}

	savepointName := fmt.Sprintf("sp_%d", time.Now().UnixNano())

	if err := t.tx.SavePoint(savepointName).Error; err != nil {
		return fmt.Errorf("gormz/advanced: create savepoint failed: %w", err)
	}

	if err := fn(t); err != nil {
		if rbErr := t.tx.RollbackTo(savepointName).Error; rbErr != nil {
			return fmt.Errorf("gormz/advanced: rollback to savepoint failed: %w (original: %v)", rbErr, err)
		}
		return err
	}

	return nil
}

// ═══════════════════════════════════════════════
// Manual Savepoint Control
// ═══════════════════════════════════════════════

// Savepoint ينشئ savepoint.
//
// ⚠️ الاسم يجب أن يطابق: ^[a-zA-Z_][a-zA-Z0-9_]*$
func (t *Tx) Savepoint(name string) error {
	if t.tx == nil {
		return errors.New("gormz/advanced: transaction not active")
	}
	if err := validateSavepointName(name); err != nil {
		return err
	}
	return t.tx.SavePoint(name).Error
}

// RollbackTo يتراجع إلى savepoint.
func (t *Tx) RollbackTo(name string) error {
	if t.tx == nil {
		return errors.New("gormz/advanced: transaction not active")
	}
	if err := validateSavepointName(name); err != nil {
		return err
	}
	return t.tx.RollbackTo(name).Error
}

// ReleaseSavepoint يحذف savepoint.
func (t *Tx) ReleaseSavepoint(name string) error {
	if t.tx == nil {
		return errors.New("gormz/advanced: transaction not active")
	}
	if err := validateSavepointName(name); err != nil {
		return err
	}
	return t.tx.Exec("RELEASE SAVEPOINT " + name).Error
}

// ═══════════════════════════════════════════════
// Convenience Wrappers
// ═══════════════════════════════════════════════

// MustBegin مثل Begin لكن يpanic عند الخطأ.
func MustBegin(ctx context.Context, cfg TxConfig) *Tx {
	tx, err := Begin(ctx, cfg)
	if err != nil {
		panic(err)
	}
	return tx
}

// SimpleTransaction ينفّذ معاملة بإعدادات افتراضية.
func SimpleTransaction(ctx context.Context, fn func(tx *Tx) error) error {
	return WithTransaction(ctx, DefaultTxConfig(), fn)
}