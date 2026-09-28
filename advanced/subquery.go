package advanced

import (
	"fmt"

	"github.com/light-tech-dev/gormz"
	"github.com/light-tech-dev/gormz/internal"
)

// ═══════════════════════════════════════════════
// SubQuery — الاستعلامات الفرعية
// ═══════════════════════════════════════════════

// SubQuery يمثل استعلامًا فرعيًا.
type SubQuery struct {
	SQL  string
	Args []any
}

// ═══════════════════════════════════════════════
// Builders
// ═══════════════════════════════════════════════

// SubFrom ينشئ SubQuery من QuerySet.
//
//	subq := advanced.SubFrom[Order](
//	    gormz.New[Order]().
//	        Select("user_id").
//	        Filter("status", "paid"),
//	    "user_id",
//	)
func SubFrom[T any](q *gormz.QuerySet[T], field string) *SubQuery {
	if q == nil {
		return &SubQuery{}
	}

	sql, args := q.ToSQL()
	return &SubQuery{
		SQL:  sql,
		Args: args,
	}
}

// SubFromColumns ينشئ SubQuery مع تحديد الأعمدة.
func SubFromColumns[T any](q *gormz.QuerySet[T], columns ...string) *SubQuery {
	if q == nil {
		return &SubQuery{}
	}

	cloned := q.Clone().Select(columns...)
	sql, args := cloned.ToSQL()
	return &SubQuery{
		SQL:  sql,
		Args: args,
	}
}

// SubRaw ينشئ SubQuery من SQL خام.
//
//	subq := advanced.SubRaw("SELECT id FROM users WHERE age > ?", 18)
func SubRaw(sql string, args ...any) *SubQuery {
	return &SubQuery{SQL: sql, Args: args}
}

// ═══════════════════════════════════════════════
// Operators
// ═══════════════════════════════════════════════

// In يستخدم SubQuery في IN.
func In(field string, sub *SubQuery) internal.RawClause {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	if sub == nil || sub.SQL == "" {
		return internal.RawClause{SQL: "1=0"}
	}
	return internal.RawClause{
		SQL:  field + " IN (" + sub.SQL + ")",
		Args: sub.Args,
	}
}

// NotIn يستخدم SubQuery في NOT IN.
func NotIn(field string, sub *SubQuery) internal.RawClause {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	if sub == nil || sub.SQL == "" {
		return internal.RawClause{SQL: "1=1"}
	}
	return internal.RawClause{
		SQL:  field + " NOT IN (" + sub.SQL + ")",
		Args: sub.Args,
	}
}

// Exists يفحص وجود.
func Exists(sub *SubQuery) internal.RawClause {
	if sub == nil || sub.SQL == "" {
		return internal.RawClause{SQL: "1=0"}
	}
	return internal.RawClause{
		SQL:  "EXISTS (" + sub.SQL + ")",
		Args: sub.Args,
	}
}

// NotExists يفحص عدم الوجود.
func NotExists(sub *SubQuery) internal.RawClause {
	if sub == nil || sub.SQL == "" {
		return internal.RawClause{SQL: "1=1"}
	}
	return internal.RawClause{
		SQL:  "NOT EXISTS (" + sub.SQL + ")",
		Args: sub.Args,
	}
}

// GtSub يفحص field > (subquery).
func GtSub(field string, sub *SubQuery) internal.RawClause {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	return internal.RawClause{
		SQL:  field + " > (" + sub.SQL + ")",
		Args: sub.Args,
	}
}

// LtSub يفحص field < (subquery).
func LtSub(field string, sub *SubQuery) internal.RawClause {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	return internal.RawClause{
		SQL:  field + " < (" + sub.SQL + ")",
		Args: sub.Args,
	}
}

// EqSub يفحص field = (subquery).
func EqSub(field string, sub *SubQuery) internal.RawClause {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	return internal.RawClause{
		SQL:  field + " = (" + sub.SQL + ")",
		Args: sub.Args,
	}
}

// GteSub يفحص field >= (subquery).
func GteSub(field string, sub *SubQuery) internal.RawClause {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	return internal.RawClause{
		SQL:  field + " >= (" + sub.SQL + ")",
		Args: sub.Args,
	}
}

// LteSub يفحص field <= (subquery).
func LteSub(field string, sub *SubQuery) internal.RawClause {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	return internal.RawClause{
		SQL:  field + " <= (" + sub.SQL + ")",
		Args: sub.Args,
	}
}

// ═══════════════════════════════════════════════
// String / Helpers
// ═══════════════════════════════════════════════

// String للتصحيح.
func (sq *SubQuery) String() string {
	return fmt.Sprintf("SubQuery{SQL: %s, Args: %v}", sq.SQL, sq.Args)
}

// IsEmpty يفحص إذا كان فارغًا.
func (sq *SubQuery) IsEmpty() bool {
	return sq == nil || sq.SQL == ""
}

// ═══════════════════════════════════════════════
// Correlated SubQuery
// ═══════════════════════════════════════════════

// CorrelatedSubQuery يمثل subquery مترابطًا.
type CorrelatedSubQuery struct {
	SQL  string
	Args []any
}

// NewCorrelated ينشئ correlated subquery.
func NewCorrelated(sql string, args ...any) *CorrelatedSubQuery {
	return &CorrelatedSubQuery{SQL: sql, Args: args}
}

// GtCorr يفحص field > (correlated subquery).
func GtCorr(field string, corr *CorrelatedSubQuery) internal.RawClause {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	return internal.RawClause{
		SQL:  field + " > (" + corr.SQL + ")",
		Args: corr.Args,
	}
}

// LtCorr يفحص field < (correlated subquery).
func LtCorr(field string, corr *CorrelatedSubQuery) internal.RawClause {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	return internal.RawClause{
		SQL:  field + " < (" + corr.SQL + ")",
		Args: corr.Args,
	}
}

// EqCorr يفحص field = (correlated subquery).
func EqCorr(field string, corr *CorrelatedSubQuery) internal.RawClause {
	if err := internal.ValidateField(field); err != nil {
		panic(err)
	}
	return internal.RawClause{
		SQL:  field + " = (" + corr.SQL + ")",
		Args: corr.Args,
	}
}

// String للتصحيح.
func (cs *CorrelatedSubQuery) String() string {
	return fmt.Sprintf("CorrelatedSubQuery{SQL: %s, Args: %v}", cs.SQL, cs.Args)
}