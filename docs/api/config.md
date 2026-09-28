# Config — الإعدادات

> إعدادات الاتصال والسلوك.

---

## 📖 API

```go
type Config struct {
    // Connection pool
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration

    // Logging
    LogLevel  LogLevel
    SlowQuery time.Duration

    // Behavior
    PrepareStmt bool
    DryRun      bool
}
```

---

## 🎯 Presets

### `DefaultConfig()`

```go
func DefaultConfig() Config {
    return Config{
        MaxOpenConns:    25,
        MaxIdleConns:    5,
        ConnMaxLifetime: time.Hour,
        ConnMaxIdleTime: 10 * time.Minute,
        LogLevel:        LogSilent,
        SlowQuery:       200 * time.Millisecond,
        PrepareStmt:     false,
        DryRun:          false,
    }
}
```

**مناسب لـ**: Production.

### `DevelopmentConfig()`

```go
func DevelopmentConfig() Config {
    return Config{
        MaxOpenConns:    10,
        MaxIdleConns:    2,
        ConnMaxLifetime: 30 * time.Minute,
        ConnMaxIdleTime: 5 * time.Minute,
        LogLevel:        LogInfo,
        SlowQuery:       100 * time.Millisecond,
        PrepareStmt:     false,
        DryRun:          false,
    }
}
```

**مناسب لـ**: Development.

### `TestingConfig()`

```go
func TestingConfig() Config {
    return Config{
        MaxOpenConns:    5,
        MaxIdleConns:    1,
        ConnMaxLifetime: time.Minute,
        LogLevel:        LogSilent,
        SlowQuery:       0,
        PrepareStmt:     false,
        DryRun:          false,
    }
}
```

**مناسب لـ**: Tests.

---

## 💡 أمثلة

### مثال 1: Global Config

```go
db, _ := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
gormz.SetDB(db)

err := gormz.ConfigureDB(gormz.DefaultConfig())
if err != nil {
    log.Fatal(err)
}
```

### مثال 2: Per-Instance

```go
app := gormz.NewInstance(db)

err := app.Configure(gormz.Config{
    MaxOpenConns:    50,
    MaxIdleConns:    10,
    ConnMaxLifetime: time.Hour,
})
```

### مثال 3: Development Setup

```go
cfg := gormz.DevelopmentConfig()
cfg.LogLevel = gormz.LogInfo

err := gormz.ConfigureDB(cfg)
```

### مثال 4: Slow Query Detection

```go
cfg := gormz.DefaultConfig()
cfg.SlowQuery = 100 * time.Millisecond

// ستُسجّل الاستعلامات > 100ms
gormz.ConfigureDB(cfg)
```

### مثال 5: Prepare Statements

```go
cfg := gormz.DefaultConfig()
cfg.PrepareStmt = true

// أسرع للاستعلامات المتكررة، لكن ذاكرة أكثر
```

---

## 📊 Log Levels

| Level | ماذا يسجّل |
|-------|-----------|
| `LogSilent` | لا شيء |
| `LogError` | أخطاء فقط |
| `LogWarn` | أخطاء + تحذيرات |
| `LogInfo` | كل شيء (بطيء) |

**Production**: `LogSilent` أو `LogError`
**Development**: `LogInfo`

---

## 🔧 Connection Pool

### MaxOpenConns

الحد الأقصى للاتصالات المتزامنة.

**القاعدة**:
```
MaxOpenConns = (2 × CPU cores) + effective_spindle_count
```

**مثال**: 4 cores → ~10 connections.

### MaxIdleConns

عدد الاتصالات الخاملة المحتفظ بها.

**القاعدة**: `MaxIdleConns < MaxOpenConns`.

### ConnMaxLifetime

العمر الأقصى للاتصال.

**التوصية**: 1 hour. يتجنب:
- Connection leaks
- Load balancer timeouts
- MySQL wait_timeout

### ConnMaxIdleTime

العمر الأقصى للاتصال الخامل.

**التوصية**: 10 minutes.

---

## 💾 Preset للـ Database

### PostgreSQL

```go
cfg := gormz.DefaultConfig()
cfg.MaxOpenConns = 50
cfg.MaxIdleConns = 10
cfg.ConnMaxLifetime = time.Hour
cfg.PrepareStmt = true
```

### MySQL

```go
cfg := gormz.DefaultConfig()
cfg.MaxOpenConns = 30
cfg.MaxIdleConns = 5
cfg.ConnMaxLifetime = 30 * time.Minute  // < wait_timeout
```

### SQLite

```go
cfg := gormz.DefaultConfig()
cfg.MaxOpenConns = 1  // SQLite: writer واحد
cfg.MaxIdleConns = 1
```

---

## 📝 Best Practices

### ✅ Do

- استخدم preset حسب البيئة
- اضبط pool حسب DB
- `PrepareStmt = true` للـ PostgreSQL/MySQL
- `SlowQuery` للتشخيص

### ❌ Don't

- لا تستخدم `MaxOpenConns = 0` (unlimited)
- لا تضع `PrepareStmt = true` مع SQLite
- لا تستخدم `LogInfo` في production