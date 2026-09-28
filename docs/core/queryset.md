# QuerySet — الدليل الكامل

> كل ما يمكنك فعله بـ `QuerySet[T]`.

---

## 📖 نظرة عامة

`QuerySet[T]` هو **builder** للاستعلامات:

```go
users, _ := gormz.New[User]().
    Filter("active", true).
    Filter("age__gte", 18).
    OrderBy("-created_at").
    Limit(10).
    All()
```

**المبادئ**:
- ✅ **Immutable** — كل method يرجّع `QuerySet` جديد
- ✅ **Thread-safe** — آمن للاستخدام المتزامن
- ✅ **Type-safe** — Generics
- ✅ **Chainable** — تسلسل طبيعي

---

## 🚀 الإنشاء

### `New[T]()`

```go
q := gormz.New[User]()
```

### `NewWith[T](instance)`

```go
app := gormz.NewInstance(db)
q := gormz.NewWith[User](app)
```

### `FromContext[T](ctx)`

```go
ctx := gormz.WithDB(context.Background(), app)
q := gormz.FromContext[User](ctx)
```

---

## 🔍 الفلاتر

### `Filter(field, value)`

```go
.Filter("name", "Ali")           // name = 'Ali'
.Filter("age__gt", 18)           // age > 18
.Filter("name__contains", "Ali") // name LIKE '%Ali%'
.Filter("id__in", []int{1,2,3})  // id IN (1,2,3)
```

**Lookups المدعومة**:
- `__gt`, `__gte`, `__lt`, `__lte`, `__ne`
- `__contains`, `__icontains`
- `__startswith`, `__istartswith`
- `__endswith`, `__iendswith`
- `__in`, `__notin`
- `__isnull`
- `__between`
- `__year`, `__month`, `__day`

### `Exclude(field, value)`

عكس Filter:

```go
.Exclude("status", "deleted")  // status != 'deleted'
```

### `Where(sql, args...)`

SQL خام:

```go
.Where("age > ? AND status = ?", 18, "active")
```

### `Q(builder)`

شروط معقدة:

```go
q := gormz.QOr(
    gormz.Eq("status", "active"),
    gormz.Eq("status", "pending"),
)
users, _ := gormz.New[User]().Q(q).All()
```

---

## 📊 الترتيب

### `OrderBy(fields...)`

```go
.OrderBy("name")              // name ASC
.OrderBy("-created_at")       // created_at DESC
.OrderBy("status", "-date")   // status ASC, date DESC
```

---

## 🎯 الحدود

### `Limit(n)`

```go
.Limit(10)  // أول 10
```

### `Offset(n)`

```go
.Offset(20)  // تخطّي 20
```

### `Paginate(page, perPage)`

```go
page, _ := gormz.New[User]().Paginate(1, 20)
// page.Items, page.Total, page.HasNext, ...
```

---

## 🎨 الأعمدة

### `Select(fields...)`

```go
.Select("id", "name", "email")
```

### `Omit(fields...)`

```go
.Omit("password", "secret")
```

---

## 🔗 العلاقات

### `Preload(relations...)`

```go
.Preload("Orders")
.Preload("Orders.Items")
```

---

## 🗑️ Soft Delete

### `WithDeleted()`

يشمل السجلات المحذوفة:

```go
users, _ := gormz.New[User]().WithDeleted().All()
```

### `OnlyDeleted()`

السجلات المحذوفة فقط:

```go
users, _ := gormz.New[User]().OnlyDeleted().All()
```

---

## 📖 Terminal Methods

### `All()`

```go
users, err := gormz.New[User]().All()
// users = []User
```

### `First()`

```go
user, err := gormz.New[User]().First()
// user = *User (nil if not found)
// err = ErrNotFound if not found
```

### `FirstOrNil()`

```go
user, err := gormz.New[User]().FirstOrNil()
// user = nil if not found
// err = nil
```

### `Last()`

```go
user, err := gormz.New[User]().Last()
```

### `Get(id)`

```go
user, err := gormz.New[User]().Get(1)
```

### `Find(field, value)`

```go
user, err := gormz.New[User]().Find("email", "ali@test.com")
```

### `Count()`

```go
count, err := gormz.New[User]().Count()
// count = int64
```

### `Exists()`

```go
exists, err := gormz.New[User]().Filter("id", 1).Exists()
// exists = bool
```

### `Pluck(field, dest)`

```go
var names []string
err := gormz.New[User]().Pluck("name", &names)
```

### `ScanInto(dest)`

```go
var results []CustomStruct
err := gormz.New[User]().ScanInto(&results)
```

---

## 💾 الكتابة

### `Create(item)`

```go
user := &User{Name: "Ali"}
err := gormz.New[User]().Create(user)
// user.ID معيّن
```

### `CreateMany(items)`

```go
users := []User{{Name: "Ali"}, {Name: "Sara"}}
err := gormz.New[User]().CreateMany(users)
```

### `Save(item)`

```go
err := gormz.New[User]().Save(user)
```

### `Update(id, field, value)`

```go
err := gormz.New[User]().Update(1, "name", "New Name")
```

### `UpdateMany(values)`

⚠️ **يتطلب conditions**:

```go
affected, err := gormz.New[User]().
    Filter("active", false).
    UpdateMany(map[string]any{"deleted": true})
```

### `Delete(id)`

```go
err := gormz.New[User]().Delete(1)  // soft delete
```

### `DeleteMany()`

⚠️ **يتطلب conditions**:

```go
affected, err := gormz.New[User]().
    Filter("active", false).
    DeleteMany()
```

### `HardDelete(id)`

```go
err := gormz.New[User]().HardDelete(1)  // حذف نهائي
```

### `Restore(id)`

```go
err := gormz.New[User]().Restore(1)  // استعادة
```

---

## 📊 Aggregations

### `Sum(field)`

```go
total, err := gormz.New[Order]().Sum("total")
```

### `Avg(field)`

```go
avg, err := gormz.New[Order]().Avg("total")
```

### `Min(field)` / `Max(field)`

```go
min, _ := gormz.New[Order]().Min("total")
max, _ := gormz.New[Order]().Max("total")
```

---

## 🔄 Immutability

**مهم**: كل method يرجّع نسخة جديدة:

```go
base := gormz.New[User]().Filter("active", true)

adults := base.Filter("age__gte", 18)
young := base.Filter("age__lt", 18)

// base لم يتغير
baseCount, _ := base.Count()      // active only
adultsCount, _ := adults.Count()  // active + adults
youngCount, _ := young.Count()    // active + young
```

---

## ⚠️ Panic vs Error

**Methods بادئتها Try** ترجع error:

```go
// Panic (للاستخدام الواثق)
q := gormz.New[User]().Filter("bad field", "x")  // panic

// Error (للمدخلات)
q, err := gormz.New[User]().TryFilter("bad field", "x")
if err != nil {
    log.Fatal(err)
}
```

| Panic Version | Error Version |
|--------------|--------------|
| `Filter` | `TryFilter` |
| `Exclude` | `TryExclude` |
| `OrderBy` | `TryOrderBy` |
| `Select` | `TrySelect` |
| `Omit` | `TryOmit` |

---

## 📖 SQL Debugging

### `ToSQL()`

```go
sql, args := gormz.New[User]().Filter("active", true).ToSQL()
fmt.Println(sql)   // SELECT * FROM users WHERE active = ?
fmt.Println(args)  // [true]
```

### `String()`

```go
fmt.Println(gormz.New[User]().Filter("active", true).String())
```

### `DryRun()`

```go
db := gormz.New[User]().Filter("active", true).DryRun()
// db.Statement.SQL.String()
```

---

## 🎯 مثال كامل

```go
// استعلام معقّد
users, err := gormz.New[User]().
    // فلترة
    Filter("active", true).
    Filter("age__gte", 18).
    Filter("status__in", []string{"verified", "premium"}).
    // Q builder
    Q(gormz.QOr(
        gormz.Eq("role", "admin"),
        gormz.Eq("role", "user"),
    )).
    // ترتيب
    OrderBy("-created_at", "name").
    // حد
    Limit(20).
    Offset(0).
    // علاقات
    Preload("Roles").
    // select
    Select("id", "name", "email", "age").
    // تنفيذ
    All()

if err != nil {
    log.Fatal(err)
}

for _, u := range users {
    fmt.Printf("%s (%d)\n", u.Name, u.Age)
}
```