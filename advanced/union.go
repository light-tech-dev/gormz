package advanced

import (
	"strconv"
	"strings"

	"github.com/light-tech-dev/gormz"
)

// ═══════════════════════════════════════════════
// Union — دمج نتائج
// ═══════════════════════════════════════════════

// UnionType نوع Union.
type UnionType string

const (
	UnionDistinct UnionType = "UNION"
	UnionAllType  UnionType = "UNION ALL"
)

// ═══════════════════════════════════════════════
// UnionQuery
// ═══════════════════════════════════════════════

// UnionQuery يمثل استعلام Union.
type UnionQuery[T any] struct {
	queries []*gormz.QuerySet[T]
	utype   UnionType
	orders  []string
	limit   int
	offset  int
}

// Union ينشئ Union (يحذف المكرر).
func Union[T any](queries ...*gormz.QuerySet[T]) *UnionQuery[T] {
	return &UnionQuery[T]{
		queries: queries,
		utype:   UnionDistinct,
	}
}

// UnionAll ينشئ Union All (يحتفظ بالمكرر).
func UnionAll[T any](queries ...*gormz.QuerySet[T]) *UnionQuery[T] {
	return &UnionQuery[T]{
		queries: queries,
		utype:   UnionAllType,
	}
}

// Add يضيف استعلامًا.
func (uq *UnionQuery[T]) Add(q *gormz.QuerySet[T]) *UnionQuery[T] {
	uq.queries = append(uq.queries, q)
	return uq
}

// OrderBy يضيف ترتيبًا للنتيجة النهائية.
func (uq *UnionQuery[T]) OrderBy(fields ...string) *UnionQuery[T] {
	uq.orders = append(uq.orders, fields...)
	return uq
}

// Limit يحدد العدد.
func (uq *UnionQuery[T]) Limit(n int) *UnionQuery[T] {
	uq.limit = n
	return uq
}

// Offset يحدد البداية.
func (uq *UnionQuery[T]) Offset(n int) *UnionQuery[T] {
	uq.offset = n
	return uq
}

// All ينفّذ Union.
func (uq *UnionQuery[T]) All() ([]T, error) {
	if len(uq.queries) == 0 {
		return nil, nil
	}

	var zero T
	db := gormz.DB().Model(&zero)

	var parts []string
	var allArgs []any

	for i, q := range uq.queries {
		if q == nil {
			continue
		}
		sql, args := q.ToSQL()
		if i == 0 {
			parts = append(parts, sql)
		} else {
			parts = append(parts, string(uq.utype)+" "+sql)
		}
		allArgs = append(allArgs, args...)
	}

	if len(parts) == 0 {
		return nil, nil
	}

	finalSQL := strings.Join(parts, " ")

	if len(uq.orders) > 0 || uq.limit > 0 || uq.offset > 0 {
		finalSQL = "SELECT * FROM (" + finalSQL + ") AS union_result"
		if len(uq.orders) > 0 {
			finalSQL += " ORDER BY " + strings.Join(uq.orders, ", ")
		}
		if uq.limit > 0 {
			finalSQL += " LIMIT " + strconv.Itoa(uq.limit)
		}
		if uq.offset > 0 {
			finalSQL += " OFFSET " + strconv.Itoa(uq.offset)
		}
	}

	var results []T
	err := db.Raw(finalSQL, allArgs...).Scan(&results).Error
	return results, err
}

// Count يرجّع العدد.
func (uq *UnionQuery[T]) Count() (int64, error) {
	items, err := uq.All()
	return int64(len(items)), err
}