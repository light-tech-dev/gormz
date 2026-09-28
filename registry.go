package gormz

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
)

// ═══════════════════════════════════════════════
// Registry — سجل الموديلات
// ═══════════════════════════════════════════════

var registry = struct {
	sync.RWMutex
	models map[string]any
	types  map[string]string // name → type name
}{
	models: make(map[string]any),
	types:  make(map[string]string),
}

// Register يسجّل QuerySet لموديل باسم معين.
// يpanic عند التكرار — استخدم TryRegister للتحكم.
//
//	var Objects = gormz.Register[Product]("product")
func Register[T any](name string) *QuerySet[T] {
	q, err := TryRegister[T](name)
	if err != nil {
		panic(err)
	}
	return q
}

// TryRegister مثل Register لكن يرجّع الخطأ.
//
//	q, err := gormz.TryRegister[Product]("product")
//	if err != nil {
//	    log.Fatal(err)
//	}
func TryRegister[T any](name string) (*QuerySet[T], error) {
	if name == "" {
		name = defaultModelName[T]()
	}

	registry.Lock()
	defer registry.Unlock()

	if _, exists := registry.models[name]; exists {
		return nil, fmt.Errorf("%w: %q", ErrAlreadyRegistered, name)
	}

	q := New[T]()
	registry.models[name] = q
	registry.types[name] = defaultModelName[T]()
	return q, nil
}

// MustRegister مثل Register.
//
// Deprecated: استخدم Register.
func MustRegister[T any](name string) *QuerySet[T] {
	return Register[T](name)
}

// RegisterWith يسجّل QuerySet على Instance معين.
func RegisterWith[T any](name string, i *Instance) (*QuerySet[T], error) {
	if i == nil {
		return nil, ErrNilDB
	}
	if name == "" {
		name = defaultModelName[T]()
	}

	registry.Lock()
	defer registry.Unlock()

	if _, exists := registry.models[name]; exists {
		return nil, fmt.Errorf("%w: %q", ErrAlreadyRegistered, name)
	}

	q := newQuerySet[T](i.DB())
	registry.models[name] = q
	registry.types[name] = defaultModelName[T]()
	return q, nil
}

// Lookup يرجّع QuerySet مسجّلًا.
func Lookup[T any](name string) (*QuerySet[T], bool) {
	registry.RLock()
	defer registry.RUnlock()

	v, ok := registry.models[name]
	if !ok {
		return nil, false
	}

	q, ok := v.(*QuerySet[T])
	return q, ok
}

// MustLookup مثل Lookup لكن يpanic إذا لم يوجد.
func MustLookup[T any](name string) *QuerySet[T] {
	q, ok := Lookup[T](name)
	if !ok {
		panic(fmt.Errorf("%w: %q", ErrNotFoundInRegistry, name))
	}
	return q
}

// Has يفحص إذا كان موديل مسجّلًا.
func Has(name string) bool {
	registry.RLock()
	defer registry.RUnlock()
	_, ok := registry.models[name]
	return ok
}

// Unregister يزيل موديلًا من السجل.
func Unregister(name string) {
	registry.Lock()
	defer registry.Unlock()
	delete(registry.models, name)
	delete(registry.types, name)
}

// RegisteredNames يرجّع كل الأسماء المسجّلة (مرتبة).
func RegisteredNames() []string {
	registry.RLock()
	defer registry.RUnlock()

	names := make([]string, 0, len(registry.models))
	for name := range registry.models {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RegisteredCount يرجّع عدد الموديلات المسجّلة.
func RegisteredCount() int {
	registry.RLock()
	defer registry.RUnlock()
	return len(registry.models)
}

// ClearRegistry يمسح كل السجل (مفيد في الاختبارات).
//
// ⚠️ للاختبارات فقط.
func ClearRegistry() {
	registry.Lock()
	defer registry.Unlock()
	registry.models = make(map[string]any)
	registry.types = make(map[string]string)
}

// defaultModelName يستخرج اسم الموديل من نوعه.
func defaultModelName[T any]() string {
	t := reflect.TypeOf(new(T)).Elem()
	return t.Name()
}

// ═══════════════════════════════════════════════
// Type-based Lookup — بحث حسب النوع
// ═══════════════════════════════════════════════

// RegistryEntry يمثل مدخلًا في السجل.
type RegistryEntry struct {
	Name     string `json:"name"`
	TypeName string `json:"type_name"`
	Model    any    `json:"-"`
}

// AllEntries يرجّع كل المدخلات مع أنواعها.
func AllEntries() []RegistryEntry {
	registry.RLock()
	defer registry.RUnlock()

	out := make([]RegistryEntry, 0, len(registry.models))
	for name, m := range registry.models {
		typeName := registry.types[name]
		if typeName == "" {
			typeName = typeNameOf(m) // fallback
		}
		out = append(out, RegistryEntry{
			Name:     name,
			TypeName: typeName,
			Model:    m,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})

	return out
}

// HasType يفحص إذا كان نوع مسجلًا (بأي اسم).
func HasType[T any]() bool {
	_, ok := LookupByType[T]()
	return ok
}

// LookupByType يبحث حسب نوع T.
//
// يبحث عن أي اسم مسجل بنفس النوع T.
//
//	Register[User]("user")
//	Register[User]("admin")
//	LookupByType[User]() → (QuerySet[User], true)  // ✅
func LookupByType[T any]() (*QuerySet[T], bool) {
	typeName := defaultModelName[T]()

	registry.RLock()
	defer registry.RUnlock()

	for name, savedType := range registry.types {
		if savedType != typeName {
			continue
		}
		v, ok := registry.models[name]
		if !ok {
			continue
		}
		q, ok := v.(*QuerySet[T])
		if ok {
			return q, true
		}
	}

	return nil, false
}

// MustLookupByType مثل LookupByType لكن يpanic.
func MustLookupByType[T any]() *QuerySet[T] {
	q, ok := LookupByType[T]()
	if !ok {
		panic(fmt.Errorf("%w: %q", ErrNotFoundInRegistry, defaultModelName[T]()))
	}
	return q
}

// NamesByType يرجّع كل الأسماء لنوع معين.
//
// مفيد عندما يُسجَّل نفس الموديل بأسماء متعددة.
func NamesByType[T any]() []string {
	typeName := defaultModelName[T]()

	registry.RLock()
	defer registry.RUnlock()

	var out []string
	for name := range registry.models {
		savedType := registry.types[name]
		if savedType == "" {
			if m, ok := registry.models[name]; ok {
				savedType = typeNameOf(m)
			}
		}
		if savedType == typeName {
			out = append(out, name)
		}
	}

	sort.Strings(out)
	return out
}

// typeNameOf يستخرج اسم النوع من any.
//
// يتعامل مع generic types مثل QuerySet[User]:
//
//	*QuerySet[User] → "User"
//	*QuerySet[Order] → "Order"
//	*User → "User"
//
// إذا كان النوع بدون اسم (generic)، يستخرج الاسم من الـ brackets.
func typeNameOf(v any) string {
	if v == nil {
		return ""
	}

	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// للأنواع البسيطة (non-generic)
	if name := t.Name(); name != "" {
		return name
	}

	// للأنواع الـ generic مثل QuerySet[User]
	fullName := t.String()

	start := strings.Index(fullName, "[")
	end := strings.LastIndex(fullName, "]")

	if start > 0 && end > start {
		inner := fullName[start+1 : end]

		// احذف package path
		// github.com/light-tech-dev/gormz/tests/fixtures.User → User
		if idx := strings.LastIndex(inner, "."); idx >= 0 {
			inner = inner[idx+1:]
		}

		return inner
	}

	// fallback — احذف package path من الاسم الكامل
	if idx := strings.LastIndex(fullName, "."); idx >= 0 {
		return fullName[idx+1:]
	}

	return fullName
}

// ═══════════════════════════════════════════════
// Summary — ملخص السجل
// ═══════════════════════════════════════════════

// RegistrySummary يمثل ملخص السجل.
type RegistrySummary struct {
	Total  int            `json:"total"`
	ByName []string       `json:"by_name"`
	ByType map[string]int `json:"by_type"`
}

// Summary يرجّع ملخص السجل.
func Summary() RegistrySummary {
	registry.RLock()
	defer registry.RUnlock()

	summary := RegistrySummary{
		Total:  len(registry.models),
		ByName: make([]string, 0, len(registry.models)),
		ByType: make(map[string]int),
	}

	for name := range registry.models {
		typeName := registry.types[name]
		if typeName == "" {
			if m, ok := registry.models[name]; ok {
				typeName = typeNameOf(m)
			}
		}
		summary.ByName = append(summary.ByName, name)
		summary.ByType[typeName]++
	}

	sort.Strings(summary.ByName)
	return summary
}

// ═══════════════════════════════════════════════
// Bulk Register/Unregister
// ═══════════════════════════════════════════════

// RegisterBatch يسجّل دفعة من الموديلات.
//
// ⚠️ إذا فشل أحدها، يُعيد أول خطأ بدون rollback.
func RegisterBatch(entries map[string]any) error {
	if len(entries) == 0 {
		return nil
	}

	registry.Lock()
	defer registry.Unlock()

	// تحقق أولًا (atomicity)
	for name := range entries {
		if name == "" {
			return fmt.Errorf("gormz: empty registry name")
		}
		if _, exists := registry.models[name]; exists {
			return fmt.Errorf("%w: %q", ErrAlreadyRegistered, name)
		}
	}

	// ثم أضف
	for name, m := range entries {
		registry.models[name] = m
	}

	return nil
}

// UnregisterBatch يزيل دفعة من الموديلات.
//
// يرجّع عدد الموديلات المُزالة.
func UnregisterBatch(names ...string) int {
	if len(names) == 0 {
		return 0
	}

	registry.Lock()
	defer registry.Unlock()

	count := 0
	for _, name := range names {
		if _, exists := registry.models[name]; exists {
			delete(registry.models, name)
			count++
		}
	}
	return count
}

// ═══════════════════════════════════════════════
// Registry Snapshot — للنسخ الاحتياطي
// ═══════════════════════════════════════════════

// Snapshot يمثل نسخة من السجل.
type Snapshot struct {
	Entries map[string]any
}

// TakeSnapshot يأخذ نسخة من السجل الحالي.
func TakeSnapshot() Snapshot {
	registry.RLock()
	defer registry.RUnlock()

	entries := make(map[string]any, len(registry.models))
	for k, v := range registry.models {
		entries[k] = v
	}

	return Snapshot{Entries: entries}
}

// Restore يستعيد السجل من نسخة.
//
// ⚠️ يستبدل كل السجل الحالي.
func (s Snapshot) Restore() {
	registry.Lock()
	defer registry.Unlock()

	registry.models = make(map[string]any, len(s.Entries))
	for k, v := range s.Entries {
		registry.models[k] = v
	}
}

// Merge يدمج نسخة مع السجل الحالي.
//
// ⚠️ المفاتيح الموجودة لا تُستبدل.
func (s Snapshot) Merge() int {
	registry.Lock()
	defer registry.Unlock()

	count := 0
	for k, v := range s.Entries {
		if _, exists := registry.models[k]; !exists {
			registry.models[k] = v
			count++
		}
	}
	return count
}
