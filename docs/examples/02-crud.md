# Example 2: CRUD كامل

> كل عمليات CRUD مع gormz.

---

## 📦 Setup

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/light-tech-dev/gormz"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

type User struct {
    ID        uint           `gorm:"primaryKey"`
    Name      string         `gorm:"size:255;not null"`
    Email     string         `gorm:"uniqueIndex;size:255"`
    Age       int
    Active    bool           `gorm:"default:true"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

func main() {
    // Setup
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    gormz.SetDB(db)
    gormz.MustMigrate[User]()

    fmt.Println("✅ DB ready")

    // Run examples
    createExample()
    readExample()
    updateExample()
    deleteExample()
}
```

---

## 🎯 1. Create

```go
func createExample() {
    fmt.Println("\n=== CREATE ===")

    // إنشاء واحد
    user := &User{
        Name:   "علي",
        Email:  "ali@test.com",
        Age:    30,
        Active: true,
    }
    if err := gormz.New[User]().Create(user); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created user ID: %d\n", user.ID)
    // Output: Created user ID: 1

    // إنشاء كثير
    users := []User{
        {Name: "سارة", Email: "sara@test.com", Age: 25},
        {Name: "عمر", Email: "omar@test.com", Age: 35},
    }
    if err := gormz.New[User]().CreateMany(users); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created %d users\n", len(users))
    // Output: Created 2 users
}
```

---

## 📖 2. Read

```go
func readExample() {
    fmt.Println("\n=== READ ===")

    // الحصول بالـ ID
    user, err := gormz.New[User]().Get(1)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Got: %s (%s)\n", user.Name, user.Email)
    // Output: Got: علي (ali@test.com)

    // الحصول بشروط
    sara, _ := gormz.New[User]().Find("email", "sara@test.com")
    fmt.Printf("Found: %s\n", sara.Name)
    // Output: Found: سارة

    // كل السجلات
    all, _ := gormz.New[User]().All()
    fmt.Printf("Total: %d users\n", len(all))
    // Output: Total: 3 users

    // مع فلتر
    adults, _ := gormz.New[User]().
        Filter("age__gte", 30).
        OrderBy("name").
        All()
    fmt.Printf("Adults (30+): %d\n", len(adults))
    // Output: Adults (30+): 2

    // Count
    count, _ := gormz.New[User]().Count()
    fmt.Printf("Count: %d\n", count)
    // Output: Count: 3

    // Exists
    exists, _ := gormz.New[User]().Filter("email", "ali@test.com").Exists()
    fmt.Printf("Ali exists: %v\n", exists)
    // Output: Ali exists: true
}
```

---

## ✏️ 3. Update

```go
func updateExample() {
    fmt.Println("\n=== UPDATE ===")

    // تحديث حقل واحد
    if err := gormz.New[User]().Update(1, "age", 31); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Updated age")

    // التحقق
    user, _ := gormz.New[User]().Get(1)
    fmt.Printf("New age: %d\n", user.Age)
    // Output: New age: 31

    // تحديث متعدد
    affected, err := gormz.New[User]().
        Filter("age__lt", 30).
        UpdateMany(map[string]any{
            "active": false,
        })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Deactivated %d users\n", affected)
    // Output: Deactivated 1 users

    // Save (upsert)
    user.Name = "علي المحدث"
    if err := gormz.New[User]().Save(user); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Saved user")
}
```

---

## 🗑️ 4. Delete

```go
func deleteExample() {
    fmt.Println("\n=== DELETE ===")

    // Soft delete
    if err := gormz.New[User]().Delete(3); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Deleted user 3 (soft)")

    // لم يعد مرئيًا
    all, _ := gormz.New[User]().All()
    fmt.Printf("Visible: %d\n", len(all))
    // Output: Visible: 2

    // لكن موجود في DB
    all, _ = gormz.New[User]().WithDeleted().All()
    fmt.Printf("With deleted: %d\n", len(all))
    // Output: With deleted: 3

    // استعادة
    if err := gormz.New[User]().Restore(3); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Restored user 3")

    // Hard delete
    if err := gormz.New[User]().HardDelete(2); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Hard deleted user 2")

    // تأكيد
    _, err := gormz.New[User]().WithDeleted().Get(2)
    fmt.Printf("User 2 gone: %v\n", gormz.IsNotFound(err))
    // Output: User 2 gone: true
}
```

---

## 📊 Output المتوقع

```
✅ DB ready

=== CREATE ===
Created user ID: 1
Created 2 users

=== READ ===
Got: علي (ali@test.com)
Found: سارة
Total: 3 users
Adults (30+): 2
Count: 3
Ali exists: true

=== UPDATE ===
Updated age
New age: 31
Deactivated 1 users
Saved user

=== DELETE ===
Deleted user 3 (soft)
Visible: 2
With deleted: 3
Restored user 3
Hard deleted user 2
User 2 gone: true
```

---

## 🎯 النقاط الرئيسية

1. ✅ **Soft Delete** تلقائي (مع `gorm.DeletedAt`)
2. ✅ **WithDeleted()** للمحذوفين
3. ✅ **Restore()** للاستعادة
4. ✅ **HardDelete()** للحذف النهائي
5. ✅ **UpdateMany()** يتطلب conditions

---

## 🔍 التفاصيل

### `Create` vs `CreateMany`

- `Create`: سجل واحد
- `CreateMany`: عدة سجلات (أسرع)

### `Get` vs `Find`

- `Get(1)`: بالـ ID
- `Find("email", "x")`: بأي حقل

### `First` vs `FirstOrNil`

- `First`: يرجّع `ErrNotFound`
- `FirstOrNil`: يرجّع `(nil, nil)`