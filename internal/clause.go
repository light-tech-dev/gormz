// Package internal provides internal helpers for gormz.
package internal

// ═══════════════════════════════════════════════
// Clauses — أنواع الشروط الداخلية
// ═══════════════════════════════════════════════

// RawClause يمثل شرط SQL خام.
//
// يُستخدم داخل gormz لتمرير SQL + args بدون validation.
//
// ⚠️ للاستخدام الداخلي فقط.
type RawClause struct {
	SQL  string
	Args []any
}

// NewRawClause ينشئ RawClause.
func NewRawClause(sql string, args ...any) RawClause {
	return RawClause{SQL: sql, Args: args}
}

// IsEmpty يفحص إذا كان الشرط فارغًا.
func (r RawClause) IsEmpty() bool {
	return r.SQL == ""
}

// String يرجّع تمثيل نصي.
func (r RawClause) String() string {
	if len(r.Args) == 0 {
		return r.SQL
	}
	return r.SQL
}

// ToSQL يرجّع SQL + args.
//
// ✅ يطبّق واجهة gormz.Clause.
func (r RawClause) ToSQL() (string, []any) {
	return r.SQL, r.Args
}

// IsNegated يرجّع false.
//
// ✅ يطبّق واجهة gormz.Clause.
func (r RawClause) IsNegated() bool {
	return false
}
