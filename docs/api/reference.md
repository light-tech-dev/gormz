# API Reference — المرجع الكامل

> كل الدوال والأنواع في gormz.

---

## 🌍 Global Functions

### Setup

| Function | الوصف |
|----------|-------|
| `SetDB(db *gorm.DB)` | ربط DB عالميًا |
| `DB() *gorm.DB` | الحصول على DB |
| `IsReady() bool` | هل DB جاهز؟ |
| `ConfigureDB(cfg Config) error` | ضبط pool |
| `Ping() error` | فحص الاتصال |
| `Close() error` | إغلاق الاتصال |

### Migration

| Function | الوصف |
|----------|-------|
| `Migrate[T]() error` | ترحيل موديل |
| `MustMigrate[T]()` | ترحيل + panic |

### Transaction

| Function | الوصف |
|----------|-------|
| `Transaction(ctx, fn) error` | معاملة |

### Info

| Function | الوصف |
|----------|-------|
| `TableNameOf[T]() string` | اسم الجدول |
| `FieldNamesOf[T]() []string` | أسماء الأعمدة |
| `ModelNameOf[T]() string` | اسم الموديل |

---

## 🎯 QuerySet

### Constructors

| Function | الوصف |
|----------|-------|
| `New[T]() *QuerySet[T]` | جديد |
| `NewWith[T](inst) *QuerySet[T]` | على Instance |
| `FromContext[T](ctx) *QuerySet[T]` | من context |

### Filters

| Method | الوصف |
|--------|-------|
| `.Filter(field, value)` | فلتر |
| `.TryFilter(field, value)` | فلتر + error |
| `.Exclude(field, value)` | عكس |
| `.TryExclude(field, value)` | + error |
| `.Where(sql, args...)` | SQL |
| `.Q(builder)` | شروط معقدة |

### Ordering

| Method | الوصف |
|--------|-------|
| `.OrderBy(fields...)` | ترتيب |
| `.TryOrderBy(fields...)` | + error |
| `.Limit(n)` | حد |
| `.Offset(n)` | بداية |

### Select

| Method | الوصف |
|--------|-------|
| `.Select(fields...)` | تحديد |
| `.Omit(fields...)` | استثناء |
| `.Preload(relations...)` | تحميل علاقات |

### Soft Delete

| Method | الوصف |
|--------|-------|
| `.WithDeleted()` | مع المحذوف |
| `.OnlyDeleted()` | المحذوف فقط |

### Reads

| Method | الوصف |
|--------|-------|
| `.All() ([]T, error)` | كل النتائج |
| `.First() (*T, error)` | أول سجل |
| `.FirstOrNil() (*T, error)` | أول أو nil |
| `.Last() (*T, error)` | آخر سجل |
| `.Get(id) (*T, error)` | بالـ ID |
| `.Find(field, value) (*T, error)` | بحقل |
| `.Count() (int64, error)` | عدد |
| `.Exists() (bool, error)` | موجود؟ |
| `.Pluck(field, dest) error` | عمود واحد |
| `.ScanInto(dest) error` | struct |

### Writes

| Method | الوصف |
|--------|-------|
| `.Create(item) error` | إنشاء |
| `.CreateMany(items) error` | إنشاء متعدد |
| `.Save(item) error` | حفظ |
| `.Update(id, field, value) error` | تحديث |
| `.UpdateMany(values) (int64, error)` | تحديث متعدد |
| `.Delete(id) error` | حذف (soft) |
| `.DeleteMany() (int64, error)` | حذف متعدد |
| `.HardDelete(id) error` | حذف نهائي |
| `.Restore(id) error` | استعادة |

### Aggregates

| Method | الوصف |
|--------|-------|
| `.Sum(field) (float64, error)` | مجموع |
| `.Avg(field) (float64, error)` | متوسط |
| `.Min(field) (float64, error)` | حد أدنى |
| `.Max(field) (float64, error)` | حد أقصى |

### Pagination

| Method | الوصف |
|--------|-------|
| `.Paginate(page, perPage)` | صفحة |

### Debug

| Method | الوصف |
|--------|-------|
| `.ToSQL() (string, []any)` | SQL |
| `.String() string` | SQL كنص |
| `.DB() *gorm.DB` | GORM أساسي |
| `.Build() *gorm.DB` | مبني |
| `.DryRun() *gorm.DB` | بدون تنفيذ |

### Context

| Method | الوصف |
|--------|-------|
| `.WithContext(ctx)` | ربط context |

---

## 🎨 Q Builder

### Constructors

| Function | الوصف |
|----------|-------|
| `Qb() *Q` | جديد (AND) |
| `QOr(children...)` | OR |
| `QAnd(children...)` | AND |

### Conditions

| Function | SQL |
|----------|-----|
| `Eq(field, v)` | `= v` |
| `Ne(field, v)` | `!= v` |
| `Gt(field, v)` | `> v` |
| `Gte(field, v)` | `>= v` |
| `Lt(field, v)` | `< v` |
| `Lte(field, v)` | `<= v` |
| `Contains(field, s)` | `LIKE '%s%'` |
| `StartsWith(field, s)` | `LIKE 's%'` |
| `EndsWith(field, s)` | `LIKE '%s'` |
| `In(field, values)` | `IN (...)` |
| `IsNull(field)` | `IS NULL` |
| `NotNull(field)` | `IS NOT NULL` |

### Methods

| Method | الوصف |
|--------|-------|
| `.And(children...)` | AND |
| `.Or(children...)` | OR |
| `.AndGroup(children...)` | AND group |
| `.OrGroup(children...)` | OR group |
| `.ToSQL() (string, []any)` | SQL |
| `.Len() int` | عدد الشروط |
| `.IsEmpty() bool` | فارغ؟ |

---

## 📦 PaginatedResult

### Fields

```go
type PaginatedResult[T any] struct {
    Items      []T   `json:"items"`
    Total      int64 `json:"total"`
    Page       int   `json:"page"`
    PerPage    int   `json:"per_page"`
    TotalPages int   `json:"total_pages"`
    HasNext    bool  `json:"has_next"`
    HasPrev    bool  `json:"has_prev"`
}
```

### Methods

| Method | الوصف |
|--------|-------|
| `.IsEmpty() bool` | فارغ؟ |
| `.Len() int` | عدد |
| `.First() (T, bool)` | أول |
| `.Last() (T, bool)` | آخر |
| `.ForEach(fn)` | تكرار |

---

## 🏢 Instance

| Method | الوصف |
|--------|-------|
| `NewInstance(db) *Instance` | جديد |
| `.DB() *gorm.DB` | GORM |
| `.Ping() error` | فحص |
| `.Close() error` | إغلاق |
| `.Configure(cfg) error` | ضبط |
| `.Transaction(ctx, fn) error` | معاملة |
| `.Query[T]() *QuerySet[T]` | QuerySet |
| `.Migrate(models...) error` | ترحيل |

### Global

| Function | الوصف |
|----------|-------|
| `GlobalInstance() *Instance` | Instance عام |
| `NewWith[T](inst) *QuerySet[T]` | QuerySet على Instance |

---

## 📚 Registry

| Function | الوصف |
|----------|-------|
| `Register[T](name) *QuerySet[T]` | تسجيل |
| `TryRegister[T](name) (*QuerySet[T], error)` | + error |
| `RegisterWith[T](name, inst)` | على Instance |
| `Lookup[T](name) (*QuerySet[T], bool)` | بحث |
| `MustLookup[T](name) *QuerySet[T]` | بحث + panic |
| `Has(name) bool` | مسجّل؟ |
| `Unregister(name)` | إزالة |
| `RegisteredNames() []string` | الأسماء |
| `RegisteredCount() int` | العدد |
| `ClearRegistry()` | مسح |

---

## 🌐 Context

| Function | الوصف |
|----------|-------|
| `WithDB(ctx, inst) ctx` | ربط |
| `WithGormDB(ctx, db) ctx` | ربط GORM |
| `DBFromContext(ctx) (*Instance, bool)` | استخراج |
| `FromContext[T](ctx) *QuerySet[T]` | QuerySet |

---

## 🚀 Advanced (gormz/advanced)

### Transactions

| Function | الوصف |
|----------|-------|
| `Begin(ctx, cfg) (*Tx, error)` | بدء |
| `WithTransaction(ctx, cfg, fn) error` | معاملة |
| `DefaultTxConfig() TxConfig` | إعداد |

### CTE

| Function | الوصف |
|----------|-------|
| `NewCTE[T](name, q) *CTE` | جديد |
| `NewRecursiveCTE[T](name, q) *CTE` | عودي |
| `With[T](ctes...) *CTEBuilder[T]` | builder |

### Locking

| Function | الوصف |
|----------|-------|
| `WithLock[T](q) *PessimisticQuery[T]` | قفل |
| `.ForUpdate()` | FOR UPDATE |
| `.ForShare()` | FOR SHARE |

### Joins

| Function | الوصف |
|----------|-------|
| `WithJoins[T](q) *JoinQuery[T]` | joins |
| `.Inner(table, on)` | INNER |
| `.Left(table, on)` | LEFT |
| `.Right(table, on)` | RIGHT |
| `.Cross(table)` | CROSS |

### Bulk

| Function | الوصف |
|----------|-------|
| `BulkInsert[T](ctx, items, cfg) error` | إدراج |
| `BulkUpsert[T](ctx, items, cfg) error` | upsert |
| `BulkUpdate[T](ctx, idCol, items) error` | تحديث |
| `BulkDeleteByIDs[T](ctx, ids, size) (int64, error)` | حذف |

### Aggregates

| Function | الوصف |
|----------|-------|
| `GroupBy[T](fields...) *AggregateQuery[T]` | GroupBy |
| `.Count(field, alias)` | Count |
| `.Sum(field, alias)` | Sum |
| `.Avg(field, alias)` | Avg |
| `.Having(sql, args...)` | Having |

### SubQuery

| Function | الوصف |
|----------|-------|
| `SubFrom[T](q, field) *SubQuery` | من QuerySet |
| `SubRaw(sql, args...) *SubQuery` | SQL |
| `In(field, sub)` | IN |
| `Exists(sub)` | EXISTS |

### Union

| Function | الوصف |
|----------|-------|
| `Union[T](queries...) *UnionQuery[T]` | Union |
| `UnionAll[T](queries...) *UnionQuery[T]` | Union All |

### Window

| Function | الوصف |
|----------|-------|
| `RowNumber(alias, partition...)` | ROW_NUMBER |
| `Rank(alias, orderBy, partition...)` | RANK |
| `Lag(field, offset, alias, partition...)` | LAG |
| `Lead(field, offset, alias, partition...)` | LEAD |
| `RunningSum(field, alias, orderBy, partition...)` | SUM OVER |

### Retry

| Function | الوصف |
|----------|-------|
| `Retry(ctx, cfg, fn) error` | إعادة |
| `DefaultRetryConfig() RetryConfig` | إعداد |
| `NewCircuitBreaker(max, timeout)` | CB |
| `ExponentialBackoff(init, max)` | Backoff |

### Batch

| Function | الوصف |
|----------|-------|
| `ProcessBatch[T](ctx, items, cfg, fn)` | دفعي |
| `Stream[T](ctx, size, fn) error` | streaming |
| `Parallel() *ParallelQuery` | متوازي |

---

## 🎯 Errors

```go
// Sentinel
ErrNotFound
ErrNotInitialized
ErrNilDB
ErrInvalidField
ErrInvalidQuery
ErrDangerousOperation
ErrAlreadyRegistered
ErrNotFoundInRegistry

// Typed
type NotFoundError { Model string; ID any }
type ValidationError { Field string; Reason string }

// Helpers
IsNotFound(err) bool
IsValidation(err) bool
```

---

## 📊 Config

```go
type Config struct {
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
    LogLevel        LogLevel
    SlowQuery       time.Duration
    PrepareStmt     bool
    DryRun          bool
}

// Presets
DefaultConfig() Config
DevelopmentConfig() Config
TestingConfig() Config
```

---

## 🎨 Types

```go
// Pagination
type PaginatedResult[T any]

// Lookups
type Lookup struct { SQL string; Args []any; Negate bool; Dialect Dialect }

// Hooks (interfaces)
type BeforeCreator interface { BeforeCreate() error }
type AfterCreator  interface { AfterCreate() error }
// ... (كل الـ hooks)

// Context
type contextKey struct
```

---

## 📖 Version

```go
const Version = "0.1.0"
const Author  = "Sanad Team"
const License = "MIT"
const URL     = "https://github.com/light-tech-dev/gormz"
```