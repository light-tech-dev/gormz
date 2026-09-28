package advanced

import (
	"fmt"
	"strings"

	"github.com/light-tech-dev/gormz"
	"github.com/light-tech-dev/gormz/internal"
	"gorm.io/gorm"
)

// ═══════════════════════════════════════════════
// Aggregates — GroupBy, Having, Distinct
// ═══════════════════════════════════════════════

// AggregateQuery يمثل استعلامًا مع تجميعات.
type AggregateQuery[T any] struct {
	db         *gorm.DB
	groupBy    []string
	having     []HavingClause
	selects    []string
	distinct   bool
	conditions []internal.RawClause
	orders     []string
	limit      int
	offset     int
}

// HavingClause يمثل HAVING.
type HavingClause struct {
	SQL  string
	Args []any
}

// ═══════════════════════════════════════════════
// Entry Point
// ═══════════════════════════════════════════════

// GroupBy ينشئ استعلامًا مع GroupBy.
func GroupBy[T any](fields ...string) *AggregateQuery[T] {
	for _, f := range fields {
		if err := internal.ValidateField(f); err != nil {
			panic(err)
		}
	}
	return &AggregateQuery[T]{
		db:      gormz.DB(),
		groupBy: fields,
	}
}

// GroupByWithInstance ينشئ استعلامًا على Instance معين.
func GroupByWithInstance[T any](i *gormz.Instance, fields ...string) *AggregateQuery[T] {
	if i == nil {
		panic(gormz.ErrNilDB)
	}
	for _, f := range fields {
		if err := internal.ValidateField(f); err != nil {
			panic(err)
		}
	}
	return &AggregateQuery[T]{
		db:      i.DB(),
		groupBy: fields,
	}
}

// Distinct يفعّل DISTINCT.
func (aq *AggregateQuery[T]) Distinct() *AggregateQuery[T] {
	aq.distinct = true
	return aq
}

// ═══════════════════════════════════════════════
// Aggregations
// ═══════════════════════════════════════════════

// Count يضيف COUNT.
func (aq *AggregateQuery[T]) Count(field, alias string) *AggregateQuery[T] {
	if field == "" {
		field = "*"
	}
	if field != "*" {
		if err := internal.ValidateField(field); err != nil {
			panic(err)
		}
	}
	if alias == "" {
		alias = "count"
	}
	if err := internal.ValidateField(alias); err != nil {
		panic(err)
	}
	aq.selects = append(aq.selects, fmt.Sprintf("COUNT(%s) AS %s", field, alias))
	return aq
}

// CountDistinct يضيف COUNT(DISTINCT).
func (aq *AggregateQuery[T]) CountDistinct(field, alias string) *AggregateQuery[T] {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	if alias == "" {
		alias = "count_distinct"
	}
	if err := internal.ValidateField(alias); err != nil {
		panic(err)
	}
	aq.selects = append(aq.selects, fmt.Sprintf("COUNT(DISTINCT %s) AS %s", field, alias))
	return aq
}

// Sum يضيف SUM.
func (aq *AggregateQuery[T]) Sum(field, alias string) *AggregateQuery[T] {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	if alias == "" {
		alias = "sum"
	}
	if err := internal.ValidateField(alias); err != nil {
		panic(err)
	}
	aq.selects = append(aq.selects, fmt.Sprintf("COALESCE(SUM(%s), 0) AS %s", field, alias))
	return aq
}

// Avg يضيف AVG.
func (aq *AggregateQuery[T]) Avg(field, alias string) *AggregateQuery[T] {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	if alias == "" {
		alias = "avg"
	}
	if err := internal.ValidateField(alias); err != nil {
		panic(err)
	}
	aq.selects = append(aq.selects, fmt.Sprintf("COALESCE(AVG(%s), 0) AS %s", field, alias))
	return aq
}

// Min يضيف MIN.
func (aq *AggregateQuery[T]) Min(field, alias string) *AggregateQuery[T] {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	if alias == "" {
		alias = "min"
	}
	if err := internal.ValidateField(alias); err != nil {
		panic(err)
	}
	aq.selects = append(aq.selects, fmt.Sprintf("MIN(%s) AS %s", field, alias))
	return aq
}

// Max يضيف MAX.
func (aq *AggregateQuery[T]) Max(field, alias string) *AggregateQuery[T] {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	if alias == "" {
		alias = "max"
	}
	if err := internal.ValidateField(alias); err != nil {
		panic(err)
	}
	aq.selects = append(aq.selects, fmt.Sprintf("MAX(%s) AS %s", field, alias))
	return aq
}

// Select يضيف عمودًا.
//
// ⚠️ المسؤولية على المستخدم.
func (aq *AggregateQuery[T]) Select(expr string) *AggregateQuery[T] {
	aq.selects = append(aq.selects, expr)
	return aq
}

// SelectAs يضيف عمودًا مع alias.
func (aq *AggregateQuery[T]) SelectAs(field, alias string) *AggregateQuery[T] {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	if alias == "" {
		aq.selects = append(aq.selects, field)
	} else {
		if err := internal.ValidateField(alias); err != nil {
			panic(err)
		}
		aq.selects = append(aq.selects, field+" AS "+alias)
	}
	return aq
}

// ═══════════════════════════════════════════════
// Filtering
// ═══════════════════════════════════════════════

// Where يضيف شرط WHERE بـ SQL خام.
func (aq *AggregateQuery[T]) Where(sql string, args ...any) *AggregateQuery[T] {
	aq.conditions = append(aq.conditions, internal.RawClause{SQL: sql, Args: args})
	return aq
}

// Filter يضيف فلتر بـ lookup syntax.
func (aq *AggregateQuery[T]) Filter(field string, value any) *AggregateQuery[T] {
	nq, err := aq.TryFilter(field, value)
	if err != nil {
		panic(err)
	}
	return nq
}

// TryFilter مثل Filter لكن يرجّع خطأ.
func (aq *AggregateQuery[T]) TryFilter(field string, value any) (*AggregateQuery[T], error) {
	clause, err := internal.BuildLookup(field, value, false)
	if err != nil {
		return nil, err
	}
	aq.conditions = append(aq.conditions, internal.RawClause{
		SQL:  clause.SQL,
		Args: clause.Args,
	})
	return aq, nil
}

// Exclude يستثني شرطًا.
func (aq *AggregateQuery[T]) Exclude(field string, value any) *AggregateQuery[T] {
	nq, err := aq.TryExclude(field, value)
	if err != nil {
		panic(err)
	}
	return nq
}

// TryExclude مثل Exclude لكن يرجّع خطأ.
func (aq *AggregateQuery[T]) TryExclude(field string, value any) (*AggregateQuery[T], error) {
	clause, err := internal.BuildLookup(field, value, true)
	if err != nil {
		return nil, err
	}
	aq.conditions = append(aq.conditions, internal.RawClause{
		SQL:  clause.SQL,
		Args: clause.Args,
	})
	return aq, nil
}

// Having يضيف HAVING.
func (aq *AggregateQuery[T]) Having(sql string, args ...any) *AggregateQuery[T] {
	aq.having = append(aq.having, HavingClause{SQL: sql, Args: args})
	return aq
}

// ═══════════════════════════════════════════════
// Ordering / Pagination
// ═══════════════════════════════════════════════

// OrderBy يضيف ترتيبًا.
func (aq *AggregateQuery[T]) OrderBy(fields ...string) *AggregateQuery[T] {
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if strings.HasPrefix(f, "-") {
			col := strings.TrimPrefix(f, "-")
			aq.orders = append(aq.orders, col+" DESC")
		} else {
			aq.orders = append(aq.orders, f+" ASC")
		}
	}
	return aq
}

// Limit يحدد العدد.
func (aq *AggregateQuery[T]) Limit(n int) *AggregateQuery[T] {
	aq.limit = n
	return aq
}

// Offset يحدد البداية.
func (aq *AggregateQuery[T]) Offset(n int) *AggregateQuery[T] {
	aq.offset = n
	return aq
}

// ═══════════════════════════════════════════════
// Terminal
// ═══════════════════════════════════════════════

// All ينفّذ الاستعلام ويرجّع النتائج كـ maps.
func (aq *AggregateQuery[T]) All() ([]map[string]any, error) {
	var results []map[string]any
	err := aq.build().Find(&results).Error
	return results, err
}

// ScanInto يقرأ في struct مخصص.
func (aq *AggregateQuery[T]) ScanInto(dest any) error {
	if dest == nil {
		return fmt.Errorf("gormz/advanced: nil destination")
	}
	return aq.build().Find(dest).Error
}

// CountGroups يرجّع عدد المجموعات.
func (aq *AggregateQuery[T]) CountGroups() (int64, error) {
	var zero T

	if len(aq.groupBy) > 0 {
		var count int64
		subq := aq.build().Select(strings.Join(aq.groupBy, ", "))
		err := gormz.DB().Model(&zero).
			Table("(?) as sub", subq).
			Count(&count).Error
		return count, err
	}

	var count int64
	err := aq.build().Model(&zero).Count(&count).Error
	return count, err
}

// ═══════════════════════════════════════════════
// Internals
// ═══════════════════════════════════════════════

func (aq *AggregateQuery[T]) build() *gorm.DB {
	var zero T
	db := aq.db.Model(&zero)

	selects := make([]string, 0, len(aq.selects)+len(aq.groupBy))
	selects = append(selects, aq.selects...)

	for _, g := range aq.groupBy {
		if !containsString(selects, g) {
			selects = append([]string{g}, selects...)
		}
	}

	if len(selects) > 0 {
		db = db.Select(strings.Join(selects, ", "))
	}

	if aq.distinct {
		db = db.Distinct()
	}

	for _, c := range aq.conditions {
		db = db.Where(c.SQL, c.Args...)
	}

	if len(aq.groupBy) > 0 {
		db = db.Group(strings.Join(aq.groupBy, ", "))
	}

	for _, h := range aq.having {
		db = db.Having(h.SQL, h.Args...)
	}

	for _, o := range aq.orders {
		db = db.Order(o)
	}

	if aq.limit > 0 {
		db = db.Limit(aq.limit)
	}
	if aq.offset > 0 {
		db = db.Offset(aq.offset)
	}

	return db
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// ═══════════════════════════════════════════════
// Convenience Functions
// ═══════════════════════════════════════════════

// Distinct يرجّع قيمًا فريدة لعمود.
func Distinct[T any](field string) ([]any, error) {
	if err := internal.ValidateField(field); err != nil {
		return nil, err
	}

	var zero T
	var results []any

	err := gormz.DB().Model(&zero).
		Distinct(field).
		Pluck(field, &results).Error

	return results, err
}

// CountDistinctValues يحسب عدد القيم الفريدة.
func CountDistinctValues[T any](field string) (int64, error) {
	if err := internal.ValidateField(field); err != nil {
		return 0, err
	}

	var zero T
	var count int64

	err := gormz.DB().Model(&zero).
		Select("COUNT(DISTINCT " + field + ")").
		Scan(&count).Error

	return count, err
}

// GroupConcat يجمع قيم عمود في string.
//
// ⚠️ SQLite/MySQL فقط. PostgreSQL يستخدم STRING_AGG.
func GroupConcat[T any](field, separator string) (string, error) {
	if err := internal.ValidateField(field); err != nil {
		return "", err
	}

	var zero T
	var result *string

	expr := fmt.Sprintf("GROUP_CONCAT(%s, ?)", field)
	err := gormz.DB().Model(&zero).
		Select(expr, separator).
		Scan(&result).Error

	if err != nil {
		return "", err
	}
	if result == nil {
		return "", nil
	}
	return *result, nil
}