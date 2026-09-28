package gormz

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/light-tech-dev/gormz/internal"
	"gorm.io/gorm"
)

// ═══════════════════════════════════════════════
// QuerySet — استعلام متسلسل type-safe
// ═══════════════════════════════════════════════

// QuerySet[T] هو builder للاستعلامات لموديل T.
//
// كل method يعيد QuerySet جديد (immutable)، لذا آمن للاستخدام المتزامن.
//
//	users, _ := gormz.New[User]().
//	    Filter("active", true).
//	    Filter("age__gte", 18).
//	    OrderBy("-created_at").
//	    Limit(10).
//	    All()
//
// # Panic vs Error
//
// طرق بادئتها Try ترجع الخطأ بدلًا من panic.
// استخدمها عند التعامل مع مدخلات المستخدم:
//
//	q, err := gormz.New[User]().TryFilter(field, value)
type QuerySet[T any] struct {
	db          *gorm.DB
	ctx         context.Context
	conditions  []Clause
	orders      []string
	selects     []string
	omits       []string
	preloads    []string
	groupBy     []string
	having      []rawHaving
	joins       []joinClause
	windows     []string
	distinct    bool
	distinctOn  []string
	limit       int
	offset      int
	unscoped    bool
	onlyDeleted bool
}

// newQuerySet ينشئ QuerySet جديد (داخلي).
func newQuerySet[T any](d *gorm.DB) *QuerySet[T] {
	if d == nil {
		panic(ErrNilDB)
	}
	return &QuerySet[T]{
		db:  d,
		ctx: context.Background(),
	}
}

// New ينشئ QuerySet جديد للموديل T.
//
//	users, _ := gormz.New[User]().All()
func New[T any]() *QuerySet[T] {
	return newQuerySet[T](DB())
}

// ═══════════════════════════════════════════════
// Context
// ═══════════════════════════════════════════════

// WithContext يربط context.
//
//	users, _ := gormz.New[User]().
//	    WithContext(ctx).
//	    Filter("active", true).
//	    All()
func (q *QuerySet[T]) WithContext(ctx context.Context) *QuerySet[T] {
	if ctx == nil {
		ctx = context.Background()
	}
	nq := q.clone()
	nq.ctx = ctx
	return nq
}

// Context يرجّع context الحالي.
func (q *QuerySet[T]) Context() context.Context {
	return q.ctx
}

// ═══════════════════════════════════════════════
// Filters — الفلاتر
// ═══════════════════════════════════════════════

// Filter يضيف شرطًا. يpanic عند خطأ في الحقل.
//
//	.Filter("name", "Ali")             → name = 'Ali'
//	.Filter("age__gt", 18)             → age > 18
//	.Filter("title__contains", "Go")   → title LIKE '%Go%'
//	.Filter("id__in", []int{1,2,3})    → id IN (1,2,3)
//
// استخدم TryFilter لتفادي panic.
func (q *QuerySet[T]) Filter(field string, value any) *QuerySet[T] {
	nq, err := q.TryFilter(field, value)
	if err != nil {
		panic(err)
	}
	return nq
}

// TryFilter مثل Filter لكن يرجّع خطأ بدلًا من panic.
func (q *QuerySet[T]) TryFilter(field string, value any) (*QuerySet[T], error) {
	clause, err := internal.BuildLookup(field, value, false)
	if err != nil {
		return nil, err
	}

	nq := q.clone()
	nq.conditions = append(nq.conditions, WhereClause{
		sql:  clause.SQL,
		args: clause.Args,
	})
	return nq, nil
}

// FilterIf يضيف شرطًا فقط إذا كان cond صحيحًا.
//
//	.FilterIf(search != "", "name__icontains", search)
func (q *QuerySet[T]) FilterIf(cond bool, field string, value any) *QuerySet[T] {
	if !cond {
		return q
	}
	return q.Filter(field, value)
}

// TryFilterIf مثل FilterIf لكن يرجّع خطأ.
func (q *QuerySet[T]) TryFilterIf(cond bool, field string, value any) (*QuerySet[T], error) {
	if !cond {
		return q, nil
	}
	return q.TryFilter(field, value)
}

// Exclude عكس Filter. يpanic عند خطأ.
func (q *QuerySet[T]) Exclude(field string, value any) *QuerySet[T] {
	nq, err := q.TryExclude(field, value)
	if err != nil {
		panic(err)
	}
	return nq
}

// TryExclude مثل Exclude لكن يرجّع خطأ.
func (q *QuerySet[T]) TryExclude(field string, value any) (*QuerySet[T], error) {
	clause, err := internal.BuildLookup(field, value, true)
	if err != nil {
		return nil, err
	}

	nq := q.clone()
	nq.conditions = append(nq.conditions, WhereClause{
		sql:  clause.SQL,
		args: clause.Args,
	})
	return nq, nil
}

// Where يقبل SQL خام.
//
//	.Where("age > ? AND status = ?", 18, "active")
//
// ⚠️ المسؤولية على المستخدم — لا يوجد validation.
func (q *QuerySet[T]) Where(sql string, args ...any) *QuerySet[T] {
	nq := q.clone()
	nq.conditions = append(nq.conditions, internal.RawClause{SQL: sql, Args: args})
	return nq
}

// Q يضيف مجموعة شروط معقدة.
//
//	q := gormz.Or(
//	    gormz.Eq("status", "active"),
//	    gormz.Eq("status", "pending"),
//	)
//	users, _ := gormz.New[User]().Q(q).All()
func (q *QuerySet[T]) Q(builder *Q) *QuerySet[T] {
	if builder == nil {
		return q
	}
	nq := q.clone()
	nq.conditions = append(nq.conditions, builder)
	return nq
}

// ═══════════════════════════════════════════════
// Ordering & Limits
// ═══════════════════════════════════════════════

// OrderBy يضيف ترتيبًا. يpanic عند خطأ.
//
//	.OrderBy("name")                       → name ASC
//	.OrderBy("-created_at")                → created_at DESC
//	.OrderBy("status", "-date")            → status ASC, date DESC
//	.OrderBy("name NULLS FIRST")           → name ASC NULLS FIRST
func (q *QuerySet[T]) OrderBy(fields ...string) *QuerySet[T] {
	nq, err := q.TryOrderBy(fields...)
	if err != nil {
		panic(err)
	}
	return nq
}

// TryOrderBy مثل OrderBy لكن يرجّع خطأ.
func (q *QuerySet[T]) TryOrderBy(fields ...string) (*QuerySet[T], error) {
	nq := q.clone()
	for _, f := range fields {
		parsed, err := parseOrderField(f)
		if err != nil {
			return nil, err
		}
		if parsed == "" {
			continue
		}
		nq.orders = append(nq.orders, parsed)
	}
	return nq, nil
}

// parseOrderField يحوّل "field", "-field", "field NULLS FIRST" إلى SQL ORDER BY.
func parseOrderField(f string) (string, error) {
	f = strings.TrimSpace(f)
	if f == "" {
		return "", nil
	}

	// DESC prefix
	desc := false
	if strings.HasPrefix(f, "-") {
		desc = true
		f = strings.TrimPrefix(f, "-")
	}

	// تحقق من وجود " NULLS FIRST" أو " NULLS LAST"
	upperF := strings.ToUpper(f)
	if idx := strings.Index(upperF, " NULLS "); idx > 0 {
		col := f[:idx]
		nullsClause := f[idx+1:] // "NULLS FIRST"
		if err := internal.ValidateField(col); err != nil {
			return "", err
		}
		dir := "ASC"
		if desc {
			dir = "DESC"
		}
		return col + " " + dir + " " + strings.ToUpper(nullsClause), nil
	}

	if err := internal.ValidateField(f); err != nil {
		return "", err
	}

	if desc {
		return f + " DESC", nil
	}
	return f + " ASC", nil
}

// Limit يحدد عدد النتائج.
func (q *QuerySet[T]) Limit(n int) *QuerySet[T] {
	nq := q.clone()
	nq.limit = n
	return nq
}

// Offset يحدد البداية.
func (q *QuerySet[T]) Offset(n int) *QuerySet[T] {
	nq := q.clone()
	nq.offset = n
	return nq
}

// Page يحدد الصفحة (بديل لـ Limit + Offset).
//
//	.Page(2, 20)  // الصفحة 2، 20 عنصر
func (q *QuerySet[T]) Page(page, perPage int) *QuerySet[T] {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	nq := q.clone()
	nq.limit = perPage
	nq.offset = (page - 1) * perPage
	return nq
}

// ═══════════════════════════════════════════════
// Field selection
// ═══════════════════════════════════════════════

// Select يحدد الأعمدة. يpanic عند خطأ.
//
//	.Select("id", "name", "email")
func (q *QuerySet[T]) Select(fields ...string) *QuerySet[T] {
	nq, err := q.TrySelect(fields...)
	if err != nil {
		panic(err)
	}
	return nq
}

// TrySelect مثل Select لكن يرجّع خطأ.
func (q *QuerySet[T]) TrySelect(fields ...string) (*QuerySet[T], error) {
	for _, f := range fields {
		if err := internal.ValidateField(f); err != nil {
			return nil, err
		}
	}
	nq := q.clone()
	nq.selects = append(nq.selects, fields...)
	return nq, nil
}

// SelectRaw يضيف تعبير SELECT خام (للـ COUNT، SUM، إلخ).
//
//	.SelectRaw("COUNT(*) as total")
//
// ⚠️ المسؤولية على المستخدم — لا يوجد validation.
func (q *QuerySet[T]) SelectRaw(exprs ...string) *QuerySet[T] {
	nq := q.clone()
	nq.selects = append(nq.selects, exprs...)
	return nq
}

// Omit يستثني الأعمدة. يpanic عند خطأ.
//
//	.Omit("password", "secret")
func (q *QuerySet[T]) Omit(fields ...string) *QuerySet[T] {
	nq, err := q.TryOmit(fields...)
	if err != nil {
		panic(err)
	}
	return nq
}

// TryOmit مثل Omit لكن يرجّع خطأ.
func (q *QuerySet[T]) TryOmit(fields ...string) (*QuerySet[T], error) {
	for _, f := range fields {
		if err := internal.ValidateField(f); err != nil {
			return nil, err
		}
	}
	nq := q.clone()
	nq.omits = append(nq.omits, fields...)
	return nq, nil
}

// ═══════════════════════════════════════════════
// Relations
// ═══════════════════════════════════════════════

// Preload يحمّل العلاقات.
//
//	.Preload("Orders")
//	.Preload("Orders.Items")
func (q *QuerySet[T]) Preload(relations ...string) *QuerySet[T] {
	nq := q.clone()
	nq.preloads = append(nq.preloads, relations...)
	return nq
}

// ═══════════════════════════════════════════════
// Soft delete
// ═══════════════════════════════════════════════

// WithDeleted يشمل السجلات المحذوفة (soft delete).
func (q *QuerySet[T]) WithDeleted() *QuerySet[T] {
	nq := q.clone()
	nq.unscoped = true
	return nq
}

// OnlyDeleted السجلات المحذوفة فقط.
func (q *QuerySet[T]) OnlyDeleted() *QuerySet[T] {
	nq := q.clone()
	nq.unscoped = true
	nq.onlyDeleted = true
	return nq
}

// ═══════════════════════════════════════════════
// GroupBy / Having / Distinct
// ═══════════════════════════════════════════════

// GroupBy يضيف GROUP BY.
func (q *QuerySet[T]) GroupBy(fields ...string) *QuerySet[T] {
	nq, err := q.TryGroupBy(fields...)
	if err != nil {
		panic(err)
	}
	return nq
}

// TryGroupBy مثل GroupBy لكن يرجّع خطأ.
func (q *QuerySet[T]) TryGroupBy(fields ...string) (*QuerySet[T], error) {
	for _, f := range fields {
		if err := internal.ValidateField(f); err != nil {
			return nil, err
		}
	}
	nq := q.clone()
	nq.groupBy = append(nq.groupBy, fields...)
	return nq, nil
}

// Having يضيف HAVING.
//
//	q.GroupBy("user_id").
//	    Having("SUM(total) > ?", 1000)
//
// ⚠️ المسؤولية على المستخدم — HAVING يستخدم تعبيرات معقدة.
func (q *QuerySet[T]) Having(sql string, args ...any) *QuerySet[T] {
	nq := q.clone()
	nq.having = append(nq.having, rawHaving{sql: sql, args: args})
	return nq
}

// Distinct يفعّل DISTINCT.
func (q *QuerySet[T]) Distinct() *QuerySet[T] {
	nq := q.clone()
	nq.distinct = true
	return nq
}

// DistinctOn يفعّل DISTINCT ON (PostgreSQL).
//
// ⚠️ PostgreSQL فقط.
func (q *QuerySet[T]) DistinctOn(fields ...string) *QuerySet[T] {
	for _, f := range fields {
		if err := internal.ValidateField(f); err != nil {
			panic(err)
		}
	}
	nq := q.clone()
	nq.distinct = true
	nq.distinctOn = append(nq.distinctOn, fields...)
	return nq
}

// ═══════════════════════════════════════════════
// Joins — الانضمامات
// ═══════════════════════════════════════════════

// Join يضيف JOIN مخصص.
//
//	q.Join("INNER", "orders", "orders.user_id = users.id")
//
// ⚠️ table و on مسؤولية المستخدم.
func (q *QuerySet[T]) Join(joinType, table, on string) *QuerySet[T] {
	nq := q.clone()
	nq.joins = append(nq.joins, joinClause{
		joinType: strings.ToUpper(strings.TrimSpace(joinType)),
		table:    table,
		on:       on,
	})
	return nq
}

// InnerJoin يضيف INNER JOIN.
func (q *QuerySet[T]) InnerJoin(table, on string) *QuerySet[T] {
	return q.Join("INNER", table, on)
}

// LeftJoin يضيف LEFT JOIN.
func (q *QuerySet[T]) LeftJoin(table, on string) *QuerySet[T] {
	return q.Join("LEFT", table, on)
}

// RightJoin يضيف RIGHT JOIN.
func (q *QuerySet[T]) RightJoin(table, on string) *QuerySet[T] {
	return q.Join("RIGHT", table, on)
}

// CrossJoin يضيف CROSS JOIN.
func (q *QuerySet[T]) CrossJoin(table string) *QuerySet[T] {
	return q.Join("CROSS", table, "")
}

// ═══════════════════════════════════════════════
// Window Functions
// ═══════════════════════════════════════════════

// Window يضيف window function للـ SELECT.
//
//	q.Window("ROW_NUMBER() OVER (PARTITION BY department_id) as rn")
//
// ⚠️ المسؤولية على المستخدم.
func (q *QuerySet[T]) Window(exprs ...string) *QuerySet[T] {
	nq := q.clone()
	nq.windows = append(nq.windows, exprs...)
	return nq
}

// ═══════════════════════════════════════════════
// Terminal: reads
// ═══════════════════════════════════════════════

// All يرجّع كل النتائج.
//
//	users, err := gormz.New[User]().Filter("active", true).All()
func (q *QuerySet[T]) All() ([]T, error) {
	var results []T
	err := q.build().Find(&results).Error
	return results, err
}

// First يرجّع أول سجل.
//
// يرجّع NotFoundError إذا لم يوجد.
func (q *QuerySet[T]) First() (*T, error) {
	var result T
	err := q.build().First(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewNotFoundError(internal.ModelName[T](), nil)
		}
		return nil, err
	}
	return &result, nil
}

// FirstOrNil يرجّع أول سجل أو (nil, nil).
func (q *QuerySet[T]) FirstOrNil() (*T, error) {
	var result T
	err := q.build().First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Last يرجّع آخر سجل.
func (q *QuerySet[T]) Last() (*T, error) {
	var result T
	err := q.build().Last(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewNotFoundError(internal.ModelName[T](), nil)
		}
		return nil, err
	}
	return &result, nil
}

// Get يرجّع سجلًا بالـ ID.
//
//	user, err := gormz.New[User]().Get(1)
//
// يرجّع NotFoundError إذا لم يوجد.
func (q *QuerySet[T]) Get(id any) (*T, error) {
	var result T
	err := q.build().Where("id = ?", id).First(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewNotFoundError(internal.ModelName[T](), id)
		}
		return nil, err
	}
	return &result, nil
}

// GetOrNil يرجّع سجلًا بالـ ID أو (nil, nil).
func (q *QuerySet[T]) GetOrNil(id any) (*T, error) {
	var result T
	err := q.build().Where("id = ?", id).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Find يبحث بحقل = قيمة.
//
//	user, err := gormz.New[User]().Find("email", "ali@example.com")
func (q *QuerySet[T]) Find(field string, value any) (*T, error) {
	if err := internal.ValidateField(field); err != nil {
		return nil, err
	}

	var result T
	err := q.build().Where(field+" = ?", value).First(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewNotFoundError(internal.ModelName[T](), value)
		}
		return nil, err
	}
	return &result, nil
}

// FindOrNil يبحث أو (nil, nil).
func (q *QuerySet[T]) FindOrNil(field string, value any) (*T, error) {
	if err := internal.ValidateField(field); err != nil {
		return nil, err
	}

	var result T
	err := q.build().Where(field+" = ?", value).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Count يرجّع العدد.
func (q *QuerySet[T]) Count() (int64, error) {
	var zero T
	var count int64
	err := q.build().Model(&zero).Count(&count).Error
	return count, err
}

// Exists يفحص وجود أي سجل.
//
// يستخدم LIMIT 1 + COUNT (أسرع من subquery).
func (q *QuerySet[T]) Exists() (bool, error) {
	var zero T
	var count int64

	err := q.build().
		Model(&zero).
		Limit(1).
		Count(&count).Error

	return count > 0, err
}

// Pluck يستخرج عمودًا واحدًا.
//
//	var names []string
//	err := gormz.New[User]().Filter("active", true).Pluck("name", &names)
func (q *QuerySet[T]) Pluck(field string, dest any) error {
	if err := internal.ValidateField(field); err != nil {
		return err
	}
	return q.build().Pluck(field, dest).Error
}

// ScanInto يقرأ في struct مخصص.
//
//	type UserStats struct {
//	    Status string
//	    Count  int64
//	}
//	var stats []UserStats
//	gormz.New[User]().SelectRaw("status", "COUNT(*) as count").
//	    ScanInto(&stats)
func (q *QuerySet[T]) ScanInto(dest any) error {
	if dest == nil {
		return fmt.Errorf("gormz: nil destination")
	}
	return q.build().Scan(dest).Error
}

// Take يرجّع سجلًا واحدًا بدون ترتيب.
func (q *QuerySet[T]) Take() (*T, error) {
	var result T
	err := q.build().Take(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewNotFoundError(internal.ModelName[T](), nil)
		}
		return nil, err
	}
	return &result, nil
}

// ═══════════════════════════════════════════════
// Terminal: writes
// ═══════════════════════════════════════════════

// Create ينشئ سجلًا.
//
//	user := &User{Name: "Ali", Email: "ali@test.com"}
//	err := gormz.New[User]().Create(user)
func (q *QuerySet[T]) Create(item *T) error {
	if item == nil {
		return fmt.Errorf("gormz: nil item")
	}
	return q.db.WithContext(q.ctx).Create(item).Error
}

// CreateMany ينشئ عدة سجلات.
func (q *QuerySet[T]) CreateMany(items []T) error {
	if len(items) == 0 {
		return nil
	}
	return q.db.WithContext(q.ctx).Create(&items).Error
}

// CreateInBatches ينشئ عدة سجلات بدفعات.
//
//	users := make([]User, 100000)
//	err := gormz.New[User]().CreateInBatches(users, 1000)
func (q *QuerySet[T]) CreateInBatches(items []T, batchSize int) error {
	if len(items) == 0 {
		return nil
	}
	if batchSize <= 0 {
		batchSize = 1000
	}
	return q.db.WithContext(q.ctx).CreateInBatches(&items, batchSize).Error
}

// Save يحفظ سجلًا (create أو update).
func (q *QuerySet[T]) Save(item *T) error {
	if item == nil {
		return fmt.Errorf("gormz: nil item")
	}
	return q.db.WithContext(q.ctx).Save(item).Error
}

// Update يحدّث سجلًا بالـ ID.
func (q *QuerySet[T]) Update(id any, field string, value any) error {
	if err := internal.ValidateField(field); err != nil {
		return err
	}

	var zero T
	return q.db.WithContext(q.ctx).
		Model(&zero).
		Where("id = ?", id).
		Update(field, value).Error
}

// UpdateMany يحدّث عدة سجلات.
//
// ⚠️ يتطلب conditions.
func (q *QuerySet[T]) UpdateMany(values map[string]any) (int64, error) {
	if len(q.conditions) == 0 {
		return 0, NewDangerousError("UpdateMany", "requires at least one condition")
	}
	if len(values) == 0 {
		return 0, nil
	}

	for f := range values {
		if err := internal.ValidateField(f); err != nil {
			return 0, err
		}
	}

	var zero T
	res := q.build().Model(&zero).Updates(values)
	return res.RowsAffected, res.Error
}

// UpdateColumn يحدّث عمودًا بدون hooks.
func (q *QuerySet[T]) UpdateColumn(id any, field string, value any) error {
	if err := internal.ValidateField(field); err != nil {
		return err
	}

	var zero T
	return q.db.WithContext(q.ctx).
		Model(&zero).
		Where("id = ?", id).
		UpdateColumn(field, value).Error
}

// Delete يحذف سجلًا بالـ ID (soft delete).
func (q *QuerySet[T]) Delete(id any) error {
	var zero T
	return q.db.WithContext(q.ctx).Delete(&zero, id).Error
}

// DeleteMany يحذف عدة سجلات.
//
// ⚠️ يتطلب conditions.
// DeleteMany يحذف عدة سجلات.
//
// ⚠️ يتطلب conditions.
func (q *QuerySet[T]) DeleteMany() (int64, error) {
	if len(q.conditions) == 0 {
		return 0, NewDangerousError("DeleteMany", "requires at least one condition")
	}

	var zero T
	conn := q.db.WithContext(q.ctx).Model(&zero)

	// Conditions
	for _, c := range q.conditions {
		if c == nil {
			continue
		}
		switch v := c.(type) {
		case WhereClause:
			conn = conn.Where(v.sql, v.args...)

		case NotClause:
			sql, args := v.ToSQL()
			conn = conn.Where(sql, args...)

		case *Q:
			sql, args := v.toSQL()
			conn = conn.Where(sql, args...)

		default:
			// أي Clause عام (مثل internal.RawClause)
			sql, args := c.ToSQL()
			if sql != "" {
				conn = conn.Where(sql, args...)
			}
		}
	}

	res := conn.Delete(&zero)
	return res.RowsAffected, res.Error
}

// HardDelete يحذف نهائيًا (يتجاوز soft delete).
func (q *QuerySet[T]) HardDelete(id any) error {
	var zero T
	return q.db.WithContext(q.ctx).Unscoped().Delete(&zero, id).Error
}

// Restore يستعيد سجلًا محذوفًا.
func (q *QuerySet[T]) Restore(id any) error {
	var zero T
	return q.db.WithContext(q.ctx).
		Unscoped().
		Model(&zero).
		Where("id = ?", id).
		UpdateColumn("deleted_at", nil).Error
}

// RestoreAll يستعيد كل السجلات المحذوفة المطابقة.
//
// ⚠️ يتطلب conditions.
func (q *QuerySet[T]) RestoreAll() (int64, error) {
	if len(q.conditions) == 0 {
		return 0, NewDangerousError("RestoreAll", "requires at least one condition")
	}

	var zero T
	res := q.db.WithContext(q.ctx).
		Unscoped().
		Model(&zero).
		Where("deleted_at IS NOT NULL").
		UpdateColumn("deleted_at", nil)
	return res.RowsAffected, res.Error
}

// ═══════════════════════════════════════════════
// Aggregates — التجميعات
// ═══════════════════════════════════════════════

// Sum يجمع عمودًا رقميًا.
func (q *QuerySet[T]) Sum(field string) (float64, error) {
	if err := internal.ValidateField(field); err != nil {
		return 0, err
	}
	return q.scalar("COALESCE(SUM(" + field + "), 0)")
}

// Avg يحسب المتوسط.
func (q *QuerySet[T]) Avg(field string) (float64, error) {
	if err := internal.ValidateField(field); err != nil {
		return 0, err
	}
	return q.scalar("COALESCE(AVG(" + field + "), 0)")
}

// Min يجد الحد الأدنى.
func (q *QuerySet[T]) Min(field string) (float64, error) {
	if err := internal.ValidateField(field); err != nil {
		return 0, err
	}
	return q.scalar("COALESCE(MIN(" + field + "), 0)")
}

// Max يجد الحد الأقصى.
func (q *QuerySet[T]) Max(field string) (float64, error) {
	if err := internal.ValidateField(field); err != nil {
		return 0, err
	}
	return q.scalar("COALESCE(MAX(" + field + "), 0)")
}

func (q *QuerySet[T]) scalar(expr string) (float64, error) {
	var zero T
	var out struct{ V float64 }
	err := q.build().Model(&zero).Select(expr + " as v").Scan(&out).Error
	return out.V, err
}

// ═══════════════════════════════════════════════
// Pagination — التصفح
// ═══════════════════════════════════════════════

// Paginate يرجّع صفحة من النتائج.
func (q *QuerySet[T]) Paginate(page, perPage int) (*PaginatedResult[T], error) {
	p := NewPage(page, perPage)

	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	items, err := q.clone().
		Limit(p.PerPage).
		Offset(p.Offset()).
		All()
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(p.PerPage) - 1) / int64(p.PerPage))

	return &PaginatedResult[T]{
		Items:      items,
		Total:      total,
		Page:       p.Number,
		PerPage:    p.PerPage,
		TotalPages: totalPages,
		HasNext:    p.Number < totalPages,
		HasPrev:    p.Number > 1,
	}, nil
}

// ═══════════════════════════════════════════════
// Debugging — للتصحيح
// ═══════════════════════════════════════════════

// ToSQL يرجّع SQL والـ args (للتصحيح).
func (q *QuerySet[T]) ToSQL() (string, []any) {
	var zero T
	var empty []T

	stmt := q.build().
		Model(&zero).
		Session(&gorm.Session{DryRun: true}).
		Find(&empty)

	return stmt.Statement.SQL.String(), stmt.Statement.Vars
}

// String يرجّع SQL كنص (للتصحيح).
func (q *QuerySet[T]) String() string {
	sql, _ := q.ToSQL()
	return sql
}

// DryRun يرجّع *gorm.DB في وضع DryRun.
func (q *QuerySet[T]) DryRun() *gorm.DB {
	return q.build().Session(&gorm.Session{DryRun: true})
}

// ═══════════════════════════════════════════════
// Introspection — للاستخدام المتقدم
// ═══════════════════════════════════════════════

// Conditions يرجّع نسخة من الشروط.
func (q *QuerySet[T]) Conditions() []Clause {
	out := make([]Clause, len(q.conditions))
	copy(out, q.conditions)
	return out
}

// AddCondition يضيف شرطًا داخليًا.
func (q *QuerySet[T]) AddCondition(c Clause) *QuerySet[T] {
	nq := q.clone()
	nq.conditions = append(nq.conditions, c)
	return nq
}

// Orders يرجّع نسخة من الترتيبات.
func (q *QuerySet[T]) Orders() []string {
	out := make([]string, len(q.orders))
	copy(out, q.orders)
	return out
}

// Selects يرجّع نسخة من الأعمدة المحددة.
func (q *QuerySet[T]) Selects() []string {
	out := make([]string, len(q.selects))
	copy(out, q.selects)
	return out
}

// Preloads يرجّع نسخة من العلاقات.
func (q *QuerySet[T]) Preloads() []string {
	out := make([]string, len(q.preloads))
	copy(out, q.preloads)
	return out
}

// LimitValue يرجّع قيمة Limit الحالية.
func (q *QuerySet[T]) LimitValue() int {
	return q.limit
}

// OffsetValue يرجّع قيمة Offset الحالية.
func (q *QuerySet[T]) OffsetValue() int {
	return q.offset
}

// IsUnscoped يفحص إذا كان unscoped.
func (q *QuerySet[T]) IsUnscoped() bool {
	return q.unscoped
}

// IsOnlyDeleted يفحص إذا كان onlyDeleted.
func (q *QuerySet[T]) IsOnlyDeleted() bool {
	return q.onlyDeleted
}

// IsDistinct يفحص إذا كان DISTINCT مفعّلًا.
func (q *QuerySet[T]) IsDistinct() bool {
	return q.distinct
}

// ═══════════════════════════════════════════════
// Advanced — للاستخدام الداخلي/المتقدم
// ═══════════════════════════════════════════════

// DB يرجّع *gorm.DB الأساسي.
func (q *QuerySet[T]) DB() *gorm.DB {
	return q.db
}

// Build يرجّع *gorm.DB مع التطبيق.
func (q *QuerySet[T]) Build() *gorm.DB {
	return q.build()
}

// Clone ينسخ QuerySet.
func (q *QuerySet[T]) Clone() *QuerySet[T] {
	return q.clone()
}

// ═══════════════════════════════════════════════
// Internals
// ═══════════════════════════════════════════════

// rawHaving يمثل HAVING clause داخلي.
type rawHaving struct {
	sql  string
	args []any
}

// joinClause يمثل JOIN clause داخلي.
type joinClause struct {
	joinType string
	table    string
	on       string
}

// clone ينسخ QuerySet (immutable).
func (q *QuerySet[T]) clone() *QuerySet[T] {
	nq := *q
	nq.conditions = append([]Clause{}, q.conditions...)
	nq.orders = append([]string{}, q.orders...)
	nq.selects = append([]string{}, q.selects...)
	nq.omits = append([]string{}, q.omits...)
	nq.preloads = append([]string{}, q.preloads...)
	nq.groupBy = append([]string{}, q.groupBy...)
	nq.having = append([]rawHaving{}, q.having...)
	nq.joins = append([]joinClause{}, q.joins...)
	nq.windows = append([]string{}, q.windows...)
	nq.distinctOn = append([]string{}, q.distinctOn...)
	return &nq
}

// build يبني *gorm.DB النهائي.
func (q *QuerySet[T]) build() *gorm.DB {
	var zero T
	conn := q.db.WithContext(q.ctx).Model(&zero)

	if q.unscoped {
		conn = conn.Unscoped()
	}
	if q.onlyDeleted {
		conn = conn.Where("deleted_at IS NOT NULL")
	}

	// Joins (قبل Where)
	for _, j := range q.joins {
		if j.joinType == "CROSS" || j.on == "" {
			conn = conn.Joins(j.joinType + " JOIN " + j.table)
		} else {
			conn = conn.Joins(j.joinType + " JOIN " + j.table + " ON " + j.on)
		}
	}

	// Conditions
	// Conditions
	for _, c := range q.conditions {
		if c == nil {
			continue
		}
		switch v := c.(type) {
		case WhereClause:
			conn = conn.Where(v.sql, v.args...)

		case NotClause:
			sql, args := v.ToSQL()
			conn = conn.Where(sql, args...)

		case *Q:
			sql, args := v.toSQL()
			conn = conn.Where(sql, args...)

		default:
			// أي Clause عام (مثل internal.RawClause)
			sql, args := c.ToSQL()
			if sql != "" {
				conn = conn.Where(sql, args...)
			}
		}
	}

	// Distinct
	if q.distinct {
		if len(q.distinctOn) > 0 {
			conn = conn.Distinct(q.distinctOn)
		} else {
			conn = conn.Distinct()
		}
	}

	// Select (مع windows)
	selects := make([]string, 0, len(q.selects)+len(q.windows))
	selects = append(selects, q.selects...)
	selects = append(selects, q.windows...)

	if len(selects) > 0 {
		conn = conn.Select(strings.Join(selects, ", "))
	}

	if len(q.omits) > 0 {
		conn = conn.Omit(q.omits...)
	}

	// Group By
	if len(q.groupBy) > 0 {
		conn = conn.Group(strings.Join(q.groupBy, ", "))
	}

	// Having
	for _, h := range q.having {
		conn = conn.Having(h.sql, h.args...)
	}

	// Preloads
	for _, p := range q.preloads {
		conn = conn.Preload(p)
	}

	// Orders
	for _, o := range q.orders {
		conn = conn.Order(o)
	}

	// Limit/Offset
	if q.limit > 0 {
		conn = conn.Limit(q.limit)
	}
	if q.offset > 0 {
		conn = conn.Offset(q.offset)
	}

	return conn
}
