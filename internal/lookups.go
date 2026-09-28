package internal

import (
	"fmt"
	"reflect"
	"strings"
)

// ═══════════════════════════════════════════════
// Lookups — معالجة __lookup
// ═══════════════════════════════════════════════

// Lookup يمثل lookup واحد.
type Lookup struct {
	SQL     string
	Args    []any
	Negate  bool
	Dialect Dialect
}

// BuildLookup يحوّل "field__op" إلى SQL + args.
//
// يستخدم DialectSQLite افتراضيًا.
func BuildLookup(field string, value any, negate bool) (*Lookup, error) {
	return BuildLookupWithDialect(field, value, negate, DialectSQLite)
}

// BuildLookupWithDialect مثل BuildLookup لكن مع dialect.
func BuildLookupWithDialect(field string, value any, negate bool, dialect Dialect) (*Lookup, error) {
	col, op := SplitFieldLookup(field)

	if err := ValidateField(col); err != nil {
		return nil, err
	}

	if op == "" {
		return &Lookup{
			SQL:     col + " = ?",
			Args:    []any{value},
			Negate:  negate,
			Dialect: dialect,
		}, nil
	}

	if err := ValidateLookup(op); err != nil {
		return nil, err
	}

	switch op {
	case "gt":
		return newLookup(col+" > ?", value, negate, dialect), nil
	case "gte":
		return newLookup(col+" >= ?", value, negate, dialect), nil
	case "lt":
		return newLookup(col+" < ?", value, negate, dialect), nil
	case "lte":
		return newLookup(col+" <= ?", value, negate, dialect), nil
	case "ne":
		return newLookup(col+" != ?", value, negate, dialect), nil

	case "contains":
		return newLookup(col+" LIKE ?", "%"+toString(value)+"%", negate, dialect), nil
	case "icontains":
		return newLookup(lowerFunc(dialect, col)+" LIKE ?",
			"%"+strings.ToLower(toString(value))+"%", negate, dialect), nil
	case "startswith":
		return newLookup(col+" LIKE ?", toString(value)+"%", negate, dialect), nil
	case "istartswith":
		return newLookup(lowerFunc(dialect, col)+" LIKE ?",
			strings.ToLower(toString(value))+"%", negate, dialect), nil
	case "endswith":
		return newLookup(col+" LIKE ?", "%"+toString(value), negate, dialect), nil
	case "iendswith":
		return newLookup(lowerFunc(dialect, col)+" LIKE ?",
			"%"+strings.ToLower(toString(value)), negate, dialect), nil

	case "in":
		return buildInLookup(col, value, negate, dialect)
	case "notin":
		return buildInLookup(col, value, !negate, dialect)

	case "isnull":
		isNull, ok := value.(bool)
		if !ok {
			return nil, NewValidationError(field, "isnull expects bool")
		}
		if isNull {
			return newLookup(col+" IS NULL", nil, negate, dialect), nil
		}
		return newLookup(col+" IS NOT NULL", nil, negate, dialect), nil

	case "between":
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return nil, NewValidationError(field, "between expects slice/array")
		}
		if rv.Len() != 2 {
			return nil, NewValidationError(field, "between expects exactly 2 elements")
		}
		return newLookup(
			col+" BETWEEN ? AND ?",
			[]any{rv.Index(0).Interface(), rv.Index(1).Interface()},
			negate,
			dialect,
		), nil

	// ═══ Regex lookups ═══
	case "regex":
		return newLookup(
			regexFunc(dialect, col),
			toString(value),
			negate,
			dialect,
		), nil

	case "iregex":
		return newLookup(
			regexFunc(dialect, lowerFunc(dialect, col)),
			strings.ToLower(toString(value)),
			negate,
			dialect,
		), nil

	// ═══ Date lookups ═══
	case "year":
		return newLookup(dateExtract(dialect, "year", col)+" = ?", value, negate, dialect), nil
	case "month":
		return newLookup(dateExtract(dialect, "month", col)+" = ?", value, negate, dialect), nil
	case "day":
		return newLookup(dateExtract(dialect, "day", col)+" = ?", value, negate, dialect), nil

	default:
		return nil, NewValidationError(field, fmt.Sprintf("unknown lookup %q", op))
	}
}

// newLookup ينشئ Lookup مع args موحّدة.
func newLookup(sql string, args any, negate bool, dialect Dialect) *Lookup {
	l := &Lookup{SQL: sql, Negate: negate, Dialect: dialect}
	if args == nil {
		return l
	}
	switch v := args.(type) {
	case []any:
		l.Args = v
	default:
		l.Args = []any{args}
	}
	return l
}

// buildInLookup يبني IN/NOT IN lookup.
func buildInLookup(col string, value any, negate bool, dialect Dialect) (*Lookup, error) {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, NewValidationError("in", "expects slice/array")
	}

	n := rv.Len()
	if n == 0 {
		// IN () دائما false
		return &Lookup{SQL: "1=0", Negate: negate, Dialect: dialect}, nil
	}

	args := make([]any, n)
	for i := 0; i < n; i++ {
		args[i] = rv.Index(i).Interface()
	}

	placeholders := strings.Repeat("?,", n)
	placeholders = placeholders[:len(placeholders)-1]

	return &Lookup{
		SQL:     col + " IN (" + placeholders + ")",
		Args:    args,
		Negate:  negate,
		Dialect: dialect,
	}, nil
}

// ═══════════════════════════════════════════════
// Dialect-aware Helpers
// ═══════════════════════════════════════════════

// lowerFunc يرجّع دالة lower حسب dialect.
//
// جميع dialects تدعم LOWER() — لا حاجة لتبديل.
func lowerFunc(dialect Dialect, col string) string {
	return "LOWER(" + col + ")"
}

// regexFunc يرجّع تعبير regex حسب dialect.
//
// ⚠️ SQLite:
//   - الافتراضي: REGEXP — لكن يحتاج دالة مخصصة (register_function)
//   - البديل: GLOB — لا يحتاج دالة، لكن syntax مختلف
//
// ⚠️ PostgreSQL: ~ (case-sensitive)، ~* (case-insensitive)
//
// ⚠️ MySQL: REGEXP، REGEXP BINARY
func regexFunc(dialect Dialect, col string) string {
	switch dialect {
	case DialectSQLite:
		// SQLite لا يدعم REGEXP افتراضيًا
		// المستخدم يحتاج تسجيل دالة REGEXP عبر:
		//   db.ConnPool.(*sqlite3.SQLiteConn).RegisterFunc("regexp", ...)
		//
		// أو استخدام GLOB (لكن syntax مختلف)
		//
		// نستخدم REGEXP — المسؤولية على المستخدم
		return col + " REGEXP ?"

	case DialectPostgres:
		// PostgreSQL: ~ للمطابقة الحساسة للحالة
		return col + " ~ ?"

	case DialectMySQL:
		// MySQL: REGEXP للمطابقة الحساسة للحالة
		return col + " REGEXP ?"

	default:
		// ISO SQL لا يعرّف regex
		// نستخدم REGEXP كافتراض
		return col + " REGEXP ?"
	}
}

// dateExtract يرجّع تعبير استخراج التاريخ حسب dialect.
func dateExtract(dialect Dialect, part, col string) string {
	switch dialect {
	case DialectSQLite:
		format := map[string]string{
			"year":  "%Y",
			"month": "%m",
			"day":   "%d",
		}[part]
		return "CAST(strftime('" + format + "', " + col + ") AS INTEGER)"

	case DialectPostgres:
		return "EXTRACT(" + strings.ToUpper(part) + " FROM " + col + ")"

	case DialectMySQL:
		switch part {
		case "year":
			return "YEAR(" + col + ")"
		case "month":
			return "MONTH(" + col + ")"
		case "day":
			return "DAY(" + col + ")"
		}
		return "EXTRACT(" + strings.ToUpper(part) + " FROM " + col + ")"

	default:
		return "EXTRACT(" + strings.ToUpper(part) + " FROM " + col + ")"
	}
}

// toString يحوّل value إلى string.
func toString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
