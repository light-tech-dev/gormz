package advanced

import (
	"github.com/light-tech-dev/gormz"
	"gorm.io/gorm"
)

// ═══════════════════════════════════════════════
// Joins — الانضمامات (Advanced)
// ═══════════════════════════════════════════════

// JoinType نوع الانضمام.
type JoinType string

const (
	JoinInner JoinType = "INNER"
	JoinLeft  JoinType = "LEFT"
	JoinRight JoinType = "RIGHT"
	JoinFull  JoinType = "FULL"
	JoinCross JoinType = "CROSS"
)

// ═══════════════════════════════════════════════
// Join
// ═══════════════════════════════════════════════

// Join يمثل join واحد.
type Join struct {
	Type  JoinType
	Table string
	On    string
}

// JoinSQL يبني SQL للـ join.
//
//	advanced.JoinSQL(advanced.JoinInner, "orders", "orders.user_id = users.id")
//	→ "INNER JOIN orders ON orders.user_id = users.id"
func JoinSQL(jt JoinType, table, on string) string {
	if jt == JoinCross || on == "" {
		return string(jt) + " JOIN " + table
	}
	return string(jt) + " JOIN " + table + " ON " + on
}

// ═══════════════════════════════════════════════
// JoinQuery
// ═══════════════════════════════════════════════

// JoinQuery يمثل استعلامًا مع joins.
type JoinQuery[T any] struct {
	q     *gormz.QuerySet[T]
	joins []Join
}

// WithJoins ينشئ JoinQuery.
//
// ⚠️ deprecated — استخدم QuerySet methods مباشرة.
func WithJoins[T any](q *gormz.QuerySet[T]) *JoinQuery[T] {
	return &JoinQuery[T]{q: q}
}

// Inner يضيف INNER JOIN.
func (jq *JoinQuery[T]) Inner(table, on string) *JoinQuery[T] {
	jq.joins = append(jq.joins, Join{Type: JoinInner, Table: table, On: on})
	return jq
}

// Left يضيف LEFT JOIN.
func (jq *JoinQuery[T]) Left(table, on string) *JoinQuery[T] {
	jq.joins = append(jq.joins, Join{Type: JoinLeft, Table: table, On: on})
	return jq
}

// Right يضيف RIGHT JOIN.
func (jq *JoinQuery[T]) Right(table, on string) *JoinQuery[T] {
	jq.joins = append(jq.joins, Join{Type: JoinRight, Table: table, On: on})
	return jq
}

// Cross يضيف CROSS JOIN.
func (jq *JoinQuery[T]) Cross(table string) *JoinQuery[T] {
	jq.joins = append(jq.joins, Join{Type: JoinCross, Table: table})
	return jq
}

// Filter يضيف فلتر.
func (jq *JoinQuery[T]) Filter(field string, value any) *JoinQuery[T] {
	jq.q = jq.q.Filter(field, value)
	return jq
}

// Where يضيف SQL خام.
func (jq *JoinQuery[T]) Where(sql string, args ...any) *JoinQuery[T] {
	jq.q = jq.q.Where(sql, args...)
	return jq
}

// Select يحدد الأعمدة.
func (jq *JoinQuery[T]) Select(fields ...string) *JoinQuery[T] {
	jq.q = jq.q.Select(fields...)
	return jq
}

// OrderBy يضيف ترتيبًا.
func (jq *JoinQuery[T]) OrderBy(fields ...string) *JoinQuery[T] {
	jq.q = jq.q.OrderBy(fields...)
	return jq
}

// Limit يحدد العدد.
func (jq *JoinQuery[T]) Limit(n int) *JoinQuery[T] {
	jq.q = jq.q.Limit(n)
	return jq
}

// Offset يحدد البداية.
func (jq *JoinQuery[T]) Offset(n int) *JoinQuery[T] {
	jq.q = jq.q.Offset(n)
	return jq
}

// ═══════════════════════════════════════════════
// Terminal
// ═══════════════════════════════════════════════

// All ينفّذ الاستعلام.
func (jq *JoinQuery[T]) All() ([]T, error) {
	var results []T
	err := jq.build().Find(&results).Error
	return results, err
}

// ScanInto يقرأ في struct مخصص.
func (jq *JoinQuery[T]) ScanInto(dest any) error {
	if dest == nil {
		return gormz.NewValidationError("dest", "nil destination")
	}
	return jq.build().Scan(dest).Error
}

// Count يرجّع العدد.
func (jq *JoinQuery[T]) Count() (int64, error) {
	var zero T
	var count int64
	err := jq.build().Model(&zero).Count(&count).Error
	return count, err
}

// ═══════════════════════════════════════════════
// Internals
// ═══════════════════════════════════════════════

func (jq *JoinQuery[T]) build() *gorm.DB {
	var zero T
	q := jq.q
	for _, j := range jq.joins {
		q = q.Join(string(j.Type), j.Table, j.On)
	}
	conn := q.Build()
	return conn.Model(&zero)
}