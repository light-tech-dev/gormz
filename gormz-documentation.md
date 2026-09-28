╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║                         g o r m x                                ║
║                                                                  ║
║          Type-safe, Type-safe ORM for Go                   ║
║                     Built on top of GORM                         ║
║                                                                  ║
║                     Documentation v0.1.0                         ║
║                                                                  ║
║                     By Sanad Team                                ║
║                     License: MIT                                 ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝


═══════════════════════════════════════════════════════════════════
                        جدول المحتويات
═══════════════════════════════════════════════════════════════════

1.  مقدمة (Introduction)
2.  التثبيت والإعداد (Installation & Setup)
3.  البداية السريعة (Quick Start)
4.  المفاهيم الأساسية (Core Concepts)
5.  QuerySet — الاستعلامات
6.  Q Builder — الشروط المعقدة
7.  Modern-Style Lookups — الفلاتر المتقدمة
8.  Pagination — الترقيم
9.  Model Registry — سجل الموديلات
10. Multi-DB — دعم قواعد بيانات متعددة
11. Context — السياق
12. Errors — معالجة الأخطاء
13. Advanced Features — الميزات المتقدمة
    13.1  Subqueries
    13.2  CTEs
    13.3  Window Functions
    13.4  Unions
    13.5  Aggregations
    13.6  Joins
    13.7  Locking
    13.8  Transactions
    13.9  Savepoints
    13.10 Retry
    13.11 Bulk Operations
    13.12 Batch Processing
14. Security — الأمان
15. Best Practices — أفضل الممارسات
16. Testing — الاختبارات
17. Migration from GORM — الانتقال من GORM
18. Examples — أمثلة واقعية
19. FAQ — الأسئلة الشائعة
20. API Reference — مرجع API


═══════════════════════════════════════════════════════════════════
1. مقدمة (Introduction)
═══════════════════════════════════════════════════════════════════

ما هو gormz؟
─────────────

gormz هي مكتبة ORM (Object-Relational Mapping) للغة Go، مبنية فوق GORM
ومستوحاة من modern ORM. تقدّم واجهة برمجية نظيفة وآمنة من حيث الأنواع
(type-safe) باستخدام Go Generics.

الفلسفة
───────

الفكرة الأساسية وراء gormz هي:

    "اجعل الاستعلامات البسيطة بسيطة، والمعقدة ممكنة."

بدلًا من كتابة:

    var users []User
    db.Model(&User{}).
        Where("age > ?", 18).
        Where("active = ?", true).
        Order("created_at DESC").
        Limit(10).
        Find(&users)

يمكنك كتابة:

    users, _ := gormz.New[User]().
        Filter("age__gt", 18).
        Filter("active", true).
        OrderBy("-created_at").
        Limit(10).
        All()


الميزات الرئيسية
────────────────

✅ Type-safe       — Generic QuerySet[T] مع دعم كامل للـ IDE
✅ advanced lookups  — __gt، __in، __contains، __icontains، ...
✅ Immutable       — كل method يعيد نسخة جديدة (thread-safe)
✅ Multi-DB        — دعم قواعد بيانات متعددة عبر Instance
✅ Context-first   — دعم كامل لـ context.Context
✅ Registry        — تجنب import cycles
✅ Pagination      — دعم مدمج للترقيم
✅ Typed Errors    — NotFoundError، ValidationError، DangerousOperationError
✅ SQL Protection  — حماية من SQL injection
✅ Dangerous Ops   — حماية من عمليات خطرة (بدون conditions)

الميزات المتقدمة (advanced/):
✅ CTEs
✅ Window Functions
✅ Unions
✅ Subqueries (مع correlated)
✅ Aggregations (GroupBy, Having)
✅ Locking (Pessimistic + Optimistic)
✅ Retries (Exponential, Linear, Constant)
✅ Savepoints
✅ Bulk Operations
✅ Batch Processing
✅ Nested Transactions


المتطلبات
─────────

- Go 1.22+
- GORM v1.25+
- قاعدة بيانات مدعومة من GORM (SQLite, PostgreSQL, MySQL, ...)


═══════════════════════════════════════════════════════════════════
2. التثبيت والإعداد (Installation & Setup)
═══════════════════════════════════════════════════════════════════

التثبيت
───────

    go get github.com/light-tech-dev/gormz

الإعداد الأساسي
───────────────

    package main

    import (
        "log"

        "github.com/light-tech-dev/gormz"
        "gorm.io/driver/sqlite"
        "gorm.io/gorm"
    )

    type User struct {
        ID     uint   `gorm:"primaryKey"`
        Name   string `gorm:"size:255"`
        Email  string `gorm:"uniqueIndex"`
        Age    int
        Active bool
    }

    func main() {
        // 1. افتح الاتصال بـ GORM كما تفعل عادة
        db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
        if err != nil {
            log.Fatal(err)
        }

        // 2. اربط الاتصال بـ gormz
        gormz.SetDB(db)

        // 3. شغّل الترحيلات
        gormz.MustMigrate[User]()

        // 4. ابدأ الاستخدام
        user := &User{Name: "Ali", Email: "ali@test.com"}
        if err := gormz.New[User]().Create(user); err != nil {
            log.Fatal(err)
        }

        log.Printf("Created user with ID: %d", user.ID)
    }

إعدادات الاتصال
───────────────

لضبط إعدادات connection pool:

    // للإنتاج
    cfg := gormz.DefaultConfig()
    cfg.MaxOpenConns = 50
    cfg.MaxIdleConns = 10
    cfg.ConnMaxLifetime = 2 * time.Hour

    if err := gormz.Configure(cfg); err != nil {
        log.Fatal(err)
    }

    // للتطوير
    cfg = gormz.DevelopmentConfig()  // 10 connections، LogInfo

    // للاختبارات
    cfg = gormz.TestingConfig()  // 5 connections، Silent

فحص الاتصال
───────────

    if err := gormz.Ping(); err != nil {
        log.Fatal("DB is down:", err)
    }

    if !gormz.IsReady() {
        log.Fatal("DB not initialized")
    }

إغلاق الاتصال
─────────────

    defer gormz.Close()


═══════════════════════════════════════════════════════════════════
3. البداية السريعة (Quick Start)
═══════════════════════════════════════════════════════════════════

تعريف الموديل
─────────────

    type User struct {
        ID        uint      `gorm:"primaryKey"`
        Name      string    `gorm:"size:255;not null"`
        Email     string    `gorm:"uniqueIndex;size:255"`
        Age       int       `gorm:"default:0"`
        Active    bool      `gorm:"default:true"`
        CreatedAt time.Time
        UpdatedAt time.Time
    }

    func (User) TableName() string {
        return "users"
    }

CRUD — إنشاء
────────────

    // سجل واحد
    user := &User{Name: "Ali", Email: "ali@test.com", Age: 30}
    err := gormz.New[User]().Create(user)
    // user.ID الآن مملوء

    // عدة سجلات
    users := []User{
        {Name: "Ali", Email: "ali@test.com"},
        {Name: "Sara", Email: "sara@test.com"},
    }
    err = gormz.New[User]().CreateMany(users)

    // دفعات (لعدد كبير)
    err = gormz.New[User]().CreateInBatches(users, 1000)

CRUD — قراءة
────────────

    // بالـ ID
    user, err := gormz.New[User]().Get(1)

    // أول سجل مطابق
    user, err := gormz.New[User]().Filter("email", "ali@test.com").First()

    // كل النتائج
    users, err := gormz.New[User]().Filter("active", true).All()

    // عدد
    count, err := gormz.New[User]().Filter("active", true).Count()

    // وجود
    exists, err := gormz.New[User]().Filter("email", "ali@test.com").Exists()

CRUD — تحديث
────────────

    // حقل واحد
    err := gormz.New[User]().Update(1, "age", 31)

    // عدة حقول (يتطلب conditions)
    affected, err := gormz.New[User]().
        Filter("age__lt", 18).
        UpdateMany(map[string]any{"active": false})

CRUD — حذف
──────────

    // حذف واحد
    err := gormz.New[User]().Delete(1)

    // حذف جماعي (يتطلب conditions)
    affected, err := gormz.New[User]().
        Filter("active", false).
        DeleteMany()


═══════════════════════════════════════════════════════════════════
4. المفاهيم الأساسية (Core Concepts)
═══════════════════════════════════════════════════════════════════

QuerySet[T]
───────────

QuerySet[T] هو builder للاستعلامات. Generic على نوع الموديل T.

    q := gormz.New[User]()  // QuerySet[User]

كل method يعيد QuerySet جديدًا (immutable):

    base := gormz.New[User]().Filter("active", true)

    adults := base.Filter("age__gte", 18)     // نسخة جديدة
    young := base.Filter("age__lt", 18)       // نسخة أخرى

    // base لا يتأثر!

هذا يجعل QuerySet آمنًا للاستخدام المتزامن (thread-safe).

Panic vs Error
──────────────

معظم methods تُطلق panic عند الخطأ:

    gormz.New[User]().Filter("bad field", "x")  // panic

هذا مقصود للاستخدام السريع عندما تعرف أن الحقل صالح.

للمدخلات الخارجية (user input)، استخدم النسخة التي تبدأ بـ Try:

    q, err := gormz.New[User]().TryFilter(userInput, "x")
    if err != nil {
        // handle validation error
    }

Immutable by Design
───────────────────

كل method يعيد نسخة جديدة، لذا يمكنك:

    base := gormz.New[User]().Filter("active", true)

    // في goroutine 1
    go func() {
        users, _ := base.Filter("age__gte", 18).All()
    }()

    // في goroutine 2
    go func() {
        users, _ := base.Filter("age__lt", 18).All()
    }()

    // آمن تمامًا — لا race conditions


═══════════════════════════════════════════════════════════════════
5. QuerySet — الاستعلامات
═══════════════════════════════════════════════════════════════════

الفلاتر الأساسية
────────────────

    // مساواة
    q.Filter("name", "Ali")

    // مقارنات
    q.Filter("age__gt", 18)    // >
    q.Filter("age__gte", 18)   // >=
    q.Filter("age__lt", 65)    // <
    q.Filter("age__lte", 65)   // <=
    q.Filter("age__ne", 30)    // !=

    // سلاسل
    q.Filter("name__contains", "ali")     // LIKE '%ali%'
    q.Filter("name__icontains", "ALI")    // case-insensitive
    q.Filter("name__startswith", "A")     // LIKE 'A%'
    q.Filter("name__istartswith", "a")
    q.Filter("name__endswith", "m")       // LIKE '%m'
    q.Filter("name__iendswith", "M")

    // قوائم
    q.Filter("id__in", []int{1, 2, 3})
    q.Filter("id__notin", []int{4, 5})

    // null
    q.Filter("deleted_at__isnull", true)
    q.Filter("deleted_at__isnull", false)

    // نطاق
    q.Filter("age__between", []int{18, 65})

    // تواريخ
    q.Filter("created_at__year", 2025)
    q.Filter("created_at__month", 12)
    q.Filter("created_at__day", 25)

    // سلاسل مترابطة
    q.Filter("active", true).
      Filter("age__gte", 18).
      Filter("name__icontains", "ali")

Exclude
───────

    q.Exclude("name", "Ali")
    q.Exclude("status__in", []string{"banned", "deleted"})

Where (SQL خام)
───────────────

    // ⚠️ المسؤولية على المستخدم — لا يوجد validation
    q.Where("age > ? AND status = ?", 18, "active")

    // مفيد لـ expressions معقدة
    q.Where("LOWER(email) = LOWER(?)", userEmail)

الترتيب
───────

    q.OrderBy("name")              // name ASC
    q.OrderBy("-created_at")       // created_at DESC
    q.OrderBy("status", "-age")    // status ASC، age DESC

التحديد
───────

    q.Select("id", "name", "email")
    q.Omit("password", "secret")

    // SQL خام (للتجميعات)
    q.SelectRaw("status", "COUNT(*) as count")

العلاقات
────────

    q.Preload("Orders")
    q.Preload("Orders.Items")
    q.Preload("Profile", "status = ?", "active")

الحد والبداية
─────────────

    q.Limit(10)
    q.Offset(20)
    q.Page(2, 20)  // page 2، 20 items

Soft Delete
───────────

    q.WithDeleted()   // ضمّن المحذوفات
    q.OnlyDeleted()   // المحذوفة فقط

التجميعات
─────────

    q.Count()          // int64
    q.Exists()         // bool
    q.Sum("total")     // float64
    q.Avg("age")       // float64
    q.Min("age")       // float64
    q.Max("age")       // float64

GroupBy
───────

    var results []struct {
        Status string
        Count  int64
    }

    err := gormz.New[Order]().
        SelectRaw("status", "COUNT(*) as count").
        GroupBy("status").
        ScanInto(&results)

النتائج (Terminal methods)
──────────────────────────

    q.All()            // []T, error
    q.First()          // *T, error (يرجع NotFoundError)
    q.FirstOrNil()     // *T, error (nil إذا لم يوجد)
    q.Last()           // *T, error
    q.Take()           // *T, error (بدون ترتيب)
    q.Get(id)          // *T, error (بالـ ID)
    q.GetOrNil(id)     // *T, error
    q.Find(field, val) // *T, error
    q.FindOrNil(field, val)
    q.Pluck("name", &names)
    q.ScanInto(&dest)


═══════════════════════════════════════════════════════════════════
6. Q Builder — الشروط المعقدة
═══════════════════════════════════════════════════════════════════

Q Builder يسمح ببناء شروط AND/OR/NOT معقدة.

QOr — مجموعة OR
───────────────

    q := gormz.QOr(
        gormz.Eq("status", "active"),
        gormz.Eq("status", "pending"),
    )

    users, _ := gormz.New[User]().Q(q).All()
    // WHERE status = 'active' OR status = 'pending'

QAnd — مجموعة AND
─────────────────

    q := gormz.QAnd(
        gormz.Eq("active", true),
        gormz.Gt("age", 18),
    )

    // WHERE active = true AND age > 18

Nested — متداخل
───────────────

    // (status = 'active' OR status = 'pending') AND age > 18
    q := gormz.Qb().And(
        gormz.QOr(
            gormz.Eq("status", "active"),
            gormz.Eq("status", "pending"),
        ),
        gormz.Gt("age", 18),
    )

NOT
───

    q := gormz.QOr(
        gormz.Not(gormz.Eq("status", "deleted")),
        gormz.Eq("status", "active"),
    )

    // WHERE NOT (status = 'deleted') OR status = 'active'

Groups — مجموعات
────────────────

    q := gormz.Qb().
        And(gormz.Eq("x", 1)).
        AndGroup(
            gormz.Eq("a", 2),
            gormz.Eq("b", 3),
        )
    // WHERE x = 1 AND (a = 2 AND b = 3)

    q := gormz.Qb().
        And(gormz.Eq("x", 1)).
        OrGroup(
            gormz.Eq("a", 2),
            gormz.Eq("b", 3),
        )
    // WHERE x = 1 OR (a = 2 OR b = 3)

المتاحة
───────

    gormz.Eq(field, value)         // =
    gormz.Ne(field, value)         // !=
    gormz.Gt(field, value)         // >
    gormz.Gte(field, value)        // >=
    gormz.Lt(field, value)         // <
    gormz.Lte(field, value)        // <=
    gormz.Contains(field, value)   // LIKE '%value%'
    gormz.StartsWith(field, value) // LIKE 'value%'
    gormz.EndsWith(field, value)   // LIKE '%value'
    gormz.In(field, []any{...})    // IN (...)
    gormz.IsNull(field)            // IS NULL
    gormz.NotNull(field)           // IS NOT NULL
    gormz.Raw(sql, args...)        // SQL خام
    gormz.Not(clause)              // NOT (...)


═══════════════════════════════════════════════════════════════════
7. Modern-Style Lookups — الفلاتر المتقدمة
═══════════════════════════════════════════════════════════════════

gormz يدعم صيغة modern ORM الشهيرة: field__lookup

    q.Filter("age__gt", 18)

الجدول الكامل:

    المساواة:
        field              → field = value
        field__ne          → field != value

    المقارنات:
        field__gt          → field > value
        field__gte         → field >= value
        field__lt          → field < value
        field__lte         → field <= value

    السلاسل:
        field__contains    → field LIKE '%value%'
        field__icontains   → LOWER(field) LIKE '%value%'
        field__startswith  → field LIKE 'value%'
        field__istartswith → LOWER(field) LIKE 'value%'
        field__endswith    → field LIKE '%value'
        field__iendswith   → LOWER(field) LIKE '%value'

    القوائم:
        field__in          → field IN (values)
        field__notin       → field NOT IN (values)

    Null:
        field__isnull      → field IS NULL / IS NOT NULL

    النطاق:
        field__between     → field BETWEEN a AND b

    التواريخ:
        field__year        → YEAR(field) = value
        field__month       → MONTH(field) = value
        field__day         → DAY(field) = value

أمثلة عملية:

    // ابحث عن المستخدمين النشطين فوق 18 سنة
    q.Filter("active", true).Filter("age__gte", 18)

    // ابحث باسم يحتوي "ali" (case-insensitive)
    q.Filter("name__icontains", "ali")

    // ابحث في قائمة IDs
    q.Filter("id__in", []uint{1, 2, 3, 4, 5})

    // ابحث عن الطلبات في نطاق سعري
    q.Filter("total__between", []float64{100, 500})

    // المستخدمون الذين لم يسجلوا الدخول بعد
    q.Filter("last_login__isnull", true)

    // المستخدمون المسجلون في 2025
    q.Filter("created_at__year", 2025)


═══════════════════════════════════════════════════════════════════
8. Pagination — الترقيم
═══════════════════════════════════════════════════════════════════

Paginate — الترقيم الكامل
─────────────────────────

    page, err := gormz.New[User]().
        Filter("active", true).
        OrderBy("name").
        Paginate(1, 20)  // page 1، 20 items

    // page.Items       []User
    // page.Total       int64
    // page.Page        int  (1)
    // page.PerPage     int  (20)
    // page.TotalPages  int
    // page.HasNext     bool
    // page.HasPrev     bool

helpers

    page.IsEmpty()           // bool
    page.Len()               // int
    first, ok := page.First() // (T, bool)
    last, ok := page.Last()   // (T, bool)
    page.ForEach(func(i int, u User) {
        fmt.Println(i, u.Name)
    })

MapPage — تحويل النتائج
───────────────────────

    names := gormz.MapPage(page, func(u User) string {
        return u.Name
    })

FilterPage — تصفية النتائج
──────────────────────────

    actives := gormz.FilterPage(page, func(u User) bool {
        return u.Active
    })

Page type — helper
──────────────────

    p := gormz.NewPage(2, 20)
    // p.Number = 2
    // p.PerPage = 20
    // p.Offset() = 20

    q.Page(p.Number, p.PerPage)
    // أو
    q.Limit(p.PerPage).Offset(p.Offset())


═══════════════════════════════════════════════════════════════════
9. Model Registry — سجل الموديلات
═══════════════════════════════════════════════════════════════════

Model Registry يحل مشكلة import cycles.

المشكلة:
    package a → package b → package a  ❌

الحل:
    package a → registry
    package b → registry
    main → registry + a + b  ✅

الاستخدام:

    // في package a
    var Users = gormz.Register[User]("user")

    // في package b
    var Orders = gormz.Register[Order]("order")

    // في أي مكان
    Users.Filter("active", true).All()

    // أو من registry
    q := gormz.MustLookup[User]("user")
    users, _ := q.Filter("active", true).All()

الواجهات:

    gormz.Register[T](name)           // يpanic عند التكرار
    gormz.TryRegister[T](name)        // يعيد error
    gormz.Lookup[T](name)             // (q، ok)
    gormz.MustLookup[T](name)         // panic إذا لم يوجد
    gormz.Has(name)                   // bool
    gormz.Unregister(name)            // حذف
    gormz.RegisteredNames()           // []string
    gormz.RegisteredCount()           // int
    gormz.ClearRegistry()             // مسح الكل

Batch:

    entries := map[string]any{
        "user":  gormz.New[User](),
        "order": gormz.New[Order](),
    }
    err := gormz.RegisterBatch(entries)

    count := gormz.UnregisterBatch("user", "order")

Type-based:

    gormz.HasType[User]()             // bool
    gormz.LookupByType[User]()        // (q، ok)
    gormz.MustLookupByType[User]()    // panic
    gormz.NamesByType[User]()         // []string

Snapshot (للاختبارات):

    snap := gormz.TakeSnapshot()
    // ... تعديلات ...
    snap.Restore()
    snap.Merge()


═══════════════════════════════════════════════════════════════════
10. Multi-DB — دعم قواعد بيانات متعددة
═══════════════════════════════════════════════════════════════════

Instance يمثل اتصالًا مستقلًا:

    // اتصال أساسي (كتابة)
    writeDB, _ := gorm.Open(...)
    writeApp := gormz.NewInstance(writeDB)

    // اتصال ثانوي (قراءة)
    readDB, _ := gorm.Open(...)
    readApp := gormz.NewInstance(readDB)

    // استخدم كل واحد
    users, _ := writeApp.Query[User]().All()
    reports, _ := readApp.Query[Report]().All()

Global Instance:

    gormz.SetDB(db)
    app := gormz.GlobalInstance()
    users, _ := app.Query[User]().All()

    // أو
    users, _ := gormz.NewWith[User](app).All()

Transaction على Instance:

    err := app.Transaction(ctx, func(tx *gorm.DB) error {
        if err := tx.Create(&user).Error; err != nil {
            return err
        }
        return tx.Create(&order).Error
    })


═══════════════════════════════════════════════════════════════════
11. Context — السياق
═══════════════════════════════════════════════════════════════════

ربط DB بـ context:

    ctx := gormz.WithDB(context.Background(), app)
    users, _ := gormz.FromContext[User](ctx).All()

استخدام QuerySet مع context:

    q := gormz.New[User]().WithContext(ctx)
    users, _ := q.Filter("active", true).All()

استخراج DB:

    app, ok := gormz.DBFromContext(ctx)
    if ok {
        // استخدم app.DB()
    }

WithGormDB:

    ctx := gormz.WithGormDB(ctx, db)
    // يعمل مثل WithDB لكن يقبل *gorm.DB مباشرة

MustFromContext:

    q := gormz.MustFromContext[User](ctx)  // panic إذا لم يوجد


═══════════════════════════════════════════════════════════════════
12. Errors — معالجة الأخطاء
═══════════════════════════════════════════════════════════════════

gormz يستخدم typed errors:

    var (
        ErrNotFound           = ...
        ErrNotInitialized     = ...
        ErrNilDB              = ...
        ErrInvalidField       = ...
        ErrInvalidQuery       = ...
        ErrDangerousOperation = ...
        ErrAlreadyRegistered  = ...
        ErrNotFoundInRegistry = ...
    )

Typed Errors:

    NotFoundError          // السجل غير موجود
    ValidationError        // حقل غير صحيح
    DangerousOperationError // عملية بدون conditions

فحص الأخطاء:

    if gormz.IsNotFound(err) {
        // السجل غير موجود
    }

    if gormz.IsValidation(err) {
        // خطأ في الحقل
    }

    if gormz.IsDangerous(err) {
        // عملية خطرة
    }

استخراج التفاصيل:

    if ne, ok := gormz.AsNotFound(err); ok {
        fmt.Println(ne.Model)  // "User"
        fmt.Println(ne.ID)     // 42
    }

    if ve, ok := gormz.AsValidation(err); ok {
        fmt.Println(ve.Field)   // "bad_field"
        fmt.Println(ve.Reason)  // "invalid characters"
    }

إنشاء أخطاء:

    err := gormz.NewNotFoundError("User", 42)
    err := gormz.NewValidationError("field", "reason")
    err := gormz.NewDangerousError("DeleteMany", "requires conditions")

استخدام errors.Is/As:

    // يعمل مع errors.Is
    if errors.Is(err, gormz.ErrNotFound) { ... }

    // يعمل مع errors.As
    var ne *gormz.NotFoundError
    if errors.As(err, &ne) { ... }


═══════════════════════════════════════════════════════════════════
13. Advanced Features — الميزات المتقدمة
═══════════════════════════════════════════════════════════════════

كل الميزات أدناه موجودة في الحزمة الفرعية:

    import "github.com/light-tech-dev/gormz/advanced"


───────────────────────────────────────────────────────────────────
13.1 Subqueries
───────────────────────────────────────────────────────────────────

    // بناء subquery من QuerySet
    subq := advanced.SubFrom[Order](
        gormz.New[Order]().Select("user_id").Filter("status", "paid"),
        "user_id",
    )

    // استخدامه
    users, _ := gormz.New[User]().
        Where("id IN (" + subq.SQL + ")", subq.Args...).
        All()

    // أو باستخدام helper
    users, _ := gormz.New[User]().Q(
        advanced.In("id", subq),
    ).All()

المتاحة:

    advanced.SubFrom[T](q, field)          // SubQuery من QuerySet
    advanced.SubRaw(sql, args...)          // SubQuery من SQL
    advanced.In(field, sub)                // IN (subquery)
    advanced.NotIn(field, sub)             // NOT IN
    advanced.Exists(sub)                   // EXISTS
    advanced.NotExists(sub)                // NOT EXISTS
    advanced.GtSub(field, sub)             // field > (subquery)
    advanced.LtSub(field, sub)             // field < (subquery)
    advanced.EqSub(field, sub)             // field = (subquery)

Correlated Subquery:

    corr := advanced.NewCorrelated(
        "SELECT AVG(age) FROM users WHERE department_id = users.department_id",
    )

    users, _ := gormz.New[User]().
        Q(advanced.GtCorr("age", corr)).
        All()


───────────────────────────────────────────────────────────────────
13.2 CTEs
───────────────────────────────────────────────────────────────────

    // CTE بسيط
    activeUsers := advanced.NewCTE("active_users",
        gormz.New[User]().Filter("active", true))

    results, _ := advanced.With[Order](activeUsers).
        Query(gormz.New[Order]().
            Where("user_id IN (SELECT id FROM active_users)")).
        All()

Recursive CTE (شجرة):

    tree := advanced.NewRecursiveCTE[Category]("tree",
        gormz.New[Category]().Filter("id", rootID))

    tree.UnionRaw(
        "SELECT c.* FROM categories c JOIN tree t ON c.parent_id = t.id",
    )

    results, _ := advanced.With[Category](tree).
        Query(gormz.New[Category]()).
        All()


───────────────────────────────────────────────────────────────────
13.3 Window Functions
───────────────────────────────────────────────────────────────────

    // ROW_NUMBER
    results, _ := gormz.New[User]().
        Select("id", "name").
        Window(advanced.RowNumber("rn", "department_id")).
        All()

    // RANK
    Window(advanced.Rank("rnk", "salary DESC", "department_id"))

    // LAG / LEAD
    Window(advanced.Lag("salary", 1, "prev_salary", "department_id"))
    Window(advanced.Lead("salary", 1, "next_salary", "department_id"))

    // Running aggregates
    Window(advanced.RunningSum("amount", "running_total", "date"))
    Window(advanced.RunningAvg("amount", "running_avg", "date"))

    // Percentiles
    Window(advanced.PercentRank("pct", "salary DESC"))
    Window(advanced.NTile(4, "quartile", "salary DESC"))

المتاحة:

    advanced.RowNumber(alias، partition...)
    advanced.Rank(alias، orderBy، partition...)
    advanced.DenseRank(alias، orderBy، partition...)
    advanced.Lag(field، offset، alias، partition...)
    advanced.Lead(field، offset، alias، partition...)
    advanced.RunningSum(field، alias، orderBy، partition...)
    advanced.RunningCount(alias، orderBy، partition...)
    advanced.RunningAvg(field، alias، orderBy، partition...)
    advanced.SumOver(field، alias، partition...)
    advanced.CountOver(alias، partition...)
    advanced.AvgOver(field، alias، partition...)
    advanced.NTile(n، alias، orderBy، partition...)
    advanced.FirstValue(field، alias، orderBy، partition...)
    advanced.LastValue(field، alias، orderBy، partition...)
    advanced.NthValue(field، n، alias، orderBy، partition...)
    advanced.PercentRank(alias، orderBy، partition...)
    advanced.CumeDist(alias، orderBy، partition...)


───────────────────────────────────────────────────────────────────
13.4 Unions
───────────────────────────────────────────────────────────────────

    // UNION (يحذف المكرر)
    q1 := gormz.New[User]().Filter("status", "active")
    q2 := gormz.New[User]().Filter("status", "pending")

    users, _ := advanced.Union(q1, q2).All()

    // UNION ALL (يحتفظ بالمكرر)
    users, _ := advanced.UnionAll(q1, q2).All()

    // مع OrderBy/Limit
    users, _ := advanced.Union(q1, q2).
        OrderBy("-created_at").
        Limit(10).
        All()


───────────────────────────────────────────────────────────────────
13.5 Aggregations
───────────────────────────────────────────────────────────────────

    results, _ := advanced.GroupBy[Order]("user_id").
        Count("id", "order_count").
        Sum("total", "total_amount").
        Avg("total", "avg_amount").
        Having("COUNT(id) > ?", 5).
        OrderBy("-total_amount").
        All()

    // أو ScanInto
    type UserStats struct {
        UserID      uint
        OrderCount  int64
        TotalAmount float64
    }

    var stats []UserStats
    err := advanced.GroupBy[Order]("user_id").
        Count("id", "order_count").
        Sum("total", "total_amount").
        ScanInto(&stats)

المتاحة:

    advanced.GroupBy[T](fields...)        // يبدأ aggregation query
    .Count(field، alias)
    .CountDistinct(field، alias)
    .Sum(field، alias)
    .Avg(field، alias)
    .Min(field، alias)
    .Max(field، alias)
    .Filter(field، value)
    .Where(sql، args...)
    .Having(sql، args...)
    .OrderBy(fields...)
    .Limit(n)
    .Offset(n)
    .All()
    .ScanInto(dest)
    .CountGroups()

Convenience:

    advanced.Distinct[User]("email")           // []any
    advanced.CountDistinctValues[User]("age")  // int64
    advanced.GroupConcat[User]("name", ",")    // string (SQLite/MySQL)


───────────────────────────────────────────────────────────────────
13.6 Joins
───────────────────────────────────────────────────────────────────

    users, _ := gormz.New[User]().
        InnerJoin("orders", "orders.user_id = users.id").
        Filter("orders.status", "paid").
        Select("users.id", "users.name", "COUNT(orders.id) as order_count").
        GroupBy("users.id").
        All()

أنواع joins:

    .InnerJoin(table, on)  // INNER JOIN
    .LeftJoin(table, on)   // LEFT JOIN
    .RightJoin(table, on)  // RIGHT JOIN
    .CrossJoin(table)      // CROSS JOIN
    .Join(type, table, on) // نوع مخصص


───────────────────────────────────────────────────────────────────
13.7 Locking
───────────────────────────────────────────────────────────────────

Pessimistic Locking:

    advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
        func(tx *advanced.Tx) error {
            user, err := advanced.WithLock[User](tx.Query[User]()).
                ForUpdate().
                Get(1)
            if err != nil {
                return err
            }

            user.Balance -= 100
            return tx.Query[User]().Save(user)
        })

أنواع الأقفال:

    .ForUpdate()        // FOR UPDATE
    .ForShare()         // FOR SHARE
    .ForNoKeyUpdate()   // FOR NO KEY UPDATE (Postgres)
    .ForKeyShare()      // FOR KEY SHARE (Postgres)

Optimistic Locking:

    type Product struct {
        ID      uint
        Name    string
        Version int `gorm:"default:1"`
    }

    // GORM يدعم optimistic locking تلقائيًا إذا وجد الحقل "Version"


───────────────────────────────────────────────────────────────────
13.8 Transactions
───────────────────────────────────────────────────────────────────

Simple Transaction:

    err := gormz.Transaction(ctx, func(tx *gorm.DB) error {
        if err := tx.Create(&user).Error; err != nil {
            return err
        }
        return tx.Create(&order).Error
    })

Advanced Transaction (مع retry):

    err := advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
        func(tx *advanced.Tx) error {
            if err := tx.Query[User]().Create(&user); err != nil {
                return err
            }
            return tx.Query[Order]().Create(&order)
        })

Custom Config:

    cfg := advanced.TxConfig{
        Isolation:  advanced.IsolationSerializable,
        ReadOnly:   false,
        Timeout:    30 * time.Second,
        Retries:    3,
        RetryDelay: 100 * time.Millisecond,
    }

    err := advanced.WithTransaction(ctx, cfg, func(tx *advanced.Tx) error {
        // ...
    })

Manual Control:

    tx, err := advanced.Begin(ctx, advanced.DefaultTxConfig())
    if err != nil {
        return err
    }
    defer tx.RollbackIfActive()

    if err := tx.Query[User]().Create(&user); err != nil {
        return err
    }

    return tx.Commit()

Nested Transactions (Savepoints):

    advanced.WithTransaction(ctx, cfg, func(tx *advanced.Tx) error {
        // العملية الرئيسية
        if err := tx.Query[User]().Create(&user); err != nil {
            return err
        }

        // عملية متداخلة (savepoint)
        err := tx.Nested(func(inner *advanced.Tx) error {
            return inner.Query[Order]().Create(&order)
        })
        // إذا فشلت، تُلغى فقط العملية الداخلية

        return err
    })


───────────────────────────────────────────────────────────────────
13.9 Savepoints
───────────────────────────────────────────────────────────────────

    advanced.WithTransaction(ctx, cfg, func(tx *advanced.Tx) error {
        // Savepoint يدوي
        if err := tx.Savepoint("step1"); err != nil {
            return err
        }

        if err := tx.Query[User]().Create(&user); err != nil {
            // التراجع إلى step1
            if rbErr := tx.RollbackTo("step1"); rbErr != nil {
                return rbErr
            }
        }

        // حذف savepoint
        return tx.ReleaseSavepoint("step1")
    })

Savepoint Object:

    sp := advanced.NewSavepoint("my_sp", tx.DB())
    if err := sp.Create(); err != nil { ... }
    // ...
    if err := sp.Rollback(); err != nil { ... }
    // أو
    if err := sp.Release(); err != nil { ... }

SavepointStack:

    stack := advanced.NewSavepointStack(tx.DB())
    if _, err := stack.Push("step1"); err != nil { ... }
    if _, err := stack.Push("step2"); err != nil { ... }

    // rollback للأخير
    stack.RollbackLast()

    // أو إلى واحد محدد
    stack.RollbackTo("step1")


───────────────────────────────────────────────────────────────────
13.10 Retry
───────────────────────────────────────────────────────────────────

    // Retry بسيط
    err := advanced.Retry(ctx, advanced.DefaultRetryConfig(),
        func() error {
            return gormz.New[User]().Create(&user)
        })

    // Retry مع backoff
    backoff := advanced.ExponentialBackoff(100*time.Millisecond, 10*time.Second)

    err := advanced.RetryWithBackoff(ctx, 5, backoff, func() error {
        return gormz.New[User]().Create(&user)
    })

    // Retry DB operation
    err := advanced.RetryDB(ctx, func(db *gorm.DB) error {
        return db.Create(&user).Error
    })

Backoff Strategies:

    advanced.ExponentialBackoff(initial، max)  // 100ms، 200ms، 400ms، ...
    advanced.LinearBackoff(step، max)          // 100ms، 200ms، 300ms، ...
    advanced.ConstantBackoff(delay)            // 500ms، 500ms، 500ms، ...

Custom Retry Config:

    cfg := advanced.RetryConfig{
        MaxAttempts:  5,
        InitialDelay: 50 * time.Millisecond,
        MaxDelay:     5 * time.Second,
        Multiplier:   2.0,
        Jitter:       0.1,
        RetryIf:      func(err error) bool {
            return errors.Is(err، sql.ErrConnDone)
        },
    }


───────────────────────────────────────────────────────────────────
13.11 Bulk Operations
───────────────────────────────────────────────────────────────────

    // Bulk Insert
    users := make([]User, 100000)
    err := advanced.BulkInsert[User](ctx, users, advanced.BulkConfig{
        BatchSize: 1000,
    })

    // Bulk Upsert (INSERT ON CONFLICT)
    err := advanced.BulkUpsert[User](ctx, users, advanced.BulkConfig{
        ConflictColumns: []string{"email"},
        UpdateColumns:   []string{"name", "age"},
    })

    // Bulk Update (بقيم مختلفة لكل سجل)
    updates := []advanced.UpdateItem[User]{
        {ID: 1, Values: map[string]any{"age": 31}},
        {ID: 2, Values: map[string]any{"age": 26}},
    }
    err := advanced.BulkUpdate[User](ctx, "id", updates)

    // Bulk Delete
    deleted, err := advanced.BulkDeleteByIDs[User](ctx, ids, 1000)
    deleted, err := advanced.BulkDeleteWhere[User](ctx, "age < ?", 18)

    // Bulk Count
    count, err := advanced.BulkCount[User](ctx, "active = ?", true)


───────────────────────────────────────────────────────────────────
13.12 Batch Processing
───────────────────────────────────────────────────────────────────

    // Parallel processing
    users := []User{...}

    result := advanced.ProcessBatch(ctx, users,
        advanced.DefaultBatchConfig(),
        func(batch []User) error {
            return gormz.New[User]().CreateMany(batch)
        })

    fmt.Println(result.Total)    // 1000
    fmt.Println(result.Success)  // 998
    fmt.Println(result.Failed)   // 2
    fmt.Println(result.Errors)   // [errors...]

    // Streaming from DB
    err := advanced.Stream[User](ctx, 1000, func(batch []User) error {
        for _, u := range batch {
            // process u
        }
        return nil
    })

    // Parallel tasks
    p := advanced.Parallel().WithWorkers(4)
    p.Add(func() error { return task1() })
    p.Add(func() error { return task2() })
    p.Add(func() error { return task3() })

    err := p.Run(ctx)


═══════════════════════════════════════════════════════════════════
14. Security — الأمان
═══════════════════════════════════════════════════════════════════

gormz مصمّمة مع الأمان في الاعتبار.

حماية من SQL Injection
──────────────────────

كل اسم حقل يُتحقق منه:

    // ✅ آمن — يتحقق من الحقل
    q.Filter("name", "Ali")

    // ❌ panic — حقل غير صحيح
    q.Filter("name; DROP TABLE users", "Ali")

للمدخلات الخارجية:

    q, err := gormz.New[User]().TryFilter(userInput, value)
    if err != nil {
        // handle validation error
    }

الحماية من العمليات الخطرة
──────────────────────────

    // ❌ خطأ — لا conditions
    _, err := gormz.New[User]().DeleteMany()
    // err: DangerousOperationError

    // ✅ يعمل — مع conditions
    _, err := gormz.New[User]().Filter("active", false).DeleteMany()

نفس الشيء لـ UpdateMany:

    // ❌ خطأ
    gormz.New[User]().UpdateMany(map[string]any{"active": false})

    // ✅ يعمل
    gormz.New[User]().
        Filter("age__lt", 18).
        UpdateMany(map[string]any{"active": false})

SQL الخام
─────────

    // ⚠️ مسؤوليتك — لا يوجد validation
    q.Where("age > ? AND status = ?", 18, "active")

    // استخدم parameterized queries دائمًا — لا تدمج strings
    // ❌ خطأ:
    q.Where(fmt.Sprintf("name = '%s'", userInput))

    // ✅ صحيح:
    q.Where("name = ?", userInput)


═══════════════════════════════════════════════════════════════════
15. Best Practices — أفضل الممارسات
═══════════════════════════════════════════════════════════════════

1. استخدم Try* للمدخلات الخارجية
─────────────────────────────────

    // للمستخدم
    q, err := gormz.New[User]().TryFilter(userInput, value)

    // للكود الداخلي (تثق بالحقل)
    q := gormz.New[User]().Filter("name", "Ali")

2. استخدم Context دائمًا
─────────────────────────

    // ✅
    gormz.New[User]().WithContext(ctx).Filter("active", true).All()

    // ❌ (يفتقد cancellation)
    gormz.New[User]().Filter("active", true).All()

3. Prefer Batch Operations
──────────────────────────

    // ❌ بطيء
    for _, u := range users {
        gormz.New[User]().Create(&u)
    }

    // ✅ سريع
    gormz.New[User]().CreateInBatches(users, 1000)

4. استخدم Registry لتجنب Cycles
───────────────────────────────

    var Users = gormz.Register[User]("user")

5. Explicit is Better
─────────────────────

    // ❌ غامض
    gormz.New[User]().Filter("active", true).All()

    // ✅ واضح
    activeUsers, err := gormz.New[User]().
        Filter("active", true).
        OrderBy("name").
        All()

6. استخدم Transactions للعمليات المتعددة
────────────────────────────────────────

    err := advanced.WithTransaction(ctx, cfg,
        func(tx *advanced.Tx) error {
            if err := tx.Query[User]().Create(&user); err != nil {
                return err
            }
            if err := tx.Query[Order]().Create(&order); err != nil {
                return err
            }
            return nil
        })

7. Handle Errors Properly
─────────────────────────

    user, err := gormz.New[User]().Get(1)
    if err != nil {
        if gormz.IsNotFound(err) {
            return ErrUserNotFound
        }
        return err
    }

8. استخدم Indexes
─────────────────

    type User struct {
        ID    uint   `gorm:"primaryKey"`
        Email string `gorm:"uniqueIndex;size:255"`
        Name  string `gorm:"index;size:255"`
    }


═══════════════════════════════════════════════════════════════════
16. Testing — الاختبارات
═══════════════════════════════════════════════════════════════════

Setup اختبار:

    func setupTestDB(t *testing.T) {
        t.Helper()

        db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
            Logger: logger.Default.LogMode(logger.Silent),
        })
        require.NoError(t, err)
        require.NoError(t, db.AutoMigrate(&User{}))

        gormz.SetDB(db)
        gormz.ClearRegistry()

        t.Cleanup(func() {
            if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
                _ = sqlDB.Close()
            }
            gormz.ResetDB()
            gormz.ClearRegistry()
        })
    }

Test CRUD:

    func TestUserCRUD(t *testing.T) {
        setupTestDB(t)

        // Create
        user := &User{Name: "Ali", Email: "ali@test.com"}
        require.NoError(t, gormz.New[User]().Create(user))
        assert.NotZero(t, user.ID)

        // Read
        got, err := gormz.New[User]().Get(user.ID)
        require.NoError(t, err)
        assert.Equal(t, "Ali", got.Name)

        // Update
        require.NoError(t, gormz.New[User]().Update(user.ID, "age", 30))

        // Delete
        require.NoError(t, gormz.New[User]().Delete(user.ID))

        _, err = gormz.New[User]().Get(user.ID)
        assert.True(t, gormz.IsNotFound(err))
    }

Test Filter:

    func TestFilter(t *testing.T) {
        setupTestDB(t)

        users := []User{
            {Name: "Ali", Age: 30, Active: true},
            {Name: "Sara", Age: 25, Active: true},
            {Name: "Omar", Age: 17, Active: false},
        }
        require.NoError(t, gormz.New[User]().CreateMany(users))

        result, err := gormz.New[User]().
            Filter("active", true).
            Filter("age__gte", 18).
            All()

        require.NoError(t, err)
        assert.Len(t, result, 2)
    }

Benchmarks:

    func BenchmarkCreate(b *testing.B) {
        setupBenchDB(b)

        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            user := &User{
                Name:  "User",
                Email: fmt.Sprintf("user%d@test.com", i),
            }
            if err := gormz.New[User]().Create(user); err != nil {
                b.Fatal(err)
            }
        }
    }


═══════════════════════════════════════════════════════════════════
17. Migration from GORM — الانتقال من GORM
═══════════════════════════════════════════════════════════════════

جدول التحويل:

    GORM                                  gormz
    ─────────────────────────────────────────────────────────────
    db.Where("age > ?", 18).Find(&u)      gormz.New[U]().Filter("age__gt", 18).All()
    db.First(&user, 1)                    gormz.New[U]().Get(1)
    db.First(&user, "email = ?", "x")     gormz.New[U]().Find("email", "x")
    db.Create(&user)                      gormz.New[U]().Create(&user)
    db.Save(&user)                        gormz.New[U]().Save(&user)
    db.Delete(&user, 1)                   gormz.New[U]().Delete(1)
    db.Model(&u).Update("x", "y")         gormz.New[U]().Update(id, "x", "y")
    db.Model(&u).Updates(m)               gormz.New[U]().Filter(...).UpdateMany(m)
    db.Model(&u).Count(&c)                gormz.New[U]().Count()

مثال كامل قبل/بعد:

    // ─── GORM ───
    var users []User
    query := db.Model(&User{}).Where("active = ?", true)
    if search != "" {
        query = query.Where("name LIKE ?", "%"+search+"%")
    }
    if minAge > 0 {
        query = query.Where("age >= ?", minAge)
    }
    err := query.Order("created_at DESC").
        Limit(20).
        Offset(0).
        Preload("Roles").
        Find(&users).Error

    // ─── gormz ───
    q := gormz.New[User]().Filter("active", true)
    if search != "" {
        q = q.Filter("name__contains", search)
    }
    if minAge > 0 {
        q = q.Filter("age__gte", minAge)
    }
    users, err := q.OrderBy("-created_at").
        Limit(20).
        Preload("Roles").
        All()

المزايا:

    ✅ Type-safe — لا string concat
    ✅ Immutable — لا مشاكل في shared state
    ✅ أقصر وأوضح
    ✅ حماية من SQL injection
    ✅ نفس GORM تحت الغطاء


═══════════════════════════════════════════════════════════════════
18. Examples — أمثلة واقعية
═══════════════════════════════════════════════════════════════════

مثال 1: نظام مستخدمين كامل
───────────────────────────

    package main

    import (
        "context"
        "log"
        "time"

        "github.com/light-tech-dev/gormz"
        "github.com/light-tech-dev/gormz/advanced"
        "gorm.io/driver/sqlite"
        "gorm.io/gorm"
    )

    type User struct {
        ID        uint      `gorm:"primaryKey"`
        Username  string    `gorm:"uniqueIndex;size:100"`
        Email     string    `gorm:"uniqueIndex;size:255"`
        FullName  string    `gorm:"size:255"`
        IsActive  bool      `gorm:"default:true"`
        LastLogin *time.Time
        CreatedAt time.Time
        UpdatedAt time.Time
    }

    type UserService struct {
        ctx context.Context
    }

    func (s *UserService) Create(username, email, fullName string) (*User, error) {
        user := &User{
            Username: username,
            Email:    email,
            FullName: fullName,
            IsActive: true,
        }

        if err := gormz.New[User]().
            WithContext(s.ctx).
            Create(user); err != nil {
            return nil, err
        }

        return user, nil
    }

    func (s *UserService) GetByID(id uint) (*User, error) {
        return gormz.New[User]().
            WithContext(s.ctx).
            Get(id)
    }

    func (s *UserService) List(page, perPage int) (*gormz.PaginatedResult[User], error) {
        return gormz.New[User]().
            WithContext(s.ctx).
            Filter("is_active", true).
            OrderBy("-created_at").
            Paginate(page, perPage)
    }

    func (s *UserService) Search(query string) ([]User, error) {
        return gormz.New[User]().
            WithContext(s.ctx).
            Filter("is_active", true).
            Q(gormz.QOr(
                gormz.Contains("username", query),
                gormz.Contains("email", query),
                gormz.Contains("full_name", query),
            )).
            Limit(20).
            All()
    }

    func (s *UserService) Deactivate(id uint) error {
        _, err := gormz.New[User]().
            WithContext(s.ctx).
            Filter("id", id).
            UpdateMany(map[string]any{"is_active": false})
        return err
    }

    func (s *UserService) RecordLogin(id uint) error {
        now := time.Now()
        return gormz.New[User]().
            WithContext(s.ctx).
            Update(id, "last_login", &now)
    }

مثال 2: نظام طلبات مع Transactions
──────────────────────────────────

    type Order struct {
        ID         uint      `gorm:"primaryKey"`
        UserID     uint      `gorm:"index"`
        Total      float64   `gorm:"type:decimal(10,2)"`
        Status     string    `gorm:"size:50"`
        CreatedAt  time.Time
    }

    type OrderItem struct {
        ID         uint    `gorm:"primaryKey"`
        OrderID    uint    `gorm:"index"`
        ProductID  uint    `gorm:"index"`
        Quantity   int
        UnitPrice  float64 `gorm:"type:decimal(10,2)"`
    }

    type OrderService struct {
        ctx context.Context
    }

    func (s *OrderService) CreateOrder(userID uint, items []OrderItem) (*Order, error) {
        var order *Order

        err := advanced.WithTransaction(s.ctx, advanced.DefaultTxConfig(),
            func(tx *advanced.Tx) error {
                // حساب الإجمالي
                var total float64
                for _, item := range items {
                    total += float64(item.Quantity) * item.UnitPrice
                }

                // إنشاء الطلب
                order = &Order{
                    UserID: userID,
                    Total:  total,
                    Status: "pending",
                }

                if err := tx.Query[Order]().Create(order); err != nil {
                    return err
                }

                // ربط العناصر
                for i := range items {
                    items[i].OrderID = order.ID
                }

                if err := tx.Query[OrderItem]().CreateMany(items); err != nil {
                    return err
                }

                return nil
            })

        if err != nil {
            return nil, err
        }

        return order, nil
    }

مثال 3: تقارير مع Aggregations
───────────────────────────────

    func GetUserOrderStats(ctx context.Context) ([]UserStats, error) {
        type UserStats struct {
            UserID      uint
            OrderCount  int64
            TotalAmount float64
            AvgAmount   float64
        }

        var stats []UserStats

        err := advanced.GroupBy[Order]("user_id").
            Count("id", "order_count").
            Sum("total", "total_amount").
            Avg("total", "avg_amount").
            Filter("status", "paid").
            Having("COUNT(id) > ?", 5).
            OrderBy("-total_amount").
            ScanInto(&stats)

        return stats, err
    }

مثال 4: شجرة فئات مع Recursive CTE
───────────────────────────────────

    type Category struct {
        ID       uint   `gorm:"primaryKey"`
        Name     string `gorm:"size:255"`
        ParentID *uint  `gorm:"index"`
    }

    func GetCategoryTree(ctx context.Context, rootID uint) ([]Category, error) {
        tree := advanced.NewRecursiveCTE[Category]("tree",
            gormz.New[Category]().Filter("id", rootID))

        tree.UnionRaw(
            "SELECT c.* FROM categories c JOIN tree t ON c.parent_id = t.id",
        )

        results, err := advanced.With[Category](tree).
            Query(gormz.New[Category]()).
            All()

        return results, err
    }

مثال 5: Bulk Import مع Validation
──────────────────────────────────

    func ImportUsers(ctx context.Context, filepath string) error {
        file, err := os.Open(filepath)
        if err != nil {
            return err
        }
        defer file.Close()

        reader := csv.NewReader(file)
        reader.Read() // skip header

        var users []User
        for {
            record, err := reader.Read()
            if err == io.EOF {
                break
            }
            if err != nil {
                return err
            }

            users = append(users, User{
                Username: record[0],
                Email:    record[1],
                FullName: record[2],
                IsActive: true,
            })
        }

        return advanced.BulkInsert[User](ctx, users, advanced.BulkConfig{
            BatchSize: 1000,
        })
    }


═══════════════════════════════════════════════════════════════════
19. FAQ — الأسئلة الشائعة
═══════════════════════════════════════════════════════════════════

Q: هل gormz بديل لـ GORM؟
A: لا. gormz مكتبة مبنية فوق GORM. تستخدم GORM تحت الغطاء، لكن
   تقدّم API أنظف. يمكنك استخدام الاثنين معًا.

Q: هل يمكنني استخدام GORM hooks؟
A: نعم! gormz تستخدم GORM تحت الغطاء، لذا كل hooks GORM تعمل:

   func (u *User) BeforeCreate(tx *gorm.DB) error {
       u.Name = strings.TrimSpace(u.Name)
       return nil
   }

Q: كيف أتعامل مع الأخطاء؟
A: استخدم typed errors:

   if gormz.IsNotFound(err) { ... }
   if ne, ok := gormz.AsNotFound(err); ok { ... }

Q: هل QuerySet آمن للاستخدام المتزامن؟
A: نعم. QuerySet immutable — كل method يعيد نسخة جديدة.

Q: كيف أستخدم مع PostgreSQL؟
A: gormz يدعم أي driver يدعمه GORM:

   import "gorm.io/driver/postgres"
   db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
   gormz.SetDB(db)

Q: ما الفرق بين Filter و Where؟
A:
   - Filter: validation + advanced lookups (آمن)
   - Where: SQL خام (مسؤوليتك)

Q: كيف أعمل pagination؟
A:

   page, _ := gormz.New[User]().Paginate(1, 20)
   // page.Items، page.Total، page.Page، ...

Q: كيف أستخدم transactions؟
A:

   err := advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
       func(tx *advanced.Tx) error {
           return tx.Query[User]().Create(&user)
       })

Q: هل هناك دعم لـ multi-tenancy؟
A: يمكن استخدام Instance منفصل لكل tenant:

   tenantDB := gormz.NewInstance(getTenantDB(tenantID))
   users, _ := tenantDB.Query[User]().All()

Q: كيف أنتقل من GORM؟
A: راجع قسم "Migration from GORM" أعلاه. العملية تدريجية —
   يمكنك استخدام الاثنين في نفس المشروع.

Q: هل المكتبة مُختبَرة؟
A: نعم — ~90% coverage. راجع مجلد `gormx_test/` و `advanced_test/`.

Q: كيف أساهم؟
A: راجع CONTRIBUTING.md. نرحّب بكل مساهمة!

Q: ما الترخيص؟
A: MIT — استخدمها بحرية في المشاريع التجارية.


═══════════════════════════════════════════════════════════════════
20. API Reference — مرجع API
═══════════════════════════════════════════════════════════════════

إعداد (Setup)
─────────────

    gormz.SetDB(db *gorm.DB)              — ربط الاتصال
    gormz.DB() *gorm.DB                    — الاتصال الحالي
    gormz.IsReady() bool                   — فحص الجاهزية
    gormz.ResetDB()                        — إعادة تعيين
    gormz.Configure(cfg Config) error      — إعدادات pool
    gormz.Ping() error                     — فحص الاتصال
    gormz.Close() error                    — إغلاق

ترحيل (Migrate)
───────────────

    gormz.Migrate[T]() error               — ترحيل موديل
    gormz.MustMigrate[T]()                 — ترحيل + panic
    gormz.MigrateAll(models...) error      — ترحيل متعدد
    gormz.MustMigrateAll(models...)        — panic
    gormz.DropTable[T]() error             — حذف جدول
    gormz.HasTable[T]() bool               — فحص وجود جدول

QuerySet — الإنشاء
──────────────────

    gormz.New[T]() *QuerySet[T]            — QuerySet جديد
    gormz.NewWith[T](i *Instance)          — من Instance
    gormz.FromContext[T](ctx)              — من context
    gormz.MustFromContext[T](ctx)          — + panic

QuerySet — الفلاتر
──────────────────

    .Filter(field, value) *QuerySet[T]
    .TryFilter(field, value) (*QuerySet[T], error)
    .Exclude(field, value) *QuerySet[T]
    .TryExclude(field, value) (*QuerySet[T], error)
    .Where(sql, args...) *QuerySet[T]
    .Q(q *Q) *QuerySet[T]

QuerySet — الترتيب
──────────────────

    .OrderBy(fields...) *QuerySet[T]
    .TryOrderBy(fields...) (*QuerySet[T], error)

QuerySet — الحد
───────────────

    .Limit(n) *QuerySet[T]
    .Offset(n) *QuerySet[T]
    .Page(page, perPage) *QuerySet[T]

QuerySet — التحديد
──────────────────

    .Select(fields...) *QuerySet[T]
    .TrySelect(fields...) (*QuerySet[T], error)
    .SelectRaw(exprs...) *QuerySet[T]
    .Omit(fields...) *QuerySet[T]

QuerySet — العلاقات
───────────────────

    .Preload(relations...) *QuerySet[T]

QuerySet — Soft Delete
──────────────────────

    .WithDeleted() *QuerySet[T]
    .OnlyDeleted() *QuerySet[T]

QuerySet — Grouping
───────────────────

    .GroupBy(fields...) *QuerySet[T]
    .Having(sql, args...) *QuerySet[T]
    .Distinct() *QuerySet[T]
    .DistinctOn(fields...) *QuerySet[T]

QuerySet — Joins
────────────────

    .Join(joinType, table, on) *QuerySet[T]
    .InnerJoin(table, on) *QuerySet[T]
    .LeftJoin(table, on) *QuerySet[T]
    .RightJoin(table, on) *QuerySet[T]
    .CrossJoin(table) *QuerySet[T]

QuerySet — Window
─────────────────

    .Window(exprs...) *QuerySet[T]

QuerySet — Context
──────────────────

    .WithContext(ctx) *QuerySet[T]
    .Context() context.Context

QuerySet — النتائج
──────────────────

    .All() ([]T, error)
    .First() (*T, error)
    .FirstOrNil() (*T, error)
    .Last() (*T, error)
    .Take() (*T, error)
    .Get(id) (*T, error)
    .GetOrNil(id) (*T, error)
    .Find(field, val) (*T, error)
    .FindOrNil(field, val) (*T, error)
    .Count() (int64, error)
    .Exists() (bool, error)
    .Pluck(field, dest) error
    .ScanInto(dest) error
    .Sum(field) (float64, error)
    .Avg(field) (float64, error)
    .Min(field) (float64, error)
    .Max(field) (float64, error)
    .Paginate(page, perPage) (*PaginatedResult[T], error)

QuerySet — الكتابة
──────────────────

    .Create(item *T) error
    .CreateMany(items []T) error
    .CreateInBatches(items []T, batchSize) error
    .Save(item *T) error
    .Update(id, field, value) error
    .UpdateColumn(id, field, value) error
    .UpdateMany(values map[string]any) (int64, error)
    .Delete(id) error
    .DeleteMany() (int64, error)
    .HardDelete(id) error
    .Restore(id) error
    .RestoreAll() (int64, error)

QuerySet — Debug
────────────────

    .ToSQL() (string, []any)
    .String() string
    .DryRun() *gorm.DB

Q Builder
─────────

    gormz.Qb() *Q
    gormz.QOr(children...) *Q
    gormz.QAnd(children...) *Q

    gormz.Eq(field, value) whereClause
    gormz.Ne(field, value)
    gormz.Gt(field, value)
    gormz.Gte(field, value)
    gormz.Lt(field, value)
    gormz.Lte(field, value)
    gormz.Contains(field, value)
    gormz.StartsWith(field, value)
    gormz.EndsWith(field, value)
    gormz.In(field, values)
    gormz.IsNull(field)
    gormz.NotNull(field)
    gormz.Raw(sql, args...)
    gormz.Not(clause)

Registry
────────

    gormz.Register[T](name) *QuerySet[T]
    gormz.TryRegister[T](name) (*QuerySet[T], error)
    gormz.Lookup[T](name) (*QuerySet[T], bool)
    gormz.MustLookup[T](name) *QuerySet[T]
    gormz.Has(name) bool
    gormz.Unregister(name)
    gormz.RegisteredNames() []string
    gormz.RegisteredCount() int
    gormz.ClearRegistry()
    gormz.RegisterBatch(entries) error
    gormz.UnregisterBatch(names...) int
    gormz.HasType[T]() bool
    gormz.LookupByType[T]() (*QuerySet[T], bool)
    gormz.MustLookupByType[T]() *QuerySet[T]
    gormz.NamesByType[T]() []string
    gormz.TakeSnapshot() Snapshot
    gormz.Summary() RegistrySummary

Errors
──────

    gormz.IsNotFound(err) bool
    gormz.IsValidation(err) bool
    gormz.IsDangerous(err) bool
    gormz.IsAlreadyRegistered(err) bool
    gormz.IsNotFoundInRegistry(err) bool
    gormz.AsNotFound(err) (*NotFoundError, bool)
    gormz.AsValidation(err) (*ValidationError, bool)
    gormz.AsDangerous(err) (*DangerousOperationError, bool)

Instance
────────

    gormz.NewInstance(db) *Instance
    .DB() *gorm.DB
    .Ping() error
    .Close() error
    .Configure(cfg) error
    .Transaction(ctx, fn) error
    .Query[T]() *QuerySet[T]
    .QueryWithContext[T](ctx) *QuerySet[T]
    .Migrate(models...) error

    gormz.GlobalInstance() *Instance


advanced/ Package
─────────────────

Subqueries:
    advanced.SubFrom[T](q, field) *SubQuery
    advanced.SubRaw(sql, args...) *SubQuery
    advanced.In(field, sub) internal.RawClause
    advanced.NotIn(field, sub)
    advanced.Exists(sub)
    advanced.NotExists(sub)
    advanced.GtSub(field, sub)
    advanced.LtSub(field, sub)
    advanced.EqSub(field, sub)
    advanced.GteSub(field, sub)
    advanced.LteSub(field, sub)

CTEs:
    advanced.NewCTE[T](name, q) *CTE
    advanced.NewRecursiveCTE[T](name, base) *CTE
    advanced.With[T](ctes...) *CTEBuilder[T]
    .Query(q) *CTEBuilder[T]
    .All() ([]T, error)
    .ScanInto(dest) error

Window:
    advanced.RowNumber، Rank، DenseRank، Lag، Lead،
    RunningSum، RunningCount، RunningAvg، SumOver،
    CountOver، AvgOver، NTile، FirstValue، LastValue،
    NthValue، PercentRank، CumeDist

Union:
    advanced.Union[T](queries...) *UnionQuery[T]
    advanced.UnionAll[T](queries...)
    .OrderBy(fields...)
    .Limit(n)
    .Offset(n)
    .All() ([]T, error)

Aggregates:
    advanced.GroupBy[T](fields...) *AggregateQuery[T]
    .Count، .CountDistinct، .Sum، .Avg، .Min، .Max
    .Select، .SelectAs
    .Filter، .Exclude، .Where، .Having
    .OrderBy، .Limit، .Offset
    .All() ([]map[string]any, error)
    .ScanInto(dest) error
    .CountGroups() (int64, error)

    advanced.Distinct[T](field) ([]any, error)
    advanced.CountDistinctValues[T](field) (int64, error)
    advanced.GroupConcat[T](field, sep) (string, error)

Locking:
    advanced.WithLock[T](q) *PessimisticQuery[T]
    .ForUpdate، .ForShare، .ForNoKeyUpdate، .ForKeyShare
    advanced.LockClause(mode) clause.Locking
    advanced.LockClauseWithTable(mode, table)
    advanced.WithLockedTransaction(ctx, cfg, fn) error

Retry:
    advanced.Retry(ctx, cfg, fn) error
    advanced.RetryDB(ctx, fn) error
    advanced.RetryWithBackoff(ctx, attempts, backoff, fn) error
    advanced.IsRetryableError(err) bool
    advanced.ExponentialBackoff(initial, max) BackoffStrategy
    advanced.LinearBackoff(step, max) BackoffStrategy
    advanced.ConstantBackoff(delay) BackoffStrategy

Transactions:
    advanced.Begin(ctx, cfg) (*Tx, error)
    advanced.WithTransaction(ctx, cfg, fn) error
    advanced.SimpleTransaction(ctx, fn) error
    advanced.MustBegin(ctx, cfg) *Tx
    .Query[T]() *QuerySet[T]
    .DB() *gorm.DB
    .Commit() error
    .Rollback() error
    .RollbackIfActive()
    .Nested(fn) error
    .Savepoint(name) error
    .RollbackTo(name) error
    .ReleaseSavepoint(name) error

Savepoints:
    advanced.NewSavepoint(name, tx) *Savepoint
    .Create، .Rollback، .Release
    advanced.NewSavepointStack(tx) *SavepointStack
    .Push، .Pop، .RollbackLast، .RollbackTo، .Depth

Bulk:
    advanced.BulkInsert[T](ctx, items, cfg) error
    advanced.BulkUpsert[T](ctx, items, cfg) error
    advanced.BulkUpdate[T](ctx, idCol, items) error
    advanced.BulkDeleteByIDs[T](ctx, ids, batchSize) (int64, error)
    advanced.BulkDeleteWhere[T](ctx, where, args...) (int64, error)
    advanced.BulkCount[T](ctx, where, args...) (int64, error)
    advanced.UpdateItem[T]{ID, Values}
    advanced.BulkConfig{...}

Batch:
    advanced.ProcessBatch[T](ctx, items, cfg, fn) *BatchResult
    advanced.Stream[T](ctx, batchSize, fn) error
    advanced.Parallel() *ParallelQuery
    .WithWorkers، .Add، .Run
    advanced.BatchConfig{...}
    advanced.BatchResult{...}


═══════════════════════════════════════════════════════════════════
                           الخاتمة
═══════════════════════════════════════════════════════════════════

شكرًا لاستخدامك gormz!

نأمل أن تجعل هذه المكتبة تطويرك أسرع، وكودك أنظف، وحياتك أسهل.

للأسئلة، الاقتراحات، أو الإبلاغ عن أخطاء:
    GitHub: https://github.com/light-tech-dev/gormz
    Issues: https://github.com/light-tech-dev/gormz/issues
    Discussions: https://github.com/light-tech-dev/gormz/discussions

═══════════════════════════════════════════════════════════════════

    "اجعل الاستعلامات البسيطة بسيطة، والمعقدة ممكنة."

    — Sanad Team

═══════════════════════════════════════════════════════════════════

                          نهاية الدوكيمنتشن

                          gormz v0.1.0
                          MIT License
                          © 2025 Sanad Team

═══════════════════════════════════════════════════════════════════