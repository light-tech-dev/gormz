package advanced

import (
	"context"
	"fmt"

	"github.com/light-tech-dev/gormz"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ═══════════════════════════════════════════════
// LockMode — وضع القفل
// ═══════════════════════════════════════════════

// LockMode وضع القفل الـ Pessimistic.
type LockMode string

const (
	// LockUpdate → FOR UPDATE
	LockUpdate LockMode = "UPDATE"

	// LockNoKeyUpdate → FOR NO KEY UPDATE (PostgreSQL فقط)
	LockNoKeyUpdate LockMode = "NO KEY UPDATE"

	// LockShare → FOR SHARE
	LockShare LockMode = "SHARE"

	// LockKeyShare → FOR KEY SHARE (PostgreSQL فقط)
	LockKeyShare LockMode = "KEY SHARE"
)

// ═══════════════════════════════════════════════
// Pessimistic Locking — FOR UPDATE / FOR SHARE
// ═══════════════════════════════════════════════

// PessimisticQuery يمثل استعلامًا مع قفل pessimistic.
//
//	users, err := advanced.WithLock[User](
//	    gormz.New[User]().Filter("id", 1),
//	).ForUpdate().All()
type PessimisticQuery[T any] struct {
	q    *gormz.QuerySet[T]
	mode LockMode
}

// WithLock ينشئ PessimisticQuery.
func WithLock[T any](q *gormz.QuerySet[T]) *PessimisticQuery[T] {
	return &PessimisticQuery[T]{q: q, mode: LockUpdate}
}

// ForUpdate يفعّل FOR UPDATE.
func (pq *PessimisticQuery[T]) ForUpdate() *PessimisticQuery[T] {
	pq.mode = LockUpdate
	return pq
}

// ForShare يفعّل FOR SHARE.
func (pq *PessimisticQuery[T]) ForShare() *PessimisticQuery[T] {
	pq.mode = LockShare
	return pq
}

// ForNoKeyUpdate يفعّل FOR NO KEY UPDATE (PostgreSQL).
func (pq *PessimisticQuery[T]) ForNoKeyUpdate() *PessimisticQuery[T] {
	pq.mode = LockNoKeyUpdate
	return pq
}

// ForKeyShare يفعّل FOR KEY SHARE (PostgreSQL).
func (pq *PessimisticQuery[T]) ForKeyShare() *PessimisticQuery[T] {
	pq.mode = LockKeyShare
	return pq
}

// Filter يضيف فلتر.
func (pq *PessimisticQuery[T]) Filter(field string, value any) *PessimisticQuery[T] {
	pq.q = pq.q.Filter(field, value)
	return pq
}

// Where يضيف SQL خام.
func (pq *PessimisticQuery[T]) Where(sql string, args ...any) *PessimisticQuery[T] {
	pq.q = pq.q.Where(sql, args...)
	return pq
}

// OrderBy يضيف ترتيبًا.
func (pq *PessimisticQuery[T]) OrderBy(fields ...string) *PessimisticQuery[T] {
	pq.q = pq.q.OrderBy(fields...)
	return pq
}

// Limit يحدد العدد.
func (pq *PessimisticQuery[T]) Limit(n int) *PessimisticQuery[T] {
	pq.q = pq.q.Limit(n)
	return pq
}

// All ينفّذ.
func (pq *PessimisticQuery[T]) All() ([]T, error) {
	var results []T
	err := pq.build().Find(&results).Error
	return results, err
}

// First يرجّع أول سجل.
func (pq *PessimisticQuery[T]) First() (*T, error) {
	var result T
	err := pq.build().First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get يرجّع سجلًا بالـ ID.
func (pq *PessimisticQuery[T]) Get(id any) (*T, error) {
	var result T
	err := pq.build().Where("id = ?", id).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update يحدّث السجلات المقفلة.
func (pq *PessimisticQuery[T]) Update(values map[string]any) (int64, error) {
	var zero T
	res := pq.build().Model(&zero).Updates(values)
	return res.RowsAffected, res.Error
}

// Delete يحذف السجلات المقفلة.
func (pq *PessimisticQuery[T]) Delete() (int64, error) {
	var zero T
	res := pq.build().Delete(&zero)
	return res.RowsAffected, res.Error
}

// build يبني الاستعلام مع القفل.
func (pq *PessimisticQuery[T]) build() *gorm.DB {
	db := pq.q.Build()
	strength := string(pq.mode)
	if strength == "" {
		strength = string(LockUpdate)
	}
	return db.Clauses(clause.Locking{Strength: strength})
}

// ═══════════════════════════════════════════════
// Optimistic Locking — Version-based
// ═══════════════════════════════════════════════

// DefaultVersionField اسم حقل الإصدار الافتراضي.
const DefaultVersionField = "version"

// OptimisticQuery يمثل استعلامًا مع قفل optimistic.
//
// ⚠️ الموديل يجب أن يحتوي على حقل version (int64 أو int).
//
//	type Order struct {
//	    ID      uint
//	    Version int64  // ← required
//	    Status  string
//	}
type OptimisticQuery[T any] struct {
	q     *gormz.QuerySet[T]
	field string
}

// WithOptimisticLock ينشئ OptimisticQuery.
func WithOptimisticLock[T any](q *gormz.QuerySet[T]) *OptimisticQuery[T] {
	return &OptimisticQuery[T]{
		q:     q,
		field: DefaultVersionField,
	}
}

// Field يحدد اسم حقل الإصدار (افتراضي: "version").
func (oq *OptimisticQuery[T]) Field(field string) *OptimisticQuery[T] {
	if field != "" {
		oq.field = field
	}
	return oq
}

// UpdateIfVersion يحدّث فقط إذا كان الإصدار مطابقًا.
//
// يرجّع خطأ إذا تغيّر الإصدار (تعارض).
//
//	err := advanced.WithOptimisticLock[Order](gormz.New[Order]()).
//	    UpdateIfVersion(ctx, 1, 5, map[string]any{"status": "paid"})
//	// UPDATE orders SET status='paid', version=version+1
//	// WHERE id=1 AND version=5
func (oq *OptimisticQuery[T]) UpdateIfVersion(
	ctx context.Context,
	id any,
	expectedVersion int64,
	updates map[string]any,
) error {
	if len(updates) == 0 {
		return nil
	}

	// تحقق من الحقول
	for k := range updates {
		if k == oq.field {
			return gormz.NewValidationError(k, "cannot update version field directly")
		}
		if err := gormz.ValidateField(k); err != nil {
			return err
		}
	}

	// نضيف زيادة على version
	versionField := oq.field
	updates[versionField] = gorm.Expr(versionField + " + 1")

	var zero T
	res := oq.q.DB().
		WithContext(ctx).
		Model(&zero).
		Where("id = ? AND "+versionField+" = ?", id, expectedVersion).
		Updates(updates)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return gormz.NewValidationError(
			versionField,
			fmt.Sprintf("version mismatch for id=%v (expected %d)", id, expectedVersion),
		)
	}

	return nil
}

// DeleteIfVersion يحذف فقط إذا كان الإصدار مطابقًا.
func (oq *OptimisticQuery[T]) DeleteIfVersion(
	ctx context.Context,
	id any,
	expectedVersion int64,
) error {
	var zero T
	res := oq.q.DB().
		WithContext(ctx).
		Where("id = ? AND "+oq.field+" = ?", id, expectedVersion).
		Delete(&zero)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return gormz.NewValidationError(
			oq.field,
			fmt.Sprintf("version mismatch for id=%v (expected %d)", id, expectedVersion),
		)
	}

	return nil
}

// IncrementVersion يزيد الإصدار يدويًا (نادرًا ما يُستخدم).
//
// مفيد إذا أردت أن تزيد الإصدار بدون تغيير حقول أخرى:
//
//	err := oq.IncrementVersion(ctx, 1)
func (oq *OptimisticQuery[T]) IncrementVersion(
	ctx context.Context,
	id any,
) error {
	var zero T
	res := oq.q.DB().
		WithContext(ctx).
		Model(&zero).
		Where("id = ?", id).
		UpdateColumn(oq.field, gorm.Expr(oq.field+" + 1"))

	return res.Error
}

// ═══════════════════════════════════════════════
// Lock Clause Helpers
// ═══════════════════════════════════════════════

// LockClause ينشئ clause.Locking.
//
// مفيد للاستخدام المباشر مع gorm.DB.
func LockClause(mode LockMode) clause.Locking {
	if mode == "" {
		mode = LockUpdate
	}
	return clause.Locking{Strength: string(mode)}
}

// LockClauseWithTable ينشئ clause.Locking مع جدول محدد.
func LockClauseWithTable(mode LockMode, table string) clause.Locking {
	l := LockClause(mode)
	l.Table = clause.Table{Name: table}
	return l
}
