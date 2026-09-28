# Transactions — المعاملات

> معاملات آمنة مع retry, savepoints, و nested.

---

## 📖 API

### `WithTransaction`

```go
err := advanced.WithTransaction(ctx, cfg, fn)
```

### `Begin` + Manual

```go
tx, err := advanced.Begin(ctx, cfg)
defer tx.RollbackIfActive()
// ...
tx.Commit()
```

### TxConfig

```go
type TxConfig struct {
    Isolation  Isolation       // مستوى العزل
    ReadOnly   bool            // للقراءة فقط
    Timeout    time.Duration   // مهلة
    Retries    int             // عدد المحاولات
    RetryDelay time.Duration   // التأخير
}
```

### Isolation Levels

| Level | PostgreSQL | MySQL |
|-------|-----------|-------|
| `IsolationReadUncommitted` | ✅ | ✅ |
| `IsolationReadCommitted` | ✅ | ✅ |
| `IsolationRepeatableRead` | ✅ | ✅ |
| `IsolationSerializable` | ✅ | ✅ |

---

## 💡 أمثلة

### مثال 1: Simple Transaction

```go
err := advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
    func(tx *advanced.Tx) error {
        user := &User{Name: "Ali"}
        if err := tx.Query[User]().Create(user); err != nil {
            return err
        }

        order := &Order{UserID: user.ID, Total: 100}
        return tx.Query[Order]().Create(order)
    })
```

### مثال 2: با Retry

```go
cfg := advanced.TxConfig{
    Isolation:  advanced.IsolationSerializable,
    Retries:    5,
    RetryDelay: 100 * time.Millisecond,
}

err := advanced.WithTransaction(ctx, cfg, func(tx *advanced.Tx) error {
    // ...
    return nil
})
```

### مثال 3: Nested (Savepoints)

```go
err := advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
    func(tx *advanced.Tx) error {
        // المستوى 1
        tx.Query[User]().Create(&user)

        // nested — savepoint
        err := tx.Nested(func(inner *advanced.Tx) error {
            return inner.Query[Log]().Create(&log)
        })
        if err != nil {
            log.Printf("Nested failed (ignored): %v", err)
        }

        return nil
    })
```

### مثال 4: Transfer بين حسابين

```go
func Transfer(ctx context.Context, fromID, toID uint, amount float64) error {
    return advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
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
                return errors.New("insufficient balance")
            }

            from.Balance -= amount
            to.Balance += amount

            if err := tx.Query[Account]().Save(from); err != nil {
                return err
            }
            return tx.Query[Account]().Save(to)
        })
}
```

### مثال 5: Savepoint Stack

```go
advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
    func(tx *advanced.Tx) error {
        stack := advanced.NewSavepointStack(tx.DB())

        // نقطة 1
        stack.Push("step1")
        tx.Query[User]().Create(&user1)

        // نقطة 2
        stack.Push("step2")
        tx.Query[User]().Create(&user2)

        // تراجع للـ step1
        stack.RollbackTo("step1")

        return nil
    })
```

---

## ⚠️ تحذيرات

### 1. Isolation Levels

- **PostgreSQL/MySQL فقط** (SQLite لا يدعم `SET TRANSACTION`)
- **ReadCommitted**: default في Postgres
- **RepeatableRead**: default في MySQL
- **Serializable**: الأكثر أمانًا، الأبطأ

### 2. SQLite

- `savepoints` مدعومة
- `SET TRANSACTION` غير مدعوم
- `nested transactions` تعمل عبر savepoints

### 3. Panic Safety

```go
err := advanced.WithTransaction(ctx, cfg, func(tx *advanced.Tx) error {
    panic("oops")  // ← rollback تلقائي
})
// err = "gormz/advanced: panic in transaction: oops"
```

### 4. Retry Logic

Retry يعمل فقط للأخطاء القابلة للمحاولة:

```go
// ✅ يـretry: deadlock, lock timeout, serialization
// ❌ لا يـretry: validation, constraint violation
```

---

## 🎯 Patterns

### 1. Long-running with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := advanced.WithTransaction(ctx, cfg, fn)
```

### 2. Retry-Aware

```go
cfg := advanced.TxConfig{
    Retries:    3,
    RetryDelay: 50 * time.Millisecond,
}

// على deadlock → يعيد المحاولة
```

### 3. Read-only

```go
cfg := advanced.TxConfig{ReadOnly: true}
advanced.WithTransaction(ctx, cfg, func(tx *advanced.Tx) error {
    // قراءة فقط
    return nil
})
```

---

## 📝 Best Practices

### ✅ Do

- استخدم `WithTransaction` بدل `Begin/Commit` اليدوي
- اجعل `fn` قصيرة قدر الإمكان
- استخدم `RollbackIfActive` في defer
- استخدم `Retry` للعمليات الحساسة

### ❌ Don't

- لا تفتح معاملة أثناء HTTP request طويل
- لا تخلط معاملات مع goroutines
- لا تتجاهل الأخطاء
- لا تنسَ `defer`