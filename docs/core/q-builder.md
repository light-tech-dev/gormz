# Q Builder — الشروط المعقدة

> بناء شروط AND/OR/NOT معقدة.

---

## 📖 الأساسيات

### Constructors

| Function | SQL |
|----------|-----|
| `Qb()` | `AND` افتراضي |
| `QOr(a, b)` | `a OR b` |
| `QAnd(a, b)` | `a AND b` |

### Conditions

| Function | SQL |
|----------|-----|
| `Eq(field, v)` | `field = v` |
| `Ne(field, v)` | `field != v` |
| `Gt(field, v)` | `field > v` |
| `Gte(field, v)` | `field >= v` |
| `Lt(field, v)` | `field < v` |
| `Lte(field, v)` | `field <= v` |
| `Contains(field, s)` | `field LIKE '%s%'` |
| `StartsWith(field, s)` | `field LIKE 's%'` |
| `EndsWith(field, s)` | `field LIKE '%s'` |
| `In(field, values)` | `field IN (...)` |
| `IsNull(field)` | `field IS NULL` |
| `NotNull(field)` | `field IS NOT NULL` |

### Negation

```go
Not(condition)
```

---

## 💡 أمثلة

### مثال 1: OR بسيط

```go
q := gormz.QOr(
    gormz.Eq("status", "active"),
    gormz.Eq("status", "pending"),
)

users, _ := gormz.New[User]().Q(q).All()
// SELECT * FROM users WHERE status = 'active' OR status = 'pending'
```

### مثال 2: AND

```go
q := gormz.QAnd(
    gormz.Eq("active", true),
    gormz.Gt("age", 18),
)

// active = true AND age > 18
```

### مثال 3: معقّد

```go
// (status = 'active' OR status = 'pending') AND age > 18
q := gormz.Qb().And(
    gormz.QOr(
        gormz.Eq("status", "active"),
        gormz.Eq("status", "pending"),
    ),
    gormz.Gt("age", 18),
)

users, _ := gormz.New[User]().Q(q).All()
```

### مثال 4: NOT

```go
q := gormz.QAnd(
    gormz.Not(gormz.Eq("status", "deleted")),
    gormz.Eq("active", true),
)

// NOT (status = 'deleted') AND active = true
// status != 'deleted' AND active = true
```

### مثال 5: NOT مع Q

```go
inner := gormz.QOr(
    gormz.Eq("a", 1),
    gormz.Eq("b", 2),
)

q := gormz.QAnd(
    gormz.Not(inner),
    gormz.Eq("c", 3),
)

// NOT (a = 1 OR b = 2) AND c = 3
```

### مثال 6: تداخل عميق

```go
q := gormz.QAnd(
    gormz.QOr(
        gormz.Eq("type", "A"),
        gormz.QAnd(
            gormz.Eq("type", "B"),
            gormz.Gt("value", 100),
        ),
    ),
    gormz.Eq("active", true),
)

// (type = 'A' OR (type = 'B' AND value > 100)) AND active = true
```

### مثال 7: دمج مع Filter

```go
q := gormz.QOr(
    gormz.Eq("role", "admin"),
    gormz.Eq("role", "moderator"),
)

users, _ := gormz.New[User]().
    Filter("active", true).       // AND
    Q(q).                          // AND (role = 'admin' OR role = 'moderator')
    Filter("age__gte", 18).
    All()
// active = true AND (role = 'admin' OR role = 'moderator') AND age >= 18
```

---

## 🔧 Methods

### `And(children...)`

```go
q := gormz.Qb().
    And(gormz.Eq("a", 1)).
    And(gormz.Eq("b", 2))
// a = 1 AND b = 2
```

### `Or(children...)`

```go
q := gormz.Qb().
    Or(gormz.Eq("a", 1)).
    Or(gormz.Eq("b", 2))
// a = 1 OR b = 2
```

⚠️ **مهم**: كلاهما **immutable**:

```go
base := gormz.Qb().And(gormz.Eq("a", 1))
extended := base.And(gormz.Eq("b", 2))

// base لم يتغير
```

### `AndGroup(children...)`

```go
q := gormz.Qb().
    And(gormz.Eq("a", 1)).
    AndGroup(
        gormz.Eq("b", 2),
        gormz.Eq("c", 3),
    )
// a = 1 AND (b = 2 AND c = 3)
```

### `OrGroup(children...)`

```go
q := gormz.Qb().
    And(gormz.Eq("a", 1)).
    OrGroup(
        gormz.Eq("b", 2),
        gormz.Eq("c", 3),
    )
// a = 1 OR (b = 2 OR c = 3)
```

### `ToSQL()`

```go
sql, args := q.ToSQL()
fmt.Println(sql)
```

---

## 🎯 مثال واقعي

### Search API

```go
func SearchUsers(filters SearchFilters) ([]User, error) {
    q := gormz.New[User]().Filter("active", true)

    // بحث نصي (OR على عدة أعمدة)
    if filters.Query != "" {
        searchQ := gormz.QOr(
            gormz.Contains("name", filters.Query),
            gormz.Contains("email", filters.Query),
            gormz.Contains("phone", filters.Query),
        )
        q = q.Q(searchQ)
    }

    // فلترة حسب الدور
    if len(filters.Roles) > 0 {
        q = q.Filter("role__in", filters.Roles)
    }

    // فلترة حسب العمر
    if filters.MinAge > 0 {
        q = q.Filter("age__gte", filters.MinAge)
    }
    if filters.MaxAge > 0 {
        q = q.Filter("age__lte", filters.MaxAge)
    }

    // حالة خاصة: admin أو verified
    if filters.SpecialOnly {
        specialQ := gormz.QOr(
            gormz.Eq("role", "admin"),
            gormz.Eq("verified", true),
        )
        q = q.Q(specialQ)
    }

    return q.OrderBy("-created_at").All()
}
```

---

## ⚠️ Performance

- `OR` مع عدة أعمدة → لا يستخدم الفهارس جيدًا
- `NOT` قد يكون بطيئًا
- استخدم `IN` بدل `OR` عند الإمكان:

```go
// ❌ بطيء
q := gormz.QOr(
    gormz.Eq("status", "active"),
    gormz.Eq("status", "pending"),
    gormz.Eq("status", "verified"),
)

// ✅ أسرع
q := gormz.In("status", []any{"active", "pending", "verified"})
```

---

## 📝 Best Practices

### ✅ Do

- استخدم `IN` بدل `OR` متكرر على نفس العمود
- استخدم `Q` للشروط المترابطة
- اختبر SQL النهائي بـ `ToSQL()`

### ❌ Don't

- لا تخلط `And`/`Or` بدون grouping
- لا تنسَ parenthesization
- لا تستخدم `NOT` مع فهارس