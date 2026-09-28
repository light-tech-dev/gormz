package advanced_test

import (
	"context"
	"testing"

	"github.com/light-tech-dev/gormz"
	"github.com/light-tech-dev/gormz/advanced"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ═══════════════════════════════════════════════
// Pessimistic Locking
// ═══════════════════════════════════════════════

func TestWithLock_ForUpdate(t *testing.T) {
	setupTestDB(t)

	require.NoError(t, gormz.New[User]().Create(&User{
		Name: "Ali", Email: "ali@test.com",
	}))

	// في SQLite، FOR UPDATE لا يعمل — لكن الاستعلام يعمل
	users, err := advanced.WithLock[User](
		gormz.New[User]().Filter("name", "Ali"),
	).ForUpdate().All()

	require.NoError(t, err)
	assert.Len(t, users, 1)
}

func TestPessimisticQuery_Filter(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	users, err := advanced.WithLock[User](
		gormz.New[User](),
	).Filter("active", true).All()

	require.NoError(t, err)
	assert.Len(t, users, 4) // 4 active users
}

// ═══════════════════════════════════════════════
// Optimistic Locking
// ═══════════════════════════════════════════════

// UserWithVersion نموذج مع version.
type UserWithVersion struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"size:255"`
	Email   string `gorm:"size:255"`
	Version int64  `gorm:"default:1"`
}

func (UserWithVersion) TableName() string { return "users_with_version" }

func setupVersionDB(t *testing.T) {
	t.Helper()

	db, err := gormOpenTest()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&UserWithVersion{}))
	gormz.SetDB(db)

	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		gormz.ResetDB()
	})
}

func TestUpdateIfVersion_Success(t *testing.T) {
	setupVersionDB(t)
	ctx := context.Background()

	user := &UserWithVersion{
		Name: "Ali", Email: "ali@test.com", Version: 1,
	}
	require.NoError(t, gormz.New[UserWithVersion]().Create(user))

	// تحديث بالإصدار الصحيح
	err := advanced.WithOptimisticLock[UserWithVersion](
		gormz.New[UserWithVersion](),
	).UpdateIfVersion(ctx, user.ID, 1, map[string]any{
		"name": "Ali Updated",
	})

	require.NoError(t, err)

	// تحقق
	got, _ := gormz.New[UserWithVersion]().Get(user.ID)
	assert.Equal(t, "Ali Updated", got.Name)
	assert.Equal(t, int64(2), got.Version) // incremented
}

func TestUpdateIfVersion_Mismatch(t *testing.T) {
	setupVersionDB(t)
	ctx := context.Background()

	user := &UserWithVersion{
		Name: "Ali", Email: "ali@test.com", Version: 1,
	}
	require.NoError(t, gormz.New[UserWithVersion]().Create(user))

	// تحديث بإصدار خاطئ
	err := advanced.WithOptimisticLock[UserWithVersion](
		gormz.New[UserWithVersion](),
	).UpdateIfVersion(ctx, user.ID, 99, map[string]any{
		"name": "Ali Updated",
	})

	assert.Error(t, err)
	assert.True(t, gormz.IsValidation(err))

	// تحقق — لم يُحدَّث
	got, _ := gormz.New[UserWithVersion]().Get(user.ID)
	assert.Equal(t, "Ali", got.Name)
	assert.Equal(t, int64(1), got.Version)
}

func TestDeleteIfVersion_Success(t *testing.T) {
	setupVersionDB(t)
	ctx := context.Background()

	user := &UserWithVersion{
		Name: "Ali", Email: "ali@test.com", Version: 1,
	}
	require.NoError(t, gormz.New[UserWithVersion]().Create(user))

	err := advanced.WithOptimisticLock[UserWithVersion](
		gormz.New[UserWithVersion](),
	).DeleteIfVersion(ctx, user.ID, 1)

	require.NoError(t, err)

	// تحقق
	_, err = gormz.New[UserWithVersion]().Get(user.ID)
	assert.True(t, gormz.IsNotFound(err))
}

func TestDeleteIfVersion_Mismatch(t *testing.T) {
	setupVersionDB(t)
	ctx := context.Background()

	user := &UserWithVersion{
		Name: "Ali", Email: "ali@test.com", Version: 1,
	}
	require.NoError(t, gormz.New[UserWithVersion]().Create(user))

	err := advanced.WithOptimisticLock[UserWithVersion](
		gormz.New[UserWithVersion](),
	).DeleteIfVersion(ctx, user.ID, 99)

	assert.Error(t, err)
	assert.True(t, gormz.IsValidation(err))

	// تحقق — لم يُحذف
	_, err = gormz.New[UserWithVersion]().Get(user.ID)
	assert.NoError(t, err)
}

func TestOptimisticQuery_CustomField(t *testing.T) {
	// موديل مع حقل مخصص
	type Doc struct {
		ID      uint   `gorm:"primaryKey"`
		Title   string
		Rev     int64 `gorm:"default:1"` // ← اسم مخصص
	}

	db, err := gormOpenTest()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Doc{}))
	gormz.SetDB(db)

	ctx := context.Background()

	doc := &Doc{Title: "Draft", Rev: 1}
	require.NoError(t, gormz.New[Doc]().Create(doc))

	err = advanced.WithOptimisticLock[Doc](gormz.New[Doc]()).
		Field("rev").
		UpdateIfVersion(ctx, doc.ID, 1, map[string]any{
			"title": "Published",
		})

	require.NoError(t, err)

	got, _ := gormz.New[Doc]().Get(doc.ID)
	assert.Equal(t, "Published", got.Title)
	assert.Equal(t, int64(2), got.Rev)
}