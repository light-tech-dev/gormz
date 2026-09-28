package advanced_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/light-tech-dev/gormz"
	"github.com/light-tech-dev/gormz/advanced"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBulkInsert(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	users := make([]User, 5000)
	for i := range users {
		users[i] = User{
			Name:   fmt.Sprintf("User %d", i),
			Email:  fmt.Sprintf("user%d@test.com", i),
			Age:    20 + (i % 50),
			Active: true,
		}
	}

	err := advanced.BulkInsert[User](ctx, users, advanced.BulkConfig{
		BatchSize: 1000,
	})
	require.NoError(t, err)

	count, _ := gormz.New[User]().Count()
	assert.Equal(t, int64(5000), count)
}

func TestBulkInsertEmpty(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	err := advanced.BulkInsert[User](ctx, nil, advanced.DefaultBulkConfig())
	require.NoError(t, err)
}

func TestBulkUpsert(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	users := []User{
		{Name: "Ali", Email: "ali@test.com", Age: 30},
		{Name: "Sara", Email: "sara@test.com", Age: 25},
	}
	err := advanced.BulkUpsert[User](ctx, users, advanced.BulkConfig{
		ConflictColumns: []string{"email"},
		UpdateColumns:   []string{"name", "age"},
	})
	require.NoError(t, err)

	// upsert — تعديل
	users[0].Age = 31
	err = advanced.BulkUpsert[User](ctx, users, advanced.BulkConfig{
		ConflictColumns: []string{"email"},
		UpdateColumns:   []string{"name", "age"},
	})
	require.NoError(t, err)

	ali, _ := gormz.New[User]().Find("email", "ali@test.com")
	assert.Equal(t, 31, ali.Age)

	count, _ := gormz.New[User]().Count()
	assert.Equal(t, int64(2), count)
}

func TestBulkUpsertRequiresConflictColumns(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	err := advanced.BulkUpsert[User](ctx, []User{{Name: "Ali"}}, advanced.BulkConfig{})
	assert.Error(t, err)
}

func TestBulkDeleteByIDs(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		gormz.New[User]().Create(&User{
			Name:  fmt.Sprintf("User %d", i),
			Email: fmt.Sprintf("user%d@test.com", i),
		})
	}

	ids := make([]any, 50)
	for i := range ids {
		ids[i] = i + 1
	}

	deleted, err := advanced.BulkDeleteByIDs[User](ctx, ids, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(50), deleted)

	count, _ := gormz.New[User]().Count()
	assert.Equal(t, int64(50), count)
}

func TestBulkDeleteWhere(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	seedUsers(t)

	deleted, err := advanced.BulkDeleteWhere[User](ctx, "age < ?", 18)
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted)
}

func TestBulkDeleteWhereRequiresCondition(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	_, err := advanced.BulkDeleteWhere[User](ctx, "")
	assert.Error(t, err)
}

func TestBulkCount(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	seedUsers(t)

	count, err := advanced.BulkCount[User](ctx, "active = ?", true)
	require.NoError(t, err)
	assert.Equal(t, int64(4), count)
}