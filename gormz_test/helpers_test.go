package gormz_test

import (
	"testing"

	"github.com/light-tech-dev/gormz"
	"github.com/light-tech-dev/gormz/tests/fixtures"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Aliases للتوافق مع الاختبارات القديمة
type User = fixtures.User
type Order = fixtures.Order
type Product = fixtures.Product

// setupTestDB ينشئ DB جديدة + migrations + cleanup.
func setupTestDB(t *testing.T) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// IMPORTANT: SQLite :memory: creates a NEW database per connection.
	// Force a single connection so all queries hit the same DB.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, db.AutoMigrate(&User{}, &Order{}, &Product{}))

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

func seedUsers(t *testing.T) []User {
	t.Helper()

	users := []User{
		{Name: "Ali", Email: "ali@test.com", Age: 30, Active: true},
		{Name: "Sara", Email: "sara@test.com", Age: 25, Active: true},
		{Name: "Omar", Email: "omar@test.com", Age: 17, Active: false},
		{Name: "Layla", Email: "layla@test.com", Age: 35, Active: true},
		{Name: "Hassan", Email: "hassan@test.com", Age: 40, Active: true},
	}

	for i := range users {
		require.NoError(t, gormz.New[User]().Create(&users[i]))
	}

	return users
}

func seedOrders(t *testing.T, userID uint) []Order {
	t.Helper()

	orders := []Order{
		{UserID: userID, Total: 100.50, Status: "paid"},
		{UserID: userID, Total: 200.00, Status: "pending"},
		{UserID: userID, Total: 50.00, Status: "paid"},
	}

	require.NoError(t, gormz.New[Order]().CreateMany(orders))
	return orders
}