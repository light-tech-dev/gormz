# Window Functions — دوال النافذة

> ROW_NUMBER, RANK, LAG, LEAD, ...

---

## 📖 API

### Basic Functions

| Function | SQL |
|----------|-----|
| `RowNumber(alias, partition...)` | `ROW_NUMBER() OVER (...)` |
| `Rank(alias, orderBy, partition...)` | `RANK() OVER (...)` |
| `DenseRank(alias, orderBy, partition...)` | `DENSE_RANK() OVER (...)` |
| `NTile(n, alias, orderBy, partition...)` | `NTILE(n) OVER (...)` |

### Lead/Lag

| Function | SQL |
|----------|-----|
| `Lag(field, offset, alias, partition...)` | `LAG(field, offset) OVER (...)` |
| `Lead(field, offset, alias, partition...)` | `LEAD(field, offset) OVER (...)` |

### Aggregates

| Function | SQL |
|----------|-----|
| `RunningSum(field, alias, orderBy, partition...)` | `SUM(field) OVER (...)` |
| `RunningCount(alias, orderBy, partition...)` | `COUNT(*) OVER (...)` |
| `AvgOver(field, alias, partition...)` | `AVG(field) OVER (...)` |

### Value

| Function | SQL |
|----------|-----|
| `FirstValue(field, alias, orderBy, partition...)` | `FIRST_VALUE(field) OVER (...)` |
| `LastValue(field, alias, orderBy, partition...)` | `LAST_VALUE(field) OVER (...)` |

---

## 💡 أمثلة

### مثال 1: ROW_NUMBER

```go
// ترتيب الطلبات لكل مستخدم
type Result struct {
    UserID    uint
    OrderID   uint
    RowNumber int
}

var results []Result

err := gormz.New[Order]().
    Select(
        "user_id",
        "id as order_id",
        advanced.RowNumber("row_number", "user_id"),
    ).
    OrderBy("user_id", "-created_at").
    ScanInto(&results)
```

### مثال 2: RANK

```go
// ترتيب الطلاب في كل صف
rows := gormz.New[Student]().
    Select(
        "name",
        "score",
        "class_id",
        advanced.Rank("rank", "score DESC", "class_id"),
    ).
    All()
```

### مثال 3: LAG/LEAD

```go
// مقارنة كل شهر بالسابق
type MonthlyStat struct {
    Month     string
    Revenue   float64
    PrevRev   *float64
    NextRev   *float64
}

var stats []MonthlyStat

err := gormz.New[Sale]().
    Select(`
        strftime('%Y-%m', created_at) as month,
        SUM(amount) as revenue,
        `+advanced.Lag("SUM(amount)", 1, "prev_rev", "DATE_TRUNC('month', created_at)")+`,
        `+advanced.Lead("SUM(amount)", 1, "next_rev", "DATE_TRUNC('month', created_at)")+`
    `).
    Group("strftime('%Y-%m', created_at)").
    Order("month").
    ScanInto(&stats)
```

### مثال 4: Running Total

```go
// إجمالي تراكمي
type DailySales struct {
    Date         string
    DailyAmount  float64
    RunningTotal float64
}

var sales []DailySales

err := gormz.New[Sale]().
    Select(`
        DATE(created_at) as date,
        SUM(amount) as daily_amount,
        `+advanced.RunningSum("amount", "running_total", "created_at")+`
    `).
    Group("DATE(created_at)").
    Order("date").
    ScanInto(&sales)
```

### مثال 5: Top N per Group

```go
// أعلى 3 موظفين راتباً في كل قسم
type TopEmployee struct {
    DepartmentID uint
    Name         string
    Salary       float64
    Rank         int
}

var results []TopEmployee

// استخدم subquery
subq := gormz.DB().Raw(`
    SELECT * FROM (
        SELECT
            department_id,
            name,
            salary,
            ROW_NUMBER() OVER (PARTITION BY department_id ORDER BY salary DESC) as rank
        FROM employees
    ) ranked
    WHERE rank <= 3
`).Scan(&results)
```

---

## ⚠️ Dialect Support

| DB | Version |
|----|---------|
| SQLite | 3.25+ ✅ |
| PostgreSQL | ✅ |
| MySQL | 8.0+ ✅ |
| MariaDB | 10.2+ ✅ |

### Examples

**PostgreSQL/Standard**:
```sql
ROW_NUMBER() OVER (PARTITION BY dept ORDER BY salary DESC)
```

**SQLite** (نفس الشيء):
```sql
ROW_NUMBER() OVER (PARTITION BY dept ORDER BY salary DESC)
```

---

## 📊 Performance

### الفهارس

```sql
-- للـ PARTITION BY department_id ORDER BY salary
CREATE INDEX idx_emp_dept_salary ON employees(department_id, salary);
```

### القواعد

- Window functions قد تكون بطيئة
- استخدم الفهارس على partition/order columns
- فضّل window على subqueries متعددة

---

## 📝 Best Practices

### ✅ Do

- استخدم window بدل subqueries لـ top N
- أضف فهارس مناسبة
- استخدم `ROW_NUMBER` للترقيم

### ❌ Don't

- لا تستخدم window في WHERE (استخدم subquery)
- لا تخلط window مع GROUP BY بدون فكر
- لا تنسَ ORDER BY للـ deterministic