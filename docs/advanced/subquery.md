# SubQuery — الاستعلامات الفرعية

> استعلام داخل استعلام.

---

## 📖 API

### `SubFrom[T]`

```go
subq := advanced.SubFrom[Order](
    gormz.New[Order]().Select("user_id").Filter("total__gt", 1000),
    "user_id",
)
```

### `SubRaw`

```go
subq := advanced.SubRaw("SELECT id FROM users WHERE age > ?", 18)
```

### Operators

| Function | SQL |
|----------|-----|
| `In(field, sub)` | `field IN (subquery)` |
| `NotIn(field, sub)` | `field NOT IN (subquery)` |
| `Exists(sub)` | `EXISTS (subquery)` |
| `NotExists(sub)` | `NOT EXISTS (subquery)` |
| `GtSub(field, sub)` | `field > (subquery)` |
| `LtSub(field, sub)` | `field < (subquery)` |
| `EqSub(field, sub)` | `field = (subquery)` |

---

## 💡 أمثلة

### مثال 1: IN SubQuery

```go
// المستخدمون الذين لديهم طلبات > 1000
subq := advanced.SubFrom[Order](
    gormz.New[Order]().
        Select("user_id").
        Filter("total__gt", 1000),
    "user_id",
)

users, err := gormz.New[User]().
    Where("id IN ("+subq.SQL+")", subq.Args...).
    All()
```

### مثال 2: EXISTS

```go
// المستخدمون الذين لديهم طلبات
subq := advanced.SubRaw("SELECT 1 FROM orders WHERE orders.user_id = users.id")

users, _ := gormz.New[User]().
    Where("EXISTS ("+subq.SQL+")").
    All()
```

### مثال 3: NOT EXISTS

```go
// المستخدمون بدون طلبات
subq := advanced.SubRaw("SELECT 1 FROM orders WHERE orders.user_id = users.id")

users, _ := gormz.New[User]().
    Where("NOT EXISTS ("+subq.SQL+")").
    All()
```

### مثال 4: Comparison

```go
// الطلبات التي total > متوسط الطلبات
subq := advanced.SubRaw("SELECT AVG(total) FROM orders")

orders, _ := gormz.New[Order]().
    Where("total > ("+subq.SQL+")").
    All()
```

### مثال 5: Correlated

```go
// الموظفون الذين راتبهم > متوسط راتب قسمهم
users, _ := gormz.New[User]().
    Where(`
        salary > (
            SELECT AVG(salary)
            FROM users u2
            WHERE u2.department_id = users.department_id
        )
    `).
    All()
```

---

## ⚠️ تحذيرات

### 1. Performance

Subqueries قد تكون بطيئة. استخدم Joins بدلًا منها:

```go
// ❌ SubQuery
users := gormz.New[User]().
    Where("id IN (SELECT user_id FROM orders WHERE status = 'paid')").
    All()

// ✅ Join
users, _ := advanced.WithJoins[User](gormz.New[User]()).
    Inner("orders", "orders.user_id = users.id").
    Filter("orders.status", "paid").
    All()
```

### 2. Correlated SubQuery

أبطأ من non-correlated:

```sql
-- ❌ بطيء (correlated)
SELECT * FROM users
WHERE age > (SELECT AVG(age) FROM users WHERE department_id = users.department_id)

-- ✅ أسرع (aggregate أولًا)
WITH dept_avg AS (
    SELECT department_id, AVG(age) as avg_age FROM users GROUP BY department_id
)
SELECT * FROM users u JOIN dept_avg d ON u.department_id = d.department_id
WHERE u.age > d.avg_age
```

### 3. Safety

`SubRaw` **لا يتحقق** من SQL:

```go
// ⚠️ مسؤوليتك
subq := advanced.SubRaw("SELECT * FROM " + userTable)
```

استخدم `SubFrom` الآمن:

```go
// ✅ آمن
subq := advanced.SubFrom[User](gormz.New[User](), "id")
```

---

## 📝 Best Practices

### ✅ Do

- استخدم `EXISTS` بدل `IN` للـ large sets
- استخدم Joins عند الإمكان
- استخدم CTE للاستعلامات المعقّدة
- اختبر الأداء

### ❌ Don't

- لا تستخدم subqueries في loop
- لا تخلط correlated + aggregate بدون فهرسة
- لا تنسَ الفهارس