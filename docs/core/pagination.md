# Pagination — التصفح

> تصفح النتائج بكفاءة.

---

## 📖 API

### `Paginate(page, perPage)`

```go
page, err := gormz.New[User]().
    Filter("active", true).
    OrderBy("-created_at").
    Paginate(1, 20)
```

**الحدود**:
- `page` >= 1
- `perPage` 1-1000
- defaults: 1, 20

---

## 💡 مثال أساسي

```go
result, err := gormz.New[User]().
    Filter("active", true).
    Paginate(1, 20)

if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Items)      // []User (20 عنصر)
fmt.Println(result.Total)      // 150 (إجمالي)
fmt.Println(result.Page)       // 1
fmt.Println(result.PerPage)    // 20
fmt.Println(result.TotalPages) // 8
fmt.Println(result.HasNext)    // true
fmt.Println(result.HasPrev)    // false
fmt.Println(result.IsEmpty())  // false
fmt.Println(result.Len())      // 20
```

---

## 🔧 Methods على `PaginatedResult`

| Method | الوصف |
|--------|-------|
| `IsEmpty()` | هل فارغ؟ |
| `Len()` | عدد العناصر |
| `First()` | أول عنصر + ok |
| `Last()` | آخر عنصر + ok |
| `ForEach(fn)` | تكرار |

### `First()` / `Last()`

```go
first, ok := result.First()
if ok {
    fmt.Println(first.Name)
}
```

---

## 📖 Helpers

### `MapPage`

```go
names := advanced.MapPage(result, func(u User) string {
    return u.Name
})
// names = []string
```

### `FilterPage`

```go
adults := advanced.FilterPage(result, func(u User) bool {
    return u.Age >= 18
})
```

---

## 🌐 REST API Integration

### Handler

```go
func ListUsers(c *fiber.Ctx) error {
    page, _ := strconv.Atoi(c.Query("page", "1"))
    perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

    result, err := gormz.New[User]().
        Filter("active", true).
        OrderBy("-created_at").
        Paginate(page, perPage)

    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{
        "data": result.Items,
        "meta": fiber.Map{
            "page":        result.Page,
            "per_page":    result.PerPage,
            "total":       result.Total,
            "total_pages": result.TotalPages,
            "has_next":    result.HasNext,
            "has_prev":    result.HasPrev,
        },
    })
}
```

**Response**:
```json
{
    "data": [...],
    "meta": {
        "page": 1,
        "per_page": 20,
        "total": 150,
        "total_pages": 8,
        "has_next": true,
        "has_prev": false
    }
}
```

---

## 📊 Performance

### Count Optimization

`Paginate` يستدعي `Count()` أولًا. **الحل**:

#### 1. CountApproximate (للمشاريع الضخمة)

```go
// PostgreSQL
db.Raw("SELECT reltuples FROM pg_class WHERE relname = ?", "users")
```

#### 2. Keyset Pagination (Cursor)

بدلًا من offset:

```go
// بدلًا من
.Offset(1000).Limit(20)

// استخدم
.Where("id > ?", lastID).OrderBy("id").Limit(20)
```

أسرع بكثير للأرقام الكبيرة.

---

## ⚠️ تحذيرات

### 1. Offset الكبير

`.Offset(100000)` بطيء. استخدم cursor pagination.

### 2. perPage الكبير

`.Paginate(1, 10000)` خطر:
- Memory
- Timeout
- DB load

**الحد الأقصى**: 1000 (مُطبّق تلقائيًا).

### 3. Count مع Conditions

`Count()` مع فلاتر معقدة بطيء. استخدم:
- Cache
- Approximate count
- Cursor pagination

---

## 🎯 مثال كامل

### Multi-Filter + Pagination

```go
type UserFilter struct {
    Query   string
    Status  string
    MinAge  int
    Page    int
    PerPage int
}

func (s *Service) List(ctx context.Context, f UserFilter) (*advanced.PaginatedResult[User], error) {
    q := gormz.FromContext[User](ctx).Filter("active", true)

    if f.Query != "" {
        q = q.Filter("name__icontains", f.Query)
    }
    if f.Status != "" {
        q = q.Filter("status", f.Status)
    }
    if f.MinAge > 0 {
        q = q.Filter("age__gte", f.MinAge)
    }

    return q.
        OrderBy("-created_at").
        Paginate(f.Page, f.PerPage)
}
```

---

## 📝 Best Practices

### ✅ Do

- استخدم `perPage` = 20-50
- أضف فهارس على أعمدة `OrderBy`
- استخدم cursor pagination للضخم
- Cache `Total` إذا ثابت

### ❌ Don't

- لا تستخدم `perPage > 1000`
- لا تستخدم `Offset > 10000`
- لا تتجاهل `Count()` cost