# Lookups — المرجع الكامل

> modern-style filters: `field__lookup`.

---

## 📖 الصيغة

```
field__lookup = value
```

**أمثلة**:
```go
.Filter("age__gt", 18)     // field=age, lookup=gt
.Filter("name", "Ali")     // field=name, no lookup (=)
```

---

## 🔍 Lookups المدعومة

### Equality

| Lookup | SQL | مثال |
|--------|-----|------|
| (بدون) | `=` | `.Filter("name", "Ali")` |
| `ne` | `!=` | `.Filter("status__ne", "deleted")` |

### Comparison

| Lookup | SQL | مثال |
|--------|-----|------|
| `gt` | `>` | `.Filter("age__gt", 18)` |
| `gte` | `>=` | `.Filter("age__gte", 18)` |
| `lt` | `<` | `.Filter("age__lt", 65)` |
| `lte` | `<=` | `.Filter("age__lte", 65)` |

### Text

| Lookup | SQL | مثال |
|--------|-----|------|
| `contains` | `LIKE '%x%'` | `.Filter("name__contains", "Ali")` |
| `icontains` | `LOWER LIKE` | `.Filter("name__icontains", "ali")` |
| `startswith` | `LIKE 'x%'` | `.Filter("name__startswith", "A")` |
| `istartswith` | `LOWER LIKE` | `.Filter("name__istartswith", "a")` |
| `endswith` | `LIKE '%x'` | `.Filter("email__endswith", ".com")` |
| `iendswith` | `LOWER LIKE` | `.Filter("email__iendswith", ".COM")` |

### Lists

| Lookup | SQL | مثال |
|--------|-----|------|
| `in` | `IN` | `.Filter("status__in", []string{"active", "pending"})` |
| `notin` | `NOT IN` | `.Filter("status__notin", []string{"deleted"})` |

### NULL

| Lookup | SQL | مثال |
|--------|-----|------|
| `isnull` | `IS NULL` | `.Filter("deleted_at__isnull", true)` |

### Range

| Lookup | SQL | مثال |
|--------|-----|------|
| `between` | `BETWEEN` | `.Filter("age__between", []int{18, 65})` |

### Date

| Lookup | SQL | مثال |
|--------|-----|------|
| `year` | `EXTRACT(YEAR...)` | `.Filter("created_at__year", 2025)` |
| `month` | `EXTRACT(MONTH...)` | `.Filter("created_at__month", 12)` |
| `day` | `EXTRACT(DAY...)` | `.Filter("created_at__day", 15)` |

---

## 💡 أمثلة عملية

### مثال 1: بحث

```go
// ابحث عن مستخدمين اسمهم يحتوي "Ali"
users, _ := gormz.New[User]().
    Filter("name__icontains", "ali").
    All()
```

### مثال 2: فلترة متعددة

```go
// active + age 18-65 + verified
users, _ := gormz.New[User]().
    Filter("active", true).
    Filter("age__between", []int{18, 65}).
    Filter("status", "verified").
    All()
```

### مثال 3: قوائم

```go
// من محافظات معينة
users, _ := gormz.New[User]().
    Filter("city__in", []string{"Cairo", "Alexandria", "Giza"}).
    All()
```

### مثال 4: NULL

```go
// غير محذوفين
users, _ := gormz.New[User]().
    Filter("deleted_at__isnull", true).
    All()

// محذوفين
deleted, _ := gormz.New[User]().
    Filter("deleted_at__isnull", false).
    All()
```

### مثال 5: تواريخ

```go
// مستخدمون سُجّلوا في 2025
users, _ := gormz.New[User]().
    Filter("created_at__year", 2025).
    All()

// سُجّلوا في ديسمبر
users, _ := gormz.New[User]().
    Filter("created_at__month", 12).
    All()
```

---

## ⚠️ Dialect-specific

### `year`, `month`, `day`

| DB | SQL |
|----|-----|
| PostgreSQL | `EXTRACT(YEAR FROM col)` |
| MySQL | `YEAR(col)` |
| SQLite | `CAST(strftime('%Y', col) AS INTEGER)` |

gormz يكتشف تلقائيًا.

### `icontains` performance

`LOWER(col) LIKE` لا يستخدم الفهارس. للحالات الكثيرة:
- استخدم `citext` في PostgreSQL
- استخدم `COLLATE NOCASE` في SQLite

---

## 🎯 نصائح

### ✅ Do

- استخدم `__in` بدل `OR` متعدد
- استخدم `__between` بدل `>= AND <=`
- استخدم `__isnull` بدل `IS NULL` يدويًا

### ❌ Don't

- لا تستخدم `__contains` مع بيانات ضخمة
- لا تخلط between مع gt/lt
- لا تنسَ فحص empty slices مع `__in`