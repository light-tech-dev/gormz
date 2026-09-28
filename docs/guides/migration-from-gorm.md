# Migration من GORM إلى gormz

> دليل كامل للانتقال التدريجي.

---

## 🎯 لماذا gormz؟

| الميزة | GORM | gormz |
|--------|------|-------|
| Type Safety | ❌ `interface{}` | ✅ Generics |
| Lookups | ❌ String | ✅ `__gt`, `__in` |
| Immutable | ❌ Mutable | ✅ Immutable |
| Pagination | ❌ Manual | ✅ Built-in |
| Field Validation | ❌ | ✅ Auto |
| Multi-DB | ⚠️ Manual | ✅ `Instance` |
| Registry | ❌ | ✅ |

---

## 📊 جدول التحويل السريع

### الاستعلامات

| GORM | gormz |
|------|-------|
| `db.Where("age > ?", 18).Find(&users)` | `gormz.New[User]().Filter("age__gt", 18).All()` |
| `db.Where("name = ?", "Ali").First(&user)` | `gormz.New[User]().Find("name", "Ali")` |
| `db.Where("id = ?", 1).First(&user)` | `gormz.New[User]().Get(1)` |
| `db.Where("id IN ?", ids).Find(&users)` | `gormz.New[User]().Filter("id__in", ids).All()` |
| `db.Where("name LIKE ?", "%Ali%").Find(&users)` | `gormz.New[User]().Filter("name__contains", "Ali").All()` |
| `db.Where("deleted_at IS NULL").Find(&users)` | `gormz.New[User]().Filter("deleted_at__isnull", true).All()` |
| `db.Order("name").Find(&users)` | `gormz.New[User]().OrderBy("name").All()` |
| `db.Order("name DESC").Find(&users)` | `gormz.New[User]().OrderBy("-name").All()` |
| `db.Limit(10).Offset(20).Find(&users)` | `gormz.New[User]().Limit(10).Offset(20).All()` |
| `db.Count(&count)` | `gormz.New[User]().Count()` |
| `db.Select("id", "name").Find(&users)` | `gormz.New[User]().Select("id", "name").All()` |
| `db.Preload("Orders").Find(&users)` | `gormz.New[User]().Preload("Orders").All()` |
| `db.Unscoped().Find(&users)` | `gormz.New[User]().WithDeleted().All()` |

### الكتابة

| GORM | gormz |
|------|-------|
| `db.Create(&user)` | `gormz.New[User]().Create(&user)` |
| `db.Create(&users)` | `gormz.New[User]().CreateMany(users)` |
| `db.Save(&user)` | `gormz.New[User]().Save(&user)` |
| `db.Model(&user).Update("name", "x")` | `gormz.New[User]().Update(id, "name", "x")` |
| `db.Where("active = ?", false).Delete(&User{})` | `gormz.New[User]().Filter("active", false).DeleteMany()` |
| `db.Unscoped().Delete(&user, 1)` | `gormz.New[User]().HardDelete(1)` |
| `db.Model(&user).Update("deleted_at", nil)` | `gormz.New[User]().Restore(1)` |

### Transactions

| GORM | gormz |
|------|-------|
| `db.Transaction(func(tx) { ... })` | `gormz.Transaction(ctx, func(tx) { ... })` |
| `tx := db.Begin(); ...; tx.Commit()` | `tx, _ := advanced.Begin(ctx, cfg); tx.Commit()` |
| `tx.Rollback()` | `tx.Rollback()` |
| Retry | `advanced.WithTransaction(ctx, cfg, fn)` |

### Q Builder

| GORM | gormz |
|------|-------|
| `db.Where("a = ? OR b = ?", 1, 2)` | `gormz.QOr(gormz.Eq("a", 1), gormz.Eq("b", 2))` |
| `db.Where("a = ? AND b = ?", 1, 2)` | `gormz.QAnd(gormz.Eq("a", 1), gormz.Eq("b", 2))` |
| `db.Where("NOT (a = ?)", 1)` | `gormz.Not(gormz.Eq("a", 1))` |
| Complex nested | `gormz.Qb().And(gormz.QOr(...), ...)` |

### Aggregations

| GORM | gormz |
|------|-------|
| `db.Model(&Order{}).Select("SUM(total)").Scan(&sum)` | `gormz.New[Order]().Sum("total")` |
| `db.Model(&Order{}).Select("AVG(total)").Scan(&avg)` | `gormz.New[Order]().Avg("total")` |
| `db.Model(&Order{}).Select("MIN(total)").Scan(&min)` | `gormz.New[Order]().Min("total")` |
| `db.Model(&Order{}).Select("MAX(total)").Scan(&max)` | `gormz.New[Order]().Max("total")` |
| `db.Model(&Order{}).Group("user_id").Select("user_id, COUNT(*)").Scan(&stats)` | `advanced.GroupBy[Order]("user_id").Count("*", "cnt").ScanInto(&stats)` |

### Joins

| GORM | gormz |
|------|-------|
| `db.Joins("JOIN orders ON orders.user_id = users.id").Find(&users)` | `advanced.WithJoins[User](gormz.New[User]()).Inner("orders", "orders.user_id = users.id").All()` |

### Bulk

| GORM | gormz |
|------|-------|
| `db.CreateInBatches(users, 1000)` | `advanced.BulkInsert[User](ctx, users, cfg)` |
| `db.Clauses(clause.OnConflict{...}).Create(&users)` | `advanced.BulkUpsert[User](ctx, users, cfg)` |

---

## 🔄 Migration Steps

### المرحلة 1: التثبيت

```bash
go get github.com/light-tech-dev/gormz
```

### المرحلة 2: ربط gormz

**قبل**:
```go
func main() {
    db, _ := gorm.Open(...)
    // استخدم db مباشرة
}
```

**بعد**:
```go
func main() {
    db, _ := gorm.Open(...)
    gormz.SetDB(db)  // ← أضف هذا
}
```

### المرحلة 3: Migration تدريجي

**استراتيجية Strangler**:

```go
// 1. ابدأ بـ CRUD بسيط
// القديم:
var users []User
db.Where("age > ?", 18).Find(&users)

// الجديد:
users, _ := gormz.New[User]().Filter("age__gt", 18).All()

// 2. انتقل للاستعلامات المعقدة
// القديم:
db.Model(&Order{}).
    Where("status = ?", "paid").
    Where("total > ?", 1000).
    Group("user_id").
    Select("user_id, SUM(total) as total").
    Scan(&stats)

// الجديد:
advanced.GroupBy[Order]("user_id").
    Where("status = ?", "paid").
    Where("total > ?", 1000).
    Sum("total", "total").
    ScanInto(&stats)

// 3. أخيرًا، Transactions
// القديم:
err := db.Transaction(func(tx *gorm.DB) error {
    // ...
})

// الجديد:
err := advanced.WithTransaction(ctx, cfg, func(tx *advanced.Tx) error {
    // ...
})
```

### المرحلة 4: إزالة GORM المباشر

بعد أن تصبح 90% من الكود يستخدم gormz، احذف الاستخدام المباشر.

---

## 💡 أمثلة تفصيلية

### مثال 1: CRUD كامل

**GORM**:
```go
// Create
user := &User{Name: "Ali", Email: "ali@test.com"}
db.Create(user)

// Read
var got User
db.Where("email = ?", "ali@test.com").First(&got)

// Update
db.Model(&user).Update("age", 30)

// Delete
db.Delete(&user)

// List
var users []User
db.Where("active = ?", true).
    Order("created_at DESC").
    Limit(10).
    Find(&users)

// Count
var count int64
db.Model(&User{}).Where("active = ?", true).Count(&count)
```

**gormz**:
```go
// Create
user := &User{Name: "Ali", Email: "ali@test.com"}
gormz.New[User]().Create(user)

// Read
got, _ := gormz.New[User]().Find("email", "ali@test.com")

// Update
gormz.New[User]().Update(user.ID, "age", 30)

// Delete
gormz.New[User]().Delete(user.ID)

// List
users, _ := gormz.New[User]().
    Filter("active", true).
    OrderBy("-created_at").
    Limit(10).
    All()

// Count
count, _ := gormz.New[User]().Filter("active", true).Count()
```

### مثال 2: Complex Query

**GORM**:
```go
var results []User
db.Where("(status = ? OR status = ?)", "active", "pending").
    Where("age > ?", 18).
    Where("name LIKE ?", "%Ali%").
    Where("deleted_at IS NULL").
    Preload("Orders").
    Order("created_at DESC").
    Limit(20).
    Offset(0).
    Find(&results)
```

**gormz**:
```go
q := gormz.QOr(
    gormz.Eq("status", "active"),
    gormz.Eq("status", "pending"),
)

results, _ := gormz.New[User]().
    Q(q).
    Filter("age__gt", 18).
    Filter("name__contains", "Ali").
    Filter("deleted_at__isnull", true).
    Preload("Orders").
    OrderBy("-created_at").
    Limit(20).
    Offset(0).
    All()
```

### مثال 3: Transfer في Transaction

**GORM**:
```go
err := db.Transaction(func(tx *gorm.DB) error {
    var from, to Account
    if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
        First(&from, fromID).Error; err != nil {
        return err
    }
    if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
        First(&to, toID).Error; err != nil {
        return err
    }

    if from.Balance < amount {
        return errors.New("insufficient")
    }

    from.Balance -= amount
    to.Balance += amount

    if err := tx.Save(&from).Error; err != nil {
        return err
    }
    return tx.Save(&to).Error
})
```

**gormz**:
```go
err := advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
    func(tx *advanced.Tx) error {
        from, err := advanced.WithLock(tx.Query[Account]()).ForUpdate().Get(fromID)
        if err != nil {
            return err
        }
        to, err := advanced.WithLock(tx.Query[Account]()).ForUpdate().Get(toID)
        if err != nil {
            return err
        }

        if from.Balance < amount {
            return errors.New("insufficient")
        }

        from.Balance -= amount
        to.Balance += amount

        if err := tx.Query[Account]().Save(from); err != nil {
            return err
        }
        return tx.Query[Account]().Save(to)
    })
```

### مثال 4: Bulk Insert

**GORM**:
```go
for i := 0; i < len(users); i += 1000 {
    end := i + 1000
    if end > len(users) {
        end = len(users)
    }
    if err := db.Create(users[i:end]).Error; err != nil {
        return err
    }
}
```

**gormz**:
```go
err := advanced.BulkInsert[User](ctx, users, advanced.BulkConfig{
    BatchSize: 1000,
})
```

---

## ⚠️ الفروقات السلوكية

### 1. Default Behavior

| الحالة | GORM | gormz |
|--------|------|-------|
| Empty query | كل السجلات | كل السجلات |
| `Delete` بدون where | ❌ panic/error | ❌ error |
| `UpdateMany` بدون where | ❌ error | ❌ error |

### 2. Errors

**GORM**:
```go
err := db.Where("id = ?", 999).First(&user).Error
// err = gorm.ErrRecordNotFound
```

**gormz**:
```go
user, err := gormz.New[User]().Get(999)
// user = nil, err = gorm.ErrRecordNotFound
// استخدم gormz.IsNotFound(err)
if gormz.IsNotFound(err) {
    // ...
}
```

### 3. Hooks

**GORM**: `BeforeSave`, `BeforeCreate`, ...
**gormz**: **نفس الـ hooks** (يعيد تصديرها) ✅

### 4. Soft Delete

**GORM**: `gorm.DeletedAt` في الموديل
**gormz**: **نفس الشيء** ✅

### 5. Preload

**GORM**: `db.Preload("Orders").Find(&users)`
**gormz**: `gormz.New[User]().Preload("Orders").All()` ✅

---

## 📋 Migration Checklist

- [ ] تثبيت gormz
- [ ] إضافة `gormz.SetDB(db)` في main
- [ ] تحويل CRUD الأساسي
- [ ] تحويل الاستعلامات البسيطة
- [ ] تحويل الاستعلامات المعقدة
- [ ] تحويل Transactions
- [ ] تحويل Bulk operations
- [ ] إزالة GORM المباشر (تدريجيًا)
- [ ] إزالة `gorm.DB` من الخدمات
- [ ] تحديث الاختبارات

---

## 🎯 نصائح

### ✅ Do

- Migration تدريجي (Strangler)
- ابدأ بـ CRUD
- اختبر بعد كل مرحلة
- استخدم `QuerySet.ToSQL()` للمقارنة

### ❌ Don't

- لا تحوّل كل شيء في مرة واحدة
- لا تحذف GORM قبل الاختبار
- لا تخلط الأنماط في نفس الملف