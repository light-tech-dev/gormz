# Hooks — دورة حياة الموديل

> BeforeCreate, AfterUpdate, ... — كل hooks gormz.

---

## 📖 نظرة عامة

gormz يدعم كل hooks GORM **natively** + يعيد تصديرها كـ interfaces.

**كيف يعمل**: نفّذ أي hook على موديلك، gormz سيدعوه تلقائيًا.

---

## 🎣 Hooks المتاحة

### Creation

| Hook | متى | الاستخدام |
|------|-----|----------|
| `BeforeCreate() error` | قبل الإدراج | validation, normalize |
| `AfterCreate() error` | بعد الإدراج | إشعارات, side-effects |

### Update

| Hook | متى | الاستخدام |
|------|-----|----------|
| `BeforeUpdate() error` | قبل التحديث | validation |
| `AfterUpdate() error` | بعد التحديث | cache invalidation |

### Save (Create + Update)

| Hook | متى | الاستخدام |
|------|-----|----------|
| `BeforeSave() error` | قبل الإنشاء/التحديث | normalize |
| `AfterSave() error` | بعد الإنشاء/التحديث | audit |

### Delete

| Hook | متى | الاستخدام |
|------|-----|----------|
| `BeforeDelete() error` | قبل الحذف | منع الحذف |
| `AfterDelete() error` | بعد الحذف | cleanup |

### Find

| Hook | متى | الاستخدام |
|------|-----|----------|
| `AfterFind() error` | بعد القراءة | deserialize |

---

## 💡 أمثلة

### مثال 1: Normalize

```go
type User struct {
    ID    uint
    Name  string
    Email string
}

func (u *User) BeforeCreate() error {
    u.Name = strings.TrimSpace(u.Name)
    u.Email = strings.ToLower(strings.TrimSpace(u.Email))
    return nil
}

// استخدام
user := &User{Name: "  Ali  ", Email: "  ALI@TEST.COM  "}
gormz.New[User]().Create(user)
// user.Name = "Ali", user.Email = "ali@test.com"
```

### مثال 2: Validation

```go
func (u *User) BeforeSave() error {
    if len(u.Name) < 2 {
        return errors.New("name too short")
    }
    if !strings.Contains(u.Email, "@") {
        return errors.New("invalid email")
    }
    return nil
}
```

### مثال 3: Audit

```go
type Post struct {
    ID        uint
    Title     string
    CreatedBy uint
    UpdatedBy uint
}

func (p *Post) BeforeCreate() error {
    p.CreatedBy = currentUserID()  // من context
    p.UpdatedBy = currentUserID()
    return nil
}

func (p *Post) BeforeUpdate() error {
    p.UpdatedBy = currentUserID()
    return nil
}
```

### مثال 4: Side Effects

```go
func (o *Order) AfterCreate() error {
    // إرسال إشعار
    go notify.Created(o)
    return nil
}

func (o *Order) AfterUpdate() error {
    // cleanup cache
    cache.Delete("order:" + strconv.Itoa(int(o.ID)))
    return nil
}
```

### مثال 5: Prevent Delete

```go
func (u *User) BeforeDelete() error {
    if u.IsAdmin {
        return errors.New("cannot delete admin")
    }
    return nil
}
```

### مثال 6: Encrypt Password

```go
type User struct {
    ID       uint
    Password string
}

func (u *User) BeforeCreate() error {
    if u.Password != "" {
        hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
        if err != nil {
            return err
        }
        u.Password = string(hashed)
    }
    return nil
}

func (u *User) BeforeUpdate() error {
    // نفس الشيء عند التحديث
    if u.Password != "" && !isHashed(u.Password) {
        hashed, _ := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
        u.Password = string(hashed)
    }
    return nil
}
```

---

## 🔧 Usage مع gormz

### مع `QuerySet`

```go
// Hooks تعمل تلقائيًا
gormz.New[User]().Create(user)      // ← BeforeCreate, AfterCreate
gormz.New[User]().Save(user)        // ← BeforeSave, AfterSave
gormz.New[User]().Delete(id)        // ← BeforeDelete, AfterDelete
gormz.New[User]().All()             // ← AfterFind لكل سجل
```

### مع Bulk

⚠️ **Bulk operations قد لا تدعو hooks**:

```go
// ✅ يدعو hooks
gormz.New[User]().Create(user)

// ⚠️ قد لا يدعو (للسرعة)
advanced.BulkInsert[User](ctx, users, cfg)
```

**الحل**: استخدم `CreateMany` إذا احتجت hooks:

```go
gormz.New[User]().CreateMany(users)  // ← يدعو hooks
```

---

## ⚠️ تحذيرات

### 1. Performance

Hooks تُنفَّذ لكل سجل. للـ bulk، استخدم batch:

```go
// ❌ بطيء
for _, user := range users {
    gormz.New[User]().Create(&user)  // hooks لكل واحد
}

// ✅ أسرع
gormz.New[User]().CreateMany(users)  // hooks للأول فقط أو لا شيء
```

### 2. Transactions

Hooks داخل transaction. إذا فشل hook → rollback:

```go
func (u *User) BeforeCreate() error {
    // إذا error → rollback تلقائي
    return validate(u)
}
```

### 3. Infinite Loops

⚠️ لا تستدعي `Save` داخل hook:

```go
// ❌ infinite loop
func (u *User) AfterSave() error {
    return gormz.New[User]().Save(u).Error
}
```

### 4. Context Access

Hooks لا تحصل على context. للـ user info:

```go
// استخدم global أو thread-local
var currentUserID uint

func (u *User) BeforeCreate() error {
    u.CreatedBy = currentUserID
    return nil
}

// قبل الإنشاء
currentUserID = getCurrentUserID(c)
gormz.New[User]().Create(user)
```

أو استخدم context value مع `WithContext`:

```go
ctx := context.WithValue(context.Background(), "user_id", 1)
gormz.New[User]().WithContext(ctx).Create(user)
```

---

## 🎯 Best Practices

### ✅ Do

- Normalize قبل الحفظ
- Validate قبل الإنشاء
- Audit (CreatedBy, UpdatedBy)
- Side-effects في `After*`
- منع الحذف الحساس

### ❌ Don't

- لا تستدعي DB داخل hook (بطيء)
- لا تستدعي `Save` داخل hook
- لا تعتمد على hook في bulk
- لا تُنفّذ عمليات بطيئة في `Before*`