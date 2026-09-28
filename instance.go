package gormz

import (
	"context"

	"gorm.io/gorm"
)

// ═══════════════════════════════════════════════
// Instance — دعم قواعد بيانات متعددة
// ═══════════════════════════════════════════════

// Instance يمثل اتصالًا مستقلًا بـ GORM.
//
// مفيد لـ:
//   - Multi-database setups
//   - Testing (كل test له DB منفصل)
//   - Read/Write split
type Instance struct {
	db *gorm.DB
}

// NewInstance ينشئ Instance جديدًا.
//
// يpanic إذا كان db == nil.
func NewInstance(d *gorm.DB) *Instance {
	if d == nil {
		panic(ErrNilDB)
	}
	return &Instance{db: d}
}

// DB يرجّع *gorm.DB الأصلي.
func (i *Instance) DB() *gorm.DB {
	return i.db
}

// Ping يفحص الاتصال.
func (i *Instance) Ping() error {
	sqlDB, err := i.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Close يغلق الاتصال.
func (i *Instance) Close() error {
	sqlDB, err := i.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Configure يضبط connection pool.
func (i *Instance) Configure(cfg Config) error {
	sqlDB, err := i.db.DB()
	if err != nil {
		return err
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	return nil
}

// Transaction ينفذ عملية داخل transaction.
func (i *Instance) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	if fn == nil {
		return NewValidationError("fn", "nil transaction callback")
	}
	return i.db.WithContext(ctx).Transaction(fn)
}

// Migrate يهاجر موديلًا على هذه الـ Instance.
func (i *Instance) Migrate(models ...any) error {
	if len(models) == 0 {
		return nil
	}
	return i.db.AutoMigrate(models...)
}

// ═══════════════════════════════════════════════
// Global Instance
// ═══════════════════════════════════════════════

// GlobalInstance يرجّع Instance بناءً على الاتصال العام.
//
// يpanic إذا لم يُنادى SetDB.
func GlobalInstance() *Instance {
	if globalDB == nil {
		panic(ErrNotInitialized)
	}
	return &Instance{db: globalDB}
}

// ═══════════════════════════════════════════════
// QuerySet Constructors — Since Go Doesn't Allow
// Generic Methods, These Are Standalone Functions
// ═══════════════════════════════════════════════

// NewWith ينشئ QuerySet[T] على Instance معين.
//
//	app := gormz.NewInstance(db)
//	users, _ := gormz.NewWith[User](app).All()
func NewWith[T any](i *Instance) *QuerySet[T] {
	if i == nil {
		panic(ErrNilDB)
	}
	return newQuerySet[T](i.db)
}

// QueryOn ينشئ QuerySet جديدًا على Instance.
//
//	app := gormz.NewInstance(db)
//	users, _ := gormz.QueryOn[User](app).Filter("active", true).All()
//
// ملاحظة: هذه دالة وليست method لأن Go لا يسمح بـ generic methods.
func QueryOn[T any](i *Instance) *QuerySet[T] {
	if i == nil {
		panic(ErrNilDB)
	}
	return newQuerySet[T](i.db)
}

// QueryOnWithContext مثل QueryOn لكن مع context.
func QueryOnWithContext[T any](i *Instance, ctx context.Context) *QuerySet[T] {
	if i == nil {
		panic(ErrNilDB)
	}
	q := newQuerySet[T](i.db)
	if ctx != nil {
		q.ctx = ctx
	}
	return q
}