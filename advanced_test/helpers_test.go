package advanced_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/light-tech-dev/gormz"
	"github.com/light-tech-dev/gormz/tests/fixtures"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Aliases
type User = fixtures.User
type Order = fixtures.Order
type Product = fixtures.Product

// setupTestDB ينشئ DB جديدة + migrations + cleanup.
func setupTestDB(t *testing.T) {
	t.Helper()

	dsn := fmt.Sprintf("file:test_adv_%d?mode=memory&cache=shared", time.Now().UnixNano())

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

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

// ✅ جديد — helper لفتح DB مخصصة للاختبارات
//
// يُستخدم في اختبارات locking (UserWithVersion) التي تحتاج
// migration لموديل مخصص.
//
// الاستخدام:
//
//	db, err := gormOpenTest()
//	require.NoError(t, err)
//	require.NoError(t, db.AutoMigrate(&MyModel{}))
//	gormz.SetDB(db)
//
//	t.Cleanup(func() {
//	    gormz.ResetDB()
//	    gormz.ClearRegistry()
//	})
func gormOpenTest() (*gorm.DB, error) {
	dsn := fmt.Sprintf("file:test_gormz_%d?mode=memory&cache=shared", time.Now().UnixNano())

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// مهم لـ SQLite :memory: — استخدام اتصال واحد
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)

	return db, nil
}

// seedUsers inserts 5 users.
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

// seedOrders inserts 5 orders.
func seedOrders(t *testing.T) []Order {
	t.Helper()

	orders := []Order{
		{UserID: 1, Total: 100, Status: "paid"},
		{UserID: 1, Total: 200, Status: "paid"},
		{UserID: 2, Total: 150, Status: "paid"},
		{UserID: 2, Total: 50, Status: "pending"},
		{UserID: 3, Total: 300, Status: "paid"},
	}

	for i := range orders {
		require.NoError(t, gormz.New[Order]().Create(&orders[i]))
	}

	return orders
}
