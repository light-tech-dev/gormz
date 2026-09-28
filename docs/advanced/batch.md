# Batch — المعالجة الدفعية

> معالجة آلاف السجلات بكفاءة.

---

## 📖 API

### ProcessBatch

```go
result := advanced.ProcessBatch[T](ctx, items, cfg, fn)
```

### Stream

```go
err := advanced.Stream[T](ctx, batchSize, fn)
```

### Parallel

```go
err := advanced.Parallel().Add(fn1).Add(fn2).Run(ctx)
```

---

## 💡 أمثلة

### مثال 1: Basic Batch

```go
users := fetchUsers()  // 100K users

result := advanced.ProcessBatch(ctx, users, advanced.DefaultBatchConfig(),
    func(batch []User) error {
        return gormz.New[User]().CreateMany(batch)
    })

fmt.Printf("Success: %d, Failed: %d\n", result.Success, result.Failed)
```

### مثال 2: Custom Config

```go
cfg := advanced.BatchConfig{
    BatchSize:   500,
    Workers:     8,
    StopOnError: false,
}

advanced.ProcessBatch(ctx, items, cfg, fn)
```

### مثال 3: Streaming

```go
// معالجة في streaming (لا يحمّل الكل)
err := advanced.Stream[User](ctx, 1000, func(batch []User) error {
    fmt.Printf("Processing batch of %d\n", len(batch))
    return processBatch(batch)
})
```

### مثال 4: Parallel Queries

```go
err := advanced.Parallel().
    Add(func() error {
        return refreshCache()
    }).
    Add(func() error {
        return updateStats()
    }).
    Add(func() error {
        return sendNotifications()
    }).
    Run(ctx)
```

### مثال 5: Import CSV

```go
func ImportUsers(ctx context.Context, csvPath string) error {
    file, _ := os.Open(csvPath)
    defer file.Close()

    reader := csv.NewReader(file)
    _, _ = reader.Read()  // skip header

    batch := make([]User, 0, 1000)

    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }

        batch = append(batch, User{
            Name:  record[0],
            Email: record[1],
        })

        if len(batch) >= 1000 {
            if err := gormz.New[User]().CreateMany(batch); err != nil {
                return err
            }
            batch = batch[:0]
        }
    }

    // آخر دفعة
    if len(batch) > 0 {
        return gormz.New[User]().CreateMany(batch)
    }

    return nil
}
```

---

## 📊 Performance

| BatchSize | Workers | 100K records |
|-----------|---------|--------------|
| 100 | 1 | 60s |
| 1000 | 1 | 20s |
| 1000 | 4 | 8s |
| 1000 | 8 | 6s |
| 5000 | 4 | 10s |

**القواعد**:
- BatchSize 1000 → توازن جيد
- Workers 4-8 → أفضل
- Workers > CPUs → لا فائدة

---

## ⚠️ تحذيرات

### 1. Memory

```go
// ⚠️ 100M سجل في slice → OOM
items := make([]User, 100_000_000)

// ✅ Stream بدلًا
advanced.Stream[User](ctx, 1000, processBatch)
```

### 2. Errors

`ProcessBatch` **لا يتوقف** على الخطأ (إلا `StopOnError`):

```go
result := advanced.ProcessBatch(...)
if len(result.Errors) > 0 {
    log.Printf("Failed batches: %d", len(result.Errors))
}
```

### 3. Transactions

كل batch له transaction منفصل:

```go
// ✅ batch واحد
advanced.BulkInsert[User](ctx, batch, cfg)

// ❌ transaction ضخم
db.Transaction(func(tx *gorm.DB) error {
    for _, u := range allUsers {
        tx.Create(&u)
    }
})
```

### 4. DB Limits

- PostgreSQL: max 65535 parameters
- MySQL: max_allowed_packet
- SQLite: SQLITE_MAX_VARIABLE_NUMBER

`BatchSize` يجب أن يكون أقل من الحدود.

---

## 🎯 Patterns

### 1. Resumable Import

```go
func ImportResumable(ctx context.Context, startID uint) error {
    batchSize := 1000
    offset := 0

    for {
        items, err := gormz.New[User]().
            Filter("id__gt", startID).
            Limit(batchSize).
            All()
        if err != nil {
            return err
        }
        if len(items) == 0 {
            break
        }

        if err := processBatch(items); err != nil {
            log.Printf("Failed at offset %d, resumable", offset)
            return err
        }

        startID = items[len(items)-1].ID
        offset += len(items)
    }

    return nil
}
```

### 2. Progress

```go
total := len(items)
processed := atomic.Int64{}

advanced.ProcessBatch(ctx, items, cfg, func(batch []User) error {
    // process
    n := processed.Add(int64(len(batch)))
    pct := float64(n) / float64(total) * 100
    fmt.Printf("\rProgress: %.1f%%", pct)
    return nil
})
```

---

## 📝 Best Practices

### ✅ Do

- استخدم `Stream` للضخم
- `Workers` = عدد CPUs
- اختبر `BatchSize`
- Log errors

### ❌ Don't

- لا تحمّل كل شيء في slice
- لا تستخدم workers ضخمة
- لا تتجاهل errors