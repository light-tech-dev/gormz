# gormz — التوثيق

> طبقة ORM بسيطة وقوية فوق GORM، مستوحاة من modern ORM.

---

## 📖 جدول المحتويات

### 🚀 البداية
- [البدء السريع](01-getting-started.md)
- [التثبيت](02-installation.md)
- [أول استعلام](03-quickstart.md)

### 🎯 Core
- [QuerySet](core/queryset.md) — الأساس
- [Lookups](core/lookups.md) — الفلاتر
- [Q Builder](core/q-builder.md) — الشروط المعقدة
- [Pagination](core/pagination.md) — التصفح
- [Hooks](core/hooks.md) — دورة الحياة
- [Registry](core/registry.md) — السجل
- [Context](core/context.md) — السياق

### 🚀 Advanced
- [Transactions](advanced/transactions.md)
- [CTE](advanced/cte.md)
- [Locking](advanced/locking.md)
- [Joins](advanced/joins.md)
- [Bulk](advanced/bulk.md)
- [Aggregates](advanced/aggregates.md)
- [SubQuery](advanced/subquery.md)
- [Union](advanced/union.md)
- [Window](advanced/window.md)
- [Retry](advanced/retry.md)
- [Batch](advanced/batch.md)

### 📚 Guides
- [Multi-DB](guides/multi-db.md)
- [Testing](guides/testing.md)
- [Performance](guides/performance.md)
- [Security](guides/security.md)
- [Migration من GORM](guides/migration-from-gorm.md)
- [Best Practices](guides/best-practices.md)

### 📖 API
- [Errors](api/errors.md)
- [Config](api/config.md)
- [Reference](api/reference.md)

### 💡 Examples
- [كل الأمثلة](examples/README.md)

---

## 🎯 لمن هذا التوثيق؟

- **المطوّر الجديد**: ابدأ بـ [البداية السريعة](01-getting-started.md)
- **المطوّر المتمرس**: انتقل إلى [Advanced](advanced/README.md)
- **المهاجر من GORM**: اقرأ [Migration Guide](guides/migration-from-gorm.md)

---

## 🔍 بحث سريع

### أريد أن...

| المهمة | الملف |
|--------|------|
| أُنشئ سجلًا | [QuerySet](core/queryset.md#create) |
| أبحث بفلتر | [Lookups](core/lookups.md) |
| أُصفّح النتائج | [Pagination](core/pagination.md) |
| أستخدم transaction | [Transactions](advanced/transactions.md) |
| أُدرج 100K سجل | [Bulk](advanced/bulk.md) |
| أعمل join | [Joins](advanced/joins.md) |
| أستخدم CTE | [CTE](advanced/cte.md) |
| أستخدم عدة DBs | [Multi-DB](guides/multi-db.md) |
| أنتقل من GORM | [Migration](guides/migration-from-gorm.md) |

---

## 📊 حالة التوثيق

| القسم | الحالة |
|-------|--------|
| Core | ✅ 100% |
| Advanced | ✅ 100% |
| Guides | ✅ 100% |
| Examples | ✅ 100% |
| API | ✅ 100% |