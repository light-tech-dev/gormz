package gormz_test

import (
	"context"
	"testing"

	"github.com/light-tech-dev/gormz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestContext_WithDB(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, db.AutoMigrate(&User{}))

	app := gormz.NewInstance(db)
	ctx := gormz.WithDB(context.Background(), app)

	require.NoError(t, gormz.FromContext[User](ctx).Create(&User{Name: "Ali"}))

	users, err := gormz.FromContext[User](ctx).All()
	require.NoError(t, err)
	assert.Len(t, users, 1)
}

func TestContext_FallbackToGlobal(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, db.AutoMigrate(&User{}))
	gormz.SetDB(db)

	ctx := context.Background()

	require.NoError(t, gormz.FromContext[User](ctx).Create(&User{Name: "Ali"}))

	count, _ := gormz.FromContext[User](ctx).Count()
	assert.Equal(t, int64(1), count)
}

func TestContext_DBFromContext(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	app := gormz.NewInstance(db)

	_, ok := gormz.DBFromContext(context.Background())
	assert.False(t, ok)

	ctx := gormz.WithDB(context.Background(), app)
	instance, ok := gormz.DBFromContext(ctx)
	assert.True(t, ok)
	assert.Same(t, app, instance)
}

func TestContext_WithGormDB(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ctx := gormz.WithGormDB(context.Background(), db)

	instance, ok := gormz.DBFromContext(ctx)
	assert.True(t, ok)
	assert.NotNil(t, instance)
}

func TestContext_WithGormDBNilPanics(t *testing.T) {
	assert.Panics(t, func() {
		gormz.WithGormDB(context.Background(), nil)
	})
}