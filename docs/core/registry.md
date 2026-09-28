# Registry — سجل الموديلات

> تسجيل QuerySets لتجنّب import cycles.

---

## 🎯 لماذا Registry؟

### المشكلة: Import Cycle

```
pkg/user   →  import  →  pkg/order
pkg/order  →  import  →  pkg/user  ❌ cycle!
```

### الحل: Registry

```go
// في pkg/user
var Users = gormz.Register[User]("user")

// في pkg/order
q, _ := gormz.Lookup[User]("user")  // ← لا import!
```

---

## 📖 API

### `Register[T](name)`

```go
var Users = gormz.Register[User]("user")
```

**يرفع panic** عند التكرار.

### `TryRegister[T](name)`

```go
q, err := gormz.TryRegister[User]("user")
if err != nil {
    log.Fatal(err)
}
```

### `Lookup[T](name)`

```go
q, ok := gormz.Lookup[User]("user")
if !ok {
    log.Fatal("user not registered")
}
```

### `MustLookup[T](name)`

```go
q := gormz.MustLookup[User]("user")  // panic if not found
```

### Helpers

| Function | الوصف |
|----------|-------|
| `Has(name)` | هل مسجّل؟ |
| `Unregister(name)` | إزالة |
| `RegisteredNames()` | كل الأسماء |
| `RegisteredCount()` | العدد |
| `ClearRegistry()` | مسح كل شيء (للاختبارات) |

---

## 💡 أمثلة

### مثال 1: Basic

```go
// apps/user/objects.go
package user

import "github.com/light-tech-dev/gormz"

var Users = gormz.Register[User]("user")

// apps/order/service.go
package order

import "github.com/light-tech-dev/gormz"

func FindUserOrders(userID uint) ([]Order, error) {
    // احصل على QuerySet من Registry
    users := gormz.MustLookup[User]("user")

    // استخدمه
    user, err := users.Get(userID)
    if err != nil {
        return nil, err
    }

    return gormz.New[Order]().Filter("user_id", user.ID).All()
}
```

### مثال 2: Cross-Package

```
pkg/
├── user/
│   └── objects.go    (يُسجّل)
├── order/
│   └── service.go    (يستخدم)
└── report/
    └── service.go    (يستخدم)
```

```go
// pkg/user/objects.go
var Users = gormz.Register[User]("user")

// pkg/order/service.go
func CreateOrder(userID uint, total float64) (*Order, error) {
    // احصل على Users من registry
    users := gormz.MustLookup[User]("user")
    user, err := users.Get(userID)
    if err != nil {
        return nil, err
    }

    order := &Order{UserID: user.ID, Total: total}
    return order, gormz.New[Order]().Create(order)
}

// pkg/report/service.go
func TopUsers(limit int) ([]User, error) {
    users := gormz.MustLookup[User]("user")
    return users.OrderBy("-created_at").Limit(limit).All()
}
```

### مثال 3: Init Pattern

```go
// apps/user/manifest.go
package user

import "github.com/light-tech-dev/gormz"

var Objects *gormz.QuerySet[User]

func init() {
    Objects = gormz.Register[User]("user")
}

// استخدام لاحقًا
func Find(email string) (*User, error) {
    return Objects.Find("email", email)
}
```

### مثال 4: Testing

```go
func TestSomething(t *testing.T) {
    setupTestDB(t)
    gormz.ClearRegistry()  // ← نظّف قبل

    // سجّل
    Users := gormz.Register[User]("user")

    // استخدم
    Users.Create(&User{Name: "Ali"})

    // تأكد
    assert.Equal(t, 1, gormz.RegisteredCount())
}
```

---

## ⚠️ تحذيرات

### 1. Duplicate Registration

```go
gormz.Register[User]("user")  // ok
gormz.Register[User]("user")  // panic!
```

**الحل**: استخدم `TryRegister` أو `Has()`:

```go
if !gormz.Has("user") {
    gormz.Register[User]("user")
}
```

### 2. Type Mismatch

```go
gormz.Register[User]("user")

// ❌ خطأ نوع
q, ok := gormz.Lookup[Order]("user")
// q = nil, ok = false

// ✅ صحيح
q, ok := gormz.Lookup[User]("user")
```

### 3. Thread Safety

Registry **thread-safe**:

```go
// آمن من goroutines متعددة
go gormz.Register[User]("user")
go gormz.Register[Order]("order")
```

### 4. Global State

Registry **global**. للاختبارات:

```go
func TestMain(m *testing.M) {
    // Setup
    code := m.Run()

    // Cleanup
    gormz.ClearRegistry()
    os.Exit(code)
}
```

---

## 🎯 Best Practices

### ✅ Do

- اسم الموديل بصيغة `singular` (`"user"` بدل `"users"`)
- استخدم `init()` للتسجيل التلقائي
- `ClearRegistry` في الاختبارات
- `Has` قبل `Register` إن أمكن

### ❌ Don't

- لا تسجّل نفس الاسم مرتين
- لا تخلط الأنواع
- لا تعتمد على ترتيب التسجيل