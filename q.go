package gormz

import (
	"strings"

	"github.com/light-tech-dev/gormz/internal"
)

// ═══════════════════════════════════════════════
// Clause — Interface لكل الشروط
// ═══════════════════════════════════════════════

// Clause هو أي شرط يمكن تحويله إلى SQL.
//
// كل من WhereClause, NotClause, RawClause, *Q يطبّق هذه الواجهة.
type Clause interface {
	ToSQL() (string, []any)
	IsNegated() bool
}

// ═══════════════════════════════════════════════
// WhereClause — شرط WHERE
// ═══════════════════════════════════════════════

// WhereClause يمثل شرط WHERE بسيط.
//
// يُنشأ عبر Eq, Ne, Gt, ...
type WhereClause struct {
	sql  string
	args []any
}

// ToSQL يرجّع SQL + args.
func (w WhereClause) ToSQL() (string, []any) {
	return w.sql, w.args
}

// IsNegated يرجّع false.
func (w WhereClause) IsNegated() bool { return false }

// IsEmpty يفحص إذا كان فارغًا.
func (w WhereClause) IsEmpty() bool { return w.sql == "" }

// SQL يرجّع SQL الخام (للتصحيح).
func (w WhereClause) SQL() string { return w.sql }

// Args يرجّع الـ args (للتصحيح).
func (w WhereClause) Args() []any { return w.args }

// ═══════════════════════════════════════════════
// NotClause — نفي شرط
// ═══════════════════════════════════════════════

// NotClause يمثل NOT (...) على شرط.
type NotClause struct {
	inner Clause
}

// ToSQL يرجّع SQL + args.
func (n NotClause) ToSQL() (string, []any) {
	if n.inner == nil {
		return "1=1", nil
	}
	sql, args := n.inner.ToSQL()
	return "NOT (" + sql + ")", args
}

// IsNegated يرجّع true.
func (n NotClause) IsNegated() bool { return true }

// ═══════════════════════════════════════════════
// RawClause — SQL خام (exported alias)
// ═══════════════════════════════════════════════

// RawClause يمثل SQL خام.
//
// ⚠️ المسؤولية على المستخدم — لا validation.
type RawClause = internal.RawClause

// Raw ينشئ RawClause.
//
// مثال:
//
//	gormz.Raw("age > ? AND status = ?", 18, "active")
func Raw(sql string, args ...any) RawClause {
	return internal.RawClause{SQL: sql, Args: args}
}

// ═══════════════════════════════════════════════
// Q — مجموعة شروط
// ═══════════════════════════════════════════════

// Q يمثل مجموعة شروط.
//
// مثال:
//
//	q := gormz.Or(
//	    gormz.Eq("status", "active"),
//	    gormz.Eq("status", "pending"),
//	)
//	users, _ := gormz.New[User]().Q(q).All()
type Q struct {
	op       string
	children []Clause
}

// ToSQL يحوّل Q إلى SQL + args.
func (q *Q) ToSQL() (string, []any) {
	return q.toSQL()
}

// IsNegated يرجّع false.
func (q *Q) IsNegated() bool { return false }

// Len يرجّع عدد الشروط.
func (q *Q) Len() int {
	if q == nil {
		return 0
	}
	return len(q.children)
}

// IsEmpty يفحص إذا كان Q فارغًا.
func (q *Q) IsEmpty() bool {
	return q == nil || len(q.children) == 0
}

// toSQL داخلي.
func (q *Q) toSQL() (string, []any) {
	if q == nil || len(q.children) == 0 {
		return "1=1", nil
	}

	parts := make([]string, 0, len(q.children))
	var args []any

	for _, c := range q.children {
		if c == nil {
			continue
		}
		sql, a := c.ToSQL()
		if sql == "" {
			continue
		}
		parts = append(parts, "("+sql+")")
		args = append(args, a...)
	}

	if len(parts) == 0 {
		return "1=1", nil
	}

	sep := " " + q.op + " "
	return strings.Join(parts, sep), args
}

// deepCopy نسخة عميقة.
func (q *Q) deepCopy() *Q {
	if q == nil {
		return nil
	}
	nq := &Q{op: q.op, children: make([]Clause, len(q.children))}
	for i, c := range q.children {
		if sub, ok := c.(*Q); ok {
			nq.children[i] = sub.deepCopy()
		} else {
			nq.children[i] = c
		}
	}
	return nq
}

// ═══════════════════════════════════════════════
// Lookup Constructors — مع panic (للاستخدام السريع)
// ═══════════════════════════════════════════════

// Eq → field = value.
//
// Panics إذا كان field غير صالح.
// استخدم EqErr للتحكم في الأخطاء.
func Eq(field string, value any) WhereClause {
	internal.MustValidateField(field)
	return WhereClause{sql: field + " = ?", args: []any{value}}
}

// EqErr مثل Eq لكن يرجّع خطأ.
func EqErr(field string, value any) (WhereClause, error) {
	if err := internal.ValidateField(field); err != nil {
		return WhereClause{}, err
	}
	return WhereClause{sql: field + " = ?", args: []any{value}}, nil
}

// Ne → field != value.
func Ne(field string, value any) WhereClause {
	internal.MustValidateField(field)
	return WhereClause{sql: field + " != ?", args: []any{value}}
}

// NeErr مثل Ne.
func NeErr(field string, value any) (WhereClause, error) {
	if err := internal.ValidateField(field); err != nil {
		return WhereClause{}, err
	}
	return WhereClause{sql: field + " != ?", args: []any{value}}, nil
}

// Gt → field > value.
func Gt(field string, value any) WhereClause {
	internal.MustValidateField(field)
	return WhereClause{sql: field + " > ?", args: []any{value}}
}

// GtErr مثل Gt.
func GtErr(field string, value any) (WhereClause, error) {
	if err := internal.ValidateField(field); err != nil {
		return WhereClause{}, err
	}
	return WhereClause{sql: field + " > ?", args: []any{value}}, nil
}

// Gte → field >= value.
func Gte(field string, value any) WhereClause {
	internal.MustValidateField(field)
	return WhereClause{sql: field + " >= ?", args: []any{value}}
}

// GteErr مثل Gte.
func GteErr(field string, value any) (WhereClause, error) {
	if err := internal.ValidateField(field); err != nil {
		return WhereClause{}, err
	}
	return WhereClause{sql: field + " >= ?", args: []any{value}}, nil
}

// Lt → field < value.
func Lt(field string, value any) WhereClause {
	internal.MustValidateField(field)
	return WhereClause{sql: field + " < ?", args: []any{value}}
}

// LtErr مثل Lt.
func LtErr(field string, value any) (WhereClause, error) {
	if err := internal.ValidateField(field); err != nil {
		return WhereClause{}, err
	}
	return WhereClause{sql: field + " < ?", args: []any{value}}, nil
}

// Lte → field <= value.
func Lte(field string, value any) WhereClause {
	internal.MustValidateField(field)
	return WhereClause{sql: field + " <= ?", args: []any{value}}
}

// LteErr مثل Lte.
func LteErr(field string, value any) (WhereClause, error) {
	if err := internal.ValidateField(field); err != nil {
		return WhereClause{}, err
	}
	return WhereClause{sql: field + " <= ?", args: []any{value}}, nil
}

// Contains → field LIKE '%value%'.
func Contains(field, value string) WhereClause {
	internal.MustValidateField(field)
	return WhereClause{sql: field + " LIKE ?", args: []any{"%" + value + "%"}}
}

// ContainsErr مثل Contains.
func ContainsErr(field, value string) (WhereClause, error) {
	if err := internal.ValidateField(field); err != nil {
		return WhereClause{}, err
	}
	return WhereClause{sql: field + " LIKE ?", args: []any{"%" + value + "%"}}, nil
}

// StartsWith → field LIKE 'value%'.
func StartsWith(field, value string) WhereClause {
	internal.MustValidateField(field)
	return WhereClause{sql: field + " LIKE ?", args: []any{value + "%"}}
}

// StartsWithErr مثل StartsWith.
func StartsWithErr(field, value string) (WhereClause, error) {
	if err := internal.ValidateField(field); err != nil {
		return WhereClause{}, err
	}
	return WhereClause{sql: field + " LIKE ?", args: []any{value + "%"}}, nil
}

// EndsWith → field LIKE '%value'.
func EndsWith(field, value string) WhereClause {
	internal.MustValidateField(field)
	return WhereClause{sql: field + " LIKE ?", args: []any{"%" + value}}
}

// EndsWithErr مثل EndsWith.
func EndsWithErr(field, value string) (WhereClause, error) {
	if err := internal.ValidateField(field); err != nil {
		return WhereClause{}, err
	}
	return WhereClause{sql: field + " LIKE ?", args: []any{"%" + value}}, nil
}

// In → field IN (values...).
func In(field string, values []any) WhereClause {
	internal.MustValidateField(field)
	if len(values) == 0 {
		return WhereClause{sql: "1=0"}
	}
	placeholders := strings.Repeat("?,", len(values))
	placeholders = placeholders[:len(placeholders)-1]
	return WhereClause{
		sql:  field + " IN (" + placeholders + ")",
		args: values,
	}
}

// InErr مثل In.
func InErr(field string, values []any) (WhereClause, error) {
	if err := internal.ValidateField(field); err != nil {
		return WhereClause{}, err
	}
	if len(values) == 0 {
		return WhereClause{sql: "1=0"}, nil
	}
	placeholders := strings.Repeat("?,", len(values))
	placeholders = placeholders[:len(placeholders)-1]
	return WhereClause{
		sql:  field + " IN (" + placeholders + ")",
		args: values,
	}, nil
}

// IsNull → field IS NULL.
func IsNull(field string) WhereClause {
	internal.MustValidateField(field)
	return WhereClause{sql: field + " IS NULL"}
}

// IsNullErr مثل IsNull.
func IsNullErr(field string) (WhereClause, error) {
	if err := internal.ValidateField(field); err != nil {
		return WhereClause{}, err
	}
	return WhereClause{sql: field + " IS NULL"}, nil
}

// NotNull → field IS NOT NULL.
func NotNull(field string) WhereClause {
	internal.MustValidateField(field)
	return WhereClause{sql: field + " IS NOT NULL"}
}

// NotNullErr مثل NotNull.
func NotNullErr(field string) (WhereClause, error) {
	if err := internal.ValidateField(field); err != nil {
		return WhereClause{}, err
	}
	return WhereClause{sql: field + " IS NOT NULL"}, nil
}

// ═══════════════════════════════════════════════
// Not — نفي شرط
// ═══════════════════════════════════════════════

// Not ينفي شرطًا.
//
//	Not(Eq("status", "deleted"))
//	→ NOT (status = ?)
func Not(c Clause) NotClause {
	return NotClause{inner: c}
}

// ═══════════════════════════════════════════════
// Q Constructors
// ═══════════════════════════════════════════════

// Or ينشئ Q مع OR.
//
//	q := gormz.Or(
//	    gormz.Eq("status", "active"),
//	    gormz.Eq("status", "pending"),
//	)
func Or(clauses ...Clause) *Q {
	return &Q{op: "OR", children: clauses}
}

// And ينشئ Q مع AND.
func And(clauses ...Clause) *Q {
	return &Q{op: "AND", children: clauses}
}

// Qb ينشئ Q فارغًا مع AND.
func Qb() *Q {
	return &Q{op: "AND"}
}

// QOr — alias للتوافق مع v0.1.0.
//
// Deprecated: استخدم Or.
func QOr(children ...any) *Q {
	return &Q{op: "OR", children: toClauses(children)}
}

// QAnd — alias للتوافق مع v0.1.0.
//
// Deprecated: استخدم And.
func QAnd(children ...any) *Q {
	return &Q{op: "AND", children: toClauses(children)}
}

// toClauses يحوّل []any إلى []Clause.
func toClauses(items []any) []Clause {
	out := make([]Clause, 0, len(items))
	for _, item := range items {
		if c, ok := item.(Clause); ok {
			out = append(out, c)
		}
	}
	return out
}

// ═══════════════════════════════════════════════
// Methods on Q
// ═══════════════════════════════════════════════

// And يضيف شروطًا بـ AND (immutable).
func (q *Q) And(clauses ...Clause) *Q {
	nq := q.deepCopy()
	if nq == nil {
		nq = &Q{op: "AND"}
	}
	nq.op = "AND"
	nq.children = append(nq.children, clauses...)
	return nq
}

// Or يضيف شروطًا بـ OR (immutable).
func (q *Q) Or(clauses ...Clause) *Q {
	nq := q.deepCopy()
	if nq == nil {
		nq = &Q{op: "OR"}
	}
	nq.op = "OR"
	nq.children = append(nq.children, clauses...)
	return nq
}

// AndGroup يضيف مجموعة AND متداخلة.
func (q *Q) AndGroup(clauses ...Clause) *Q {
	return q.And(And(clauses...))
}

// OrGroup يضيف مجموعة OR متداخلة.
func (q *Q) OrGroup(clauses ...Clause) *Q {
	return q.Or(Or(clauses...))
}
