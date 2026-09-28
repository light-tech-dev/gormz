package gormz_test

import (
	"fmt"
	"testing"

	"github.com/light-tech-dev/gormz"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ═══════════════════════════════════════════════
// Examples — Simple, Isolated, One Per Concern
// ═══════════════════════════════════════════════

// ExampleNew demonstrates the basic usage.
func ExampleNew() {
	gormz.ResetDB()
	gormz.ClearRegistry()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		panic(err)
	}
	gormz.SetDB(db)

	user := &User{Name: "Ali", Email: "ali@test.com"}
	if err := gormz.New[User]().Create(user); err != nil {
		panic(err)
	}

	users, err := gormz.New[User]().
		Filter("name", "Ali").
		All()
	if err != nil {
		panic(err)
	}

	fmt.Println(len(users))
	// Output: 1
}

// ExampleQuerySet_Filter demonstrates filters.
func ExampleQuerySet_Filter() {
	gormz.ResetDB()
	gormz.ClearRegistry()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		panic(err)
	}
	gormz.SetDB(db)

	if err := gormz.New[User]().Create(&User{Name: "Ali", Email: "ali@test.com", Age: 30, Active: true}); err != nil {
		panic(err)
	}
	if err := gormz.New[User]().Create(&User{Name: "Sara", Email: "sara@test.com", Age: 25, Active: true}); err != nil {
		panic(err)
	}
	if err := gormz.New[User]().Create(&User{Name: "Omar", Email: "omar@test.com", Age: 17, Active: false}); err != nil {
		panic(err)
	}

	adults, err := gormz.New[User]().
		Filter("active", true).
		Filter("age__gte", 18).
		All()
	if err != nil {
		panic(err)
	}

	fmt.Println(len(adults))
	// Output: 2
}

// ExampleRegister demonstrates the model registry.
func ExampleRegister() {
	gormz.ResetDB()
	gormz.ClearRegistry()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		panic(err)
	}
	gormz.SetDB(db)
	gormz.ClearRegistry()

	// Register model
	Users := gormz.Register[User]("user")
	if err := Users.Create(&User{Name: "Ali", Email: "ali@test.com"}); err != nil {
		panic(err)
	}

	// Lookup by name
	q := gormz.MustLookup[User]("user")
	users, err := q.All()
	if err != nil {
		panic(err)
	}

	fmt.Println(len(users))
	// Output: 1
}

// ═══════════════════════════════════════════════
// Real-World Example
// ═══════════════════════════════════════════════

func TestExample_RealWorld(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}, &Order{}))
	gormz.SetDB(db)

	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
		gormz.ResetDB()
		gormz.ClearRegistry()
	})

	// Create user
	user := &User{Name: "Ali", Email: "ali@test.com", Age: 30}
	require.NoError(t, gormz.New[User]().Create(user))

	// Create orders
	orders := []Order{
		{UserID: user.ID, Total: 100.50, Status: "paid"},
		{UserID: user.ID, Total: 200.00, Status: "pending"},
		{UserID: user.ID, Total: 50.00, Status: "paid"},
	}
	require.NoError(t, gormz.New[Order]().CreateMany(orders))

	// Sum paid orders
	total, err := gormz.New[Order]().
		Filter("user_id", user.ID).
		Filter("status", "paid").
		Sum("total")
	require.NoError(t, err)

	// 100.50 + 50.00 = 150.50
	assertInDelta(t, 150.50, total, 0.01)
}

// ═══════════════════════════════════════════════
// Helpers
// ═══════════════════════════════════════════════

func assertInDelta(t *testing.T, expected, actual, delta float64) {
	t.Helper()
	if actual < expected-delta || actual > expected+delta {
		t.Errorf("expected %v ± %v, got %v", expected, delta, actual)
	}
}
