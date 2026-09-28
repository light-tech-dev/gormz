// Package internal provides internal helpers for gormz.
package internal

import (
	"fmt"
	"regexp"
	"strings"
)

// ═══════════════════════════════════════════════
// Regex Patterns
// ═══════════════════════════════════════════════

var (
	// fieldRegex يطابق أسماء الحقول الصحيحة.
	//
	// يسمح بـ:
	//   - snake_case: user_id
	//   - dotted: users.id
	//   - prefixed: t1.name
	//   - numbers: field_1
	fieldRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$`)

	// lookupRegex يطابق أسماء lookups.
	lookupRegex = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

	// tableRegex يطابق أسماء الجداول الصحيحة.
	tableRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

	// identifierRegex يطابق identifiers عامة (أعمدة، جداول، aliases).
	identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

	// directionRegex يطابق direction كامل (مع NULLS FIRST/LAST).
	directionRegex = regexp.MustCompile(`^(ASC|DESC)(\s+NULLS\s+(FIRST|LAST))?$`)
)

// ═══════════════════════════════════════════════
// SQL Keywords — كلمات محجوزة
// ═══════════════════════════════════════════════

// sqlKeywords كلمات محجوزة (لا يمكن استخدامها كأسماء حقول).
var sqlKeywords = map[string]bool{
	// DML
	"select": true, "from": true, "where": true, "insert": true,
	"update": true, "delete": true, "merge": true, "into": true,
	"values": true, "set": true, "returning": true,

	// DDL
	"create": true, "alter": true, "drop": true, "truncate": true,
	"table": true, "view": true, "index": true, "sequence": true,
	"schema": true, "database": true, "column": true, "constraint": true,

	// DCL
	"grant": true, "revoke": true,

	// DML/JOIN
	"join": true, "inner": true, "outer": true, "left": true,
	"right": true, "full": true, "cross": true, "on": true, "using": true,

	// Aggregates
	"group": true, "having": true, "order": true, "by": true,
	"asc": true, "desc": true, "limit": true, "offset": true,
	"distinct": true, "union": true, "intersect": true, "except": true,
	"all": true, "any": true, "some": true,

	// Logical
	"and": true, "or": true, "not": true, "in": true,
	"exists": true, "between": true, "like": true, "ilike": true,
	"regexp": true, "case": true, "when": true, "then": true,
	"else": true, "end": true,

	// Null/Bool
	"null": true, "true": true, "false": true, "is": true,
	"unknown": true,

	// Misc
	"as": true, "with": true, "recursive": true, "over": true,
	"partition": true, "primary": true, "foreign": true,
	"key": true, "references": true, "unique": true, "check": true,
	"default": true, "auto_increment": true,
}

// ═══════════════════════════════════════════════
// Field Validation
// ═══════════════════════════════════════════════

// ValidateField يتحقق من صحة اسم حقل واحد.
//
// يعيد ValidationError إذا كان الحقل غير صالح.
//
// يقبل:
//   - snake_case: user_id
//   - dotted: users.id
//   - prefixed: t1.name
//
// يرفض:
//   - أسماء فارغة
//   - أحرف غير صالحة (; ' ")
//   - كلمات SQL محجوزة
func ValidateField(field string) error {
	if field == "" {
		return NewValidationError(field, "empty field name")
	}
	if len(field) > 128 {
		return NewValidationError(field, "field name too long (max 128)")
	}
	if !fieldRegex.MatchString(field) {
		return NewValidationError(field, "invalid characters (allowed: a-z, A-Z, 0-9, _, .)")
	}

	// فحص الكلمات المحجوزة (للجزء الأول فقط)
	base := field
	if i := strings.Index(field, "."); i > 0 {
		base = field[:i]
	}
	if sqlKeywords[strings.ToLower(base)] {
		return NewValidationError(field, "reserved SQL keyword")
	}
	return nil
}

// ValidateFields يتحقق من عدة حقول.
func ValidateFields(fields ...string) error {
	for _, f := range fields {
		if err := ValidateField(f); err != nil {
			return err
		}
	}
	return nil
}

// MustValidateField يتحقق وpanic عند الخطأ.
//
// للاستخدام الداخلي عندما نعرف أن الحقل صالح.
func MustValidateField(field string) {
	if err := ValidateField(field); err != nil {
		panic(err)
	}
}

// MustValidateFields يتحقق من عدة حقول وpanic عند الخطأ.
func MustValidateFields(fields ...string) {
	if err := ValidateFields(fields...); err != nil {
		panic(err)
	}
}

// ═══════════════════════════════════════════════
// Identifier Validation — عام
// ═══════════════════════════════════════════════

// ValidateIdentifier يتحقق من صحة identifier عام.
//
// يُستخدم لـ: أسماء الجداول، الأعمدة، aliases، CTE names.
//
// يرفض:
//   - أسماء فارغة
//   - أحرف غير صالحة
//   - كلمات SQL محجوزة
func ValidateIdentifier(id string) error {
	if id == "" {
		return NewValidationError(id, "empty identifier")
	}
	if len(id) > 64 {
		return NewValidationError(id, "identifier too long (max 64)")
	}
	if !identifierRegex.MatchString(id) {
		return NewValidationError(id, "invalid identifier characters")
	}
	if sqlKeywords[strings.ToLower(id)] {
		return NewValidationError(id, "reserved SQL keyword")
	}
	return nil
}

// MustValidateIdentifier يتحقق وpanic عند الخطأ.
func MustValidateIdentifier(id string) {
	if err := ValidateIdentifier(id); err != nil {
		panic(err)
	}
}

// ValidateIdentifiers يتحقق من عدة identifiers.
func ValidateIdentifiers(ids ...string) error {
	for _, id := range ids {
		if err := ValidateIdentifier(id); err != nil {
			return err
		}
	}
	return nil
}

// ═══════════════════════════════════════════════
// Lookup Validation
// ═══════════════════════════════════════════════

// validLookups قائمة lookups الصحيحة.
//
// ⚠️ يجب أن تتطابق مع الحالات في lookups.go
var validLookups = map[string]bool{
	// Comparison
	"gt": true, "gte": true, "lt": true, "lte": true, "ne": true,

	// String
	"contains": true, "icontains": true,
	"startswith": true, "istartswith": true,
	"endswith": true, "iendswith": true,

	// Set
	"in": true, "notin": true,

	// Null
	"isnull": true,

	// Range
	"between": true,

	// Regex
	"regex": true, "iregex": true,

	// Date
	"year": true, "month": true, "day": true,
}

// ValidateLookup يتحقق من صحة اسم lookup.
func ValidateLookup(lookup string) error {
	if lookup == "" {
		return NewValidationError(lookup, "empty lookup")
	}
	if !lookupRegex.MatchString(lookup) {
		return NewValidationError(lookup, "invalid lookup name")
	}
	if !validLookups[lookup] {
		return NewValidationError(lookup, "unknown lookup")
	}
	return nil
}

// IsValidLookup يفحص إذا كان lookup صالحًا.
func IsValidLookup(lookup string) bool {
	return validLookups[lookup]
}

// SupportedLookups يرجّع قائمة lookups المدعومة (مرتبة).
func SupportedLookups() []string {
	out := make([]string, 0, len(validLookups))
	for k := range validLookups {
		out = append(out, k)
	}
	return out
}

// ═══════════════════════════════════════════════
// Operator Validation
// ═══════════════════════════════════════════════

// validOperators قائمة operators الصحيحة.
var validOperators = map[string]bool{
	"=": true, "!=": true, "<>": true,
	">": true, ">=": true, "<": true, "<=": true,
	"LIKE": true, "NOT LIKE": true, "ILIKE": true,
	"IN": true, "NOT IN": true,
	"IS NULL": true, "IS NOT NULL": true,
	"BETWEEN": true, "NOT BETWEEN": true,
	"REGEXP": true, "NOT REGEXP": true,
}

// ValidateOperator يتحقق من صحة operator.
func ValidateOperator(op string) error {
	if !validOperators[strings.ToUpper(op)] {
		return NewValidationError(op, "invalid operator")
	}
	return nil
}

// ═══════════════════════════════════════════════
// SplitFieldLookup
// ═══════════════════════════════════════════════

// SplitFieldLookup يقسم "field__lookup" إلى جزأيه.
//
//	SplitFieldLookup("age__gt")               → ("age", "gt")
//	SplitFieldLookup("name")                  → ("name", "")
//	SplitFieldLookup("user__email__contains") → ("user__email", "contains")
func SplitFieldLookup(s string) (field, lookup string) {
	// ابحث عن آخر "__"
	if i := strings.LastIndex(s, "__"); i > 0 {
		return s[:i], s[i+2:]
	}
	return s, ""
}

// ═══════════════════════════════════════════════
// Direction Sanitization
// ═══════════════════════════════════════════════

// SanitizeDirection يتحقق من صحة direction.
//
// يقبل:
//
//   - "ASC" / "DESC" / "" (فارغ)
//
//   - "ASC NULLS FIRST" / "DESC NULLS LAST"
//
//     SanitizeDirection("asc")                → ("ASC", nil)
//     SanitizeDirection("DESC")               → ("DESC", nil)
//     SanitizeDirection("ASC NULLS FIRST")    → ("ASC NULLS FIRST", nil)
//     SanitizeDirection("bad")                → ("", error)
func SanitizeDirection(dir string) (string, error) {
	d := strings.ToUpper(strings.TrimSpace(dir))
	if d == "" {
		return "", nil
	}
	if directionRegex.MatchString(d) {
		return d, nil
	}
	return "", NewValidationError(dir, "invalid direction (allowed: ASC, DESC, ASC NULLS FIRST, DESC NULLS LAST)")
}

// ═══════════════════════════════════════════════
// Table Name Validation
// ═══════════════════════════════════════════════

// ValidateTableName يتحقق من صحة اسم جدول.
func ValidateTableName(table string) error {
	if table == "" {
		return NewValidationError(table, "empty table name")
	}
	if len(table) > 64 {
		return NewValidationError(table, "table name too long (max 64)")
	}
	if !tableRegex.MatchString(table) {
		return NewValidationError(table, "invalid table name")
	}
	if sqlKeywords[strings.ToLower(table)] {
		return NewValidationError(table, "reserved SQL keyword")
	}
	return nil
}

// MustValidateTableName يتحقق وpanic عند الخطأ.
func MustValidateTableName(table string) {
	if err := ValidateTableName(table); err != nil {
		panic(err)
	}
}

// ═══════════════════════════════════════════════
// ValidationError Type
// ═══════════════════════════════════════════════

// ValidationError خطأ تحقق.
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("gormz: invalid field %q: %s", e.Field, e.Reason)
}

// NewValidationError ينشئ خطأ تحقق.
func NewValidationError(field, reason string) *ValidationError {
	return &ValidationError{Field: field, Reason: reason}
}
