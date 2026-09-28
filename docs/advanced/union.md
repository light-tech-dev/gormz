# Union — الدمج

> دمج نتائج عدة استعلامات.

---

## 📖 API

### `Union`

```go
advanced.Union(q1, q2, q3).All()
```

يحذف المكرر.

### `UnionAll`

```go
advanced.UnionAll(q1, q2).All()
```

يحتفظ بالمكرر (أسرع).

---

## 💡 أمثلة

### مثال 1: Basic

```go
q1 := gormz.New[User]().Filter("status", "active")
q2 := gormz.New[User]().Filter("status", "pending")

users, err := advanced.Union(q1, q2).All()
// = (active ∪ pending) مع حذف المكرر
```

### مثال 2: Union All

```go
q1 := gormz.New[Log]().Filter("level", "error")
q2 := gormz.New[Log]().Filter("level", "critical")

logs, _ := advanced.UnionAll(q1, q2).All()
// = error + critical (بدون حذف)
```

### مثال 3: Order + Limit على Union

```go
users, _ := advanced.Union(q1, q2).
    OrderBy("-created_at").
    Limit(10).
    All()
```

### مثال 4: Dynamic

```go
uq := advanced.Union[User]()

if filter.Active {
    uq.Add(gormz.New[User]().Filter("status", "active"))
}
if filter.Pending {
    uq.Add(gormz.New[User]().Filter("status", "pending"))
}
if filter.Archived {
    uq.Add(gormz.New[User]().Filter("status", "archived"))
}

users, _ := uq.All()
```

---

## ⚠️ التحذيرات

### 1. Column Count

كل استعلام يجب أن يكون له **نفس عدد الأعمدة**:

```go
// ❌ خطأ
q1 := gormz.New[User]().Select("id", "name")
q2 := gormz.New[User]().Select("id")  // 1 عمود فقط
advanced.Union(q1, q2)  // error

// ✅ صحيح
q2 := gormz.New[User]().Select("id", "name")
```

### 2. Column Types

الأعمدة يجب أن تكون **متوافقة**:

```go
// ⚠️ int vs string
q1 := ... Select("id")        // int
q2 := ... Select("name")      // string
// قد يعمل لكن غريب
```

### 3. Performance

- **UNION**: يعمل DISTINCT (أبطأ)
- **UNION ALL**: أسرع، استخدمه إذا كنت واثقًا

### 4. SQLite

مدعوم منذ 3.0+.

---

## 📝 Best Practices

### ✅ Do

- استخدم `UNION ALL` عند الإمكان
- اختر عدد أعمدة متطابق
- استخدم types متوافقة

### ❌ Don't

- لا تخلط أنواع مختلفة
- لا تنسَ ORDER BY بعد Union
- لا تستخدم Union في loop