// Package internal provides internal helpers for gormz.
package internal

import (
	"fmt"
	"strings"
)

// ═══════════════════════════════════════════════
// Dialect — نوع قاعدة البيانات
// ═══════════════════════════════════════════════

// Dialect يمثل محرك قاعدة البيانات.
type Dialect string

const (
	DialectSQLite   Dialect = "sqlite"
	DialectPostgres Dialect = "postgres"
	DialectMySQL    Dialect = "mysql"
	DialectUnknown  Dialect = ""
)

// ═══════════════════════════════════════════════
// Builder — لبناء SQL بأمان
// ═══════════════════════════════════════════════

// Builder يمثل builder لبناء SQL بأمان.
//
// يستخدم لتجنب string concatenation المباشر.
//
//	b := NewBuilder()
//	b.WriteString("SELECT * FROM users")
//	b.WriteString(" WHERE active = ?", true)
//	b.WriteString(" ORDER BY name")
//	sql, args := b.Build()
type Builder struct {
	strings.Builder
	args []any
}

// NewBuilder ينشئ Builder جديد.
func NewBuilder() *Builder {
	return &Builder{}
}

// WriteString يضيف نصًا للـ SQL.
//
// إذا كان هناك args، تُضاف إلى قائمة الـ args.
func (b *Builder) WriteString(s string, args ...any) {
	b.Builder.WriteString(s)
	if len(args) > 0 {
		b.args = append(b.args, args...)
	}
}

// WriteQueryPart يضيف جزءًا من الاستعلام مع مسافة قبل.
func (b *Builder) WriteQueryPart(s string, args ...any) {
	if b.Builder.Len() > 0 && !strings.HasSuffix(b.Builder.String(), " ") {
		b.Builder.WriteString(" ")
	}
	b.WriteString(s, args...)
}

// Build يرجّع SQL النهائي + args.
func (b *Builder) Build() (string, []any) {
	return b.Builder.String(), b.args
}

// Args يرجّع الـ args.
func (b *Builder) Args() []any {
	return b.args
}

// Len يرجّع طول SQL الحالي.
func (b *Builder) Len() int {
	return b.Builder.Len()
}

// Reset يعيد تعيين Builder.
func (b *Builder) Reset() {
	b.Builder.Reset()
	b.args = nil
}

// ═══════════════════════════════════════════════
// Placeholders
// ═══════════════════════════════════════════════

// Placeholders ينشئ عدد n من "?" مفصولة بفواصل.
//
//	Placeholders(3) → "?,?,?"
//	Placeholders(0) → ""
func Placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	if n == 1 {
		return "?"
	}
	return strings.Repeat("?,", n-1) + "?"
}

// PlaceholdersWrapped ينشئ "(?,?,?)".
//
//	PlaceholdersWrapped(3) → "(?,?,?)"
//	PlaceholdersWrapped(0) → "()"
func PlaceholdersWrapped(n int) string {
	if n <= 0 {
		return "()"
	}
	return "(" + Placeholders(n) + ")"
}

// ═══════════════════════════════════════════════
// Identifier Quoting — ✅ إصلاح: dialect-aware
// ═══════════════════════════════════════════════

// QuoteIdentifier يحيط اسم الحقل بعلامات اقتباس مناسبة حسب dialect.
//
//	QuoteIdentifier("name", DialectSQLite)   → "`name`"
//	QuoteIdentifier("name", DialectPostgres) → "\"name\""
//	QuoteIdentifier("name", DialectMySQL)    → "`name`"
//
// ⚠️ يجب أن يكون الحقل صالحًا مسبقًا (لا يوجد validation هنا).
func QuoteIdentifier(name string, dialect Dialect) string {
	switch dialect {
	case DialectPostgres:
		return `"` + name + `"`
	case DialectMySQL, DialectSQLite:
		return "`" + name + "`"
	default:
		// افتراضي: ISO SQL (double quotes)
		return `"` + name + `"`
	}
}

// QuoteIdentifierSimple يحيط اسم الحقل بعلامات اقتباس (افتراضي: ISO).
//
// Deprecated: استخدم QuoteIdentifier مع dialect.
func QuoteIdentifierSimple(name string) string {
	return QuoteIdentifier(name, DialectUnknown)
}

// QuoteTableAndColumn يحيط جدول.عمود بعلامات اقتباس.
//
//	QuoteTableAndColumn("users", "id", DialectPostgres) → "users"."id"
func QuoteTableAndColumn(table, column string, dialect Dialect) string {
	return QuoteIdentifier(table, dialect) + "." + QuoteIdentifier(column, dialect)
}

// ═══════════════════════════════════════════════
// Join Helpers
// ═══════════════════════════════════════════════

// JoinConditions يدمج شروط SQL.
//
//	JoinConditions(" AND ", "a = ?", "b = ?") → "a = ? AND b = ?"
func JoinConditions(sep string, conditions ...string) string {
	return strings.Join(conditions, sep)
}

// ═══════════════════════════════════════════════
// Clause Builders
// ═══════════════════════════════════════════════

// BuildWhere يبني WHERE clause.
func BuildWhere(conditions []string, args []any) (string, []any) {
	if len(conditions) == 0 {
		return "", nil
	}
	sql := "WHERE " + strings.Join(conditions, " AND ")
	return sql, args
}

// BuildOrderBy يبني ORDER BY clause.
//
//	BuildOrderBy([]string{"name ASC", "age DESC"}) → "ORDER BY name ASC, age DESC"
func BuildOrderBy(orders []string) string {
	if len(orders) == 0 {
		return ""
	}
	return "ORDER BY " + strings.Join(orders, ", ")
}

// BuildGroupBy يبني GROUP BY clause.
func BuildGroupBy(fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	return "GROUP BY " + strings.Join(fields, ", ")
}

// BuildHaving يبني HAVING clause.
func BuildHaving(conditions []string, args []any) (string, []any) {
	if len(conditions) == 0 {
		return "", nil
	}
	sql := "HAVING " + strings.Join(conditions, " AND ")
	return sql, args
}

// BuildLimitOffset يبني LIMIT/OFFSET clause.
func BuildLimitOffset(limit, offset int) string {
	var b strings.Builder
	if limit > 0 {
		b.WriteString(fmt.Sprintf(" LIMIT %d", limit))
	}
	if offset > 0 {
		b.WriteString(fmt.Sprintf(" OFFSET %d", offset))
	}
	return b.String()
}

// ═══════════════════════════════════════════════
// Aggregations
// ═══════════════════════════════════════════════

// AggregateFunc يمثل دالة تجميع.
type AggregateFunc string

const (
	AggCount AggregateFunc = "COUNT"
	AggSum   AggregateFunc = "SUM"
	AggAvg   AggregateFunc = "AVG"
	AggMin   AggregateFunc = "MIN"
	AggMax   AggregateFunc = "MAX"
)

// Aggregate يبني تعبير تجميع.
//
//	Aggregate(AggSum, "total", "total_sum") → "SUM(total) AS total_sum"
func Aggregate(fn AggregateFunc, field, alias string) string {
	var expr string
	if field == "*" || field == "" {
		expr = string(fn) + "(*)"
	} else {
		expr = string(fn) + "(" + field + ")"
	}

	if alias != "" {
		expr += " AS " + alias
	}
	return expr
}

// AggregateDistinct يبني تعبير تجميع مع DISTINCT.
//
//	AggregateDistinct(AggCount, "user_id", "unique_users") → "COUNT(DISTINCT user_id) AS unique_users"
func AggregateDistinct(fn AggregateFunc, field, alias string) string {
	if field == "*" || field == "" {
		return Aggregate(fn, field, alias)
	}

	expr := string(fn) + "(DISTINCT " + field + ")"
	if alias != "" {
		expr += " AS " + alias
	}
	return expr
}

// ═══════════════════════════════════════════════
// Join Clauses
// ═══════════════════════════════════════════════

// JoinType نوع join.
type JoinType string

const (
	JoinInner JoinType = "INNER"
	JoinLeft  JoinType = "LEFT"
	JoinRight JoinType = "RIGHT"
	JoinFull  JoinType = "FULL"
	JoinCross JoinType = "CROSS"
)

// JoinClause يبني join clause.
//
//	JoinClause(JoinInner, "orders", "orders.user_id = users.id")
//	→ "INNER JOIN orders ON orders.user_id = users.id"
func JoinClause(jt JoinType, table, on string) string {
	if jt == JoinCross || on == "" {
		return string(jt) + " JOIN " + table
	}
	return string(jt) + " JOIN " + table + " ON " + on
}

// ═══════════════════════════════════════════════
// CTE
// ═══════════════════════════════════════════════

// CTEClause يبني CTE.
//
//	CTEClause("recent", "SELECT * FROM orders", false)
//	→ "WITH recent AS (SELECT * FROM orders)"
func CTEClause(name, sql string, recursive bool) string {
	prefix := "WITH "
	if recursive {
		prefix = "WITH RECURSIVE "
	}
	return prefix + name + " AS (" + sql + ")"
}

// ═══════════════════════════════════════════════
// CASE WHEN
// ═══════════════════════════════════════════════

// CaseWhen يمثل CASE WHEN clause.
type CaseWhen struct {
	conditions []caseCondition
	elseExpr   string
}

type caseCondition struct {
	when string
	then string
	args []any
}

// NewCaseWhen ينشئ CaseWhen جديد.
func NewCaseWhen() *CaseWhen {
	return &CaseWhen{}
}

// When يضيف شرط WHEN.
func (cw *CaseWhen) When(condition, then string, args ...any) *CaseWhen {
	cw.conditions = append(cw.conditions, caseCondition{
		when: condition,
		then: then,
		args: args,
	})
	return cw
}

// Else يضيف ELSE.
func (cw *CaseWhen) Else(expr string) *CaseWhen {
	cw.elseExpr = expr
	return cw
}

// Build يبني CASE WHEN.
func (cw *CaseWhen) Build() (string, []any) {
	var b strings.Builder
	b.WriteString("CASE")

	var allArgs []any
	for _, c := range cw.conditions {
		b.WriteString(" WHEN ")
		b.WriteString(c.when)
		b.WriteString(" THEN ")
		b.WriteString(c.then)
		allArgs = append(allArgs, c.args...)
	}

	if cw.elseExpr != "" {
		b.WriteString(" ELSE ")
		b.WriteString(cw.elseExpr)
	}

	b.WriteString(" END")
	return b.String(), allArgs
}