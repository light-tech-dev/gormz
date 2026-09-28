package gormz

import (
	"context"

	"gorm.io/gorm"
)

// ═══════════════════════════════════════════════
// Context — ربط DB بـ context
// ═══════════════════════════════════════════════

type contextKey struct {
	name string
}

var instanceKey = &contextKey{"gormz.instance"}

// WithDB يضع Instance في الـ context.
//
//	ctx := gormz.WithDB(ctx, app)
//	users, _ := gormz.FromContext[User](ctx).All()
func WithDB(ctx context.Context, i *Instance) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, instanceKey, i)
}

// WithGormDB يضع *gorm.DB في الـ context.
//
//	ctx := gormz.WithGormDB(ctx, db)
func WithGormDB(ctx context.Context, d *gorm.DB) context.Context {
	if d == nil {
		panic(ErrNilDB)
	}
	return WithDB(ctx, NewInstance(d))
}

// DBFromContext يستخرج الـ Instance من الـ context.
func DBFromContext(ctx context.Context) (*Instance, bool) {
	if ctx == nil {
		return nil, false
	}
	i, ok := ctx.Value(instanceKey).(*Instance)
	return i, ok
}

// FromContext ينشئ QuerySet من الـ context.
//
// إن لم يكن هناك DB في الـ context، يستخدم الـ global.
func FromContext[T any](ctx context.Context) *QuerySet[T] {
	if i, ok := DBFromContext(ctx); ok {
		q := newQuerySet[T](i.DB())
		q.ctx = ctx
		return q
	}
	// Fallback إلى global
	q := newQuerySet[T](DB())
	q.ctx = ctx
	return q
}

// MustFromContext مثل FromContext لكن يpanic إذا لم يوجد DB في context.
func MustFromContext[T any](ctx context.Context) *QuerySet[T] {
	i, ok := DBFromContext(ctx)
	if !ok {
		panic(ErrNotInitialized)
	}
	q := newQuerySet[T](i.DB())
	q.ctx = ctx
	return q
}