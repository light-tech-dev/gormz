
---

## 🔟 `docs/guides/testing.md`

````markdown
# Testing — اختبارات gormz

> كيف تختبر تطبيقك مع gormz.

---

## 🎯 Setup

### In-Memory SQLite

```go
func setupTestDB(t *testing.T) *gormz.Instance {
    t.Helper()

    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    require.NoError(t, err)

    require.NoError(t, db.AutoMigrate(&User{}, &Order{}))

    inst := gormz.NewInstance(db)
    gormz.SetDB(db)
    gormz.ClearRegistry()

    return inst
}
```

**الفائدة**:
- ⚡ سريع جدًا
- 🧹 نظيف لكل test
- 🚫 لا يحتاج DB خارجي

---

## 📖 Helper Functions

### Fixtures

```go
func createTestUsers(t *testing.T, count int) []User {
    t.Helper()

    users := make([]User, count)
    for i := 0; i < count; i++ {
        users[i] = User{
            Name:  fmt.Sprintf("User %d", i),
            Email: fmt.Sprintf("user%d@test.com", i),
            Age:   20 + (i % 30),
        }
    }

    require.NoError(t, gormz.New[User]().CreateMany(users))
    return users
}
```

### Cleanup

```go
func cleanupTest(t *testing.T) {
    t.Cleanup(func() {
        gormz.ClearRegistry()
    })
}
```

---

## 💡 أمثلة

### مثال 1: CRUD Test

```go
func TestUserCRUD(t *testing.T) {
    setupTestDB(t)

    // Create
    user := &User{Name: "Ali", Email: "ali@test.com"}
    err := gormz.New[User]().Create(user)
    require.NoError(t, err)
    assert.NotZero(t, user.ID)

    // Read
    got, err := gormz.New[User]().Get(user.ID)
    require.NoError(t, err)
    assert.Equal(t, "Ali", got.Name)

    // Update
    err = gormz.New[User]().Update(user.ID, "name", "Ali Updated")
    require.NoError(t, err)

    // Verify
    got, _ = gormz.New[User]().Get(user.ID)
    assert.Equal(t, "Ali Updated", got.Name)

    // Delete
    err = gormz.New[User]().Delete(user.ID)
    require.NoError(t, err)

    // Verify
    _, err = gormz.New[User]().Get(user.ID)
    assert.True(t, gormz.IsNotFound(err))
}
```

### مثال 2: Filter Test

```go
func TestUserFilters(t *testing.T) {
    setupTestDB(t)

    users := []User{
        {Name: "Ali", Age: 30, Active: true},
        {Name: "Sara", Age: 25, Active: true},
        {Name: "Omar", Age: 17, Active: false},
    }
    require.NoError(t, gormz.New[User]().CreateMany(users))

    t.Run("filter active", func(t *testing.T) {
        result, _ := gormz.New[User]().Filter("active", true).All()
        assert.Len(t, result, 2)
    })

    t.Run("filter age >= 18", func(t *testing.T) {
        result, _ := gormz.New[User]().Filter("age__gte", 18).All()
        assert.Len(t, result, 2)
    })

    t.Run("compound", func(t *testing.T) {
        result, _ := gormz.New[User]().
            Filter("active", true).
            Filter("age__gte", 18).
            All()
        assert.Len(t, result, 2)
    })
}
```

### مثال 3: Table-Driven

```go
func TestAgeFilter(t *testing.T) {
    setupTestDB(t)
    createTestUsers(t, 10)

    tests := []struct {
        name     string
        filter   string
        value    int
        expected int
    }{
        {"gt 25", "age__gt", 25, 5},
        {"gte 25", "age__gte", 25, 5},
        {"lt 25", "age__lt", 25, 5},
        {"lte 25", "age__lte", 25, 5},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := gormz.New[User]().
                Filter(tt.filter, tt.value).
                All()
            require.NoError(t, err)
            assert.Len(t, result, tt.expected)
        })
    }
}
```

### مثال 4: Transaction Test

```go
func TestTransaction_Rollback(t *testing.T) {
    setupTestDB(t)
    ctx := context.Background()

    err := advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
        func(tx *advanced.Tx) error {
            tx.Query[User]().Create(&User{Name: "Ali"})
            return errors.New("intentional")
        })
    assert.Error(t, err)

    count, _ := gormz.New[User]().Count()
    assert.Equal(t, int64(0), count)
}
```

### مثال 5: Mock Time

```go
func TestUser_TimeFilter(t *testing.T) {
    setupTestDB(t)

    // بيانات مع تواريخ محددة
    past := User{Name: "Old", CreatedAt: time.Now().AddDate(0, -1, 0)}
    recent := User{Name: "New", CreatedAt: time.Now()}
    gormz.New[User]().CreateMany([]User{past, recent})

    // فلترة
    result, _ := gormz.New[User]().
        Filter("created_at__gt", time.Now().AddDate(0, 0, -7)).
        All()
    assert.Len(t, result, 1)
    assert.Equal(t, "New", result[0].Name)
}
```

### مثال 6: Benchmark

```go
func BenchmarkCreate(b *testing.B) {
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    db.AutoMigrate(&User{})
    gormz.SetDB(db)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        user := &User{Name: "Ali", Email: fmt.Sprintf("ali%d@test.com", i)}
        gormz.New[User]().Create(user)
    }
}
```

### مثال 7: Testcontainers (PostgreSQL)

```go
import "github.com/testcontainers/testcontainers-go/modules/postgres"

func setupPostgres(t *testing.T) *gormz.Instance {
    t.Helper()

    ctx := context.Background()
    container, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:16-alpine"),
        postgres.WithDatabase("test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    require.NoError(t, err)

    t.Cleanup(func() {
        container.Terminate(ctx)
    })

    dsn, _ := container.ConnectionString(ctx, "sslmode=disable")
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    require.NoError(t, err)
    require.NoError(t, db.AutoMigrate(&User{}))

    return gormz.NewInstance(db)
}
```

---

## 🎨 Helpers Library

**اقتراح**: أنشئ `testutil` package:

```go
// testutil/db.go
package testutil

func InMemoryDB(t *testing.T, models ...any) *gormz.Instance {
    t.Helper()
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    db.AutoMigrate(models...)
    inst := gormz.NewInstance(db)
    gormz.SetDB(db)
    return inst
}

func Fixture[T any](t *testing.T, count int, fn func(int) T) []T {
    t.Helper()
    items := make([]T, count)
    for i := 0; i < count; i++ {
        items[i] = fn(i)
    }
    require.NoError(t, gormz.New[T]().CreateMany(items))
    return items
}
```

**استخدام**:
```go
func TestFoo(t *testing.T) {
    testutil.InMemoryDB(t, &User{}, &Order{})

    users := testutil.Fixture[User](t, 10, func(i int) User {
        return User{Name: fmt.Sprintf("User %d", i)}
    })
    // ...
}
```

---

## 📝 Best Practices

### ✅ Do

- استخدم `:memory:` للسرعة
- استخدم `t.Helper()` في helpers
- استخدم `t.Cleanup()`
- Test table-driven

### ❌ Don't

- لا تستخدم نفس DB بين tests
- لا تعتمد على ترتيب التنفيذ
- لا تترك DB مفتوح