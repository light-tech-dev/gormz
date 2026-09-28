# Aggregates — التجميعات

> GroupBy, Having, Distinct, Count/Sum/Avg.

---

## 📖 API

### GroupBy

```go
advanced.GroupBy[T](fields...).
    Count(field, alias).
    Sum(field, alias).
    Avg(field, alias).
    Having(sql, args...).
    ScanInto(&results)
```

### Aggregations

| Method | SQL |
|--------|-----|
| `Count(field, alias)` | `COUNT(field) AS alias` |
| `CountDistinct(field, alias)` | `COUNT(DISTINCT field)` |
| `Sum(field, alias)` | `SUM(field)` |
| `Avg(field, alias)` | `AVG(field)` |
| `Min(field, alias)` | `MIN(field)` |
| `Max(field, alias)` | `MAX(field)` |

---

## 💡 أمثلة

### مثال 1: Count per group

```go
type UserStats struct {
    UserID     uint
    OrderCount int64
}

var stats []UserStats

err := advanced.GroupBy[Order]("user_id").
    Count("id", "order_count").
    ScanInto(&stats)
```

### مثال 2: Multiple aggregates

```go
type Stats struct {
    UserID     uint
    OrderCount int64
    TotalSpent float64
    AvgOrder   float64
}

var stats []Stats

err := advanced.GroupBy[Order]("user_id").
    Count("id", "order_count").
    Sum("total", "total_spent").
    Avg("total", "avg_order").
    ScanInto(&stats)
```

### مثال 3: Having

```go
// المستخدمون الذين لديهم أكثر من 5 طلبات
err := advanced.GroupBy[Order]("user_id").
    Count("id", "order_count").
    Having("COUNT(id) > ?", 5).
    ScanInto(&stats)
```

### مثال 4: Filtered

```go
err := advanced.GroupBy[Order]("user_id").
    Where("status = ?", "paid").
    Where("created_at > ?", lastMonth).
    Sum("total", "revenue").
    ScanInto(&stats)
```

### مثال 5: Distinct

```go
// عدد المستخدمين الفريدين
var count int64
err := gormz.DB().Model(&Order{}).
    Distinct("user_id").
    Count(&count).Error
```

---

## ⚠️ تحذيرات

### 1. Select Columns

يجب أن تكون كل الأعمدة في `GROUP BY`:

```go
// ❌ خطأ (name ليس في GROUP BY)
advanced.GroupBy[Order]("user_id").
    Select("name").
    Count("id", "cnt")

// ✅ صحيح
advanced.GroupBy[Order]("user_id", "name").
    Count("id", "cnt")
```

### 2. NULL Handling

`SUM`/`AVG` **تتجاهل NULL**. استخدم `COALESCE`:

```go
advanced.GroupBy[Order]("user_id").
    Sum("total", "total")  // يستخدم COALESCE تلقائيًا
```

### 3. Performance

- أضف فهارس على group by columns
- `COUNT(*)` أسرع من `COUNT(field)`
- `COUNT(DISTINCT)` بطيء

---

## 📝 Best Practices

### ✅ Do

- استخدم `COALESCE` للـ NULL
- فهارس على `GROUP BY` columns
- `HAVING` بدل WHERE على aggregates

### ❌ Don't

- لا تخلط aggregates مع non-grouped columns
- لا تستخدم `COUNT(DISTINCT)` على جداول ضخمة