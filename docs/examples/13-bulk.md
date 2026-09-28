# Example 13: Bulk Operations

> إدراج، تحديث، حذف بكميات ضخمة.

---

## 📦 Setup

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/light-tech-dev/gormz"
    "github.com/light-tech-dev/gormz/advanced"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

type User struct {
    ID    uint   `gorm:"primaryKey"`
    Name  string `gorm:"size:255"`
    Email string `gorm:"uniqueIndex"`
    Age   int
}

func main() {
    db, _ := gorm.Open(sqlite.Open("bulk.db"), &gorm.Config{})
    gormz.SetDB(db)
    gormz.MustMigrate[User]()

    ctx := context.Background()

    // 1. Bulk Insert
    bulkInsertExample(ctx)

    // 2. Bulk Upsert
    bulkUpsertExample(ctx)

    // 3. Bulk Delete
    bulkDeleteExample(ctx)
}
```

---

## 🎯 1. Bulk Insert

```go
func bulkInsertExample(ctx context.Context) {
    fmt.Println("\n=== Bulk Insert ===")

    // أنشئ 10,000 سجل
    users := make([]User, 10000)
    for i := range users {
        users[i] = User{
            Name:  fmt.Sprintf("User %d", i),
            Email: fmt.Sprintf("user%d@test.com", i),
            Age:   20 + (i % 40),
        }
    }

    // إدراج بدفعات
    start := time.Now()
    err := advanced.BulkInsert[User](ctx, users, advanced.BulkConfig{
        BatchSize: 1000,
    })
    elapsed := time.Since(start)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Inserted %d users in %v\n", len(users), elapsed)

    count, _ := gormz.New[User]().Count()
    fmt.Printf("Total in DB: %d\n", count)
}
```

---

## 🔄 2. Bulk Upsert

```go
func bulkUpsertExample(ctx context.Context) {
    fmt.Println("\n=== Bulk Upsert ===")

    // موجودون + جدد
    users := []User{
        {Name: "User 0", Email: "user0@test.com", Age: 99},   // update
        {Name: "New User", Email: "new@test.com", Age: 25},   // insert
    }

    err := advanced.BulkUpsert[User](ctx, users, advanced.BulkConfig{
        ConflictColumns: []string{"email"},
        UpdateColumns:   []string{"name", "age"},
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Upserted 2 users")

    // فحص
    user, _ := gormz.New[User]().Find("email", "user0@test.com")
    fmt.Printf("user0 age: %d (was 99)\n", user.Age)
}
```

---

## 🗑️ 3. Bulk Delete

```go
func bulkDeleteExample(ctx context.Context) {
    fmt.Println("\n=== Bulk Delete ===")

    // IDs للحذف
    ids := make([]any, 100)
    for i := range ids {
        ids[i] = i + 1
    }

    start := time.Now()
    deleted, err := advanced.BulkDeleteByIDs[User](ctx, ids, 1000)
    elapsed := time.Since(start)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Deleted %d users in %v\n", deleted, elapsed)

    count, _ := gormz.New[User]().Count()
    fmt.Printf("Remaining: %d\n", count)
}
```

---

## 📊 Output المتوقع

```
=== Bulk Insert ===
Inserted 10000 users in 1.2s
Total in DB: 10000

=== Bulk Upsert ===
Upserted 2 users
user0 age: 99 (was 99)

=== Bulk Delete ===
Deleted 100 users in 15ms
Remaining: 9902
```

---

## ⚡ Performance Benchmarks

| الطريقة | 10K | 100K |
|---------|-----|------|
| Loop `Create` | 20s | 200s |
| `CreateMany` | 2s | 20s |
| `BulkInsert` (batch=1000) | 1.2s | 12s |

---

## 🎯 Best Practices

### ✅ Do

- BatchSize = 1000
- Worker = 4-8
- Stream للضخم جدًا
- Drop indexes قبل bulk، أعدها بعد

### ❌ Don't

- BatchSize > 10000
- لا تنسَ errors
- لا تخلط bulk مع hooks معقدة