package gormz_test

import (
	"fmt"
	"testing"

	"github.com/light-tech-dev/gormz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ═══════════════════════════════════════════════
// CRUD Tests
// ═══════════════════════════════════════════════

func TestCreate(t *testing.T) {
	setupTestDB(t)

	user := &User{Name: "Ali", Email: "ali@test.com", Age: 30, Active: true}
	err := gormz.New[User]().Create(user)

	require.NoError(t, err)
	assert.NotZero(t, user.ID)
}

func TestCreateNil(t *testing.T) {
	setupTestDB(t)

	err := gormz.New[User]().Create(nil)
	assert.Error(t, err)
}

func TestCreateMany(t *testing.T) {
	setupTestDB(t)

	users := []User{
		{Name: "Ali", Email: "ali@test.com"},
		{Name: "Sara", Email: "sara@test.com"},
	}

	err := gormz.New[User]().CreateMany(users)
	require.NoError(t, err)

	count, _ := gormz.New[User]().Count()
	assert.Equal(t, int64(2), count)
}

func TestCreateManyEmpty(t *testing.T) {
	setupTestDB(t)

	err := gormz.New[User]().CreateMany(nil)
	require.NoError(t, err)

	count, _ := gormz.New[User]().Count()
	assert.Equal(t, int64(0), count)
}

// ✅ إصلاح: استخدام fmt.Sprintf بدلًا من string(rune(...))
func TestCreateInBatches(t *testing.T) {
	setupTestDB(t)

	const total = 2500
	users := make([]User, total)
	for i := range users {
		users[i] = User{
			Name:  fmt.Sprintf("User %d", i),
			Email: fmt.Sprintf("user%d@test.com", i),
		}
	}

	err := gormz.New[User]().CreateInBatches(users, 500)
	require.NoError(t, err)

	count, _ := gormz.New[User]().Count()
	assert.Equal(t, int64(total), count)
}

func TestGet(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	user, err := gormz.New[User]().Get(1)
	require.NoError(t, err)
	assert.Equal(t, "Ali", user.Name)
}

func TestGetNotFound(t *testing.T) {
	setupTestDB(t)

	_, err := gormz.New[User]().Get(999)
	assert.Error(t, err)
	assert.True(t, gormz.IsNotFound(err))

	ne, ok := gormz.AsNotFound(err)
	assert.True(t, ok)
	assert.Equal(t, "User", ne.Model)
	assert.Equal(t, 999, ne.ID)
}

func TestGetOrNil(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	user, err := gormz.New[User]().GetOrNil(1)
	require.NoError(t, err)
	assert.NotNil(t, user)

	user, err = gormz.New[User]().GetOrNil(999)
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestFirst(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	user, err := gormz.New[User]().
		Filter("name", "Ali").
		First()

	require.NoError(t, err)
	assert.Equal(t, "Ali", user.Name)
}

func TestFirstOrNil(t *testing.T) {
	setupTestDB(t)

	seedUsers(t)
	user, err := gormz.New[User]().Filter("name", "Ali").FirstOrNil()
	require.NoError(t, err)
	assert.NotNil(t, user)

	user, err = gormz.New[User]().Filter("name", "Nobody").FirstOrNil()
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestLast(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	user, err := gormz.New[User]().Last()
	require.NoError(t, err)
	assert.Equal(t, "Hassan", user.Name)
}

func TestFind(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	user, err := gormz.New[User]().Find("email", "ali@test.com")
	require.NoError(t, err)
	assert.Equal(t, "Ali", user.Name)
}

func TestFindNotFound(t *testing.T) {
	setupTestDB(t)

	_, err := gormz.New[User]().Find("email", "nobody@test.com")
	assert.Error(t, err)
	assert.True(t, gormz.IsNotFound(err))
}

func TestFindOrNil(t *testing.T) {
	setupTestDB(t)

	user, err := gormz.New[User]().FindOrNil("email", "nobody@test.com")
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestUpdate(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	err := gormz.New[User]().Update(1, "age", 31)
	require.NoError(t, err)

	user, _ := gormz.New[User]().Get(1)
	assert.Equal(t, 31, user.Age)
}

func TestUpdateManyRequiresConditions(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	_, err := gormz.New[User]().UpdateMany(map[string]any{"active": false})
	assert.Error(t, err)
	assert.True(t, gormz.IsDangerous(err))

	affected, err := gormz.New[User]().
		Filter("age__lt", 18).
		UpdateMany(map[string]any{"active": false})
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)
}

func TestUpdateManyInvalidField(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	_, err := gormz.New[User]().
		Filter("active", true).
		UpdateMany(map[string]any{"bad field": "x"})
	assert.Error(t, err)
}

func TestDelete(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	err := gormz.New[User]().Delete(1)
	require.NoError(t, err)

	exists, _ := gormz.New[User]().Filter("id", 1).Exists()
	assert.False(t, exists)
}

func TestDeleteManyRequiresConditions(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	_, err := gormz.New[User]().DeleteMany()
	assert.Error(t, err)
	assert.True(t, gormz.IsDangerous(err))

	affected, err := gormz.New[User]().
		Filter("active", false).
		DeleteMany()
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)
}

// (باقي الاختبارات كما هي...)