# Context — السياق

> ربط DB بـ context.Context.

---

## 🎯 لماذا Context؟

1. **Request-scoped DB**: كل request له DB معيّن
2. **Multi-tenant**: DB مختلف لكل tenant
3. **Cancellation**: إلغاء عمليات DB طويلة
4. **Timeouts**: حد زمني للاستعلامات
5. **Tracing**: تتبّع الطلبات

---

## 📖 API

### `WithDB(ctx, instance)`

```go
ctx := gormz.WithDB(context.Background(), app)
```

### `WithGormDB(ctx, db)`

```go
ctx := gormz.WithGormDB(context.Background(), db)
```

### `DBFromContext(ctx)`

```go
inst, ok := gormz.DBFromContext(ctx)
```

### `FromContext[T](ctx)`

```go
q := gormz.FromContext[User](ctx)
```

---

## 💡 أمثلة

### مثال 1: HTTP Request

```go
func UserMiddleware(app *gormz.Instance) fiber.Handler {
    return func(c *fiber.Ctx) error {
        ctx := gormz.WithDB(c.UserContext(), app)
        c.SetUserContext(ctx)
        return c.Next()
    }
}

func GetUsers(c *fiber.Ctx) error {
    users, _ := gormz.FromContext[User](c.UserContext()).All()
    return c.JSON(users)
}
```

### مثال 2: Timeout

```go
func GetUsersWithTimeout() ([]User, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    return gormz.FromContext[User](ctx).All()
    // إذا تجاوز 5s → context deadline exceeded
}
```

### مثال 3: Multi-tenant

```go
func TenantMiddleware(mgr *TenantManager) fiber.Handler {
    return func(c *fiber.Ctx) error {
        tenantID := c.Get("X-Tenant-ID")

        inst, err := mgr.Get(tenantID)
        if err != nil {
            return c.Status(400).JSON(fiber.Map{"error": "invalid tenant"})
        }

        ctx := gormz.WithDB(c.UserContext(), inst)
        c.SetUserContext(ctx)
        return c.Next()
    }
}
```

### مثال 4: Tracing

```go
type TraceContext struct {
    TraceID string
    SpanID  string
}

func TracingMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        traceID := c.Get("X-Trace-ID")
        if traceID == "" {
            traceID = uuid.New().String()
        }

        ctx := context.WithValue(c.UserContext(), "trace_id", traceID)
        c.SetUserContext(ctx)
        return c.Next()
    }
}

func CreateUser(c *fiber.Ctx) error {
    ctx := c.UserContext()
    traceID, _ := ctx.Value("trace_id").(string)

    log.Printf("[%s] Creating user", traceID)
    return gormz.FromContext[User](ctx).Create(&user)
}
```

### مثال 5: Fallback

```go
// FromContext يستخدم global إذا لم يوجد
func Service(ctx context.Context) error {
    // إذا ctx فيه DB → يستخدمه
    // وإلا → global
    q := gormz.FromContext[User](ctx)
    return q.Create(&user)
}
```

---

## 🔧 Usage مع الحزم

```go
func ProcessOrder(ctx context.Context, orderID uint) error {
    // استخدم ctx في كل عمليات DB
    order, err := gormz.FromContext[Order](ctx).Get(orderID)
    if err != nil {
        return err
    }

    user, err := gormz.FromContext[User](ctx).Get(order.UserID)
    if err != nil {
        return err
    }

    // إذا ctx يُلغى → كل الاستعلامات تُلغى
    return nil
}
```

---

## ⚠️ تحذيرات

### 1. Context Immutable

```go
ctx := context.Background()
ctx = gormz.WithDB(ctx, app)  // ← ضروري إعادة التخصيص
```

### 2. Do Not Store in Structs

```go
// ❌ خطأ
type Service struct {
    ctx context.Context
}

// ✅ صحيح — تمرير كـ argument
func (s *Service) Do(ctx context.Context) error {
    // ...
}
```

### 3. Nil Context

```go
// ❌ panic
gormz.FromContext[User](nil)

// ✅
gormz.FromContext[User](context.Background())
```

### 4. Context Overhead

```go
// كل WithDB ينسخ context — رخيص لكن موجود
ctx := gormz.WithDB(context.Background(), app)
```

---

## 🎯 Patterns

### 1. Request-scoped DB

```go
func NewRequestContext(r *http.Request) context.Context {
    ctx := r.Context()
    ctx = gormz.WithDB(ctx, getDBForRequest(r))
    ctx = context.WithValue(ctx, "request_id", r.Header.Get("X-Request-ID"))
    return ctx
}
```

### 2. Background Jobs

```go
func ProcessJob(ctx context.Context, jobID uint) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()

    return gormz.FromContext[Job](ctx).Get(jobID)
}
```

### 3. Tests

```go
func setupTest(t *testing.T) context.Context {
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    db.AutoMigrate(&User{})
    inst := gormz.NewInstance(db)

    return gormz.WithDB(context.Background(), inst)
}

func TestCreate(t *testing.T) {
    ctx := setupTest(t)

    user := &User{Name: "Ali"}
    err := gormz.FromContext[User](ctx).Create(user)
    require.NoError(t, err)
}
```

---

## 📝 Best Practices

### ✅ Do

- استخدم `FromContext` في handlers/services
- `WithDB` في middleware
- `context.WithTimeout` للعمليات الطويلة
- Fallback pattern (global as default)

### ❌ Don't

- لا تخزّن context في struct
- لا تستخدم nil context
- لا تمرّر context عبر قنوات
- لا تستخدم context للبيانات الاختيارية