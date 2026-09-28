package advanced

import (
	"fmt"
	"strings"

	"github.com/light-tech-dev/gormz"
)

// ═══════════════════════════════════════════════
// CTE — Common Table Expression
// ═══════════════════════════════════════════════

// CTE يمثل Common Table Expression.
//
// مثال:
//
//	cte := &advanced.CTE{
//	    Name: "recent_orders",
//	    SQL:  "SELECT * FROM orders WHERE created_at > ?",
//	    Args: []any{cutoff},
//	}
type CTE struct {
	// Name اسم CTE (يجب أن يكون صالحًا كـ identifier).
	Name string

	// SQL جسم CTE (بدون "name AS" وبدون "WITH").
	//
	// مثال: "SELECT * FROM orders WHERE total > 100"
	SQL string

	// Args معاملات SQL.
	Args []any

	// Recursive هل CTE recursive؟
	Recursive bool

	// Columns أسماء الأعمدة (اختياري).
	Columns []string
}

// NewCTE ينشئ CTE من QuerySet.
//
//	users := gormz.New[User]().Filter("active", true)
//	cte := advanced.NewCTE[User]("active_users", users)
func NewCTE[T any](name string, q *gormz.QuerySet[T]) *CTE {
	if q == nil {
		return &CTE{Name: name}
	}
	sql, args := q.ToSQL()
	return &CTE{
		Name: name,
		SQL:  sql,
		Args: args,
	}
}

// NewRecursiveCTE ينشئ recursive CTE.
//
// ⚠️ يجب أن تُضيف الجزء التكراري عبر UnionRaw.
func NewRecursiveCTE[T any](name string, base *gormz.QuerySet[T]) *CTE {
	if base == nil {
		return &CTE{Name: name, Recursive: true}
	}
	sql, args := base.ToSQL()
	return &CTE{
		Name:      name,
		SQL:       sql,
		Args:      args,
		Recursive: true,
	}
}

// WithColumns يحدد أسماء الأعمدة.
//
//	cte.WithColumns("id", "name", "total")
func (c *CTE) WithColumns(cols ...string) *CTE {
	c.Columns = cols
	return c
}

// UnionRaw يضيف UNION ALL مع SQL خام.
//
// يُستخدم لبناء recursive CTEs:
//
//	cte := advanced.NewRecursiveCTE[Category]("tree", baseQuery).
//	    UnionRaw("SELECT c.* FROM categories c JOIN tree t ON c.parent_id = t.id", nil...)
func (c *CTE) UnionRaw(sql string, args ...any) *CTE {
	if c.SQL != "" {
		c.SQL = c.SQL + " UNION ALL " + sql
	} else {
		c.SQL = sql
	}
	c.Args = append(c.Args, args...)
	return c
}

// ToSQL يرجّع جسم CTE كامل: "name AS (SELECT ...)".
//
// ⚠️ لا يُضيف "WITH" — ذلك مسؤولية CTEBuilder.
//
// مثال:
//
//	WITH recent AS (SELECT ...)   ← "WITH" من CTEBuilder
//	     ^^^^^^^^                ← هذا ما يُرجعه ToSQL
func (c *CTE) ToSQL() (string, []any) {
	if c == nil || c.Name == "" {
		return "", nil
	}
	if c.SQL == "" {
		return "", nil
	}

	var b strings.Builder
	b.WriteString(c.Name)

	if len(c.Columns) > 0 {
		b.WriteString(" (")
		b.WriteString(strings.Join(c.Columns, ", "))
		b.WriteString(")")
	}

	b.WriteString(" AS (")
	b.WriteString(c.SQL)
	b.WriteString(")")

	return b.String(), c.Args
}

// String يرجّع تمثيلًا نصيًا (للتصحيح).
func (c *CTE) String() string {
	sql, _ := c.ToSQL()
	return sql
}

// IsEmpty يفحص إذا كان فارغًا.
func (c *CTE) IsEmpty() bool {
	return c == nil || c.Name == "" || c.SQL == ""
}

// ═══════════════════════════════════════════════
// CTEBuilder — بناء استعلام مع CTEs
// ═══════════════════════════════════════════════

// CTEBuilder يمثل استعلامًا مع CTEs.
type CTEBuilder[T any] struct {
	ctes      []*CTE
	mainQuery *gormz.QuerySet[T]
}

// With يبدأ استعلامًا مع CTE.
//
// مثال:
//
//	cte := advanced.NewCTE[Order]("recent", recentOrders)
//
//	result, err := advanced.With[Order](cte).
//	    Query(gormz.New[Order]().Filter("status", "paid")).
//	    All()
func With[T any](ctes ...*CTE) *CTEBuilder[T] {
	// تصفية CTEs الفارغة
	filtered := make([]*CTE, 0, len(ctes))
	for _, c := range ctes {
		if c != nil && !c.IsEmpty() {
			filtered = append(filtered, c)
		}
	}
	return &CTEBuilder[T]{ctes: filtered}
}

// Query يحدد الاستعلام الرئيسي.
func (cb *CTEBuilder[T]) Query(q *gormz.QuerySet[T]) *CTEBuilder[T] {
	cb.mainQuery = q
	return cb
}

// All ينفّذ الاستعلام ويرجّع النتائج.
func (cb *CTEBuilder[T]) All() ([]T, error) {
	if cb.mainQuery == nil {
		return nil, fmt.Errorf("gormz/advanced: CTEBuilder requires Query")
	}

	cteSQL, cteArgs := cb.buildCTEs()
	mainSQL, mainArgs := cb.mainQuery.ToSQL()

	// دمج: WITH ctes <mainSQL>
	var fullSQL string
	if cteSQL != "" {
		fullSQL = cteSQL + " " + mainSQL
	} else {
		fullSQL = mainSQL
	}

	allArgs := append(cteArgs, mainArgs...)

	var zero T
	var results []T
	err := gormz.DB().Model(&zero).Raw(fullSQL, allArgs...).Scan(&results).Error
	return results, err
}

// ScanInto ينفّذ ويقرأ في struct مخصص.
//
//	type Row struct {
//	    Name  string
//	    Total float64
//	}
//	var rows []Row
//	err := cb.ScanInto(&rows)
func (cb *CTEBuilder[T]) ScanInto(dest any) error {
	if cb.mainQuery == nil {
		return fmt.Errorf("gormz/advanced: CTEBuilder requires Query")
	}
	if dest == nil {
		return fmt.Errorf("gormz/advanced: nil destination")
	}

	cteSQL, cteArgs := cb.buildCTEs()
	mainSQL, mainArgs := cb.mainQuery.ToSQL()

	var fullSQL string
	if cteSQL != "" {
		fullSQL = cteSQL + " " + mainSQL
	} else {
		fullSQL = mainSQL
	}

	allArgs := append(cteArgs, mainArgs...)

	return gormz.DB().Raw(fullSQL, allArgs...).Scan(dest).Error
}

// buildCTEs يبني سلسلة CTEs كاملة: "WITH ... , ... , ...".
//
// ✅ منطق واضح:
//  1. تحديد هل يوجد recursive (مرة واحدة)
//  2. بناء كل CTE بنفس الطريقة (عبر ToSQL)
//  3. دمجها بفاصلة
//  4. إضافة البادئة (WITH / WITH RECURSIVE)
//
// مثال الناتج:
//
//	WITH RECURSIVE
//	    tree AS (SELECT ...),
//	    leaves AS (SELECT ...)
func (cb *CTEBuilder[T]) buildCTEs() (string, []any) {
	if len(cb.ctes) == 0 {
		return "", nil
	}

	var parts []string
	var allArgs []any

	// هل يوجد recursive؟
	recursive := false
	for _, cte := range cb.ctes {
		if cte.Recursive {
			recursive = true
			break
		}
	}

	// بناء كل CTE بنفس الطريقة
	for _, cte := range cb.ctes {
		sql, args := cte.ToSQL()
		if sql == "" {
			continue
		}
		parts = append(parts, sql)
		allArgs = append(allArgs, args...)
	}

	if len(parts) == 0 {
		return "", nil
	}

	// البادئة
	var prefix string
	if recursive {
		prefix = "WITH RECURSIVE "
	} else {
		prefix = "WITH "
	}

	return prefix + strings.Join(parts, ", "), allArgs
}

// ═══════════════════════════════════════════════
// RecursiveTree — CTE عودي جاهز للشجرة
// ═══════════════════════════════════════════════

// RecursiveTree يبني CTE عودي للشجرة (hierarchies).
//
// مثال — شجرة الأقسام:
//
//	// 1. الـ base query (الأب الجذري)
//	base := gormz.New[Department]().Filter("id", rootID)
//
//	// 2. بناء CTE عودي
//	cte := advanced.RecursiveTree[Department](
//	    "dept_tree",
//	    "id",         // idCol
//	    "parent_id",  // parentCol
//	    []string{"id", "name", "parent_id"}, // allCols
//	    func(cteName string) string {
//	        return "SELECT d.* FROM departments d " +
//	               "JOIN " + cteName + " t ON d.parent_id = t.id"
//	    },
//	    rootID,
//	)
//
//	// 3. استخدام
//	result, err := advanced.With[Department](cte).
//	    Query(gormz.New[Department]()).
//	    All()
//
// ⚠️ في PostgreSQL، يجب استخدام "WITH RECURSIVE" حتى لو كان CTE واحد.
func RecursiveTree[T any](
	cteName string,
	idCol, parentCol string,
	allCols []string,
	recursiveQuery func(cteName string) string,
	rootID any,
) *CTE {
	tableName := gormz.TableNameOf[T]()

	// الجزء الأول: الأب الجذري
	cols := strings.Join(allCols, ", ")
	if cols == "" {
		cols = "*"
	}

	baseSQL := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s = ?",
		cols, tableName, idCol,
	)

	// الجزء العودي
	recursive := recursiveQuery(cteName)

	return &CTE{
		Name:      cteName,
		SQL:       baseSQL + " UNION ALL " + recursive,
		Args:      []any{rootID},
		Recursive: true,
		Columns:   allCols,
	}
}

// RecursiveTreeWithDialect مثل RecursiveTree لكن يدعم تسمية الأعمدة حسب dialect.
func RecursiveTreeWithDialect[T any](
	cteName string,
	idCol, parentCol string,
	allCols []string,
	recursiveQuery func(cteName string) string,
	rootID any,
) *CTE {
	return RecursiveTree[T](
		cteName,
		idCol,
		parentCol,
		allCols,
		recursiveQuery,
		rootID,
	)
}

// ═══════════════════════════════════════════════
// Convenience: Single CTE
// ═══════════════════════════════════════════════

// WithCTE اختصار لـ CTE واحد.
//
//	result, err := advanced.WithCTE[User](cte).
//	    Query(gormz.New[User]()).
//	    All()
func WithCTE[T any](cte *CTE) *CTEBuilder[T] {
	return With[T](cte)
}

// ═══════════════════════════════════════════════
// String Helpers
// ═══════════════════════════════════════════════

// FormatCTEs يرجّع CTEs منسقة (للتصحيح).
func FormatCTEs(ctes ...*CTE) string {
	var parts []string
	for _, c := range ctes {
		if c == nil || c.IsEmpty() {
			continue
		}
		sql, _ := c.ToSQL()
		parts = append(parts, sql)
	}
	if len(parts) == 0 {
		return ""
	}

	prefix := "WITH "
	for _, c := range ctes {
		if c != nil && c.Recursive {
			prefix = "WITH RECURSIVE "
			break
		}
	}

	return prefix + strings.Join(parts, ",\n     ")
}
