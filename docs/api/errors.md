# Errors — المرجع الكامل

> كل الأخطاء في gormz.

---

## 📖 الأخطاء المعرّفة

### Sentinel Errors

```go
var (
    ErrNotFound           = gorm.ErrRecordNotFound
    ErrNotInitialized     = errors.New("gormz: DB not initialized")
    ErrNilDB              = errors.New("gormz: nil DB")
    ErrInvalidField       = errors.New("gormz: invalid field")
    ErrInvalidQuery       = errors.New("gormz: invalid query")
    ErrDangerousOperation = errors.New("gormz: dangerous operation without conditions")
    ErrAlreadyRegistered  = errors.New("gormz: model already registered")
    ErrNotFoundInRegistry = errors.New("gormz: model not found in registry")
)
```

---

## 🔍 Typed Errors

### `NotFoundError`

```go
type NotFoundError struct {
    Model string
    ID    any
}

func (e *NotFoundError) Error() string
func (e *NotFoundError) Unwrap() error  // → ErrNotFound
```

### `ValidationError`

```go
type ValidationError struct {
    Field  string
    Reason string
}

func (e *ValidationError) Error() string
func (e *ValidationError) Unwrap() error  // → ErrInvalidField
```

---

## 🛠️ Helpers

### `IsNotFound(err)`

```go
_, err := gormz.New[User]().Get(999)
if gormz.IsNotFound(err) {
    log.Println("user not found")
}
```

### `IsValidation(err)`

```go
q, err := gormz.New[User]().TryFilter("bad field", "x")
if gormz.IsValidation(err) {
    log.Println("invalid field")
}
```

---

## 💡 أمثلة

### مثال 1: Handle Not Found

```go
user, err := gormz.New[User]().Get(1)
if err != nil {
    if gormz.IsNotFound(err) {
        return c.Status(404).JSON(fiber.Map{"error": "user not found"})
    }
    return c.Status(500).JSON(fiber.Map{"error": err.Error()})
}
```

### مثال 2: Extract Validation

```go
q, err := gormz.New[User]().TryFilter(field, value)
if err != nil {
    var ve *gormz.ValidationError
    if errors.As(err, &ve) {
        return fmt.Errorf("field %q: %s", ve.Field, ve.Reason)
    }
    return err
}
```

### مثال 3: Dangerous Operation

```go
_, err := gormz.New[User]().UpdateMany(map[string]any{"active": false})
if errors.Is(err, gormz.ErrDangerousOperation) {
    log.Println("refusing to update all users!")
}
```

### مثال 4: Registry

```go
q, err := gormz.TryRegister[User]("user")
if errors.Is(err, gormz.ErrAlreadyRegistered) {
    log.Println("user already registered")
}
```

---

## 🎯 Error Wrapping

gormz يستخدم `%w` للتغليف:

```go
// في المعاملات
return fmt.Errorf("%w: %q", ErrAlreadyRegistered, name)

// الاختبار
errors.Is(err, ErrAlreadyRegistered)  // ✅
```

---

## 📊 Error Reference

| Error | متى | الحل |
|-------|-----|------|
| `ErrNotFound` | سجل غير موجود | تحقق من ID |
| `ErrNotInitialized` | `SetDB` لم يُناد | اتصل بـ DB |
| `ErrNilDB` | `nil` DB | تحقق من DB |
| `ErrInvalidField` | حقل خاطئ | استخدم حقلًا صحيحًا |
| `ErrInvalidQuery` | استعلام خاطئ | راجع SQL |
| `ErrDangerousOperation` | بدون conditions | أضف فلتر |
| `ErrAlreadyRegistered` | تسجيل مكرر | استخدم `Has()` |
| `ErrNotFoundInRegistry` | موديل غير مسجّل | سجّله أولًا |

---

## 🧪 Testing Errors

```go
func TestNotFound(t *testing.T) {
    setupTestDB(t)

    _, err := gormz.New[User]().Get(999)
    assert.True(t, gormz.IsNotFound(err))
    assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestValidationError(t *testing.T) {
    _, err := gormz.New[User]().TryFilter("bad field", "x")
    require.Error(t, err)

    var ve *gormz.ValidationError
    require.True(t, errors.As(err, &ve))
    assert.Equal(t, "bad field", ve.Field)
}
```

---

## 📝 Best Practices

### ✅ Do

- استخدم `IsNotFound` للفحص
- استخدم `errors.As` لاستخراج التفاصيل
- استخدم `errors.Is` للمقارنة
- غلّف الأخطاء بـ `%w`

### ❌ Don't

- لا تقارن مباشرة بـ `==`
- لا تتجاهل `Unwrap`
- لا تُنشئ أخطاء بدون سياق