package advanced_test

import (
	"context"
	"testing"

	"github.com/light-tech-dev/gormz"
	"github.com/light-tech-dev/gormz/advanced"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIncrementVersion(t *testing.T) {
	setupVersionDB(t)
	ctx := context.Background()

	user := &UserWithVersion{
		Name: "Ali", Email: "ali@test.com", Version: 1,
	}
	require.NoError(t, gormz.New[UserWithVersion]().Create(user))

	err := advanced.WithOptimisticLock[UserWithVersion](
		gormz.New[UserWithVersion](),
	).IncrementVersion(ctx, user.ID)

	require.NoError(t, err)

	got, _ := gormz.New[UserWithVersion]().Get(user.ID)
	assert.Equal(t, int64(2), got.Version)
}

func TestUpdateIfVersion_CannotUpdateVersionDirectly(t *testing.T) {
	setupVersionDB(t)
	ctx := context.Background()

	user := &UserWithVersion{
		Name: "Ali", Email: "ali@test.com", Version: 1,
	}
	require.NoError(t, gormz.New[UserWithVersion]().Create(user))

	err := advanced.WithOptimisticLock[UserWithVersion](
		gormz.New[UserWithVersion](),
	).UpdateIfVersion(ctx, user.ID, 1, map[string]any{
		"version": 999, // ← ممنوع
	})

	assert.Error(t, err)
	assert.True(t, gormz.IsValidation(err))
}
