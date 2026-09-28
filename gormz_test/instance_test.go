package gormz_test

import (
	"context"
	"errors"
	"testing"

	"github.com/light-tech-dev/gormz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ═══════════════════════════════════════════════
// Instance Query Tests
// ═══════════════════════════════════════════════

func TestInstance_Query(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}))

	app := gormz.NewInstance(db)

	user := &User{Name: "Ali", Email: "ali@test.com"}
	require.NoError(t, gormz.QueryOn[User](app).Create(user))

	users, err := gormz.QueryOn[User](app).All()
	require.NoError(t, err)
	assert.Len(t, users, 1)
}

func TestInstance_Multiple(t *testing.T) {
	// DB 1
	db1, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db1.AutoMigrate(&User{}))
	app1 := gormz.NewInstance(db1)

	// DB 2
	db2, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db2.AutoMigrate(&User{}))
	app2 := gormz.NewInstance(db2)

	// Insert
	require.NoError(t, gormz.QueryOn[User](app1).Create(&User{Name: "Ali", Email: "ali@test.com"}))
	require.NoError(t, gormz.QueryOn[User](app2).Create(&User{Name: "Sara", Email: "sara@test.com"}))
	require.NoError(t, gormz.QueryOn[User](app2).Create(&User{Name: "Omar", Email: "omar@test.com"}))

	// Counts
	count1, err := gormz.QueryOn[User](app1).Count()
	require.NoError(t, err)
	count2, err := gormz.QueryOn[User](app2).Count()
	require.NoError(t, err)

	assert.Equal(t, int64(1), count1)
	assert.Equal(t, int64(2), count2)
}

// ═══════════════════════════════════════════════
// Instance Transaction Tests
// ═══════════════════════════════════════════════

func TestInstance_Transaction(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}))
	app := gormz.NewInstance(db)

	// Successful transaction
	err = app.Transaction(context.Background(), func(tx *gorm.DB) error {
		return tx.Create(&User{Name: "Ali", Email: "ali@test.com"}).Error
	})
	require.NoError(t, err)

	count, err := gormz.QueryOn[User](app).Count()
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Failed transaction — should rollback
	err = app.Transaction(context.Background(), func(tx *gorm.DB) error {
		if err := tx.Create(&User{Name: "X", Email: "x@test.com"}).Error; err != nil {
			return err
		}
		return errors.New("intentional rollback")
	})
	assert.Error(t, err)

	count, err = gormz.QueryOn[User](app).Count()
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestInstance_Transaction_NilCallback(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}))
	app := gormz.NewInstance(db)

	err = app.Transaction(context.Background(), nil)
	assert.Error(t, err)
}

// ═══════════════════════════════════════════════
// Instance Ping / Close Tests
// ═══════════════════════════════════════════════

func TestInstance_Ping(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	app := gormz.NewInstance(db)

	assert.NoError(t, app.Ping())
}

func TestInstance_NilPanics(t *testing.T) {
	assert.Panics(t, func() {
		gormz.NewInstance(nil)
	})
}

// ═══════════════════════════════════════════════
// NewWith / QueryOn Tests
// ═══════════════════════════════════════════════

func TestInstance_NewWith(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}))
	app := gormz.NewInstance(db)

	q := gormz.NewWith[User](app)
	require.NoError(t, q.Create(&User{Name: "Ali", Email: "ali@test.com"}))

	users, err := q.All()
	require.NoError(t, err)
	assert.Len(t, users, 1)
}

func TestInstance_QueryOn(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}))
	app := gormz.NewInstance(db)

	q := gormz.QueryOn[User](app)
	require.NoError(t, q.Create(&User{Name: "Ali", Email: "ali@test.com"}))

	users, err := q.All()
	require.NoError(t, err)
	assert.Len(t, users, 1)
}

func TestInstance_NewWithNilPanics(t *testing.T) {
	assert.Panics(t, func() {
		gormz.NewWith[User](nil)
	})
}

func TestInstance_QueryOnNilPanics(t *testing.T) {
	assert.Panics(t, func() {
		gormz.QueryOn[User](nil)
	})
}