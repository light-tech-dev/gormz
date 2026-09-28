# Retry — إعادة المحاولة

> Retry + Circuit Breaker + Backoff strategies.

---

## 📖 API

### Retry

```go
advanced.Retry(ctx, cfg, fn)
```

### RetryConfig

```go
type RetryConfig struct {
    MaxAttempts  int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
    Jitter       float64
    RetryIf      func(error) bool
}
```

### Backoff

```go
advanced.ExponentialBackoff(initial, max)
advanced.LinearBackoff(step, max)
advanced.ConstantBackoff(delay)
```

### Circuit Breaker

```go
cb := advanced.NewCircuitBreaker(maxFailures, resetTimeout)
cb.Call(fn)
```

---

## 💡 أمثلة

### مثال 1: Basic Retry

```go
err := advanced.Retry(ctx, advanced.DefaultRetryConfig(),
    func() error {
        return gormz.New[User]().Create(&user)
    })
```

### مثال 2: Custom Config

```go
cfg := advanced.RetryConfig{
    MaxAttempts:  5,
    InitialDelay: 100 * time.Millisecond,
    MaxDelay:     5 * time.Second,
    Multiplier:   2.0,
    Jitter:       0.1,
}

err := advanced.Retry(ctx, cfg, func() error {
    return callExternalAPI()
})
```

### مثال 3: Conditional Retry

```go
cfg := advanced.DefaultRetryConfig()
cfg.RetryIf = func(err error) bool {
    // retry فقط على network errors
    return errors.Is(err, syscall.ECONNREFUSED)
}

advanced.Retry(ctx, cfg, fn)
```

### مثال 4: Circuit Breaker

```go
cb := advanced.NewCircuitBreaker(5, 30*time.Second)

err := cb.Call(func() error {
    return callPaymentAPI()
})

if err != nil && cb.State() == advanced.CircuitOpen {
    return errors.New("service unavailable, try later")
}
```

### مثال 5: Backoff

```go
backoff := advanced.ExponentialBackoff(100*time.Millisecond, 10*time.Second)

for attempt := 1; attempt <= 5; attempt++ {
    if err := try(); err == nil {
        return nil
    }
    time.Sleep(backoff(attempt))
}
```

---

## 🔄 Circuit Breaker States

```
     ┌─────────┐
     │ CLOSED  │  ← طبيعي
     └────┬────┘
          │ failures >= maxFailures
          ↓
     ┌─────────┐
     │  OPEN   │  ← كل الطلبات تفشل
     └────┬────┘
          │ resetTimeout
          ↓
     ┌─────────┐
     │HALF-OPEN│  ← اختبار
     └────┬────┘
          │ نجاح × halfOpenMax
          ↓
     ┌─────────┐
     │ CLOSED  │
     └─────────┘
```

---

## ⚠️ تحذيرات

### 1. Jitter

```go
// بدون jitter → كل retries في نفس الوقت (thundering herd)
cfg.Jitter = 0.1  // ±10%
```

### 2. Context Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// إذا ctx يُلغى → retry يتوقف
advanced.Retry(ctx, cfg, fn)
```

### 3. Idempotency

⚠️ Retry فقط للعمليات **Idempotent**:

```go
// ✅ آمن للـ retry
gormz.New[User]().Filter("id", 1).Update("name", "Ali")

// ❌ ليس آمنًا (ينشئ نسخًا)
gormz.New[User]().Create(&user)
```

---

## 📝 Best Practices

### ✅ Do

- استخدم `Retry` لعمليات الشبكة
- `Retry` للـ deadlocks
- Circuit Breaker للخدمات الخارجية
- Jitter للتجنّب thundering herd

### ❌ Don't

- لا تـretry على validation errors
- لا تـretry بدون timeout
- لا تـretry non-idempotent operations