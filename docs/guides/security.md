# Security — الأمان في gormz

> حماية تطبيقك من SQL injection وأخطاء شائعة.

---

## 🎯 الحماية المدمجة

### 1. Field Validation

كل حقل يُمرّ بفحص:

```go
// ❌ يpanic
gormz.New[User]().Filter("name; DROP TABLE users", "x")

// ✅ آمن
gormz.New[User]().Filter("name", "x")
```

**القواعد**:
- الحروف: `a-zA-Z0-9_`
- النقطة: `.` للـ joins
- ممنوع: كلمات SQL (`select`, `drop`, ...)

### 2. Parameterized Queries

القيم دائمًا parameterized:

```go
.Filter("name", userInput)  // name = ? [userInput]
```

### 3. Guards

```go
// ❌ خطأ
gormz.New[User]().UpdateMany(map[string]any{"active": false})
// err = dangerous operation

// ✅
gormz.New[User]().
    Filter("age__lt", 18).
    UpdateMany(map[string]any{"active": false})
```

---

## ⚠️ النقاط الحساسة

### 1. `Where` SQL خام

```go
// ❌ خطر
.Where(fmt.Sprintf("name = '%s'", userInput))

// ✅ آمن
.Where("name = ?", userInput)

// ⚠️ خطر إذا لم تتحقق
.Where("name = '" + userInput + "'")  // SQL injection!
```

### 2. `Raw`

```go
// ❌ خطر
gormz.Raw("SELECT * FROM users WHERE name = '" + name + "'")

// ✅ آمن
gormz.Raw("SELECT * FROM users WHERE name = ?", name)
```

### 3. `Select`/`OrderBy`

✅ آمنة (تتحقق من الحقل):

```go
.Select("name", "email")  // ✅
.OrderBy("name", "-created_at")  // ✅
```

---

## 🔐 Password Hashing

لا تخزن كلمات المرور كنص:

```go
import "golang.org/x/crypto/bcrypt"

func HashPassword(pwd string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
    return string(bytes), err
}

func CheckPassword(pwd, hash string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd)) == nil
}
```

---

## 🛡️ Session Security

```go
// استخدام UUID بدل counter
import "github.com/google/uuid"

sessionID := uuid.New().String()
```

---

## 🔒 Sensitive Fields

### Omit الحقول الحساسة

```go
type User struct {
    ID       uint
    Username string
    Password string `json:"-"`  // لا يُصدَّر في JSON
    Token    string `json:"-"`
}

// أو
users, _ := gormz.New[User]().
    Omit("password", "token").
    All()
```

---

## 🚨 Common Mistakes

### 1. SQL Injection عبر Dynamic Field

```go
// ❌ خطر
fieldName := c.Query("sort_by")
gormz.New[User]().Filter(fieldName, value)  // panic (بفضل validation)

// ✅ آمن — استخدم whitelist
allowed := map[string]bool{"name": true, "age": true, "email": true}
if !allowed[fieldName] {
    return errors.New("invalid field")
}
```

### 2. Mass Assignment

```go
// ❌ خطر — المستخدم قد يعدّل is_admin
type UpdateDTO struct {
    Name    string
    Email   string
    IsAdmin bool  // ← خطر!
}

// ✅ آمن — DTO محدود
type UpdateDTO struct {
    Name  string
    Email string
}
```

### 3. Soft Delete Bypass

```go
// ❌ المستخدم قد يستعيد محذوفًا
func DeleteUser(id uint) error {
    return gormz.New[User]().Delete(id).Error  // soft
}

// قد يستخدم
gormz.New[User]().WithDeleted().Get(id)  // يعرض المحذوف
```

**الحل**: تحقق من `DeletedAt` في handlers حساسة.

---

## 📋 Checklist

- [ ] كل user input يمر بـ validation
- [ ] Passwords hashed (bcrypt/argon2)
- [ ] Sessions مع UUID
- [ ] `Omit` للحقول الحساسة
- [ ] DTOs محدودة
- [ ] Rate limiting على endpoints
- [ ] HTTPS في production
- [ ] CSRF protection
- [ ] SQL injection tests

---

## 🧪 اختبارات الأمان

```go
func TestSQLInjection(t *testing.T) {
    setupTestDB(t)

    malicious := []string{
        "name; DROP TABLE users",
        "name' OR '1'='1",
        "1=1",
        "SELECT * FROM users",
    }

    for _, input := range malicious {
        t.Run(input, func(t *testing.T) {
            assert.Panics(t, func() {
                gormz.New[User]().Filter(input, "x")
            })
        })
    }
}
```

---

## 📝 Best Practices

### ✅ Do

- استخدم `Filter` بدل `Where` إن أمكن
- تحقق من `field names` قبل الاستخدام
- استخدم DTOs محدودة
- Omit الحقول الحساسة
- Test SQL injection

### ❌ Don't

- لا تستخدم string concatenation في SQL
- لا تخزن plaintext passwords
- لا تفترض أمان user input
- لا تستخدم `Raw` بدون validation