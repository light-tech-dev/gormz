package gormz_test

import (
	"testing"

	"github.com/light-tech-dev/gormz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ═══════════════════════════════════════════════
// Basic Registry
// ═══════════════════════════════════════════════

func TestRegistry_Register(t *testing.T) {
	setupTestDB(t)

	q := gormz.Register[User]("user")
	require.NotNil(t, q)

	got, ok := gormz.Lookup[User]("user")
	assert.True(t, ok)
	assert.Same(t, q, got)
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	setupTestDB(t)

	gormz.Register[User]("user")

	assert.Panics(t, func() {
		gormz.Register[User]("user")
	})
}

func TestRegistry_TryRegisterDuplicate(t *testing.T) {
	setupTestDB(t)

	_, err := gormz.TryRegister[User]("user")
	require.NoError(t, err)

	_, err = gormz.TryRegister[User]("user")
	assert.Error(t, err)
	assert.True(t, gormz.IsAlreadyRegistered(err))
}

func TestRegistry_LookupNotFound(t *testing.T) {
	setupTestDB(t)

	_, ok := gormz.Lookup[User]("not-registered")
	assert.False(t, ok)
}

func TestRegistry_MustLookupPanics(t *testing.T) {
	setupTestDB(t)

	assert.Panics(t, func() {
		gormz.MustLookup[User]("not-registered")
	})
}

func TestRegistry_Unregister(t *testing.T) {
	setupTestDB(t)

	gormz.Register[User]("user")
	gormz.Unregister("user")

	_, ok := gormz.Lookup[User]("user")
	assert.False(t, ok)
}

func TestRegistry_RegisteredNames(t *testing.T) {
	setupTestDB(t)

	gormz.Register[User]("user")
	gormz.Register[Order]("order")

	names := gormz.RegisteredNames()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "user")
	assert.Contains(t, names, "order")
}

func TestRegistry_RegisteredCount(t *testing.T) {
	setupTestDB(t)

	assert.Equal(t, 0, gormz.RegisteredCount())

	gormz.Register[User]("user")
	gormz.Register[Order]("order")

	assert.Equal(t, 2, gormz.RegisteredCount())
}

func TestRegistry_Usage(t *testing.T) {
	setupTestDB(t)

	Users := gormz.Register[User]("user")
	require.NoError(t, Users.Create(&User{Name: "Ali", Email: "ali@test.com"}))

	q := gormz.MustLookup[User]("user")
	users, _ := q.All()
	assert.Len(t, users, 1)
}

// ═══════════════════════════════════════════════
// Type-based Lookup
// ═══════════════════════════════════════════════

func TestRegistry_HasType(t *testing.T) {
	setupTestDB(t)

	assert.False(t, gormz.HasType[User]())

	gormz.Register[User]("user")
	assert.True(t, gormz.HasType[User]())
}

func TestRegistry_LookupByType(t *testing.T) {
	setupTestDB(t)

	gormz.Register[User]("user")
	q, ok := gormz.LookupByType[User]()
	assert.True(t, ok)
	assert.NotNil(t, q)
}

func TestRegistry_MustLookupByType(t *testing.T) {
	setupTestDB(t)

	assert.Panics(t, func() {
		gormz.MustLookupByType[User]()
	})

	gormz.Register[User]("user")
	assert.NotPanics(t, func() {
		gormz.MustLookupByType[User]()
	})
}

func TestRegistry_NamesByType(t *testing.T) {
	setupTestDB(t)

	gormz.Register[User]("user")
	gormz.Register[User]("admin_user")
	gormz.Register[Order]("order")

	names := gormz.NamesByType[User]()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "user")
	assert.Contains(t, names, "admin_user")
	assert.NotContains(t, names, "order")
}

// ═══════════════════════════════════════════════
// All Entries & Summary
// ═══════════════════════════════════════════════

func TestRegistry_AllEntries(t *testing.T) {
	setupTestDB(t)

	gormz.Register[User]("user")
	gormz.Register[Order]("order")

	entries := gormz.AllEntries()
	require.Len(t, entries, 2)

	// مرتبة حسب الاسم
	assert.Equal(t, "order", entries[0].Name)
	assert.Equal(t, "Order", entries[0].TypeName)
	assert.Equal(t, "user", entries[1].Name)
	assert.Equal(t, "User", entries[1].TypeName)
}

func TestRegistry_Summary(t *testing.T) {
	setupTestDB(t)

	gormz.Register[User]("user")
	gormz.Register[User]("admin_user")
	gormz.Register[Order]("order")

	summary := gormz.Summary()
	assert.Equal(t, 3, summary.Total)
	assert.Len(t, summary.ByName, 3)
	assert.Equal(t, 2, summary.ByType["User"])
	assert.Equal(t, 1, summary.ByType["Order"])
}

// ═══════════════════════════════════════════════
// Bulk Register
// ═══════════════════════════════════════════════

func TestRegistry_RegisterBatch(t *testing.T) {
	setupTestDB(t)

	entries := map[string]any{
		"user":  gormz.New[User](),
		"order": gormz.New[Order](),
	}

	err := gormz.RegisterBatch(entries)
	require.NoError(t, err)

	assert.Equal(t, 2, gormz.RegisteredCount())
}

func TestRegistry_RegisterBatchDuplicate(t *testing.T) {
	setupTestDB(t)

	gormz.Register[User]("user")

	entries := map[string]any{
		"user":  gormz.New[User](),
		"order": gormz.New[Order](),
	}

	err := gormz.RegisterBatch(entries)
	assert.Error(t, err)
	assert.True(t, gormz.IsAlreadyRegistered(err))

	// لم يُضف أي مدخل (atomic)
	assert.Equal(t, 1, gormz.RegisteredCount())
}

func TestRegistry_UnregisterBatch(t *testing.T) {
	setupTestDB(t)

	gormz.Register[User]("user")
	gormz.Register[Order]("order")
	gormz.Register[User]("admin")

	count := gormz.UnregisterBatch("user", "order", "missing")
	assert.Equal(t, 2, count)
	assert.Equal(t, 1, gormz.RegisteredCount())
}

// ═══════════════════════════════════════════════
// Snapshot
// ═══════════════════════════════════════════════

func TestRegistry_Snapshot(t *testing.T) {
	setupTestDB(t)

	gormz.Register[User]("user")
	gormz.Register[Order]("order")

	snap := gormz.TakeSnapshot()

	// مسح الكل
	gormz.ClearRegistry()
	assert.Equal(t, 0, gormz.RegisteredCount())

	// استعادة
	snap.Restore()
	assert.Equal(t, 2, gormz.RegisteredCount())
}

func TestRegistry_SnapshotMerge(t *testing.T) {
	setupTestDB(t)

	gormz.Register[User]("user")
	snap := gormz.TakeSnapshot()

	// إضافة order
	gormz.Register[Order]("order")

	// merge لا يستبدل
	count := snap.Merge()
	assert.Equal(t, 0, count) // user موجود مسبقًا
	assert.Equal(t, 2, gormz.RegisteredCount())
}